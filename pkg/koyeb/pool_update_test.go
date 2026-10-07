package koyeb

import (
	"strconv"
	"testing"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// liveSandboxPoolFixture returns a pool as the API serves it: a SANDBOX
// definition with the platform wiring (ports 3030/3031, routes, minted
// SANDBOX_SECRET) and a scale-to-zero scaling with a deep sleep delay.
func liveSandboxPoolFixture() koyeb.ServicePool {
	id := "123e4567-e89b-42d3-a456-426614174000"
	name := "my-pool"
	size := int64(5)
	secret := "minted-secret"
	image := "koyeb/sandbox:latest"
	instanceType := "micro"
	defType := koyeb.DEPLOYMENTDEFINITIONTYPE_SANDBOX
	minScale := int64(0)
	maxScale := int64(1)
	deepSleep := int64(1800)

	return koyeb.ServicePool{
		Id:   &id,
		Name: &name,
		Size: &size,
		Definition: &koyeb.DeploymentDefinition{
			Name: &name,
			Type: &defType,
			Docker: &koyeb.DockerSource{
				Image: &image,
			},
			InstanceTypes: []koyeb.DeploymentInstanceType{{Type: &instanceType}},
			Regions:       []string{"par"},
			Env: []koyeb.DeploymentEnv{
				{Key: koyeb.PtrString(SandboxSecretKey), Value: &secret},
			},
			Ports: []koyeb.DeploymentPort{
				{Port: koyeb.PtrInt64(3030), Protocol: koyeb.PtrString("http")},
				{Port: koyeb.PtrInt64(3031), Protocol: koyeb.PtrString("http")},
			},
			Routes: []koyeb.DeploymentRoute{
				{Port: koyeb.PtrInt64(3030), Path: koyeb.PtrString("/koyeb-sandbox/")},
				{Port: koyeb.PtrInt64(3031), Path: koyeb.PtrString("/")},
			},
			Scalings: []koyeb.DeploymentScaling{{
				Min: &minScale,
				Max: &maxScale,
				Targets: []koyeb.DeploymentScalingTarget{{
					SleepIdleDelay: &koyeb.DeploymentScalingTargetSleepIdleDelay{DeepSleepValue: &deepSleep},
				}},
			}},
		},
	}
}

// liveWebPoolFixture returns a WEB pool carrying explicit member wiring.
func liveWebPoolFixture() koyeb.ServicePool {
	pool := liveSandboxPoolFixture()
	defType := koyeb.DEPLOYMENTDEFINITIONTYPE_WEB
	pool.Definition.Type = &defType
	pool.Definition.Ports = []koyeb.DeploymentPort{
		{Port: koyeb.PtrInt64(8080), Protocol: koyeb.PtrString("http")},
		{Port: koyeb.PtrInt64(9090), Protocol: koyeb.PtrString("tcp")},
	}
	pool.Definition.Routes = []koyeb.DeploymentRoute{
		{Port: koyeb.PtrInt64(8080), Path: koyeb.PtrString("/")},
	}
	pool.Definition.Env = nil
	return pool
}

func TestBuildUpdateServicePool(t *testing.T) {
	t.Run("no flags changed resends the live size and definition", func(t *testing.T) {
		cmd := newPoolUpdateCmd()
		live := liveSandboxPoolFixture()

		req, err := buildUpdateServicePool(sandboxTestContext(&fakeAPI{}), cmd.Flags(), live, "my-pool")
		require.NoError(t, err)

		assert.Equal(t, int64(5), req.GetSize())
		def := req.GetDefinition()
		assert.Equal(t, live.Definition, &def, "the live definition must be resent verbatim")
	})

	t.Run("--size only keeps the definition untouched", func(t *testing.T) {
		cmd := newPoolUpdateCmd()
		require.NoError(t, cmd.Flags().Set("size", "12"))
		live := liveSandboxPoolFixture()

		req, err := buildUpdateServicePool(sandboxTestContext(&fakeAPI{}), cmd.Flags(), live, "my-pool")
		require.NoError(t, err)

		assert.Equal(t, int64(12), req.GetSize())
		def := req.GetDefinition()
		assert.Equal(t, live.Definition, &def)
	})

	t.Run("--docker replaces the image and keeps the wiring", func(t *testing.T) {
		cmd := newPoolUpdateCmd()
		require.NoError(t, cmd.Flags().Set("docker", "ghcr.io/acme/sandbox:v2"))
		live := liveSandboxPoolFixture()

		req, err := buildUpdateServicePool(sandboxTestContext(&fakeAPI{}), cmd.Flags(), live, "my-pool")
		require.NoError(t, err)

		def := req.GetDefinition()
		docker := def.Docker
		assert.Equal(t, "ghcr.io/acme/sandbox:v2", docker.GetImage())
		// The platform-minted secret and the sandbox wiring carry over.
		assert.Len(t, def.Env, 1)
		assert.Equal(t, SandboxSecretKey, def.Env[0].GetKey())
		assert.Len(t, def.Ports, 2)
		assert.Len(t, def.Routes, 2)
	})

	t.Run("--env merges over the live environment", func(t *testing.T) {
		cmd := newPoolUpdateCmd()
		require.NoError(t, cmd.Flags().Set("env", "LOG_LEVEL=debug"))
		live := liveSandboxPoolFixture()

		req, err := buildUpdateServicePool(sandboxTestContext(&fakeAPI{}), cmd.Flags(), live, "my-pool")
		require.NoError(t, err)

		env := req.GetDefinition().Env
		require.Len(t, env, 2)
		assert.Equal(t, SandboxSecretKey, env[0].GetKey(), "the live env comes first")
		assert.Equal(t, "LOG_LEVEL", env[1].GetKey())
		assert.Equal(t, "debug", env[1].GetValue())
	})

	t.Run("SANDBOX pools reject explicit ports", func(t *testing.T) {
		cmd := newPoolUpdateCmd()
		require.NoError(t, cmd.Flags().Set("ports", "8080"))
		live := liveSandboxPoolFixture()

		_, err := buildUpdateServicePool(sandboxTestContext(&fakeAPI{}), cmd.Flags(), live, "my-pool")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not allowed on SANDBOX pools")
		assert.Contains(t, err.Error(), "3030/3031")
	})

	t.Run("SANDBOX pools reject explicit routes", func(t *testing.T) {
		cmd := newPoolUpdateCmd()
		require.NoError(t, cmd.Flags().Set("routes", "/:8080"))
		live := liveSandboxPoolFixture()

		_, err := buildUpdateServicePool(sandboxTestContext(&fakeAPI{}), cmd.Flags(), live, "my-pool")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not allowed on SANDBOX pools")
	})

	t.Run("--type database is rejected", func(t *testing.T) {
		cmd := newPoolUpdateCmd()
		require.NoError(t, cmd.Flags().Set("type", "database"))
		live := liveSandboxPoolFixture()

		_, err := buildUpdateServicePool(sandboxTestContext(&fakeAPI{}), cmd.Flags(), live, "my-pool")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not supported")
	})

	t.Run("--type web is rejected on a live SANDBOX pool", func(t *testing.T) {
		cmd := newPoolUpdateCmd()
		require.NoError(t, cmd.Flags().Set("type", "web"))
		require.NoError(t, cmd.Flags().Set("ports", "8000:http"))
		live := liveSandboxPoolFixture()

		_, err := buildUpdateServicePool(sandboxTestContext(&fakeAPI{}), cmd.Flags(), live, "my-pool")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "pool type cannot be changed on update")
		assert.Contains(t, err.Error(), "mis-wired")
	})

	t.Run("re-typing toward SANDBOX is rejected too", func(t *testing.T) {
		cmd := newPoolUpdateCmd()
		require.NoError(t, cmd.Flags().Set("type", "sandbox"))
		live := liveWebPoolFixture()

		_, err := buildUpdateServicePool(sandboxTestContext(&fakeAPI{}), cmd.Flags(), live, "my-pool")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "pool type cannot be changed on update")
	})

	t.Run("the live type can be restated explicitly", func(t *testing.T) {
		cmd := newPoolUpdateCmd()
		require.NoError(t, cmd.Flags().Set("type", "sandbox"))
		live := liveSandboxPoolFixture()

		req, err := buildUpdateServicePool(sandboxTestContext(&fakeAPI{}), cmd.Flags(), live, "my-pool")
		require.NoError(t, err)

		def := req.GetDefinition()
		assert.Equal(t, live.Definition, &def, "restating the live type must be a no-op")
	})

	t.Run("WEB pools merge and delete declared ports", func(t *testing.T) {
		cmd := newPoolUpdateCmd()
		require.NoError(t, cmd.Flags().Set("ports", "!9090"))
		live := liveWebPoolFixture()

		req, err := buildUpdateServicePool(sandboxTestContext(&fakeAPI{}), cmd.Flags(), live, "my-pool")
		require.NoError(t, err)

		ports := req.GetDefinition().Ports
		require.Len(t, ports, 1)
		assert.Equal(t, int64(8080), ports[0].GetPort())
	})

	t.Run("a pool without a definition cannot be updated", func(t *testing.T) {
		cmd := newPoolUpdateCmd()

		_, err := buildUpdateServicePool(sandboxTestContext(&fakeAPI{}), cmd.Flags(), koyeb.ServicePool{}, "my-pool")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "has no definition to update")
	})
}

func TestBuildUpdateServicePoolScalings(t *testing.T) {
	t.Run("unchanged scalings are kept verbatim", func(t *testing.T) {
		cmd := newPoolUpdateCmd()
		live := liveSandboxPoolFixture()

		req, err := buildUpdateServicePool(sandboxTestContext(&fakeAPI{}), cmd.Flags(), live, "my-pool")
		require.NoError(t, err)

		scalings := req.GetDefinition().Scalings
		require.Len(t, scalings, 1)
		assert.Equal(t, live.Definition.Scalings, scalings)
	})

	t.Run("--deep-sleep-delay updates the live scale-to-zero target", func(t *testing.T) {
		cmd := newPoolUpdateCmd()
		require.NoError(t, cmd.Flags().Set("deep-sleep-delay", "10m"))
		live := liveSandboxPoolFixture()

		req, err := buildUpdateServicePool(sandboxTestContext(&fakeAPI{}), cmd.Flags(), live, "my-pool")
		require.NoError(t, err)

		scalings := req.GetDefinition().Scalings
		require.Len(t, scalings, 1)
		assert.Equal(t, int64(0), scalings[0].GetMin(), "the live min-scale carries over")
		assert.Equal(t, int64(1), scalings[0].GetMax())
		targets := scalings[0].GetTargets()
		require.Len(t, targets, 1)
		delays := targets[0].GetSleepIdleDelay()
		assert.Equal(t, int64(600), delays.GetDeepSleepValue())
		assert.False(t, delays.HasLightSleepValue())
	})

	t.Run("--deep-sleep-delay 0 clears the target", func(t *testing.T) {
		cmd := newPoolUpdateCmd()
		require.NoError(t, cmd.Flags().Set("deep-sleep-delay", "0"))
		live := liveSandboxPoolFixture()

		req, err := buildUpdateServicePool(sandboxTestContext(&fakeAPI{}), cmd.Flags(), live, "my-pool")
		require.NoError(t, err)

		scalings := req.GetDefinition().Scalings
		require.Len(t, scalings, 1)
		assert.Empty(t, scalings[0].GetTargets())
	})

	t.Run("--min-scale with carried sleep delays fails fast", func(t *testing.T) {
		cmd := newPoolUpdateCmd()
		require.NoError(t, cmd.Flags().Set("min-scale", "2"))
		live := liveSandboxPoolFixture()

		_, err := buildUpdateServicePool(sandboxTestContext(&fakeAPI{}), cmd.Flags(), live, "my-pool")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "can only be used when min-scale is 0")
	})

	t.Run("--min-scale works once the sleep delays are cleared", func(t *testing.T) {
		cmd := newPoolUpdateCmd()
		require.NoError(t, cmd.Flags().Set("min-scale", "2"))
		require.NoError(t, cmd.Flags().Set("deep-sleep-delay", "0"))
		live := liveSandboxPoolFixture()

		req, err := buildUpdateServicePool(sandboxTestContext(&fakeAPI{}), cmd.Flags(), live, "my-pool")
		require.NoError(t, err)

		scalings := req.GetDefinition().Scalings
		require.Len(t, scalings, 1)
		assert.Equal(t, int64(2), scalings[0].GetMin())
		assert.Equal(t, int64(1), scalings[0].GetMax())
		assert.Empty(t, scalings[0].GetTargets())
	})

	t.Run("--min-scale 1 builds a scaling for a pool without any", func(t *testing.T) {
		cmd := newPoolUpdateCmd()
		require.NoError(t, cmd.Flags().Set("min-scale", "1"))
		live := liveSandboxPoolFixture()
		live.Definition.Scalings = nil

		req, err := buildUpdateServicePool(sandboxTestContext(&fakeAPI{}), cmd.Flags(), live, "my-pool")
		require.NoError(t, err)

		scalings := req.GetDefinition().Scalings
		require.Len(t, scalings, 1)
		assert.Equal(t, int64(1), scalings[0].GetMin())
		assert.Equal(t, int64(1), scalings[0].GetMax())
	})
}

func TestPoolUpdateFlow(t *testing.T) {
	poolID := "123e4567-e89b-42d3-a456-426614174000"

	t.Run("refetches the live pool and resends size + definition", func(t *testing.T) {
		fake := &fakeAPI{pool: &koyeb.ServicePool{}}
		live := liveSandboxPoolFixture()
		fake.pool = &live
		ctx := sandboxTestContext(fake)
		cmd := newPoolUpdateCmd()
		require.NoError(t, cmd.Flags().Set("size", strconv.FormatInt(7, 10)))

		err := NewPoolHandler().Update(ctx, cmd, []string{"my-pool"}, poolID)
		require.NoError(t, err)

		assert.Equal(t, []string{poolID}, fake.poolsFetched, "the live pool is refetched first")
		assert.Equal(t, poolID, fake.updatedPoolID)

		req := fake.updatePoolReq
		require.NotNil(t, req)
		assert.Equal(t, int64(7), req.GetSize())
		require.NotNil(t, req.Definition, "the definition must travel with every update")
		assert.Equal(t, live.Definition, req.Definition)
	})

	t.Run("API errors surface as CLI errors", func(t *testing.T) {
		fake := &fakeAPI{updatePoolErr: assert.AnError}
		live := liveSandboxPoolFixture()
		fake.pool = &live
		ctx := sandboxTestContext(fake)
		cmd := newPoolUpdateCmd()

		err := NewPoolHandler().Update(ctx, cmd, []string{"my-pool"}, poolID)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Error while updating the pool `my-pool`")
	})

	t.Run("refetch errors surface as CLI errors", func(t *testing.T) {
		fake := &fakeAPI{getPoolErr: assert.AnError}
		ctx := sandboxTestContext(fake)
		cmd := newPoolUpdateCmd()

		err := NewPoolHandler().Update(ctx, cmd, []string{"my-pool"}, poolID)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Error while retrieving the pool `my-pool`")
	})
}

func TestPoolUpdateCmdFlagSet(t *testing.T) {
	cmd := newPoolUpdateCmd()

	// The update command declares the same curated flag set as create.
	for _, name := range []string{
		"size", "type", "ports", "routes",
		"docker", "docker-private-registry-secret", "docker-args",
		"docker-command", "docker-entrypoint", "instance-type", "regions",
		"env", "config-file", "min-scale", "light-sleep-delay", "deep-sleep-delay",
	} {
		assert.NotNil(t, cmd.Flags().Lookup(name), "flag --%s must be registered on pool update", name)
	}
	// No --name: renaming a pool is not supported.
	assert.Nil(t, cmd.Flags().Lookup("name"), "flag --name must not be registered on pool update")
}
