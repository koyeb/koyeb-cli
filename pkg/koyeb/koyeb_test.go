package koyeb

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Run is the exit-code seam of the CLI binary: every invocation maps to
// exactly one decision — successful runs (help, version, completion
// included) exit 0, error paths (command not found, invalid arguments,
// API errors, recovered panics) print the CLI error box and exit 1.
// These tests pin that contract so no error path can regress to exit 0.

func TestRunExitCode(t *testing.T) {
	testCases := []struct {
		name     string
		args     []string
		wantCode int
	}{
		{"no arguments prints help", []string{}, 0},
		{"help flag", []string{"--help"}, 0},
		{"help command", []string{"help"}, 0},
		{"version", []string{"version"}, 0},
		{"completion", []string{"completion", "bash"}, 0},
		{"unknown command", []string{"this-command-does-not-exist"}, 1},
		{"missing required argument", []string{"apps", "describe"}, 1},
		{"invalid flag value", []string{"apps", "list", "--output", "invalidformat"}, 1},
		{"unknown flag", []string{"apps", "list", "--no-such-flag"}, 1},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var exitCode int
			_, stderr := captureStd(t, func() { exitCode = Run(tc.args) })

			assert.Equal(t, tc.wantCode, exitCode,
				"the exit code must match the %s contract", tc.name)
			if tc.wantCode == 0 {
				assert.NotContains(t, stderr, "❌", "a successful run must not print an error box")
			} else {
				assert.Contains(t, stderr, "❌", "the error box must still be printed")
			}
		})
	}
}

func TestRunExitCodeOnAPIErrors(t *testing.T) {
	t.Run("application not found", func(t *testing.T) {
		// The ST-33 repro: the identifier resolves to no object, the CLI
		// prints the resolution error box and must exit non-zero.
		var requests atomic.Int32
		server := newKoyebAPITestServer(t, func(w http.ResponseWriter, r *http.Request) {
			requests.Add(1)
			writeJSON(t, w, http.StatusOK, map[string]any{"apps": []any{}, "count": 0})
		})
		args := append(koyebAPITestConfigArgs(t, server), "apps", "describe", "some-nonexistent-id")

		var exitCode int
		_, stderr := captureStd(t, func() { exitCode = Run(args) })

		assert.Equal(t, int32(1), requests.Load(), "the CLI must have queried the API server")
		assert.Equal(t, 1, exitCode, "a not-found error is an error path and must exit non-zero")
		assert.Contains(t, stderr, "Unable to find the application `some-nonexistent-id`")
	})

	t.Run("api server error", func(t *testing.T) {
		var requests atomic.Int32
		server := newKoyebAPITestServer(t, func(w http.ResponseWriter, r *http.Request) {
			requests.Add(1)
			writeJSON(t, w, http.StatusInternalServerError,
				map[string]any{"status": 500, "code": "internal_error", "message": "internal server error"})
		})
		args := append(koyebAPITestConfigArgs(t, server), "apps", "describe", "some-nonexistent-id")

		var exitCode int
		_, stderr := captureStd(t, func() { exitCode = Run(args) })

		assert.Equal(t, int32(1), requests.Load(), "the CLI must have queried the API server")
		assert.Equal(t, 1, exitCode, "an API error is an error path and must exit non-zero")
		assert.Contains(t, stderr, "❌", "the error box must still be printed")
	})
}

func TestRunRecoversPanicsAsNonZeroExitCode(t *testing.T) {
	// The recover handler prints the unexpected-error box; the exit-code
	// decision must treat the recovered panic as an error too.
	rootCmd := &cobra.Command{Use: "koyeb"}
	rootCmd.AddCommand(&cobra.Command{
		Use:  "boom",
		RunE: func(*cobra.Command, []string) error { panic("boom") },
	})
	rootCmd.SetArgs([]string{"boom"})

	var exitCode int
	_, stderr := captureStd(t, func() { exitCode = run(rootCmd) })

	require.Equal(t, 1, exitCode, "a recovered panic is an error path and must exit non-zero")
	assert.Contains(t, stderr, "An unexpected error occured", "the recovered panic must still print the CLI error box")
}

// captureStd runs f with os.Stdout and os.Stderr redirected to pipes, and
// returns whatever f wrote to them.
func captureStd(t *testing.T, f func()) (string, string) {
	t.Helper()

	stdoutR, stdoutW, err := os.Pipe()
	require.NoError(t, err)
	stderrR, stderrW, err := os.Pipe()
	require.NoError(t, err)

	origStdout, origStderr := os.Stdout, os.Stderr
	os.Stdout, os.Stderr = stdoutW, stderrW
	f()
	os.Stdout, os.Stderr = origStdout, origStderr

	require.NoError(t, stdoutW.Close())
	require.NoError(t, stderrW.Close())

	stdout, err := io.ReadAll(stdoutR)
	require.NoError(t, err)
	stderr, err := io.ReadAll(stderrR)
	require.NoError(t, err)
	return string(stdout), string(stderr)
}

// newKoyebAPITestServer starts an HTTP test server standing in for the
// Koyeb API, and closes it when the test finishes.
func newKoyebAPITestServer(t *testing.T, handler http.HandlerFunc) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	return server
}

// koyebAPITestConfigArgs writes a CLI configuration file pointing at the
// given API server, and returns the --config arguments selecting it.
func koyebAPITestConfigArgs(t *testing.T, server *httptest.Server) []string {
	t.Helper()

	// The login, version and completion commands are process-wide singletons
	// whose called state is sticky: once one of them has run, skipConfigLoading
	// makes every later Run call in the same process skip the configuration
	// file. Rebuild them with the state of a fresh process so the scenario
	// reads the configuration file it was given.
	loginCmd = newLoginCmd()
	versionCmd = newVersionCmd()
	completionCmd = newCompletionCmd()

	// Run initializes the apiurl and token package variables from the
	// configuration file; restore them so the API-backed scenarios don't
	// leak into the other tests of the package.
	origAPIURL, origToken := apiurl, token
	t.Cleanup(func() { apiurl, token = origAPIURL, origToken })

	configPath := filepath.Join(t.TempDir(), "koyeb.yaml")
	content := fmt.Sprintf("url: %s\ntoken: koyeb-cli-test-token\n", server.URL)
	require.NoError(t, os.WriteFile(configPath, []byte(content), 0o600))

	return []string{"--config", configPath}
}

// writeJSON writes body as a JSON HTTP response.
func writeJSON(t *testing.T, w http.ResponseWriter, statusCode int, body map[string]any) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	require.NoError(t, json.NewEncoder(w).Encode(body))
}
