package koyeb

import (
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
)

// `koyeb deploy` composes the all-sources definition bundles plus the
// archive source bundle: it deploys a local directory as a freshly built
// archive, so the git and docker source bundles stay off the surface.

func TestDeployCmdFlagSet(t *testing.T) {
	cmd := NewDeployCmd()
	flags := cmd.Flags()

	// The deploy command flags.
	for _, name := range []string{"app", "wait", "wait-timeout", "service-account-id"} {
		assert.NotNil(t, flags.Lookup(name), "flag --%s must be registered on deploy", name)
	}

	// The definition flag-set: the all-sources surface base plus the union
	// of the composed bundles, with the archive source bundle added on top.
	for _, name := range slices.Concat(
		[]string{
			"type", "deployment-strategy", "privileged", "skip-cache",
			"delete-after-delay", "delete-after-inactivity-delay",
			"routes", "ports", "auth", "auth-disable",
		},
		envConfigFilesFlagNames(serviceEnvConfigFilesFlagUsage),
		instanceTypeRegionsFlagNames(serviceInstanceTypeRegionsFlagUsage),
		scalingSleepDelayFlagNames(serviceScalingSleepDelayFlagUsage),
		networkPolicyFlagNames(),
		checksFlagNames(),
		volumesFlagNames(),
		proxyPortsFlagNames(),
		archiveSourceFlagNames(),
	) {
		assert.NotNil(t, flags.Lookup(name), "flag --%s must be registered on deploy", name)
	}

	// The git and docker source bundles must not leak onto the surface:
	// their entries derive from the bundle registrations, so the negative
	// set grows with the bundles.
	for _, name := range slices.Concat(
		gitSourceFlagNames(),
		dockerSourceFlagNames(serviceDockerSourceFlagUsage),
	) {
		assert.Nil(t, flags.Lookup(name), "flag --%s must not be registered on deploy", name)
	}
}
