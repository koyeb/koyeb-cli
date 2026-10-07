package koyeb

import (
	"testing"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/errors"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The proxy-ports flag bundle (--proxy-ports) is the single registration
// seam of the service definition's proxy-port flags. These tests pin the
// bundle contract; the per-surface suites pin the merged behavior.

func TestAddProxyPortsFlagsServiceSurface(t *testing.T) {
	flags := pflag.NewFlagSet("service", pflag.ContinueOnError)
	addProxyPortsFlags(flags)

	flag := flags.Lookup("proxy-ports")
	require.NotNil(t, flag, "--proxy-ports must be registered by the proxy-ports bundle")
	assert.Equal(t, "[]", flag.DefValue, "--proxy-ports default value")
	assert.Equal(t,
		"Update service proxy ports (available for services of type \"web\" only) using format PORT[:PROTOCOL], "+
			"for example --proxy-ports 22:tcp\n"+
			"PROTOCOL defaults to \"tcp\". Supported protocols are \"tcp\"."+
			"To delete a proxy port, prefix its number with '!', for example --proxy-ports '!80'\n",
		flag.Usage, "--proxy-ports help text")
}

// proxyPortsFlagNames returns the flag names the bundle registers — the
// bundle union the surface flag-set tests derive their expectations from.
func proxyPortsFlagNames() []string {
	flags := pflag.NewFlagSet("proxy-ports", pflag.ContinueOnError)
	addProxyPortsFlags(flags)
	names := make([]string, 0, 1)
	flags.VisitAll(func(f *pflag.Flag) { names = append(names, f.Name) })
	return names
}

// proxyPortsBundleTestFlagSet registers the bundle surface the parse step
// consumes.
func proxyPortsBundleTestFlagSet() *pflag.FlagSet {
	flags := pflag.NewFlagSet("proxy-ports", pflag.ContinueOnError)
	addProxyPortsFlags(flags)
	return flags
}

// liveProxyPorts returns a proxy port carrying a live value, so the
// changed-only merge conventions are observable.
func liveProxyPorts() []koyeb.DeploymentProxyPort {
	tcp := koyeb.PROXYPORTPROTOCOL_TCP
	return []koyeb.DeploymentProxyPort{{
		Port:     koyeb.PtrInt64(22),
		Protocol: &tcp,
	}}
}

func TestParseProxyPortsNoFlagsKeepsLiveValues(t *testing.T) {
	h := &ServiceHandler{}
	flags := proxyPortsBundleTestFlagSet()

	ports, err := h.parseProxyPorts(koyeb.DEPLOYMENTDEFINITIONTYPE_WEB, flags, liveProxyPorts())
	require.NoError(t, err)
	require.Len(t, ports, 1)
	assert.Equal(t, int64(22), *ports[0].Port)
	assert.Equal(t, koyeb.PROXYPORTPROTOCOL_TCP, *ports[0].Protocol)
}

func TestParseProxyPortsChangedFlagsMergeLiveValues(t *testing.T) {
	h := &ServiceHandler{}

	t.Run("--proxy-ports exposes a new proxy port", func(t *testing.T) {
		flags := proxyPortsBundleTestFlagSet()
		require.NoError(t, flags.Parse([]string{"--proxy-ports", "8022:tcp"}))

		ports, err := h.parseProxyPorts(koyeb.DEPLOYMENTDEFINITIONTYPE_WEB, flags, liveProxyPorts())
		require.NoError(t, err)
		require.Len(t, ports, 2)
		assert.Equal(t, int64(22), *ports[0].Port, "the live proxy port must be kept")
		assert.Equal(t, int64(8022), *ports[1].Port)
		assert.Equal(t, koyeb.PROXYPORTPROTOCOL_TCP, *ports[1].Protocol)
	})

	t.Run("--proxy-ports updates a live proxy port by number", func(t *testing.T) {
		flags := proxyPortsBundleTestFlagSet()
		// The SDK's proxy-port protocol enum accepts "tcp" only: any
		// other protocol value falls back to it, so tcp is the whole range.
		require.NoError(t, flags.Parse([]string{"--proxy-ports", "22:tcp"}))

		ports, err := h.parseProxyPorts(koyeb.DEPLOYMENTDEFINITIONTYPE_WEB, flags, liveProxyPorts())
		require.NoError(t, err)
		require.Len(t, ports, 1, "the live proxy port must be updated in place")
		assert.Equal(t, int64(22), *ports[0].Port)
		assert.Equal(t, koyeb.PROXYPORTPROTOCOL_TCP, *ports[0].Protocol)
	})

	t.Run("'!' closes a live proxy port", func(t *testing.T) {
		flags := proxyPortsBundleTestFlagSet()
		require.NoError(t, flags.Parse([]string{"--proxy-ports", "!22"}))

		ports, err := h.parseProxyPorts(koyeb.DEPLOYMENTDEFINITIONTYPE_WEB, flags, liveProxyPorts())
		require.NoError(t, err)
		assert.Empty(t, ports)
	})
}

func TestParseProxyPortsErrorPaths(t *testing.T) {
	h := &ServiceHandler{}

	t.Run("proxy ports are web-only", func(t *testing.T) {
		flags := proxyPortsBundleTestFlagSet()
		require.NoError(t, flags.Parse([]string{"--proxy-ports", "8022:tcp"}))

		_, err := h.parseProxyPorts(koyeb.DEPLOYMENTDEFINITIONTYPE_WORKER, flags, liveProxyPorts())
		var cliErr *errors.CLIError
		require.ErrorAs(t, err, &cliErr)
		assert.Equal(t,
			`your service has ports configured, which is only possible for services of type "web"`,
			cliErr.Why)
	})

	for name, tc := range map[string]struct {
		value   string
		wantWhy string
	}{
		"an unparseable port": {
			value:   "abc",
			wantWhy: "unable to parse the port \"abc\"",
		},
		"an unparseable protocol": {
			value:   "22:udp",
			wantWhy: "unable to parse the protocol from the port \"22:udp\"",
		},
		"a deletion carrying a protocol": {
			value:   "!22:tcp",
			wantWhy: "unable to parse the port \"!22:tcp\"",
		},
	} {
		t.Run(name, func(t *testing.T) {
			flags := proxyPortsBundleTestFlagSet()
			require.NoError(t, flags.Parse([]string{"--proxy-ports", tc.value}))

			_, err := h.parseProxyPorts(koyeb.DEPLOYMENTDEFINITIONTYPE_WEB, flags, liveProxyPorts())
			var cliErr *errors.CLIError
			require.ErrorAs(t, err, &cliErr)
			assert.Equal(t, tc.wantWhy, cliErr.Why)
		})
	}
}
