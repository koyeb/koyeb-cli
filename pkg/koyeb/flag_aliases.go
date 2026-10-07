package koyeb

import "github.com/spf13/pflag"

// flagAliases maps the legacy flag names to their canonical names, for
// example --port to --ports. The map is public interface: every alias is
// kept verbatim, including the historical "health-checks-graee" typo.
var flagAliases = map[string]string{
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

// normalizeFlagAlias is the flag-name normalization seam shared by the
// service umbrella (service create/update, app init, deploy) and the
// pool surfaces: it maps the legacy alias names to their canonical flags.
// Aliases whose canonical flag a surface does not register stay inert —
// normalization only affects names whose canonical flag exists there;
// they activate when their bundle lands on the surface.
func normalizeFlagAlias(_ *pflag.FlagSet, name string) pflag.NormalizedName {
	alias, exists := flagAliases[name]
	if exists {
		name = alias
	}
	return pflag.NormalizedName(name)
}
