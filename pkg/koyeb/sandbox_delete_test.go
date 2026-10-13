package koyeb

import (
	"bytes"
	"net/http"
	"os"
	"testing"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSandboxDeleteCmdRegistered(t *testing.T) {
	cmd := sandboxSubCmd(t, "delete")

	assert.Equal(t, "delete NAME", cmd.Use)
	assert.NotNil(t, cmd.ValidArgsFunction,
		"sandbox delete must wire dynamic completion like the neighboring commands")
	require.NoError(t, cmd.Args(cmd, []string{"myapp/mysandbox"}))
	require.Error(t, cmd.Args(cmd, nil), "sandbox delete takes exactly one NAME")
}

func TestSandboxDeleteFlow(t *testing.T) {
	t.Run("deletes the resolved sandbox through the port", func(t *testing.T) {
		fake := sandboxFixtureAPI()
		ctx := sandboxTestContext(fake)

		require.NoError(t, NewSandboxHandler().Delete(ctx, sandboxSubCmd(t, "delete"), []string{sandboxFixtureServiceID}))
		assert.Equal(t, []string{sandboxFixtureServiceID}, fake.deletedServices)
	})

	t.Run("API errors surface as CLI errors", func(t *testing.T) {
		fake := sandboxFixtureAPI()
		fake.deleteServiceErr = assert.AnError
		ctx := sandboxTestContext(fake)

		err := NewSandboxHandler().Delete(ctx, sandboxSubCmd(t, "delete"), []string{sandboxFixtureServiceID})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Error while deleting the sandbox `"+sandboxFixtureServiceID+"`")
	})

	t.Run("rejects non-sandbox services without deleting", func(t *testing.T) {
		fake := sandboxFixtureAPI()
		webType := koyeb.SERVICETYPE_WEB
		fake.service.Type = &webType
		ctx := sandboxTestContext(fake)

		err := NewSandboxHandler().Delete(ctx, sandboxSubCmd(t, "delete"), []string{sandboxFixtureServiceID})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not a sandbox")
		assert.Empty(t, fake.deletedServices, "the type check must fail the delete before the port call")
	})
}

// TestSandboxDeleteResolvesNameAndShortID pins the resolution contract end
// to end for the delete flow: both identifiers resolve through the real id
// mapper and the deletion reaches the resolved service.
func TestSandboxDeleteResolvesNameAndShortID(t *testing.T) {
	for _, identifier := range []string{"myapp/mysandbox", "323e4567"} {
		t.Run("resolves "+identifier, func(t *testing.T) {
			var deleted bool
			server := newKoyebAPITestServer(t, sandboxAPIChainHandler(t, &deleted))
			args := append(koyebAPITestConfigArgs(t, server), "sandbox", "delete", identifier)

			// logrus keeps the process-wide stderr it captured at init:
			// swap its output to a buffer to pin the confirmation message.
			var logs bytes.Buffer
			log.SetOutput(&logs)
			t.Cleanup(func() { log.SetOutput(os.Stderr) })

			var exitCode int
			_, _ = captureStd(t, func() { exitCode = Run(args) })

			require.Equal(t, 0, exitCode, "sandbox delete %s must succeed", identifier)
			assert.True(t, deleted, "the deletion must reach the resolved service")
			assert.Contains(t, logs.String(), "Sandbox "+identifier+" deleted.")
		})
	}
}

func TestSandboxDeleteRejectsUnknownName(t *testing.T) {
	// Resolution failure surfaces as a CLI error with a non-zero exit code.
	server := newKoyebAPITestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/services" {
			writeJSON(t, w, http.StatusOK, map[string]any{"services": []any{}, "count": 0})
			return
		}
		writeJSON(t, w, http.StatusOK, map[string]any{})
	})
	args := append(koyebAPITestConfigArgs(t, server), "sandbox", "delete", "no-such-app/no-such-sandbox")

	var exitCode int
	_, stderr := captureStd(t, func() { exitCode = Run(args) })

	assert.Equal(t, 1, exitCode, "an unknown sandbox is an error path")
	assert.Contains(t, stderr, "could not find sandbox 'no-such-app/no-such-sandbox'")
}
