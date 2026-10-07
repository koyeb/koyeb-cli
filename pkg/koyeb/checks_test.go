package koyeb

import (
	"testing"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/errors"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The checks flag bundle (--checks, --checks-grace-period) is the single
// registration seam of the service definition's healthcheck flags. These
// tests pin the bundle contract; the per-surface suites pin the merged
// behavior.

func TestAddChecksFlagsServiceSurface(t *testing.T) {
	flags := pflag.NewFlagSet("service", pflag.ContinueOnError)
	addChecksFlags(flags)

	expected := []struct {
		name     string
		defValue string
		usage    string
	}{
		{
			"checks", "[]",
			"Update service healthchecks (available for services of type \"web\" only)\n" +
				"For HTTP healthchecks, use the format <PORT>:http:<PATH>, for example --checks 8080:http:/health\n" +
				"For TCP healthchecks, use the format <PORT>:tcp, for example --checks 8080:tcp\n" +
				"To delete a healthcheck, use !PORT, for example --checks '!8080'\n",
		},
		{
			"checks-grace-period", "[]",
			"Set healthcheck grace period in seconds.\n" +
				"Use the format <healthcheck>=<seconds>, for example --checks-grace-period 8080=10\n",
		},
	}
	for _, want := range expected {
		flag := flags.Lookup(want.name)
		require.NotNil(t, flag, "--%s must be registered by the checks bundle", want.name)
		assert.Equal(t, want.defValue, flag.DefValue, "--%s default value", want.name)
		assert.Equal(t, want.usage, flag.Usage, "--%s help text", want.name)
	}
}

// checksFlagNames returns the flag names the bundle registers — the
// bundle union the surface flag-set tests derive their expectations from.
func checksFlagNames() []string {
	flags := pflag.NewFlagSet("checks", pflag.ContinueOnError)
	addChecksFlags(flags)
	names := make([]string, 0, 2)
	flags.VisitAll(func(f *pflag.Flag) { names = append(names, f.Name) })
	return names
}

// checksBundleTestFlagSet registers the bundle surface the parse step
// consumes.
func checksBundleTestFlagSet() *pflag.FlagSet {
	flags := pflag.NewFlagSet("checks", pflag.ContinueOnError)
	addChecksFlags(flags)
	return flags
}

// liveChecks returns a healthcheck list carrying a live value, so the
// changed-only merge conventions are observable.
func liveChecks() []koyeb.DeploymentHealthCheck {
	http := koyeb.NewHTTPHealthCheck()
	http.Port = koyeb.PtrInt64(8080)
	http.Path = koyeb.PtrString("/health")
	check := koyeb.NewDeploymentHealthCheckWithDefaults()
	check.SetHttp(*http)
	return []koyeb.DeploymentHealthCheck{*check}
}

func TestParseChecksNoFlagsKeepsLiveValues(t *testing.T) {
	h := &ServiceHandler{}
	flags := checksBundleTestFlagSet()

	checks, err := h.parseChecks(koyeb.DEPLOYMENTDEFINITIONTYPE_WEB, flags, liveChecks())
	require.NoError(t, err)
	require.Len(t, checks, 1)
	http, ok := checks[0].GetHttpOk()
	require.True(t, ok)
	assert.Equal(t, int64(8080), *http.Port)
	assert.Equal(t, "/health", *http.Path)
}

func TestParseChecksChangedFlagsMergeLiveValues(t *testing.T) {
	h := &ServiceHandler{}

	t.Run("--checks adds a new healthcheck", func(t *testing.T) {
		flags := checksBundleTestFlagSet()
		require.NoError(t, flags.Parse([]string{"--checks", "9000:tcp"}))

		checks, err := h.parseChecks(koyeb.DEPLOYMENTDEFINITIONTYPE_WEB, flags, liveChecks())
		require.NoError(t, err)
		require.Len(t, checks, 2)
		http, ok := checks[0].GetHttpOk()
		require.True(t, ok)
		assert.Equal(t, int64(8080), *http.Port, "the live healthcheck must be kept")
		tcp, ok := checks[1].GetTcpOk()
		require.True(t, ok)
		assert.Equal(t, int64(9000), *tcp.Port)
	})

	t.Run("--checks updates a live healthcheck by port", func(t *testing.T) {
		flags := checksBundleTestFlagSet()
		require.NoError(t, flags.Parse([]string{"--checks", "8080:tcp"}))

		checks, err := h.parseChecks(koyeb.DEPLOYMENTDEFINITIONTYPE_WEB, flags, liveChecks())
		require.NoError(t, err)
		require.Len(t, checks, 1)
		tcp, ok := checks[0].GetTcpOk()
		require.True(t, ok, "the live HTTP check must be re-typed to TCP")
		assert.Equal(t, int64(8080), *tcp.Port)
	})

	t.Run("'!' deletes a live healthcheck by port", func(t *testing.T) {
		flags := checksBundleTestFlagSet()
		require.NoError(t, flags.Parse([]string{"--checks", "!8080"}))

		checks, err := h.parseChecks(koyeb.DEPLOYMENTDEFINITIONTYPE_WEB, flags, liveChecks())
		require.NoError(t, err)
		assert.Empty(t, checks)
	})
}

func TestParseChecksGracePeriod(t *testing.T) {
	h := &ServiceHandler{}

	t.Run("the grace period applies to the matching healthcheck", func(t *testing.T) {
		flags := checksBundleTestFlagSet()
		require.NoError(t, flags.Parse([]string{"--checks-grace-period", "8080=30"}))

		checks, err := h.parseChecks(koyeb.DEPLOYMENTDEFINITIONTYPE_WEB, flags, liveChecks())
		require.NoError(t, err)
		require.Len(t, checks, 1)
		require.NotNil(t, checks[0].GracePeriod)
		assert.Equal(t, int64(30), *checks[0].GracePeriod)
	})

	t.Run("the grace period applies to a healthcheck created in the same call", func(t *testing.T) {
		flags := checksBundleTestFlagSet()
		require.NoError(t, flags.Parse([]string{
			"--checks", "9000:tcp", "--checks-grace-period", "9000=15",
		}))

		checks, err := h.parseChecks(koyeb.DEPLOYMENTDEFINITIONTYPE_WEB, flags, liveChecks())
		require.NoError(t, err)
		require.Len(t, checks, 2)
		tcp, ok := checks[1].GetTcpOk()
		require.True(t, ok)
		assert.Equal(t, int64(9000), *tcp.Port)
		require.NotNil(t, checks[1].GracePeriod)
		assert.Equal(t, int64(15), *checks[1].GracePeriod)
	})

	for name, tc := range map[string]struct {
		grace    string
		wantWhy  string
		wantWhat string
	}{
		"unmatched healthcheck port": {
			grace:    "9000=10",
			wantWhy:  "--checks-grace-period does not match any healthcheck",
			wantWhat: "Invalid grace period",
		},
		"malformed grace period": {
			grace:    "8080-10",
			wantWhy:  "--checks-grace-period should be formatted as <healthcheck port number>=<grace period in seconds>",
			wantWhat: "Invalid grace period",
		},
		"non-numeric healthcheck port": {
			grace:    "abc=10",
			wantWhy:  "the grace period should be formatted as <healthcheck port number>=<grace period in seconds>",
			wantWhat: "Invalid grace period",
		},
		"non-numeric grace period": {
			grace:    "8080=abc",
			wantWhy:  "the grace period should be a number of seconds, not abc",
			wantWhat: "Invalid grace period",
		},
	} {
		t.Run(name, func(t *testing.T) {
			flags := checksBundleTestFlagSet()
			require.NoError(t, flags.Parse([]string{"--checks-grace-period", tc.grace}))

			_, err := h.parseChecks(koyeb.DEPLOYMENTDEFINITIONTYPE_WEB, flags, liveChecks())
			var cliErr *errors.CLIError
			require.ErrorAs(t, err, &cliErr)
			assert.Equal(t, tc.wantWhat, cliErr.What)
			assert.Equal(t, tc.wantWhy, cliErr.Why)
		})
	}
}

func TestParseChecksErrorPaths(t *testing.T) {
	h := &ServiceHandler{}

	t.Run("checks are web-only", func(t *testing.T) {
		flags := checksBundleTestFlagSet()
		require.NoError(t, flags.Parse([]string{"--checks", "9000:tcp"}))

		_, err := h.parseChecks(koyeb.DEPLOYMENTDEFINITIONTYPE_WORKER, flags, liveChecks())
		var cliErr *errors.CLIError
		require.ErrorAs(t, err, &cliErr)
		assert.Equal(t, `--checks can only be specified for "web" services`, cliErr.Why)
	})

	t.Run("an unparseable healthcheck is rejected", func(t *testing.T) {
		flags := checksBundleTestFlagSet()
		require.NoError(t, flags.Parse([]string{"--checks", "8080:udp"}))

		_, err := h.parseChecks(koyeb.DEPLOYMENTDEFINITIONTYPE_WEB, flags, liveChecks())
		var cliErr *errors.CLIError
		require.ErrorAs(t, err, &cliErr)
		assert.Equal(t, "unable to parse the protocol from the check \"8080:udp\"", cliErr.Why)
	})
}
