package koyeb

import (
	"testing"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
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

// dockerBundleTestFlagSet registers the full bundle surface the parse
// step consumes: the docker flags plus --privileged, which every surface
// registers next to the bundle.
func dockerBundleTestFlagSet() *pflag.FlagSet {
	flags := pflag.NewFlagSet("docker-bundle", pflag.ContinueOnError)
	addDockerSourceFlags(flags, serviceDockerSourceFlagUsage)
	flags.Bool("privileged", false, "")
	return flags
}

// liveDockerSource returns a source carrying a value on every bundle
// field, so the changed-only merge conventions are observable.
func liveDockerSource() *koyeb.DockerSource {
	source := koyeb.NewDockerSourceWithDefaults()
	source.SetImage("live/image:v1")
	source.SetImageRegistrySecret("live-secret")
	source.SetArgs([]string{"live-arg"})
	source.SetCommand("live-cmd")
	source.SetEntrypoint([]string{"live-entry"})
	source.SetPrivileged(true)
	return source
}

func TestParseDockerSourceNoFlagsKeepsLiveValues(t *testing.T) {
	h := &ServiceHandler{}
	flags := dockerBundleTestFlagSet()

	source, changed, err := h.parseDockerSource(nil, flags, liveDockerSource(), dockerSourceParseOptions{})
	require.NoError(t, err)
	assert.False(t, changed)
	assert.Equal(t, "live/image:v1", source.GetImage())
	assert.Equal(t, "live-secret", source.GetImageRegistrySecret())
	assert.Equal(t, []string{"live-arg"}, source.GetArgs())
	assert.Equal(t, "live-cmd", source.GetCommand())
	assert.Equal(t, []string{"live-entry"}, source.GetEntrypoint())
	assert.True(t, source.GetPrivileged())
}

func TestParseDockerSourceChangedFlagsOverrideLiveValues(t *testing.T) {
	h := &ServiceHandler{}
	flags := dockerBundleTestFlagSet()
	require.NoError(t, flags.Parse([]string{
		"--docker", "nginx:latest",
		"--docker-private-registry-secret", "my-secret",
		"--docker-args", "arg1", "--docker-args", "arg2",
		"--docker-command", "nginx",
		"--docker-entrypoint", "entry.sh",
		"--privileged=false",
	}))

	// The pure option is what the pool builders use: the nil CLI context
	// proves the apply path makes no API call.
	source, changed, err := h.parseDockerSource(nil, flags, liveDockerSource(), dockerSourceParseOptions{})
	require.NoError(t, err)
	assert.True(t, changed)
	assert.Equal(t, "nginx:latest", source.GetImage())
	assert.Equal(t, "my-secret", source.GetImageRegistrySecret())
	assert.Equal(t, []string{"arg1", "arg2"}, source.GetArgs())
	assert.Equal(t, "nginx", source.GetCommand())
	assert.Equal(t, []string{"entry.sh"}, source.GetEntrypoint())
	assert.False(t, source.GetPrivileged())
}

func TestParseDockerSourcePrivilegedAloneReportsChanged(t *testing.T) {
	h := &ServiceHandler{}
	flags := dockerBundleTestFlagSet()
	require.NoError(t, flags.Parse([]string{"--privileged=false"}))

	source, changed, err := h.parseDockerSource(nil, flags, liveDockerSource(), dockerSourceParseOptions{})
	require.NoError(t, err)
	assert.True(t, changed, "--privileged alone must mark the source changed")
	assert.False(t, source.GetPrivileged())
	assert.Equal(t, "live/image:v1", source.GetImage(), "the live image must be kept")
}

func TestParseDockerSourceDefaultImageWhenDockerUnset(t *testing.T) {
	h := &ServiceHandler{}
	createOptions := dockerSourceParseOptions{defaultImage: koyebSandboxImage}

	t.Run("unset --docker defaults the image", func(t *testing.T) {
		flags := dockerBundleTestFlagSet()

		source, changed, err := h.parseDockerSource(nil, flags, koyeb.NewDockerSourceWithDefaults(), createOptions)
		require.NoError(t, err)
		assert.False(t, changed, "the default image alone does not mark the source changed")
		assert.Equal(t, koyebSandboxImage, source.GetImage())
	})

	t.Run("--docker wins over the default image", func(t *testing.T) {
		flags := dockerBundleTestFlagSet()
		require.NoError(t, flags.Parse([]string{"--docker", "nginx"}))

		source, _, err := h.parseDockerSource(nil, flags, koyeb.NewDockerSourceWithDefaults(), createOptions)
		require.NoError(t, err)
		assert.Equal(t, "nginx", source.GetImage())
	})

	t.Run("without a default the live image is kept", func(t *testing.T) {
		flags := dockerBundleTestFlagSet()

		source, changed, err := h.parseDockerSource(nil, flags, liveDockerSource(), dockerSourceParseOptions{})
		require.NoError(t, err)
		assert.False(t, changed)
		assert.Equal(t, "live/image:v1", source.GetImage())
	})
}

func TestParseDockerSourceVerifyImageOption(t *testing.T) {
	h := &ServiceHandler{}

	t.Run("verification on reaches the API call", func(t *testing.T) {
		flags := dockerBundleTestFlagSet()
		require.NoError(t, flags.Parse([]string{"--docker", "nginx"}))

		// With verification on, the bundle reaches checkDockerImage: the
		// nil CLI context cannot serve that API call, which pins the
		// service surface's verification path.
		require.Panics(t, func() {
			_, _, _ = h.parseDockerSource(nil, flags, koyeb.NewDockerSourceWithDefaults(),
				dockerSourceParseOptions{verifyImage: true})
		})
	})

	t.Run("--docker-skip-verify opts out", func(t *testing.T) {
		flags := dockerBundleTestFlagSet()
		require.NoError(t, flags.Parse([]string{"--docker", "nginx", "--docker-skip-verify"}))

		source, changed, err := h.parseDockerSource(nil, flags, koyeb.NewDockerSourceWithDefaults(),
			dockerSourceParseOptions{verifyImage: true})
		require.NoError(t, err)
		assert.True(t, changed)
		assert.Equal(t, "nginx", source.GetImage())
	})
}
