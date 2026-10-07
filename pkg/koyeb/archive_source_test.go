package koyeb

import (
	"testing"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/errors"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The archive source flag bundle (--archive and the --archive-* family)
// is the single registration seam of the archive source flags, composed
// by the services definition umbrella and by `deploy`. These tests pin
// the bundle contract; the per-surface suites pin the merged behavior.

func TestAddArchiveSourceFlagsServiceSurface(t *testing.T) {
	flags := pflag.NewFlagSet("service", pflag.ContinueOnError)
	addArchiveSourceFlags(flags)

	expected := []struct {
		name     string
		defValue string
		usage    string
	}{
		{"archive", "", "Archive ID to deploy"},
		{"archive-builder", "buildpack", `Builder to use, either "buildpack" (default) or "docker"`},
		{"archive-buildpack-build-command", "", "Buid command"},
		{"archive-buildpack-run-command", "", "Run command"},
		{"archive-docker-dockerfile", "", "Dockerfile path"},
		{"archive-docker-entrypoint", "[]", "Docker entrypoint"},
		{
			"archive-docker-command", "",
			"Set the docker CMD explicitly. To provide arguments to the command, use the --archive-docker-args flag.",
		},
		{
			"archive-docker-args", "[]",
			"Set arguments to the docker command. To provide multiple arguments, use the --archive-docker-args flag multiple times.",
		},
		{"archive-docker-target", "", "Docker target"},
		{
			"archive-ignore-dir", "[.git,node_modules,vendor]",
			"Set directories to ignore when building the archive.\n" +
				"To ignore multiple directories, use the flag multiple times.\n" +
				"To include all directories, set the flag to an empty string.",
		},
	}
	require.Len(t, expected, 10)
	for _, want := range expected {
		flag := flags.Lookup(want.name)
		require.NotNil(t, flag, "--%s must be registered by the archive source bundle", want.name)
		assert.Equal(t, want.defValue, flag.DefValue, "--%s default value", want.name)
		assert.Equal(t, want.usage, flag.Usage, "--%s help text", want.name)
	}
}

// archiveSourceFlagNames returns the flag names the bundle registers —
// the bundle union the surface flag-set tests derive their expectations
// from.
func archiveSourceFlagNames() []string {
	flags := pflag.NewFlagSet("archive-source", pflag.ContinueOnError)
	addArchiveSourceFlags(flags)
	names := make([]string, 0, 10)
	flags.VisitAll(func(f *pflag.Flag) { names = append(names, f.Name) })
	return names
}

// archiveSourceTestFlagSet registers the bundle surface the parse step
// consumes: the archive flags plus --privileged, which every composing
// surface registers next to the bundle and the builder helpers read.
func archiveSourceTestFlagSet() *pflag.FlagSet {
	flags := pflag.NewFlagSet("archive-source", pflag.ContinueOnError)
	addArchiveSourceFlags(flags)
	flags.Bool("privileged", false, "")
	return flags
}

// liveArchiveSource returns a source carrying a live value on the bundle
// fields, so the changed-only merge conventions are observable.
func liveArchiveSource() *koyeb.ArchiveSource {
	return &koyeb.ArchiveSource{
		Id:        koyeb.PtrString("live-archive"),
		Buildpack: &koyeb.BuildpackBuilder{},
	}
}

func TestParseArchiveSourceNoFlagsKeepsLiveValues(t *testing.T) {
	h := &ServiceHandler{}
	flags := archiveSourceTestFlagSet()

	source, err := h.parseArchiveSource(flags, liveArchiveSource())
	require.NoError(t, err)
	assert.Equal(t, "live-archive", source.GetId())
	assert.True(t, source.HasBuildpack(), "the live buildpack builder must be kept")
}

func TestParseArchiveSourceChangedFlagsMergeLiveValues(t *testing.T) {
	h := &ServiceHandler{}

	t.Run("--archive overrides the live archive ID", func(t *testing.T) {
		flags := archiveSourceTestFlagSet()
		require.NoError(t, flags.Parse([]string{"--archive", "my-archive"}))

		source, err := h.parseArchiveSource(flags, liveArchiveSource())
		require.NoError(t, err)
		assert.Equal(t, "my-archive", source.GetId())
	})

	t.Run("the buildpack commands merge into the live builder", func(t *testing.T) {
		flags := archiveSourceTestFlagSet()
		require.NoError(t, flags.Parse([]string{
			"--archive-buildpack-build-command", "build",
			"--archive-buildpack-run-command", "run",
		}))

		source, err := h.parseArchiveSource(flags, liveArchiveSource())
		require.NoError(t, err)
		require.True(t, source.HasBuildpack())
		buildpack := source.GetBuildpack()
		assert.Equal(t, "build", *buildpack.BuildCommand)
		assert.Equal(t, "run", *buildpack.RunCommand)
	})

	t.Run("--archive-builder docker switches to the docker builder", func(t *testing.T) {
		flags := archiveSourceTestFlagSet()
		require.NoError(t, flags.Parse([]string{
			"--archive-builder", "docker",
			"--archive-docker-dockerfile", "Dockerfile.dev",
			"--archive-docker-entrypoint", "entry.sh",
			"--archive-docker-command", "cmd",
			"--archive-docker-args", "arg1", "--archive-docker-args", "arg2",
			"--archive-docker-target", "dev",
			"--privileged",
		}))

		source, err := h.parseArchiveSource(flags, liveArchiveSource())
		require.NoError(t, err)
		require.True(t, source.HasDocker(), "the buildpack builder must be replaced")
		docker := source.GetDocker()
		assert.Equal(t, "Dockerfile.dev", *docker.Dockerfile)
		assert.Equal(t, []string{"entry.sh"}, docker.Entrypoint)
		assert.Equal(t, "cmd", *docker.Command)
		assert.Equal(t, []string{"arg1", "arg2"}, docker.Args)
		assert.Equal(t, "dev", *docker.Target)
		assert.True(t, docker.GetPrivileged())
	})
}

func TestParseArchiveSourceErrorPaths(t *testing.T) {
	h := &ServiceHandler{}

	for name, tc := range map[string]struct {
		args    []string
		source  *koyeb.ArchiveSource
		wantWhy string
	}{
		"an invalid builder": {
			args:    []string{"--archive-builder", "xxx"},
			source:  &koyeb.ArchiveSource{},
			wantWhy: "the --archive-builder is invalid",
		},
		"docker builder flags on a buildpack source": {
			args:    []string{"--archive-docker-command", "cmd"},
			source:  &koyeb.ArchiveSource{Buildpack: &koyeb.BuildpackBuilder{}},
			wantWhy: "invalid flag combination",
		},
	} {
		t.Run(name, func(t *testing.T) {
			flags := archiveSourceTestFlagSet()
			require.NoError(t, flags.Parse(tc.args))

			_, err := h.parseArchiveSource(flags, tc.source)
			var cliErr *errors.CLIError
			require.ErrorAs(t, err, &cliErr)
			assert.Equal(t, tc.wantWhy, cliErr.Why)
		})
	}
}

func TestParseArchiveIgnoreDirectories(t *testing.T) {
	h := &ServiceHandler{}

	t.Run("the flag default carries the built-in ignore set", func(t *testing.T) {
		flags := archiveSourceTestFlagSet()
		handler := &ArchiveHandler{}

		require.NoError(t, h.ParseArchiveIgnoreDirectories(flags, handler))
		assert.Equal(t, []string{".git", "node_modules", "vendor"}, handler.ignoreDirectories)
	})

	t.Run("the flag values land on the archive handler", func(t *testing.T) {
		flags := archiveSourceTestFlagSet()
		require.NoError(t, flags.Parse([]string{
			"--archive-ignore-dir", "build", "--archive-ignore-dir", "tmp",
		}))
		handler := &ArchiveHandler{}

		require.NoError(t, h.ParseArchiveIgnoreDirectories(flags, handler))
		assert.Equal(t, []string{"build", "tmp"}, handler.ignoreDirectories)
	})

	t.Run("an empty value ignores no directory", func(t *testing.T) {
		flags := archiveSourceTestFlagSet()
		require.NoError(t, flags.Parse([]string{"--archive-ignore-dir", ""}))
		handler := &ArchiveHandler{}

		require.NoError(t, h.ParseArchiveIgnoreDirectories(flags, handler))
		assert.Empty(t, handler.ignoreDirectories)
	})
}
