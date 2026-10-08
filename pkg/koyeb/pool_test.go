package koyeb

import (
	"context"
	"slices"
	"strconv"
	"testing"
	"time"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/idmapper"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildCreateServicePool(t *testing.T) {
	tests := []struct {
		name        string
		poolName    string
		size        int64
		dockerImage string
	}{
		{
			name:        "minimal declared flag set does not panic",
			poolName:    "my-pool",
			size:        1,
			dockerImage: "",
		},
		{
			name:        "explicit size and docker image",
			poolName:    "worker-pool",
			size:        5,
			dockerImage: "ghcr.io/acme/sandbox:latest",
		},
		{
			name:        "large size",
			poolName:    "big-pool",
			size:        42,
			dockerImage: "docker.io/library/alpine",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := newPoolCreateCmd()
			require.NoError(t, cmd.Flags().Set("size", strconv.FormatInt(tt.size, 10)))
			if tt.dockerImage != "" {
				require.NoError(t, cmd.Flags().Set("docker", tt.dockerImage))
			}

			req, err := buildCreateServicePool(&CLIContext{}, cmd, tt.poolName)
			require.NoError(t, err)

			assert.Equal(t, tt.poolName, req.GetName())
			assert.Equal(t, tt.size, req.GetSize())

			def := req.GetDefinition()
			assert.Equal(t, tt.poolName, def.GetName(), "definition name must be set from the positional NAME")
			assert.Equal(t, koyeb.DEPLOYMENTDEFINITIONTYPE_SANDBOX, def.GetType())
			assert.False(t, def.HasPorts(), "pool definition must not declare ports")
			assert.False(t, def.HasRoutes(), "pool definition must not declare routes")

			expectedImage := tt.dockerImage
			if expectedImage == "" {
				expectedImage = "koyeb/sandbox"
			}
			docker := def.GetDocker()
			assert.Equal(t, expectedImage, docker.GetImage())
		})
	}
}

func TestBuildCreateServicePoolDefaultsOnly(t *testing.T) {
	cmd := newPoolCreateCmd()

	req, err := buildCreateServicePool(&CLIContext{}, cmd, "default-pool")
	require.NoError(t, err)

	assert.Equal(t, "default-pool", req.GetName())
	assert.Equal(t, int64(1), req.GetSize())
	def := req.GetDefinition()
	assert.Equal(t, koyeb.DEPLOYMENTDEFINITIONTYPE_SANDBOX, def.GetType())
	assert.False(t, def.HasPorts(), "SANDBOX pool definitions must not declare ports")
	assert.False(t, def.HasRoutes(), "SANDBOX pool definitions must not declare routes")
	for _, env := range def.Env {
		assert.NotEqual(t, SandboxSecretKey, env.GetKey(),
			"the platform mints the pool secret; the CLI must never inject one")
	}
}

func TestBuildCreateServicePoolType(t *testing.T) {
	tests := []struct {
		name    string
		typeArg string
		want    koyeb.DeploymentDefinitionType
	}{
		{name: "default is sandbox", typeArg: "", want: koyeb.DEPLOYMENTDEFINITIONTYPE_SANDBOX},
		{name: "sandbox", typeArg: "sandbox", want: koyeb.DEPLOYMENTDEFINITIONTYPE_SANDBOX},
		{name: "web is case-insensitive", typeArg: "WEB", want: koyeb.DEPLOYMENTDEFINITIONTYPE_WEB},
		{name: "worker", typeArg: "worker", want: koyeb.DEPLOYMENTDEFINITIONTYPE_WORKER},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := newPoolCreateCmd()
			if tt.typeArg != "" {
				require.NoError(t, cmd.Flags().Set("type", tt.typeArg))
			}

			req, err := buildCreateServicePool(&CLIContext{}, cmd, "my-pool")
			require.NoError(t, err)
			def := req.GetDefinition()
			assert.Equal(t, tt.want, def.GetType())
		})
	}

	t.Run("database is rejected fail-fast with the rule", func(t *testing.T) {
		cmd := newPoolCreateCmd()
		require.NoError(t, cmd.Flags().Set("type", "database"))

		_, err := buildCreateServicePool(&CLIContext{}, cmd, "my-pool")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not supported")
		assert.Contains(t, err.Error(), "WEB, WORKER and SANDBOX")
	})

	t.Run("unknown types are rejected", func(t *testing.T) {
		cmd := newPoolCreateCmd()
		require.NoError(t, cmd.Flags().Set("type", "sidecar"))

		_, err := buildCreateServicePool(&CLIContext{}, cmd, "my-pool")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "--type flag is not valid")
	})
}

func TestBuildCreateServicePoolPortsAndRoutes(t *testing.T) {
	t.Run("web pools carry declared ports and routes verbatim", func(t *testing.T) {
		cmd := newPoolCreateCmd()
		require.NoError(t, cmd.Flags().Set("type", "web"))
		require.NoError(t, cmd.Flags().Set("ports", "8080:http"))
		require.NoError(t, cmd.Flags().Set("ports", "9090:tcp"))
		require.NoError(t, cmd.Flags().Set("routes", "/:8080"))

		req, err := buildCreateServicePool(&CLIContext{}, cmd, "my-pool")
		require.NoError(t, err)

		def := req.GetDefinition()
		ports := def.Ports
		require.Len(t, ports, 2)
		assert.Equal(t, int64(8080), ports[0].GetPort())
		assert.Equal(t, "http", ports[0].GetProtocol())
		assert.Equal(t, int64(9090), ports[1].GetPort())
		assert.Equal(t, "tcp", ports[1].GetProtocol())

		routes := def.GetRoutes()
		require.Len(t, routes, 1)
		assert.Equal(t, "/", routes[0].GetPath())
		assert.Equal(t, int64(8080), routes[0].GetPort())
	})

	t.Run("worker pools accept declared ports", func(t *testing.T) {
		cmd := newPoolCreateCmd()
		require.NoError(t, cmd.Flags().Set("type", "worker"))
		require.NoError(t, cmd.Flags().Set("ports", "8080"))

		req, err := buildCreateServicePool(&CLIContext{}, cmd, "my-pool")
		require.NoError(t, err)

		def := req.GetDefinition()
		ports := def.Ports
		require.Len(t, ports, 1)
		assert.Equal(t, int64(8080), ports[0].GetPort())
		assert.Equal(t, "http", ports[0].GetProtocol(), "PORT defaults to http")
	})

	t.Run("web pools without declared ports send none", func(t *testing.T) {
		cmd := newPoolCreateCmd()
		require.NoError(t, cmd.Flags().Set("type", "web"))

		req, err := buildCreateServicePool(&CLIContext{}, cmd, "my-pool")
		require.NoError(t, err)

		def := req.GetDefinition()
		assert.False(t, def.HasPorts(), "no default ports are injected for non-SANDBOX pools")
		assert.False(t, def.HasRoutes(), "no default routes are injected for non-SANDBOX pools")
	})

	t.Run("explicit ports are rejected on SANDBOX pools", func(t *testing.T) {
		cmd := newPoolCreateCmd()
		require.NoError(t, cmd.Flags().Set("ports", "8080"))

		_, err := buildCreateServicePool(&CLIContext{}, cmd, "my-pool")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not allowed on SANDBOX pools")
		assert.Contains(t, err.Error(), "3030/3031")
	})

	t.Run("explicit routes are rejected on the SANDBOX default", func(t *testing.T) {
		cmd := newPoolCreateCmd()
		require.NoError(t, cmd.Flags().Set("routes", "/:8080"))

		_, err := buildCreateServicePool(&CLIContext{}, cmd, "my-pool")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not allowed on SANDBOX pools")
	})

	t.Run("invalid port values are rejected", func(t *testing.T) {
		cmd := newPoolCreateCmd()
		require.NoError(t, cmd.Flags().Set("type", "web"))
		require.NoError(t, cmd.Flags().Set("ports", "not-a-port"))

		_, err := buildCreateServicePool(&CLIContext{}, cmd, "my-pool")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "unable to parse the port")
	})
}

func TestBuildCreateServicePoolPrivileged(t *testing.T) {
	t.Run("members run unprivileged by default", func(t *testing.T) {
		cmd := newPoolCreateCmd()

		req, err := buildCreateServicePool(&CLIContext{}, cmd, "my-pool")
		require.NoError(t, err)

		docker := req.GetDefinition().Docker
		assert.False(t, docker.GetPrivileged())
	})

	t.Run("--privileged flags the member containers", func(t *testing.T) {
		cmd := newPoolCreateCmd()
		require.NoError(t, cmd.Flags().Set("privileged", "true"))

		req, err := buildCreateServicePool(&CLIContext{}, cmd, "my-pool")
		require.NoError(t, err)

		docker := req.GetDefinition().Docker
		assert.True(t, docker.GetPrivileged())
	})
}

func TestBuildCreateServicePoolVolumes(t *testing.T) {
	t.Run("--volumes mounts the declared member volumes", func(t *testing.T) {
		cmd := newPoolCreateCmd()
		require.NoError(t, cmd.Flags().Set("volumes", "323e4567-e89b-42d3-a456-426614174000:/data"))
		require.NoError(t, cmd.Flags().Set("volumes", "423e4567-e89b-42d3-a456-426614174000:/other"))

		req, err := buildCreateServicePool(sandboxTestContext(&fakeAPI{}), cmd, "my-pool")
		require.NoError(t, err)

		def := req.GetDefinition()
		volumes := def.GetVolumes()
		require.Len(t, volumes, 2)
		assert.Equal(t, "323e4567-e89b-42d3-a456-426614174000", volumes[0].GetId())
		assert.Equal(t, "/data", volumes[0].GetPath())
		assert.Equal(t, "423e4567-e89b-42d3-a456-426614174000", volumes[1].GetId())
		assert.Equal(t, "/other", volumes[1].GetPath())
	})

	t.Run("--volume aliases --volumes, matching the service commands", func(t *testing.T) {
		cmd := newPoolCreateCmd()
		require.NoError(t, cmd.Flags().Set("volume", "323e4567-e89b-42d3-a456-426614174000:/data"))

		req, err := buildCreateServicePool(sandboxTestContext(&fakeAPI{}), cmd, "my-pool")
		require.NoError(t, err)

		def := req.GetDefinition()
		volumes := def.GetVolumes()
		require.Len(t, volumes, 1)
		assert.Equal(t, "/data", volumes[0].GetPath())
	})

	t.Run("a volume declared without a mount path is rejected", func(t *testing.T) {
		cmd := newPoolCreateCmd()
		require.NoError(t, cmd.Flags().Set("volumes", "323e4567-e89b-42d3-a456-426614174000"))

		_, err := buildCreateServicePool(sandboxTestContext(&fakeAPI{}), cmd, "my-pool")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "unable to parse the volume")
	})

	t.Run("members carry no volumes by default", func(t *testing.T) {
		cmd := newPoolCreateCmd()

		req, err := buildCreateServicePool(&CLIContext{}, cmd, "my-pool")
		require.NoError(t, err)

		def := req.GetDefinition()
		assert.False(t, def.HasVolumes())
	})
}

func TestBuildCreateServicePoolNetworkPolicy(t *testing.T) {
	t.Run("--block-network denies all member egress", func(t *testing.T) {
		cmd := newPoolCreateCmd()
		require.NoError(t, cmd.Flags().Set("block-network", "true"))

		req, err := buildCreateServicePool(&CLIContext{}, cmd, "my-pool")
		require.NoError(t, err)

		networkPolicy := req.GetDefinition().NetworkPolicy
		require.NotNil(t, networkPolicy, "the definition must carry a network policy")
		egress := networkPolicy.GetEgress()
		assert.Equal(t, koyeb.EGRESSPOLICYMODE_DENY_ALL, egress.GetMode())
	})

	t.Run("--outbound-allowlist lists member egress destinations", func(t *testing.T) {
		cmd := newPoolCreateCmd()
		require.NoError(t, cmd.Flags().Set("outbound-allowlist", "10.0.0.0/8"))

		req, err := buildCreateServicePool(&CLIContext{}, cmd, "my-pool")
		require.NoError(t, err)

		networkPolicy := req.GetDefinition().NetworkPolicy
		require.NotNil(t, networkPolicy)
		egress := networkPolicy.GetEgress()
		assert.Equal(t, koyeb.EGRESSPOLICYMODE_DENY_ALL, egress.GetMode())
		require.Len(t, egress.AllowList, 1)
	})

	t.Run("members carry no policy by default", func(t *testing.T) {
		cmd := newPoolCreateCmd()

		req, err := buildCreateServicePool(&CLIContext{}, cmd, "my-pool")
		require.NoError(t, err)

		def := req.GetDefinition()
		assert.False(t, def.HasNetworkPolicy())
	})
}

func TestBuildCreateServicePoolSandboxKnobs(t *testing.T) {
	t.Run("--exposed-port-protocol emits the wiring with the chosen protocol", func(t *testing.T) {
		cmd := newPoolCreateCmd()
		require.NoError(t, cmd.Flags().Set("exposed-port-protocol", "http2"))

		req, err := buildCreateServicePool(&CLIContext{}, cmd, "my-pool")
		require.NoError(t, err)

		def := req.GetDefinition()
		ports := def.Ports
		require.Len(t, ports, 2, "the sandbox wiring is emitted when the protocol is explicit")
		assert.Equal(t, int64(3030), ports[0].GetPort())
		assert.Equal(t, "http", ports[0].GetProtocol())
		assert.Equal(t, int64(3031), ports[1].GetPort())
		assert.Equal(t, "http2", ports[1].GetProtocol())
		routes := def.Routes
		require.Len(t, routes, 2)
		assert.Equal(t, "/koyeb-sandbox/", routes[0].GetPath())
		assert.Equal(t, "/", routes[1].GetPath())
	})

	t.Run("--enable-tcp-proxy exposes port 3031 via TCP proxy", func(t *testing.T) {
		cmd := newPoolCreateCmd()
		require.NoError(t, cmd.Flags().Set("enable-tcp-proxy", "true"))

		req, err := buildCreateServicePool(&CLIContext{}, cmd, "my-pool")
		require.NoError(t, err)

		def := req.GetDefinition()
		proxyPorts := def.ProxyPorts
		require.Len(t, proxyPorts, 1)
		assert.Equal(t, int64(3031), proxyPorts[0].GetPort())
		assert.Equal(t, koyeb.PROXYPORTPROTOCOL_TCP, proxyPorts[0].GetProtocol())
	})

	t.Run("knobs are rejected on WEB pools", func(t *testing.T) {
		cmd := newPoolCreateCmd()
		require.NoError(t, cmd.Flags().Set("type", "web"))
		require.NoError(t, cmd.Flags().Set("exposed-port-protocol", "http"))

		_, err := buildCreateServicePool(&CLIContext{}, cmd, "my-pool")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "sandbox-only options")
	})

	t.Run("the TCP proxy is rejected on WORKER pools", func(t *testing.T) {
		cmd := newPoolCreateCmd()
		require.NoError(t, cmd.Flags().Set("type", "worker"))
		require.NoError(t, cmd.Flags().Set("enable-tcp-proxy", "true"))

		_, err := buildCreateServicePool(&CLIContext{}, cmd, "my-pool")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "sandbox-only options")
	})

	t.Run("invalid protocols are rejected", func(t *testing.T) {
		cmd := newPoolCreateCmd()
		require.NoError(t, cmd.Flags().Set("exposed-port-protocol", "tcp"))

		_, err := buildCreateServicePool(&CLIContext{}, cmd, "my-pool")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Invalid exposed port protocol")
	})

	t.Run("no knobs, no wiring: the platform owns the default", func(t *testing.T) {
		cmd := newPoolCreateCmd()

		req, err := buildCreateServicePool(&CLIContext{}, cmd, "my-pool")
		require.NoError(t, err)

		def := req.GetDefinition()
		assert.False(t, def.HasPorts(), "unchanged knobs must not emit wiring")
		assert.False(t, def.HasProxyPorts())
	})
}

func TestPoolCreateCmdFlagSet(t *testing.T) {
	cmd := newPoolCreateCmd()
	flags := cmd.Flags()

	// The pool typing and wiring flags must be registered.
	for _, name := range []string{"type", "ports", "routes"} {
		assert.NotNil(t, flags.Lookup(name), "flag --%s must be registered on pool create", name)
	}
	// Service flags that have no pool equivalent must not be registered:
	// cobra rejects them as unknown flags, which guards against invalid pool
	// definitions. Pool secrets and mesh stay off the pool surface per the
	// cross-client contract.
	for _, name := range []string{
		"checks", "proxy-ports", "auth", "app", "wait", "wait-timeout",
		"sandbox-secret", "enable-mesh",
	} {
		assert.Nil(t, flags.Lookup(name), "flag --%s must not be registered on pool create", name)
	}
	// The curated flag set must be declared, matching the SDK surfaces:
	// the pool surface base plus the union of its definition flag
	// bundles, so a new bundle flag lands on the pool surface without
	// hand-updating this list.
	for _, name := range slices.Concat(
		[]string{
			"size", "type", "ports", "routes", "privileged",
			"exposed-port-protocol", "enable-tcp-proxy",
		},
		dockerSourceFlagNames(sandboxPoolDockerSourceFlagUsage),
		networkPolicyFlagNames(),
		envConfigFilesFlagNames(sandboxPoolEnvConfigFilesFlagUsage),
		volumesFlagNames(),
		instanceTypeRegionsFlagNames(sandboxPoolInstanceTypeRegionsFlagUsage),
		scalingSleepDelayFlagNames(sandboxPoolScalingSleepDelayFlagUsage),
	) {
		assert.NotNil(t, flags.Lookup(name), "flag --%s must be registered on pool create", name)
	}
}

func TestPoolCreateCmdFlagAliases(t *testing.T) {
	t.Run("--port and --route alias --ports and --routes, matching the service commands", func(t *testing.T) {
		cmd := newPoolCreateCmd()

		require.NoError(t, cmd.Flags().Set("port", "8080"))
		require.NoError(t, cmd.Flags().Set("type", "web"))

		req, err := buildCreateServicePool(&CLIContext{}, cmd, "my-pool")
		require.NoError(t, err)

		def := req.GetDefinition()
		ports := def.Ports
		require.Len(t, ports, 1)
		assert.Equal(t, int64(8080), ports[0].GetPort())
	})

	t.Run("--docker-arg aliases --docker-args, inheriting the service alias set", func(t *testing.T) {
		cmd := newPoolCreateCmd()

		require.NoError(t, cmd.Flags().Set("docker", "ghcr.io/acme/sandbox"))
		require.NoError(t, cmd.Flags().Set("docker-arg", "my-arg"))

		req, err := buildCreateServicePool(&CLIContext{}, cmd, "my-pool")
		require.NoError(t, err)

		docker := req.GetDefinition().Docker
		assert.Equal(t, "ghcr.io/acme/sandbox", docker.GetImage())
		assert.Equal(t, []string{"my-arg"}, docker.GetArgs(),
			"--docker-arg must normalize to the registered --docker-args flag")
	})

	t.Run("aliases whose canonical flag is not registered stay inert", func(t *testing.T) {
		cmd := newPoolCreateCmd()

		// --check would normalize to --checks, which pool create does not
		// register: the alias must not resolve to any flag.
		err := cmd.Flags().Set("check", "8080:http:/health")
		assert.Error(t, err, "--check must stay inert on pool create")
	})
}

func TestPoolCreateFlow(t *testing.T) {
	t.Run("sends the built request through the port", func(t *testing.T) {
		fake := &fakeAPI{}
		ctx := sandboxTestContext(fake)
		cmd := newPoolCreateCmd()
		require.NoError(t, cmd.Flags().Set("size", "3"))
		require.NoError(t, cmd.Flags().Set("docker", "ghcr.io/acme/sandbox"))

		req, err := buildCreateServicePool(ctx, cmd, "my-pool")
		require.NoError(t, err)

		require.NoError(t, NewPoolHandler().Create(ctx, cmd, []string{"my-pool"}, req))

		recorded := fake.createPoolReq
		require.NotNil(t, recorded, "the create request must travel through the port")
		assert.Equal(t, "my-pool", recorded.GetName())
		assert.Equal(t, int64(3), recorded.GetSize())
		docker := recorded.GetDefinition().Docker
		assert.Equal(t, "ghcr.io/acme/sandbox", docker.GetImage())
	})

	t.Run("API errors surface as CLI errors", func(t *testing.T) {
		fake := &fakeAPI{createPoolErr: assert.AnError}
		ctx := sandboxTestContext(fake)
		cmd := newPoolCreateCmd()

		err := NewPoolHandler().Create(ctx, cmd, []string{"my-pool"}, koyeb.CreateServicePool{})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Error while creating the pool `my-pool`")
	})
}

func TestPoolListCmdRegistersNameFilter(t *testing.T) {
	cmd := newPoolListCmd()

	nameFlag := cmd.Flags().Lookup("name")
	require.NotNil(t, nameFlag, "pool list must register --name")
	assert.Equal(t, "", nameFlag.DefValue, "the name filter is opt-in")
}

func TestPoolListCmdRegistersPaginationFlags(t *testing.T) {
	cmd := newPoolListCmd()

	for _, name := range []string{"limit", "offset"} {
		flag := cmd.Flags().Lookup(name)
		require.NotNil(t, flag, "pool list must register --%s", name)
		assert.Equal(t, "int64", flag.Value.Type(), "--%s must be an Int64 flag like pool claims list", name)
		assert.Equal(t, "0", flag.DefValue, "--%s is opt-in", name)
	}
}

func TestPoolListFlow(t *testing.T) {
	t.Run("--name filters the listing", func(t *testing.T) {
		fake := &fakeAPI{pools: []koyeb.ServicePool{servicePoolFixture()}}
		cmd := newPoolListCmd()
		require.NoError(t, cmd.Flags().Set("name", "my-pool"))

		require.NoError(t, NewPoolHandler().List(sandboxTestContext(fake), cmd, nil))
		assert.Equal(t, "my-pool", fake.listPoolsName, "the filter must reach the API request")
	})

	t.Run("without --name the listing is unfiltered", func(t *testing.T) {
		fake := &fakeAPI{pools: []koyeb.ServicePool{servicePoolFixture()}}
		cmd := newPoolListCmd()

		require.NoError(t, NewPoolHandler().List(sandboxTestContext(fake), cmd, nil))
		assert.Equal(t, "", fake.listPoolsName)
	})

	t.Run("--limit and --offset reach the request", func(t *testing.T) {
		fake := &fakeAPI{pools: []koyeb.ServicePool{servicePoolFixture()}}
		cmd := newPoolListCmd()
		require.NoError(t, cmd.Flags().Set("limit", "7"))
		require.NoError(t, cmd.Flags().Set("offset", "3"))

		require.NoError(t, NewPoolHandler().List(sandboxTestContext(fake), cmd, nil))
		assert.Equal(t, "7", fake.listPoolsLimit, "--limit must bound the request")
		assert.Equal(t, "3", fake.listPoolsOffset, "--offset must shift the request")
	})

	t.Run("--limit alone keeps the default start", func(t *testing.T) {
		fake := &fakeAPI{pools: []koyeb.ServicePool{servicePoolFixture()}}
		cmd := newPoolListCmd()
		require.NoError(t, cmd.Flags().Set("limit", "7"))

		require.NoError(t, NewPoolHandler().List(sandboxTestContext(fake), cmd, nil))
		assert.Equal(t, "7", fake.listPoolsLimit)
		assert.Equal(t, "0", fake.listPoolsOffset)
	})

	t.Run("--offset alone keeps the fetch-all page size", func(t *testing.T) {
		fake := &fakeAPI{pools: []koyeb.ServicePool{servicePoolFixture()}}
		cmd := newPoolListCmd()
		require.NoError(t, cmd.Flags().Set("offset", "3"))

		require.NoError(t, NewPoolHandler().List(sandboxTestContext(fake), cmd, nil))
		assert.Equal(t, "3", fake.listPoolsOffset)
		assert.Equal(t, "100", fake.listPoolsLimit, "without --limit the listing still pages to the end")
	})

	t.Run("unset flags keep the fetch-all page walk", func(t *testing.T) {
		fake := &fakeAPI{pools: []koyeb.ServicePool{servicePoolFixture()}}
		cmd := newPoolListCmd()

		require.NoError(t, NewPoolHandler().List(sandboxTestContext(fake), cmd, nil))
		assert.Equal(t, "0", fake.listPoolsOffset, "the walk starts at offset 0")
		assert.Equal(t, "100", fake.listPoolsLimit, "the walk keeps the fetch-all page size")
	})
}

func TestPoolGetDescribeFlow(t *testing.T) {
	poolID := "123e4567-e89b-42d3-a456-426614174000"

	// get and describe share the port's GetServicePool: one fetch
	// through the seam, two renderers over the same reply.
	t.Run("get fetches the pool through the port", func(t *testing.T) {
		live := liveSandboxPoolFixture()
		fake := &fakeAPI{pool: &live}
		ctx := sandboxTestContext(fake)

		require.NoError(t, NewPoolHandler().Get(ctx, newPoolGetCmd(), []string{poolID}))
		assert.Equal(t, []string{poolID}, fake.poolsFetched)
	})

	t.Run("describe fetches the pool through the port", func(t *testing.T) {
		live := liveSandboxPoolFixture()
		fake := &fakeAPI{pool: &live}
		ctx := sandboxTestContext(fake)

		require.NoError(t, NewPoolHandler().Describe(ctx, newPoolDescribeCmd(), []string{poolID}))
		assert.Equal(t, []string{poolID}, fake.poolsFetched)
	})

	t.Run("API errors surface as CLI errors", func(t *testing.T) {
		fake := &fakeAPI{getPoolErr: assert.AnError}
		ctx := sandboxTestContext(fake)

		err := NewPoolHandler().Get(ctx, newPoolGetCmd(), []string{poolID})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Error while retrieving the pool `"+poolID+"`")
	})
}

func TestPoolDeleteFlow(t *testing.T) {
	poolID := "123e4567-e89b-42d3-a456-426614174000"

	t.Run("deletes the resolved pool through the port", func(t *testing.T) {
		fake := &fakeAPI{}
		ctx := sandboxTestContext(fake)

		require.NoError(t, NewPoolHandler().Delete(ctx, newPoolDeleteCmd(), []string{poolID}))
		assert.Equal(t, []string{poolID}, fake.deletedPools)
	})

	t.Run("API errors surface as CLI errors", func(t *testing.T) {
		fake := &fakeAPI{deletePoolErr: assert.AnError}
		ctx := sandboxTestContext(fake)

		err := NewPoolHandler().Delete(ctx, newPoolDeleteCmd(), []string{poolID})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Error while deleting the pool `"+poolID+"`")
	})
}

func TestPoolCmdWiring(t *testing.T) {
	cmd := NewPoolCmd()

	assert.Equal(t, "pool ACTION", cmd.Use)
	assert.Equal(t, []string{"pools", "sp"}, cmd.Aliases)
	assert.NotNil(t, cmd.PersistentFlags().Lookup("project"))
	assert.NotNil(t, cmd.PersistentFlags().Lookup("workspace"))

	for _, sub := range []string{"create", "update", "list", "get", "describe", "delete", "claim", "claims"} {
		_, _, err := cmd.Find([]string{sub})
		require.NoError(t, err, "pool command must register the %q subcommand", sub)
	}
}

func servicePoolFixture() koyeb.ServicePool {
	id := "123e4567-e89b-42d3-a456-426614174000"
	name := "my-pool"
	status := koyeb.SERVICEPOOLSTATUS_READY
	size := int64(5)
	readyCount := int64(3)
	generation := "gen-42"
	createdAt := time.Date(2024, 5, 1, 12, 0, 0, 0, time.UTC)

	return koyeb.ServicePool{
		Id:         &id,
		Name:       &name,
		Status:     &status,
		Size:       &size,
		ReadyCount: &readyCount,
		Generation: &generation,
		CreatedAt:  &createdAt,
	}
}

func TestPoolReplyHeadersAndFields(t *testing.T) {
	pool := servicePoolFixture()

	reply := NewPoolReply(idmapper.NewMapper(context.Background(), nil), pool, false)

	assert.Equal(t, "Pool", reply.Title())
	assert.Equal(t, []string{"id", "name", "status", "size", "ready_count", "generation", "created_at"}, reply.Headers())

	fields := reply.Fields()
	require.Len(t, fields, 1)
	assert.Equal(t, "123e4567", fields[0]["id"])
	assert.Equal(t, "my-pool", fields[0]["name"])
	assert.Equal(t, "READY", fields[0]["status"])
	assert.Equal(t, "5", fields[0]["size"])
	assert.Equal(t, "3", fields[0]["ready_count"])
	assert.Equal(t, "gen-42", fields[0]["generation"])
	assert.NotEmpty(t, fields[0]["created_at"])
}

func TestPoolReplyFullModeDoesNotTruncateID(t *testing.T) {
	pool := servicePoolFixture()

	reply := NewPoolReply(idmapper.NewMapper(context.Background(), nil), pool, true)

	fields := reply.Fields()
	require.Len(t, fields, 1)
	assert.Equal(t, "123e4567-e89b-42d3-a456-426614174000", fields[0]["id"])
}

func TestPoolReplyNilSafe(t *testing.T) {
	reply := NewPoolReply(idmapper.NewMapper(context.Background(), nil), koyeb.ServicePool{}, false)

	fields := reply.Fields()
	require.Len(t, fields, 1)
	assert.Equal(t, "", fields[0]["id"])
	assert.Equal(t, "", fields[0]["status"]) // zero value of ServicePoolStatus is ""
}

func TestListPoolsReplyHeadersAndFields(t *testing.T) {
	pool := servicePoolFixture()
	value := &koyeb.ListServicePoolsReply{ServicePools: []koyeb.ServicePool{pool}}

	reply := NewListPoolsReply(idmapper.NewMapper(context.Background(), nil), value, false)

	assert.Equal(t, "Pools", reply.Title())
	assert.Equal(t, []string{"id", "name", "status", "size", "ready_count", "generation", "created_at"}, reply.Headers())

	fields := reply.Fields()
	require.Len(t, fields, 1)
	assert.Equal(t, "123e4567", fields[0]["id"])
	assert.Equal(t, "my-pool", fields[0]["name"])
	assert.Equal(t, "READY", fields[0]["status"])
	assert.Equal(t, "5", fields[0]["size"])
}

func TestListPoolsReplyEmpty(t *testing.T) {
	reply := NewListPoolsReply(idmapper.NewMapper(context.Background(), nil), &koyeb.ListServicePoolsReply{}, false)

	assert.Empty(t, reply.Fields())
}

func TestDescribePoolReplyHeadersAndFields(t *testing.T) {
	pool := servicePoolFixture()
	updatedAt := time.Date(2024, 5, 1, 12, 5, 0, 0, time.UTC)
	pool.UpdatedAt = &updatedAt

	instanceType := "nano"
	image := "koyeb/sandbox"
	pool.Definition = &koyeb.DeploymentDefinition{
		Docker:        &koyeb.DockerSource{Image: &image},
		InstanceTypes: []koyeb.DeploymentInstanceType{{Type: &instanceType}},
		Regions:       []string{"par", "fra"},
	}

	reply := NewDescribePoolReply(idmapper.NewMapper(context.Background(), nil), pool, false)

	assert.Equal(t, "Pool", reply.Title())
	assert.Equal(t, []string{
		"id", "name", "status", "size", "ready_count", "generation",
		"image", "instance_types", "regions", "created_at", "updated_at",
	}, reply.Headers())

	fields := reply.Fields()
	require.Len(t, fields, 1)
	assert.Equal(t, "123e4567", fields[0]["id"])
	assert.Equal(t, "koyeb/sandbox", fields[0]["image"])
	assert.Equal(t, "nano", fields[0]["instance_types"])
	assert.Equal(t, "par,fra", fields[0]["regions"])
	assert.NotEmpty(t, fields[0]["updated_at"])
}

func TestDescribePoolReplyNilSafe(t *testing.T) {
	reply := NewDescribePoolReply(idmapper.NewMapper(context.Background(), nil), koyeb.ServicePool{}, false)

	fields := reply.Fields()
	require.Len(t, fields, 1)
	assert.Equal(t, "", fields[0]["image"])
	assert.Equal(t, "", fields[0]["instance_types"])
	assert.Equal(t, "", fields[0]["regions"])
}

func TestParseSingleInstanceScaling(t *testing.T) {
	t.Run("defaults to min-scale with max 1", func(t *testing.T) {
		cmd := newPoolCreateCmd()
		flags := cmd.Flags()

		scaling, err := parseSingleInstanceScaling(flags, "pool")
		require.NoError(t, err)
		assert.Equal(t, int64(1), scaling.GetMin())
		assert.Equal(t, int64(1), scaling.GetMax())
		assert.False(t, scaling.HasTargets())
	})

	t.Run("sleep delays require min-scale 0", func(t *testing.T) {
		for _, cmdFlags := range []*struct {
			what string
			cmd  func() *cobra.Command
		}{
			{"sandbox", func() *cobra.Command { return sandboxCreateCmd(t) }},
			{"pool", func() *cobra.Command { return newPoolCreateCmd() }},
		} {
			t.Run(cmdFlags.what, func(t *testing.T) {
				cmd := cmdFlags.cmd()
				require.NoError(t, cmd.Flags().Set("light-sleep-delay", "5m"))

				_, err := parseSingleInstanceScaling(cmd.Flags(), cmdFlags.what)
				require.Error(t, err)
				assert.Contains(t, err.Error(), "can only be used when min-scale is 0")
			})
		}
	})

	t.Run("sleep delays set scale-to-zero targets", func(t *testing.T) {
		cmd := newPoolCreateCmd()
		require.NoError(t, cmd.Flags().Set("min-scale", "0"))
		require.NoError(t, cmd.Flags().Set("deep-sleep-delay", "30m"))

		scaling, err := parseSingleInstanceScaling(cmd.Flags(), "pool")
		require.NoError(t, err)
		assert.Equal(t, int64(0), scaling.GetMin())
		targets := scaling.GetTargets()
		require.Len(t, targets, 1)
		delay := targets[0].GetSleepIdleDelay()
		assert.Equal(t, int64(1800), delay.GetDeepSleepValue())
		assert.False(t, delay.HasLightSleepValue())
	})
}
