package koyeb

import (
	"testing"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/errors"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The git source flag bundle (--git and the --git-* family) is the single
// registration seam of the git source flags. These tests pin the bundle
// contract; the builder-switching matrix is pinned by the service suite
// (TestSetGitSourceBuilder), the per-surface suites pin the merged
// behavior.

func TestAddGitSourceFlagsServiceSurface(t *testing.T) {
	flags := pflag.NewFlagSet("service", pflag.ContinueOnError)
	addGitSourceFlags(flags)

	expected := []struct {
		name     string
		defValue string
		usage    string
	}{
		{"git", "", "Git repository"},
		{"git-branch", "main", "Git branch"},
		{"git-sha", "", "Git commit SHA to deploy"},
		{
			"git-no-deploy-on-push", "false",
			"Disable new deployments creation when code changes are pushed on the configured branch",
		},
		{"git-workdir", "", "Path to the sub-directory containing the code to build and deploy"},
		{"git-credential-source", "", "Source of the Git repository credentials"},
		{"git-builder", "buildpack", `Builder to use, either "buildpack" (default) or "docker"`},
		{"git-build-command", "", "Buid command (legacy, prefer git-buildpack-build-command)"},
		{"git-run-command", "", "Run command (legacy, prefer git-buildpack-run-command)"},
		{"git-buildpack-build-command", "", "Buid command"},
		{"git-buildpack-run-command", "", "Run command"},
		{"git-docker-dockerfile", "", "Dockerfile path"},
		{"git-docker-entrypoint", "[]", "Docker entrypoint"},
		{
			"git-docker-command", "",
			"Set the docker CMD explicitly. To provide arguments to the command, use the --git-docker-args flag.",
		},
		{
			"git-docker-args", "[]",
			"Set arguments to the docker command. To provide multiple arguments, use the --git-docker-args flag multiple times.",
		},
		{"git-docker-target", "", "Docker target"},
	}
	require.Len(t, expected, 16)
	for _, want := range expected {
		flag := flags.Lookup(want.name)
		require.NotNil(t, flag, "--%s must be registered by the git source bundle", want.name)
		assert.Equal(t, want.defValue, flag.DefValue, "--%s default value", want.name)
		assert.Equal(t, want.usage, flag.Usage, "--%s help text", want.name)
	}
}

// gitSourceFlagNames returns the flag names the bundle registers — the
// bundle union the surface flag-set tests derive their expectations from.
func gitSourceFlagNames() []string {
	flags := pflag.NewFlagSet("git-source", pflag.ContinueOnError)
	addGitSourceFlags(flags)
	names := make([]string, 0, 16)
	flags.VisitAll(func(f *pflag.Flag) { names = append(names, f.Name) })
	return names
}

// gitSourceTestFlagSet registers the bundle surface the parse step
// consumes: the git flags plus --privileged, which every composing
// surface registers next to the bundle and the builder helpers read.
func gitSourceTestFlagSet() *pflag.FlagSet {
	flags := pflag.NewFlagSet("git-source", pflag.ContinueOnError)
	addGitSourceFlags(flags)
	flags.Bool("privileged", false, "")
	return flags
}

// liveGitSource returns a source carrying a value on every bundle field,
// so the changed-only merge conventions are observable.
func liveGitSource() *koyeb.GitSource {
	return &koyeb.GitSource{
		Repository:       koyeb.PtrString("github.com/org/repo"),
		Branch:           koyeb.PtrString("live-branch"),
		Sha:              koyeb.PtrString("livesha"),
		NoDeployOnPush:   koyeb.PtrBool(false),
		Workdir:          koyeb.PtrString("/live"),
		CredentialSource: koyeb.PtrString("live-credentials"),
		Buildpack:        &koyeb.BuildpackBuilder{},
	}
}

func TestParseGitSourceNoFlagsKeepsLiveValues(t *testing.T) {
	h := &ServiceHandler{}
	flags := gitSourceTestFlagSet()

	source, err := h.parseGitSource(flags, liveGitSource())
	require.NoError(t, err)
	assert.Equal(t, "github.com/org/repo", source.GetRepository())
	assert.Equal(t, "live-branch", source.GetBranch())
	assert.Equal(t, "livesha", source.GetSha())
	assert.False(t, source.GetNoDeployOnPush())
	assert.Equal(t, "/live", source.GetWorkdir())
	assert.Equal(t, "live-credentials", source.GetCredentialSource())
	assert.True(t, source.HasBuildpack(), "the live buildpack builder must be kept")
}

func TestParseGitSourceChangedFlagsMergeLiveValues(t *testing.T) {
	h := &ServiceHandler{}

	t.Run("the source flags merge over the live source", func(t *testing.T) {
		flags := gitSourceTestFlagSet()
		require.NoError(t, flags.Parse([]string{
			"--git", "github.com/org/other",
			"--git-branch", "feature",
			"--git-sha", "abc123",
			"--git-no-deploy-on-push",
			"--git-workdir", "/app",
			"--git-credential-source", "connector:123e4567-e89b-12d3-a456-426614174000",
		}))

		source, err := h.parseGitSource(flags, liveGitSource())
		require.NoError(t, err)
		assert.Equal(t, "github.com/org/other", source.GetRepository())
		assert.Equal(t, "feature", source.GetBranch())
		assert.Equal(t, "abc123", source.GetSha())
		assert.True(t, source.GetNoDeployOnPush())
		assert.Equal(t, "/app", source.GetWorkdir())
		assert.Equal(t, "connector:123e4567-e89b-12d3-a456-426614174000", source.GetCredentialSource())
	})

	t.Run("a fresh source defaults the branch to the flag default", func(t *testing.T) {
		flags := gitSourceTestFlagSet()
		require.NoError(t, flags.Parse([]string{"--git", "github.com/org/repo"}))

		source, err := h.parseGitSource(flags, &koyeb.GitSource{})
		require.NoError(t, err)
		assert.Equal(t, "github.com/org/repo", source.GetRepository())
		assert.Equal(t, "main", source.GetBranch())
	})
}

func TestParseGitSourceErrorPaths(t *testing.T) {
	h := &ServiceHandler{}

	for name, tc := range map[string]struct {
		args    []string
		source  *koyeb.GitSource
		wantWhy string
	}{
		"an invalid builder": {
			args:    []string{"--git-builder", "xxx"},
			source:  &koyeb.GitSource{},
			wantWhy: "the --git-builder is invalid",
		},
		"docker builder flags on a buildpack source": {
			args:    []string{"--git-docker-command", "cmd"},
			source:  &koyeb.GitSource{Buildpack: &koyeb.BuildpackBuilder{}},
			wantWhy: "invalid flag combination",
		},
		"the legacy and buildpack build commands together": {
			args: []string{
				"--git-builder", "buildpack",
				"--git-build-command", "build",
				"--git-buildpack-build-command", "build",
			},
			source:  &koyeb.GitSource{Buildpack: &koyeb.BuildpackBuilder{}},
			wantWhy: "can't use --git-build-command and --git-buildpack-build-command together",
		},
	} {
		t.Run(name, func(t *testing.T) {
			flags := gitSourceTestFlagSet()
			require.NoError(t, flags.Parse(tc.args))

			_, err := h.parseGitSource(flags, tc.source)
			var cliErr *errors.CLIError
			require.ErrorAs(t, err, &cliErr)
			assert.Equal(t, tc.wantWhy, cliErr.Why)
		})
	}
}
