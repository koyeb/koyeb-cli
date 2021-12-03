package koyeb

import (
	"fmt"
	"os"
	"path"
	"time"

	"github.com/blang/semver"
	"github.com/rhysd/go-github-selfupdate/selfupdate"
)

const DevVersion = "develop"

func DetectUpdates() {
	if Version == DevVersion {
		return
	}
	version, err := semver.Parse(Version)
	if err != nil {
		fmt.Fprintf(os.Stderr, "unable to parse version: %v", err)
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
	latest, found, err := selfupdate.DetectLatest(GithubRepo)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	if !found {
		return
	}

	if latest.Version.Compare(version) > 0 {
		fmt.Fprintf(os.Stderr, "* A new version of the koyeb-cli (%s) is available *\nSee update instructions here: %s\n", latest.Version, latest.URL)
	}
	if dFile == nil {
		os.Create(detectUpdateFile)
	} else {
		now := time.Now().Local()
		os.Chtimes(detectUpdateFile, now, now)
	}
}
