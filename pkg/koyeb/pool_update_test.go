package koyeb

import (
	"slices"
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

// liveArchivePoolFixture returns a pool as the API serves it when the
// members boot from an archive: the definition carries an archive
// source with a buildpack builder next to the platform wiring.
func liveArchivePoolFixture() koyeb.ServicePool {
	pool := liveSandboxPoolFixture()
	archiveID := "live-archive"
	pool.Definition.Docker = nil
	pool.Definition.Archive = &koyeb.ArchiveSource{
		Id:        &archiveID,
		Buildpack: &koyeb.BuildpackBuilder{},
	}
	return pool
}

func TestBuildUpdateServicePoolArchiveSource(t *testing.T) {
	t.Run("--archive replaces the live Docker source", func(t *testing.T) {
		cmd := newPoolUpdateCmd()
		require.NoError(t, cmd.Flags().Set("archive", "my-archive"))
		live := liveSandboxPoolFixture()

		req, err := buildUpdateServicePool(sandboxTestContext(&fakeAPI{}), cmd.Flags(), live, "my-pool")
		require.NoError(t, err)

		def := req.GetDefinition()
		archive := def.Archive
		require.NotNil(t, archive, "the definition must carry the archive source")
		assert.Equal(t, "my-archive", archive.GetId())
		assert.False(t, def.HasDocker(), "the archive source replaces the live docker source")
	})

	t.Run("no flags changed resends the live archive verbatim", func(t *testing.T) {
		cmd := newPoolUpdateCmd()
		live := liveArchivePoolFixture()

		req, err := buildUpdateServicePool(sandboxTestContext(&fakeAPI{}), cmd.Flags(), live, "my-pool")
		require.NoError(t, err)

		def := req.GetDefinition()
		assert.Equal(t, live.Definition, &def, "the live archive source must be resent verbatim")
	})

	t.Run("--archive merges the changed flags over the live archive", func(t *testing.T) {
		cmd := newPoolUpdateCmd()
		require.NoError(t, cmd.Flags().Set("archive", "new-archive"))
		require.NoError(t, cmd.Flags().Set("archive-buildpack-run-command", "run"))
		live := liveArchivePoolFixture()

		req, err := buildUpdateServicePool(sandboxTestContext(&fakeAPI{}), cmd.Flags(), live, "my-pool")
		require.NoError(t, err)

		archive := req.GetDefinition().Archive
		require.NotNil(t, archive)
		assert.Equal(t, "new-archive", archive.GetId())
		buildpack := archive.GetBuildpack()
		assert.Equal(t, "run", *buildpack.RunCommand)
	})

	t.Run("--privileged applies to the live archive builder", func(t *testing.T) {
		cmd := newPoolUpdateCmd()
		require.NoError(t, cmd.Flags().Set("privileged", "true"))
		live := liveArchivePoolFixture()

		req, err := buildUpdateServicePool(sandboxTestContext(&fakeAPI{}), cmd.Flags(), live, "my-pool")
		require.NoError(t, err)

		def := req.GetDefinition()
		assert.False(t, def.HasDocker(), "--privileged must not inject a docker source")
		archive := def.Archive
		require.NotNil(t, archive)
		buildpack := archive.GetBuildpack()
		assert.True(t, buildpack.GetPrivileged())
	})

	t.Run("--docker replaces the live archive source", func(t *testing.T) {
		cmd := newPoolUpdateCmd()
		require.NoError(t, cmd.Flags().Set("docker", "ghcr.io/acme/sandbox:v2"))
		live := liveArchivePoolFixture()

		req, err := buildUpdateServicePool(sandboxTestContext(&fakeAPI{}), cmd.Flags(), live, "my-pool")
		require.NoError(t, err)

		def := req.GetDefinition()
		assert.False(t, def.HasArchive(), "the docker source replaces the live archive source")
		docker := def.Docker
		require.NotNil(t, docker)
		assert.Equal(t, "ghcr.io/acme/sandbox:v2", docker.GetImage())
	})

	t.Run("conflicting --docker and --archive flags are rejected", func(t *testing.T) {
		cmd := newPoolUpdateCmd()
		require.NoError(t, cmd.Flags().Set("docker", "ghcr.io/acme/sandbox"))
		require.NoError(t, cmd.Flags().Set("archive", "my-archive"))
		live := liveSandboxPoolFixture()

		_, err := buildUpdateServicePool(sandboxTestContext(&fakeAPI{}), cmd.Flags(), live, "my-pool")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid flag combination")
	})
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

	t.Run("--privileged flags the member containers", func(t *testing.T) {
		cmd := newPoolUpdateCmd()
		require.NoError(t, cmd.Flags().Set("privileged", "true"))
		live := liveSandboxPoolFixture()

		req, err := buildUpdateServicePool(sandboxTestContext(&fakeAPI{}), cmd.Flags(), live, "my-pool")
		require.NoError(t, err)

		docker := req.GetDefinition().Docker
		assert.True(t, docker.GetPrivileged())
	})

	t.Run("an unchanged --privileged keeps the live value", func(t *testing.T) {
		cmd := newPoolUpdateCmd()
		live := liveSandboxPoolFixture()
		live.Definition.Docker.Privileged = koyeb.PtrBool(true)

		req, err := buildUpdateServicePool(sandboxTestContext(&fakeAPI{}), cmd.Flags(), live, "my-pool")
		require.NoError(t, err)

		docker := req.GetDefinition().Docker
		assert.True(t, docker.GetPrivileged())
	})

	t.Run("--block-network replaces the live member policy", func(t *testing.T) {
		cmd := newPoolUpdateCmd()
		require.NoError(t, cmd.Flags().Set("block-network", "true"))
		live := liveSandboxPoolFixture()
		allowCIDR := "10.0.0.0/8"
		mode := koyeb.EGRESSPOLICYMODE_DENY_ALL
		live.Definition.NetworkPolicy = &koyeb.NetworkPolicy{
			Egress: &koyeb.EgressPolicy{
				Mode:      &mode,
				AllowList: []koyeb.NetworkPolicyDestination{{Cidr: &allowCIDR}},
			},
		}

		req, err := buildUpdateServicePool(sandboxTestContext(&fakeAPI{}), cmd.Flags(), live, "my-pool")
		require.NoError(t, err)

		networkPolicy := req.GetDefinition().NetworkPolicy
		require.NotNil(t, networkPolicy)
		egress := networkPolicy.GetEgress()
		assert.Equal(t, koyeb.EGRESSPOLICYMODE_DENY_ALL, egress.GetMode())
		assert.Empty(t, egress.AllowList, "--block-network drops any existing allow-list")
	})

	t.Run("--no-network-policy reverts the live member policy", func(t *testing.T) {
		cmd := newPoolUpdateCmd()
		require.NoError(t, cmd.Flags().Set("no-network-policy", "true"))
		live := liveSandboxPoolFixture()
		mode := koyeb.EGRESSPOLICYMODE_DENY_ALL
		live.Definition.NetworkPolicy = &koyeb.NetworkPolicy{
			Egress: &koyeb.EgressPolicy{Mode: &mode},
		}

		req, err := buildUpdateServicePool(sandboxTestContext(&fakeAPI{}), cmd.Flags(), live, "my-pool")
		require.NoError(t, err)

		networkPolicy := req.GetDefinition().NetworkPolicy
		require.NotNil(t, networkPolicy)
		egress := networkPolicy.GetEgress()
		assert.Equal(t, koyeb.EGRESSPOLICYMODE_DEFAULT, egress.GetMode())
	})

	t.Run("an unchanged policy is kept verbatim", func(t *testing.T) {
		cmd := newPoolUpdateCmd()
		live := liveSandboxPoolFixture()
		mode := koyeb.EGRESSPOLICYMODE_DENY_ALL
		livePolicy := &koyeb.NetworkPolicy{Egress: &koyeb.EgressPolicy{Mode: &mode}}
		live.Definition.NetworkPolicy = livePolicy

		req, err := buildUpdateServicePool(sandboxTestContext(&fakeAPI{}), cmd.Flags(), live, "my-pool")
		require.NoError(t, err)

		assert.Equal(t, livePolicy, req.GetDefinition().NetworkPolicy)
	})

	t.Run("--exposed-port-protocol rewrites the live wiring", func(t *testing.T) {
		cmd := newPoolUpdateCmd()
		require.NoError(t, cmd.Flags().Set("exposed-port-protocol", "http2"))
		live := liveSandboxPoolFixture()

		req, err := buildUpdateServicePool(sandboxTestContext(&fakeAPI{}), cmd.Flags(), live, "my-pool")
		require.NoError(t, err)

		def := req.GetDefinition()
		ports := def.Ports
		require.Len(t, ports, 2)
		assert.Equal(t, int64(3030), ports[0].GetPort())
		assert.Equal(t, "http", ports[0].GetProtocol())
		assert.Equal(t, int64(3031), ports[1].GetPort())
		assert.Equal(t, "http2", ports[1].GetProtocol())
		require.Len(t, def.Routes, 2)
	})

	t.Run("--enable-tcp-proxy surgically edits the 3031 proxy port", func(t *testing.T) {
		live := liveSandboxPoolFixture()
		port3031 := int64(3031)
		port22 := int64(22)
		tcp := koyeb.PROXYPORTPROTOCOL_TCP
		live.Definition.ProxyPorts = []koyeb.DeploymentProxyPort{
			{Port: &port22, Protocol: &tcp},
			{Port: &port3031, Protocol: &tcp},
		}

		cmd := newPoolUpdateCmd()
		require.NoError(t, cmd.Flags().Set("enable-tcp-proxy", "false"))

		req, err := buildUpdateServicePool(sandboxTestContext(&fakeAPI{}), cmd.Flags(), live, "my-pool")
		require.NoError(t, err)
		proxyPorts := req.GetDefinition().ProxyPorts
		require.Len(t, proxyPorts, 1, "only the 3031 entry is cleared; other proxy ports are kept")
		assert.Equal(t, int64(22), proxyPorts[0].GetPort())

		cmd = newPoolUpdateCmd()
		require.NoError(t, cmd.Flags().Set("enable-tcp-proxy", "true"))

		req, err = buildUpdateServicePool(sandboxTestContext(&fakeAPI{}), cmd.Flags(), live, "my-pool")
		require.NoError(t, err)
		proxyPorts = req.GetDefinition().ProxyPorts
		require.Len(t, proxyPorts, 2, "enabling adds the 3031 entry without dropping the others")
		assert.Equal(t, int64(22), proxyPorts[0].GetPort())
		assert.Equal(t, int64(3031), proxyPorts[1].GetPort())
	})

	t.Run("knobs are rejected on WEB pools", func(t *testing.T) {
		cmd := newPoolUpdateCmd()
		require.NoError(t, cmd.Flags().Set("enable-tcp-proxy", "true"))
		live := liveWebPoolFixture()

		_, err := buildUpdateServicePool(sandboxTestContext(&fakeAPI{}), cmd.Flags(), live, "my-pool")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "sandbox-only options")
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

	t.Run("--volumes mounts a member volume", func(t *testing.T) {
		cmd := newPoolUpdateCmd()
		require.NoError(t, cmd.Flags().Set("volumes", "423e4567-e89b-42d3-a456-426614174000:/other"))
		live := liveSandboxPoolFixture()

		req, err := buildUpdateServicePool(sandboxTestContext(&fakeAPI{}), cmd.Flags(), live, "my-pool")
		require.NoError(t, err)

		def := req.GetDefinition()
		volumes := def.GetVolumes()
		require.Len(t, volumes, 1)
		assert.Equal(t, "423e4567-e89b-42d3-a456-426614174000", volumes[0].GetId())
		assert.Equal(t, "/other", volumes[0].GetPath())
	})

	t.Run("unchanged --volumes keeps the live mounts verbatim", func(t *testing.T) {
		cmd := newPoolUpdateCmd()
		live := liveSandboxPoolFixture()
		live.Definition.Volumes = liveVolumes()

		req, err := buildUpdateServicePool(sandboxTestContext(&fakeAPI{}), cmd.Flags(), live, "my-pool")
		require.NoError(t, err)

		def := req.GetDefinition()
		assert.Equal(t, liveVolumes(), def.GetVolumes(),
			"unchanged flags must keep the live volume mounts")
	})

	t.Run("--volumes remounts a live volume to a new path", func(t *testing.T) {
		cmd := newPoolUpdateCmd()
		require.NoError(t, cmd.Flags().Set("volumes", "323e4567-e89b-42d3-a456-426614174000:/new"))
		live := liveSandboxPoolFixture()
		live.Definition.Volumes = liveVolumes()

		req, err := buildUpdateServicePool(sandboxTestContext(&fakeAPI{}), cmd.Flags(), live, "my-pool")
		require.NoError(t, err)

		def := req.GetDefinition()
		volumes := def.GetVolumes()
		require.Len(t, volumes, 1)
		assert.Equal(t, "323e4567-e89b-42d3-a456-426614174000", volumes[0].GetId())
		assert.Equal(t, "/new", volumes[0].GetPath())
	})

	t.Run("--volumes '!VOLUME' unmounts a live volume", func(t *testing.T) {
		cmd := newPoolUpdateCmd()
		require.NoError(t, cmd.Flags().Set("volumes", "!323e4567-e89b-42d3-a456-426614174000"))
		live := liveSandboxPoolFixture()
		live.Definition.Volumes = liveVolumes()

		req, err := buildUpdateServicePool(sandboxTestContext(&fakeAPI{}), cmd.Flags(), live, "my-pool")
		require.NoError(t, err)

		def := req.GetDefinition()
		assert.Empty(t, def.GetVolumes(), "'!' must unmount the live volume")
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

	t.Run("WEB pools with unchanged port flags resend the wiring verbatim", func(t *testing.T) {
		cmd := newPoolUpdateCmd()
		live := liveWebPoolFixture()

		req, err := buildUpdateServicePool(sandboxTestContext(&fakeAPI{}), cmd.Flags(), live, "my-pool")
		require.NoError(t, err)

		def := req.GetDefinition()
		assert.Equal(t, live.Definition, &def, "unchanged --port/--route flags must keep the live wiring")
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

	t.Run("deleting the last port empties the member wiring", func(t *testing.T) {
		cmd := newPoolUpdateCmd()
		require.NoError(t, cmd.Flags().Set("ports", "!8080"))
		live := liveWebPoolFixture()
		live.Definition.Ports = []koyeb.DeploymentPort{
			{Port: koyeb.PtrInt64(8080), Protocol: koyeb.PtrString("http")},
		}

		req, err := buildUpdateServicePool(sandboxTestContext(&fakeAPI{}), cmd.Flags(), live, "my-pool")
		require.NoError(t, err)

		assert.Empty(t, req.GetDefinition().Ports,
			"the last deletion must empty the wiring, not silently no-op")
	})

	t.Run("deleting the last route empties the member routes", func(t *testing.T) {
		cmd := newPoolUpdateCmd()
		require.NoError(t, cmd.Flags().Set("routes", "!/"))
		live := liveWebPoolFixture()
		live.Definition.Routes = []koyeb.DeploymentRoute{
			{Port: koyeb.PtrInt64(8080), Path: koyeb.PtrString("/")},
		}

		req, err := buildUpdateServicePool(sandboxTestContext(&fakeAPI{}), cmd.Flags(), live, "my-pool")
		require.NoError(t, err)

		assert.Empty(t, req.GetDefinition().Routes)
	})

	t.Run("a pool without a definition cannot be updated", func(t *testing.T) {
		cmd := newPoolUpdateCmd()

		_, err := buildUpdateServicePool(sandboxTestContext(&fakeAPI{}), cmd.Flags(), koyeb.ServicePool{}, "my-pool")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "has no definition to update")
	})
}

func TestBuildUpdateServicePoolChecks(t *testing.T) {
	t.Run("unchanged flags keep the live member healthchecks verbatim", func(t *testing.T) {
		cmd := newPoolUpdateCmd()
		live := liveWebPoolFixture()
		live.Definition.HealthChecks = liveChecks()

		req, err := buildUpdateServicePool(sandboxTestContext(&fakeAPI{}), cmd.Flags(), live, "my-pool")
		require.NoError(t, err)

		assert.Equal(t, live.Definition, req.Definition,
			"unchanged --checks flags must keep the live healthchecks")
	})

	t.Run("--checks adds a member healthcheck on WEB pools", func(t *testing.T) {
		cmd := newPoolUpdateCmd()
		require.NoError(t, cmd.Flags().Set("checks", "9000:tcp"))
		live := liveWebPoolFixture()
		live.Definition.HealthChecks = liveChecks()

		req, err := buildUpdateServicePool(sandboxTestContext(&fakeAPI{}), cmd.Flags(), live, "my-pool")
		require.NoError(t, err)

		checks := req.GetDefinition().HealthChecks
		require.Len(t, checks, 2)
		http, ok := checks[0].GetHttpOk()
		require.True(t, ok)
		assert.Equal(t, int64(8080), *http.Port, "the live healthcheck must be kept")
		tcp, ok := checks[1].GetTcpOk()
		require.True(t, ok)
		assert.Equal(t, int64(9000), *tcp.Port)
	})

	t.Run("--checks-grace-period updates the live healthcheck", func(t *testing.T) {
		cmd := newPoolUpdateCmd()
		require.NoError(t, cmd.Flags().Set("checks-grace-period", "8080=30"))
		live := liveWebPoolFixture()
		live.Definition.HealthChecks = liveChecks()

		req, err := buildUpdateServicePool(sandboxTestContext(&fakeAPI{}), cmd.Flags(), live, "my-pool")
		require.NoError(t, err)

		checks := req.GetDefinition().HealthChecks
		require.Len(t, checks, 1)
		require.NotNil(t, checks[0].GracePeriod)
		assert.Equal(t, int64(30), *checks[0].GracePeriod)
	})

	t.Run("'!' deletes a live member healthcheck", func(t *testing.T) {
		cmd := newPoolUpdateCmd()
		require.NoError(t, cmd.Flags().Set("checks", "!8080"))
		live := liveWebPoolFixture()
		live.Definition.HealthChecks = liveChecks()

		req, err := buildUpdateServicePool(sandboxTestContext(&fakeAPI{}), cmd.Flags(), live, "my-pool")
		require.NoError(t, err)

		assert.Empty(t, req.GetDefinition().HealthChecks,
			"deleting the last healthcheck must clear the list, not silently no-op")
	})

	t.Run("SANDBOX pools reject --checks", func(t *testing.T) {
		cmd := newPoolUpdateCmd()
		require.NoError(t, cmd.Flags().Set("checks", "8080:http:/health"))
		live := liveSandboxPoolFixture()

		_, err := buildUpdateServicePool(sandboxTestContext(&fakeAPI{}), cmd.Flags(), live, "my-pool")
		require.Error(t, err)
		assert.Contains(t, err.Error(), `--checks can only be specified for "web" services`)
	})
}

func TestBuildUpdateServicePoolProxyPorts(t *testing.T) {
	port22 := int64(22)
	tcp := koyeb.PROXYPORTPROTOCOL_TCP
	liveProxyPorts := []koyeb.DeploymentProxyPort{{Port: &port22, Protocol: &tcp}}

	t.Run("WEB pools merge declared proxy ports over the live values", func(t *testing.T) {
		cmd := newPoolUpdateCmd()
		require.NoError(t, cmd.Flags().Set("proxy-ports", "8022:tcp"))
		live := liveWebPoolFixture()
		live.Definition.ProxyPorts = liveProxyPorts

		req, err := buildUpdateServicePool(sandboxTestContext(&fakeAPI{}), cmd.Flags(), live, "my-pool")
		require.NoError(t, err)

		proxyPorts := req.GetDefinition().ProxyPorts
		require.Len(t, proxyPorts, 2)
		assert.Equal(t, int64(22), proxyPorts[0].GetPort(), "the live proxy port must be kept")
		assert.Equal(t, int64(8022), proxyPorts[1].GetPort())
		assert.Equal(t, koyeb.PROXYPORTPROTOCOL_TCP, proxyPorts[1].GetProtocol())
	})

	t.Run("deleting the last proxy port empties the member wiring", func(t *testing.T) {
		cmd := newPoolUpdateCmd()
		require.NoError(t, cmd.Flags().Set("proxy-ports", "!22"))
		live := liveWebPoolFixture()
		live.Definition.ProxyPorts = liveProxyPorts

		req, err := buildUpdateServicePool(sandboxTestContext(&fakeAPI{}), cmd.Flags(), live, "my-pool")
		require.NoError(t, err)

		assert.Empty(t, req.GetDefinition().ProxyPorts,
			"the last deletion must empty the wiring, not silently no-op")
	})

	t.Run("unchanged flags keep the live proxy ports verbatim", func(t *testing.T) {
		cmd := newPoolUpdateCmd()
		live := liveWebPoolFixture()
		live.Definition.ProxyPorts = liveProxyPorts

		req, err := buildUpdateServicePool(sandboxTestContext(&fakeAPI{}), cmd.Flags(), live, "my-pool")
		require.NoError(t, err)

		assert.Equal(t, liveProxyPorts, req.GetDefinition().ProxyPorts,
			"unchanged --proxy-ports flags must keep the live wiring")
	})

	t.Run("WORKER pools accept declared proxy ports", func(t *testing.T) {
		cmd := newPoolUpdateCmd()
		require.NoError(t, cmd.Flags().Set("proxy-ports", "5432:tcp"))
		live := liveWebPoolFixture()
		defType := koyeb.DEPLOYMENTDEFINITIONTYPE_WORKER
		live.Definition.Type = &defType

		req, err := buildUpdateServicePool(sandboxTestContext(&fakeAPI{}), cmd.Flags(), live, "my-pool")
		require.NoError(t, err)

		proxyPorts := req.GetDefinition().ProxyPorts
		require.Len(t, proxyPorts, 1)
		assert.Equal(t, int64(5432), proxyPorts[0].GetPort())
	})

	t.Run("SANDBOX pools reject explicit proxy ports", func(t *testing.T) {
		cmd := newPoolUpdateCmd()
		require.NoError(t, cmd.Flags().Set("proxy-ports", "22:tcp"))
		live := liveSandboxPoolFixture()

		_, err := buildUpdateServicePool(sandboxTestContext(&fakeAPI{}), cmd.Flags(), live, "my-pool")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not allowed on SANDBOX pools")
		assert.Contains(t, err.Error(), "proxy-ports")
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
		live := liveSandboxPoolFixture()
		fake := &fakeAPI{pool: &live}
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

	// The update command declares the same curated flag set as create:
	// the pool surface base plus the union of its definition flag
	// bundles.
	for _, name := range slices.Concat(
		[]string{
			"size", "type", "ports", "routes", "privileged",
			"exposed-port-protocol", "enable-tcp-proxy",
		},
		dockerSourceFlagNames(sandboxPoolDockerSourceFlagUsage),
		archiveSourceFlagNames(),
		networkPolicyFlagNames(),
		envConfigFilesFlagNames(sandboxPoolEnvConfigFilesFlagUsage),
		volumesFlagNames(),
		instanceTypeRegionsFlagNames(sandboxPoolInstanceTypeRegionsFlagUsage),
		scalingSleepDelayFlagNames(sandboxPoolScalingSleepDelayFlagUsage),
		checksFlagNames(),
		proxyPortsFlagNames(),
	) {
		assert.NotNil(t, cmd.Flags().Lookup(name), "flag --%s must be registered on pool update", name)
	}
	// No --name: renaming a pool is not supported.
	assert.Nil(t, cmd.Flags().Lookup("name"), "flag --name must not be registered on pool update")
}

func TestPoolUpdateCmdFlagAliases(t *testing.T) {
	cmd := newPoolUpdateCmd()
	require.NoError(t, cmd.Flags().Set("docker-arg", "my-arg"))
	live := liveSandboxPoolFixture()

	req, err := buildUpdateServicePool(sandboxTestContext(&fakeAPI{}), cmd.Flags(), live, "my-pool")
	require.NoError(t, err)

	docker := req.GetDefinition().Docker
	assert.Equal(t, []string{"my-arg"}, docker.GetArgs(),
		"--docker-arg must normalize to the registered --docker-args flag")
	// The changed-only merge keeps the untouched live image.
	assert.Equal(t, "koyeb/sandbox:latest", docker.GetImage())
}
