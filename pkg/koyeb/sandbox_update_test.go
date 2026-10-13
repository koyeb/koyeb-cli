package koyeb

import (
	"fmt"
	"testing"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// sandboxUpdateCmd returns the `sandbox update` command under test.
func sandboxUpdateCmd(t *testing.T) *cobra.Command {
	t.Helper()
	cmd, _, err := NewSandboxCmd().Find([]string{"update"})
	require.NoError(t, err)
	return cmd
}

func TestSandboxUpdateFlagsRegistered(t *testing.T) {
	flags := sandboxUpdateCmd(t).Flags()

	// The SDK update surface: lifecycle (update_lifecycle) and egress
	// network policy (update_network_policy).
	for _, name := range sandboxUpdateFlagNames {
		assert.NotNil(t, flags.Lookup(name), "flag --%s must be registered on sandbox update", name)
	}
}

const (
	sandboxUpdateServiceID    = "323e4567-e89b-42d3-a456-426614174000"
	sandboxUpdateDeploymentID = "123e4567-e89b-42d3-a456-426614174000"
)

// sandboxUpdateFixture wires a resolvable sandbox through the fake port:
// a SANDBOX service with a live lifecycle and a latest deployment whose
// definition carries the sandbox wiring.
func sandboxUpdateFixture() *fakeAPI {
	serviceType := koyeb.SERVICETYPE_SANDBOX
	deleteAfterCreate := int64(3600)
	secret := "sandbox-secret"
	image := "koyeb/sandbox:latest"

	service := &koyeb.Service{
		Id:                 koyeb.PtrString(sandboxUpdateServiceID),
		Type:               &serviceType,
		LatestDeploymentId: koyeb.PtrString(sandboxUpdateDeploymentID),
		LifeCycle:          &koyeb.ServiceLifeCycle{DeleteAfterCreate: &deleteAfterCreate},
	}
	definition := &koyeb.DeploymentDefinition{
		Docker: &koyeb.DockerSource{Image: &image},
		Env:    []koyeb.DeploymentEnv{{Key: koyeb.PtrString(SandboxSecretKey), Value: &secret}},
	}
	deployment := koyeb.Deployment{
		Id:         koyeb.PtrString(sandboxUpdateDeploymentID),
		Definition: definition,
	}

	fake := &fakeAPI{service: service}
	fake.deployments = map[string]koyeb.Deployment{sandboxUpdateDeploymentID: deployment}
	return fake
}

// runSandboxUpdate parses args onto the update command and runs the flow
// against the fake API, returning the recorded UpdateService request.
func runSandboxUpdate(t *testing.T, fake *fakeAPI, args []string) (*koyeb.UpdateService, error) {
	t.Helper()
	cmd := sandboxUpdateCmd(t)
	require.NoError(t, cmd.Flags().Parse(args))
	err := NewSandboxHandler().Update(sandboxTestContext(fake), cmd, []string{sandboxUpdateServiceID})
	return fake.updateServiceReq, err
}

func TestSandboxUpdateWithoutFlagsFailsFast(t *testing.T) {
	fake := sandboxUpdateFixture()

	req, err := runSandboxUpdate(t, fake, nil)

	require.Error(t, err)
	assert.Nil(t, req)
	assert.Empty(t, fake.servicesFetched, "a flag-less update must fail before any API round-trip")
	assert.Contains(t, err.Error(), "nothing to update")
}

func TestSandboxUpdateRejectsNonSandboxServices(t *testing.T) {
	fake := sandboxUpdateFixture()
	webType := koyeb.SERVICETYPE_WEB
	fake.service.Type = &webType

	_, err := runSandboxUpdate(t, fake, []string{"--delete-after-delay", "1h"})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "not a sandbox")
}

// TestSandboxUpdateLifecycleMerge pins update_lifecycle's merge-not-clobber
// semantics: only the supplied fields are updated, the live values of the
// others are preserved, and 0 disables a delay.
func TestSandboxUpdateLifecycleMerge(t *testing.T) {
	t.Run("supplied delay updates, live delay is preserved", func(t *testing.T) {
		fake := sandboxUpdateFixture()

		req, err := runSandboxUpdate(t, fake, []string{"--delete-after-inactivity-delay", "30m"})

		require.NoError(t, err)
		lifecycle := req.GetLifeCycle()
		assert.Equal(t, int64(3600), lifecycle.GetDeleteAfterCreate(), "live delete_after_create must be preserved")
		assert.Equal(t, int64(1800), lifecycle.GetDeleteAfterSleep(), "supplied delay must be applied")
	})

	t.Run("both delays update in one update", func(t *testing.T) {
		fake := sandboxUpdateFixture()

		req, err := runSandboxUpdate(t, fake,
			[]string{"--delete-after-delay", "10m", "--delete-after-inactivity-delay", "30m"})

		require.NoError(t, err)
		lifecycle := req.GetLifeCycle()
		assert.Equal(t, int64(600), lifecycle.GetDeleteAfterCreate())
		assert.Equal(t, int64(1800), lifecycle.GetDeleteAfterSleep())
	})

	t.Run("a zero duration disables the delay", func(t *testing.T) {
		fake := sandboxUpdateFixture()

		req, err := runSandboxUpdate(t, fake, []string{"--delete-after-delay", "0"})

		require.NoError(t, err)
		lifecycle := req.GetLifeCycle()
		_, ok := lifecycle.GetDeleteAfterCreateOk()
		assert.False(t, ok, "--delete-after-delay 0 must unset delete_after_create")
	})

	t.Run("no lifecycle flag resends the live lifecycle unchanged", func(t *testing.T) {
		fake := sandboxUpdateFixture()

		req, err := runSandboxUpdate(t, fake, []string{"--block-network"})

		require.NoError(t, err)
		lifecycle := req.GetLifeCycle()
		assert.Equal(t, int64(3600), lifecycle.GetDeleteAfterCreate(), "the live lifecycle must be resent")
	})

	t.Run("no live lifecycle and no lifecycle flag sends no lifecycle", func(t *testing.T) {
		fake := sandboxUpdateFixture()
		fake.service.LifeCycle = nil

		req, err := runSandboxUpdate(t, fake, []string{"--block-network"})

		require.NoError(t, err)
		_, ok := req.GetLifeCycleOk()
		assert.False(t, ok, "no lifecycle must be sent when none is live and none is supplied")
	})
}

func TestSandboxUpdateSendsLatestDeploymentDefinition(t *testing.T) {
	fake := sandboxUpdateFixture()

	req, err := runSandboxUpdate(t, fake, []string{"--delete-after-delay", "10m"})

	require.NoError(t, err)
	definition := req.GetDefinition()
	docker := definition.GetDocker()
	assert.Equal(t, "koyeb/sandbox:latest", docker.GetImage())
	assert.Equal(t, []string{sandboxUpdateServiceID}, fake.updateServiceCalls)

	found := false
	for _, env := range definition.GetEnv() {
		if env.GetKey() == SandboxSecretKey {
			found = true
		}
	}
	assert.True(t, found, "the live definition (SANDBOX_SECRET included) must be the update base")
}

func TestSandboxUpdateMissingLatestDeployment(t *testing.T) {
	fake := sandboxUpdateFixture()
	delete(fake.deployments, sandboxUpdateDeploymentID)

	_, err := runSandboxUpdate(t, fake, []string{"--delete-after-delay", "10m"})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "couldn't find the latest deployment")
	assert.Empty(t, fake.updateServiceCalls)
}

func TestSandboxUpdateAPIErrorPropagates(t *testing.T) {
	fake := sandboxUpdateFixture()
	fake.updateServiceErr = fmt.Errorf("api boom")

	_, err := runSandboxUpdate(t, fake, []string{"--delete-after-delay", "10m"})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "Error while updating the sandbox")
}

// sandboxUpdateFixtureWithPolicy returns the update fixture whose latest
// deployment carries a live DENY_ALL egress policy with two allowed
// destinations.
func sandboxUpdateFixtureWithPolicy() *fakeAPI {
	fake := sandboxUpdateFixture()
	denyAll := koyeb.EGRESSPOLICYMODE_DENY_ALL
	deployment := fake.deployments[sandboxUpdateDeploymentID]
	deployment.Definition.NetworkPolicy = &koyeb.NetworkPolicy{
		Egress: &koyeb.EgressPolicy{
			Mode: &denyAll,
			AllowList: []koyeb.NetworkPolicyDestination{
				{Cidr: koyeb.PtrString("10.0.0.0/8")},
				{Cidr: koyeb.PtrString("8.8.8.8/32")},
			},
		},
	}
	fake.deployments[sandboxUpdateDeploymentID] = deployment
	return fake
}

func sandboxUpdateEgress(t *testing.T, req *koyeb.UpdateService) koyeb.EgressPolicy {
	t.Helper()
	require.True(t, req.HasDefinition(), "the update must carry a definition")
	definition := req.GetDefinition()
	require.True(t, definition.HasNetworkPolicy(), "the update must carry a network policy")
	networkPolicy := definition.GetNetworkPolicy()
	require.True(t, networkPolicy.HasEgress(), "the network policy must carry an egress policy")
	return networkPolicy.GetEgress()
}

func sandboxUpdateAllowlistCIDRs(egress koyeb.EgressPolicy) []string {
	cidrs := make([]string, 0, len(egress.GetAllowList()))
	for _, destination := range egress.GetAllowList() {
		cidrs = append(cidrs, destination.GetCidr())
	}
	return cidrs
}

// TestSandboxUpdateNetworkPolicy pins update_network_policy's behavior:
// --block-network denies all egress, --outbound-allowlist denies by default
// with normalized destinations, --no-network-policy reverts to the platform
// default, and the flag combinations raise the same validation errors as
// the SDKs.
func TestSandboxUpdateNetworkPolicy(t *testing.T) {
	t.Run("--block-network denies all outbound traffic", func(t *testing.T) {
		fake := sandboxUpdateFixture()

		req, err := runSandboxUpdate(t, fake, []string{"--block-network"})

		require.NoError(t, err)
		egress := sandboxUpdateEgress(t, req)
		assert.Equal(t, koyeb.EGRESSPOLICYMODE_DENY_ALL, egress.GetMode())
		assert.Empty(t, egress.GetAllowList())
	})

	t.Run("--outbound-allowlist normalizes bare IPs and clears host bits", func(t *testing.T) {
		fake := sandboxUpdateFixture()

		req, err := runSandboxUpdate(t, fake, []string{
			"--outbound-allowlist", "10.0.0.1/8,203.0.113.42,2001:db8::1",
		})

		require.NoError(t, err)
		egress := sandboxUpdateEgress(t, req)
		assert.Equal(t, koyeb.EGRESSPOLICYMODE_DENY_ALL, egress.GetMode())
		assert.Equal(t, []string{"10.0.0.0/8", "203.0.113.42/32", "2001:db8::1/128"},
			sandboxUpdateAllowlistCIDRs(egress))
	})

	t.Run("--outbound-allowlist merges over the live rules and ! removes", func(t *testing.T) {
		fake := sandboxUpdateFixtureWithPolicy()

		req, err := runSandboxUpdate(t, fake, []string{"--outbound-allowlist", "!10.0.0.0/8"})

		require.NoError(t, err)
		egress := sandboxUpdateEgress(t, req)
		assert.Equal(t, koyeb.EGRESSPOLICYMODE_DENY_ALL, egress.GetMode())
		assert.Equal(t, []string{"8.8.8.8/32"}, sandboxUpdateAllowlistCIDRs(egress))
	})

	t.Run("--no-network-policy reverts to the platform default", func(t *testing.T) {
		fake := sandboxUpdateFixtureWithPolicy()

		req, err := runSandboxUpdate(t, fake, []string{"--no-network-policy"})

		require.NoError(t, err)
		egress := sandboxUpdateEgress(t, req)
		assert.Equal(t, koyeb.EGRESSPOLICYMODE_DEFAULT, egress.GetMode())
		assert.Empty(t, egress.GetAllowList())
	})

	t.Run("unchanged flags keep the live policy", func(t *testing.T) {
		fake := sandboxUpdateFixtureWithPolicy()

		req, err := runSandboxUpdate(t, fake, []string{"--delete-after-delay", "10m"})

		require.NoError(t, err)
		egress := sandboxUpdateEgress(t, req)
		assert.Equal(t, koyeb.EGRESSPOLICYMODE_DENY_ALL, egress.GetMode())
		assert.Equal(t, []string{"10.0.0.0/8", "8.8.8.8/32"}, sandboxUpdateAllowlistCIDRs(egress))
	})

	t.Run("--block-network and --outbound-allowlist are mutually exclusive", func(t *testing.T) {
		fake := sandboxUpdateFixture()

		_, err := runSandboxUpdate(t, fake, []string{"--block-network", "--outbound-allowlist", "1.2.3.4"})

		require.Error(t, err)
		assert.Contains(t, err.Error(), "--block-network, --outbound-allowlist, and --no-network-policy are mutually exclusive")
		assert.Empty(t, fake.updateServiceCalls, "an invalid update must not reach the API")
	})

	t.Run("invalid destinations are rejected", func(t *testing.T) {
		for _, entry := range []string{"notanip", "fe80::1%eth0", "10.0.0.0/33"} {
			t.Run(entry, func(t *testing.T) {
				fake := sandboxUpdateFixture()

				_, err := runSandboxUpdate(t, fake, []string{"--outbound-allowlist", entry})

				require.Error(t, err)
				assert.Contains(t, err.Error(), "could not parse")
				assert.Contains(t, err.Error(), "as a CIDR or IP")
				assert.Empty(t, fake.updateServiceCalls, "an invalid update must not reach the API")
			})
		}
	})
}
