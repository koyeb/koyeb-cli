package koyeb

import (
	"context"
	"testing"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/errors"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/idmapper"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The volumes flag bundle (--volumes) is the single registration seam of
// the service definition's volume flags. These tests pin the bundle
// contract; the per-surface suites pin the merged behavior.

func TestAddVolumesFlagsServiceSurface(t *testing.T) {
	flags := pflag.NewFlagSet("service", pflag.ContinueOnError)
	addVolumesFlags(flags)

	flag := flags.Lookup("volumes")
	require.NotNil(t, flag, "--volumes must be registered by the volumes bundle")
	assert.Equal(t, "[]", flag.DefValue, "--volumes default value")
	assert.Equal(t,
		"Update service volumes using the format VOLUME:PATH, for example --volume myvolume:/data."+
			"To delete a volume, use !VOLUME, for example --volume '!myvolume'\n",
		flag.Usage, "--volumes help text")
}

// volumesFlagNames returns the flag names the bundle registers — the
// bundle union the surface flag-set tests derive their expectations from.
func volumesFlagNames() []string {
	flags := pflag.NewFlagSet("volumes", pflag.ContinueOnError)
	addVolumesFlags(flags)
	names := make([]string, 0, 1)
	flags.VisitAll(func(f *pflag.Flag) { names = append(names, f.Name) })
	return names
}

// volumesBundleTestFlagSet registers the bundle surface the parse step
// consumes.
func volumesBundleTestFlagSet() *pflag.FlagSet {
	flags := pflag.NewFlagSet("volumes", pflag.ContinueOnError)
	addVolumesFlags(flags)
	return flags
}

// liveVolumes returns a mounted volume carrying a live value, so the
// changed-only merge conventions are observable. The mapper resolves
// volume IDs: a full UUID passes through without any API call.
func liveVolumes() []koyeb.DeploymentVolume {
	return []koyeb.DeploymentVolume{{
		Id:   koyeb.PtrString("323e4567-e89b-42d3-a456-426614174000"),
		Path: koyeb.PtrString("/data"),
	}}
}

func volumesTestContext() *CLIContext {
	return &CLIContext{
		Context: context.Background(),
		Mapper:  idmapper.NewMapper(context.Background(), nil),
	}
}

func TestParseVolumesNoFlagsKeepsLiveValues(t *testing.T) {
	h := &ServiceHandler{}
	flags := volumesBundleTestFlagSet()

	volumes, err := h.parseVolumes(volumesTestContext(), flags, liveVolumes())
	require.NoError(t, err)
	require.Len(t, volumes, 1)
	assert.Equal(t, "323e4567-e89b-42d3-a456-426614174000", volumes[0].GetId())
	assert.Equal(t, "/data", volumes[0].GetPath())
}

func TestParseVolumesChangedFlagsMergeLiveValues(t *testing.T) {
	h := &ServiceHandler{}

	t.Run("--volumes mounts a new volume", func(t *testing.T) {
		flags := volumesBundleTestFlagSet()
		require.NoError(t, flags.Parse([]string{
			"--volumes", "423e4567-e89b-42d3-a456-426614174000:/other",
		}))

		volumes, err := h.parseVolumes(volumesTestContext(), flags, liveVolumes())
		require.NoError(t, err)
		require.Len(t, volumes, 2)
		assert.Equal(t, "/data", volumes[0].GetPath(), "the live mount must be kept")
		assert.Equal(t, "423e4567-e89b-42d3-a456-426614174000", volumes[1].GetId())
		assert.Equal(t, "/other", volumes[1].GetPath())
	})

	t.Run("--volumes remounts a live volume to a new path", func(t *testing.T) {
		flags := volumesBundleTestFlagSet()
		require.NoError(t, flags.Parse([]string{
			"--volumes", "323e4567-e89b-42d3-a456-426614174000:/new",
		}))

		volumes, err := h.parseVolumes(volumesTestContext(), flags, liveVolumes())
		require.NoError(t, err)
		require.Len(t, volumes, 1)
		assert.Equal(t, "323e4567-e89b-42d3-a456-426614174000", volumes[0].GetId())
		assert.Equal(t, "/new", volumes[0].GetPath())
	})

	t.Run("'!' unmounts a live volume", func(t *testing.T) {
		flags := volumesBundleTestFlagSet()
		require.NoError(t, flags.Parse([]string{
			"--volumes", "!323e4567-e89b-42d3-a456-426614174000",
		}))

		volumes, err := h.parseVolumes(volumesTestContext(), flags, liveVolumes())
		require.NoError(t, err)
		assert.Empty(t, volumes)
	})
}

func TestParseVolumesErrorPaths(t *testing.T) {
	h := &ServiceHandler{}

	for name, tc := range map[string]struct {
		value    string
		wantWhy  string
		wantWhat string
	}{
		"a mount without a path": {
			value:    "323e4567-e89b-42d3-a456-426614174000",
			wantWhy:  "unable to parse the volume \"323e4567-e89b-42d3-a456-426614174000\"",
			wantWhat: "Error while configuring the service",
		},
		"an unmount carrying a path": {
			value:    "!323e4567-e89b-42d3-a456-426614174000:/data",
			wantWhy:  "unable to parse the volume \"!323e4567-e89b-42d3-a456-426614174000:/data\"",
			wantWhat: "Error while configuring the service",
		},
	} {
		t.Run(name, func(t *testing.T) {
			flags := volumesBundleTestFlagSet()
			require.NoError(t, flags.Parse([]string{"--volumes", tc.value}))

			_, err := h.parseVolumes(volumesTestContext(), flags, liveVolumes())
			var cliErr *errors.CLIError
			require.ErrorAs(t, err, &cliErr)
			assert.Equal(t, tc.wantWhat, cliErr.What)
			assert.Equal(t, tc.wantWhy, cliErr.Why)
		})
	}
}
