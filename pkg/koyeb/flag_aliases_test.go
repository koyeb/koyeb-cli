package koyeb

import (
	"testing"

	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
)

// The shared flag-alias map is the single name-normalization seam of
// the service umbrella (service create/update, app init, deploy) and
// the pool surfaces. This test pins the map itself — every legacy
// alias is public interface, the historical "health-checks-graee" typo
// included — and that every entry normalizes to its canonical name.

func TestFlagAliases(t *testing.T) {
	// The pinned contract: alias name to canonical flag name.
	want := map[string]string{
		"port":  "ports",
		"proxy": "proxy-ports",
		"check": "checks",

		"healthcheck":              "checks",
		"healthcheck-grace":        "checks-grace-period",
		"healthcheck-grace-period": "checks-grace-period",

		"health-check":              "checks",
		"health-check-grace":        "checks-grace-period",
		"health-check-grace-period": "checks-grace-period",

		"healthchecks":              "checks",
		"healthchecks-grace":        "checks-grace-period",
		"healthchecks-grace-period": "checks-grace-period",

		"health-checks":              "checks",
		"health-checks-graee":        "checks-grace-period",
		"health-checks-grace-period": "checks-grace-period",

		"strategy": "deployment-strategy",

		"route":              "routes",
		"volume":             "volumes",
		"region":             "regions",
		"git-docker-arg":     "git-docker-args",
		"docker-arg":         "docker-args",
		"archive-docker-arg": "archive-docker-args",
	}

	assert.Equal(t, want, flagAliases,
		"the shared alias map is public interface: no entry may be added, dropped or retargeted")

	t.Run("every entry normalizes to its canonical name", func(t *testing.T) {
		for alias, canonical := range want {
			assert.Equal(t, pflag.NormalizedName(canonical), normalizeFlagAlias(nil, alias),
				"alias --%s must normalize to --%s", alias, canonical)
		}
	})

	t.Run("canonical and unknown names pass through untouched", func(t *testing.T) {
		for _, name := range []string{"ports", "docker-args", "size"} {
			assert.Equal(t, pflag.NormalizedName(name), normalizeFlagAlias(nil, name),
				"non-alias --%s must pass through untouched", name)
		}
	})
}
