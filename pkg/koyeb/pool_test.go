package koyeb

import (
	"context"
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
}

func TestPoolCreateCmdDoesNotRegisterWebOrSandboxOnlyFlags(t *testing.T) {
	cmd := newPoolCreateCmd()
	flags := cmd.Flags()

	// Web service configuration flags must not be registered: cobra rejects
	// them as unknown flags, which guards against invalid pool definitions.
	for _, name := range []string{"type", "ports", "routes", "checks"} {
		assert.Nil(t, flags.Lookup(name), "flag --%s must not be registered on pool create", name)
	}
	// Sandbox-specific flags have no pool equivalent.
	for _, name := range []string{"app", "wait", "wait-timeout"} {
		assert.Nil(t, flags.Lookup(name), "flag --%s must not be registered on pool create", name)
	}
	// The curated flag set must be declared.
	for _, name := range []string{
		"size", "docker", "docker-private-registry-secret", "docker-args",
		"docker-command", "docker-entrypoint", "instance-type", "regions",
		"env", "config-file", "min-scale", "light-sleep-delay", "deep-sleep-delay",
	} {
		assert.NotNil(t, flags.Lookup(name), "flag --%s must be registered on pool create", name)
	}
}

func TestPoolCmdWiring(t *testing.T) {
	cmd := NewPoolCmd()

	assert.Equal(t, "pool ACTION", cmd.Use)
	assert.Equal(t, []string{"pools", "sp"}, cmd.Aliases)
	assert.NotNil(t, cmd.PersistentFlags().Lookup("project"))
	assert.NotNil(t, cmd.PersistentFlags().Lookup("workspace"))

	for _, sub := range []string{"create", "list", "get", "describe", "delete", "claim", "claims"} {
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
