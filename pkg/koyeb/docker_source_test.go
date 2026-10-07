package koyeb

import (
	"testing"

	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The docker source flag bundle is the single registration seam shared by
// the service, sandbox and pool surfaces. These tests pin the bundle
// contract; the per-surface suites pin the merged behavior.

func TestAddDockerSourceFlagsServiceSurface(t *testing.T) {
	flags := pflag.NewFlagSet("service", pflag.ContinueOnError)
	addDockerSourceFlags(flags, serviceDockerSourceFlagUsage)

	expected := []struct {
		name     string
		defValue string
		usage    string
	}{
		{"docker", "", "Docker image"},
		{"docker-private-registry-secret", "", "Docker private registry secret"},
		{"docker-skip-verify", "false", "Skip docker image verification"},
		{
			"docker-entrypoint", "[]",
			"Docker entrypoint. To provide multiple arguments, use the --docker-entrypoint flag multiple times.",
		},
		{
			"docker-command", "",
			"Set the docker CMD explicitly. To provide arguments to the command, use the --docker-args flag.",
		},
		{
			"docker-args", "[]",
			"Set arguments to the docker command. To provide multiple arguments, use the --docker-args flag multiple times.",
		},
	}
	for _, want := range expected {
		flag := flags.Lookup(want.name)
		require.NotNil(t, flag, "--%s must be registered by the docker bundle", want.name)
		assert.Equal(t, want.defValue, flag.DefValue, "--%s default value", want.name)
		assert.Equal(t, want.usage, flag.Usage, "--%s help text", want.name)
	}
}

func TestAddDockerSourceFlagsSandboxPoolSurface(t *testing.T) {
	flags := pflag.NewFlagSet("sandbox-pool", pflag.ContinueOnError)
	addDockerSourceFlags(flags, sandboxPoolDockerSourceFlagUsage)

	expected := []struct {
		name     string
		defValue string
		usage    string
	}{
		{"docker", "", "Docker image (default: koyeb/sandbox)"},
		{"docker-private-registry-secret", "", "Docker private registry secret"},
		{"docker-entrypoint", "[]", "Docker entrypoint"},
		{"docker-command", "", "Docker command"},
		{"docker-args", "[]", "Docker command arguments"},
	}
	for _, want := range expected {
		flag := flags.Lookup(want.name)
		require.NotNil(t, flag, "--%s must be registered by the docker bundle", want.name)
		assert.Equal(t, want.defValue, flag.DefValue, "--%s default value", want.name)
		assert.Equal(t, want.usage, flag.Usage, "--%s help text", want.name)
	}

	// The sandbox and pool surfaces do not offer --docker-skip-verify.
	assert.Nil(t, flags.Lookup("docker-skip-verify"),
		"--docker-skip-verify must stay a service-surface flag only")
}
