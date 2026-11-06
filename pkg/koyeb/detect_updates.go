package koyeb

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"

	"github.com/blang/semver"
	"github.com/rhysd/go-github-selfupdate/selfupdate"
	log "github.com/sirupsen/logrus"
)

const DevVersion = "develop"

var (
	ghAuthToken = func() ([]byte, error) {
		return exec.Command("gh", "auth", "token").Output()
	}
	detectLatestRelease = func(slug string) (*selfupdate.Release, bool, error) {
		updater, err := selfupdate.NewUpdater(selfupdate.Config{APIToken: githubAPIToken()})
		if err != nil {
			return nil, false, err
		}
		return updater.DetectLatest(slug)
	}
)

func githubAPIToken() string {
	for _, env := range []string{"GH_TOKEN", "GITHUB_TOKEN"} {
		if token := strings.TrimSpace(os.Getenv(env)); token != "" {
			return token
		}
	}

	token, err := ghAuthToken()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(token))
}

func DetectUpdates() {
	if Version == DevVersion {
		return
	}
	version, err := semver.Parse(Version)
	if err != nil {
		log.Errorf("unable to parse version: %v", err)
		return
	}

	detectUpdateFile := path.Join(os.TempDir(), "koyeb-cli-detect-update")
	dFile, _ := os.Stat(detectUpdateFile)

	if dFile != nil {
		oneHourAgo := time.Now().Add(-time.Hour)
		if dFile.ModTime().After(oneHourAgo) {
			return
		}
	}
	latest, found, err := detectLatestRelease(GithubRepo)
	if err != nil {
		log.Debugf("unable to detect latest version: %v", err)
		return
	}
	if !found {
		return
	}

	if latest.Version.Compare(version) > 0 {
		fmt.Fprintf(os.Stderr, "* A new version of the koyeb-cli (%s) is available *\nSee update instructions here: %s\n", latest.Version, latest.URL)
	}
	if dFile == nil {
		if _, err := os.Create(detectUpdateFile); err != nil {
			log.Debugf("Unable to create detect update file: %v", err)
			return
		}
	} else {
		now := time.Now().Local()
		if err := os.Chtimes(detectUpdateFile, now, now); err != nil {
			log.Debugf("Unable to update detect update file: %v", err)
			return
		}
	}
}
