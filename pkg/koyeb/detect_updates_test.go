package koyeb

import (
	"errors"
	"io"
	"os"
	"path"
	"testing"

	"github.com/rhysd/go-github-selfupdate/selfupdate"
	"github.com/stretchr/testify/require"
)

func TestDetectUpdatesIgnoresLatestVersionCheckErrors(t *testing.T) {
	origVersion := Version
	origGithubRepo := GithubRepo
	origStderr := os.Stderr
	origDetectLatestRelease := detectLatestRelease
	detectUpdateFile := path.Join(os.TempDir(), "koyeb-cli-detect-update")
	t.Cleanup(func() {
		Version = origVersion
		GithubRepo = origGithubRepo
		os.Stderr = origStderr
		detectLatestRelease = origDetectLatestRelease
		_ = os.Remove(detectUpdateFile)
	})

	Version = "1.0.0"
	GithubRepo = "koyeb/koyeb-cli"
	detectLatestRelease = func(string) (*selfupdate.Release, bool, error) {
		return nil, false, errors.New("github unavailable")
	}
	require.NoError(t, os.RemoveAll(detectUpdateFile))

	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stderr = w

	DetectUpdates()

	require.NoError(t, w.Close())
	stderr, err := io.ReadAll(r)
	require.NoError(t, err)
	require.Empty(t, string(stderr))
}

func TestGithubAPITokenFromEnvironment(t *testing.T) {
	origGHAuthToken := ghAuthToken
	t.Cleanup(func() {
		ghAuthToken = origGHAuthToken
	})
	t.Setenv("GH_TOKEN", "gh-token")
	t.Setenv("GITHUB_TOKEN", "github-token")
	ghAuthToken = func() ([]byte, error) {
		return nil, errors.New("gh should not be called")
	}

	require.Equal(t, "gh-token", githubAPIToken())
}

func TestGithubAPITokenFromGHCLI(t *testing.T) {
	origGHAuthToken := ghAuthToken
	t.Cleanup(func() {
		ghAuthToken = origGHAuthToken
	})
	t.Setenv("GH_TOKEN", "")
	t.Setenv("GITHUB_TOKEN", "")
	ghAuthToken = func() ([]byte, error) {
		return []byte("gh-cli-token\n"), nil
	}

	require.Equal(t, "gh-cli-token", githubAPIToken())
}

func TestGithubAPITokenIgnoresGHCLIError(t *testing.T) {
	origGHAuthToken := ghAuthToken
	t.Cleanup(func() {
		ghAuthToken = origGHAuthToken
	})
	t.Setenv("GH_TOKEN", "")
	t.Setenv("GITHUB_TOKEN", "")
	ghAuthToken = func() ([]byte, error) {
		return nil, errors.New("gh is unavailable")
	}

	require.Empty(t, githubAPIToken())
}
