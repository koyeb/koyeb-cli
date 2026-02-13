package koyeb

import (
	"testing"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestSetGitSourceBuilder(t *testing.T) {
	tests := map[string]struct {
		args     []string
		source   *koyeb.GitSource
		expected *koyeb.GitSource
	}{
		"invalid_git_builder": {
			args:     []string{"--git-builder", "xxx"},
			source:   &koyeb.GitSource{},
			expected: nil,
		},
		"source_git_default_branch": {
			args:   []string{"--git", "github.com/org/repo"},
			source: &koyeb.GitSource{},
			expected: &koyeb.GitSource{
				Repository: koyeb.PtrString("github.com/org/repo"),
				Branch:     koyeb.PtrString("main"),
			},
		},
		"source_buildpack_no_arg": {
			args: []string{},
			source: &koyeb.GitSource{
				Buildpack: &koyeb.BuildpackBuilder{},
			},
			expected: &koyeb.GitSource{
				Branch:    koyeb.PtrString("main"),
				Buildpack: &koyeb.BuildpackBuilder{},
			},
		},
		"source_docker_no_arg": {
			args: []string{},
			source: &koyeb.GitSource{
				Docker: &koyeb.DockerBuilder{},
			},
			expected: &koyeb.GitSource{
				Branch: koyeb.PtrString("main"),
				Docker: &koyeb.DockerBuilder{},
			},
		},
		"source_buildpack_set_builder_to_buildpack": {
			args: []string{"--git-builder", "buildpack"},
			source: &koyeb.GitSource{
				Buildpack: &koyeb.BuildpackBuilder{},
			},
			expected: &koyeb.GitSource{
				Branch:    koyeb.PtrString("main"),
				Buildpack: &koyeb.BuildpackBuilder{},
			},
		},
		"source_buildpack_set_builder_to_docker": {
			args: []string{"--git-builder", "docker"},
			source: &koyeb.GitSource{
				Buildpack: &koyeb.BuildpackBuilder{},
			},
			expected: &koyeb.GitSource{
				Branch: koyeb.PtrString("main"),
				Docker: &koyeb.DockerBuilder{},
			},
		},
		"source_docker_set_builder_to_buildpack": {
			args: []string{"--git-builder", "buildpack"},
			source: &koyeb.GitSource{
				Docker: &koyeb.DockerBuilder{},
			},
			expected: &koyeb.GitSource{
				Branch:    koyeb.PtrString("main"),
				Buildpack: &koyeb.BuildpackBuilder{},
			},
		},
		"source_docker_set_builder_to_docker": {
			args: []string{"--git-builder", "docker"},
			source: &koyeb.GitSource{
				Docker: &koyeb.DockerBuilder{},
			},
			expected: &koyeb.GitSource{
				Branch: koyeb.PtrString("main"),
				Docker: &koyeb.DockerBuilder{},
			},
		},
		"source_buildpack_set_docker_args": {
			args: []string{"--git-docker-dockerfile", "Dockerfile.dev"},
			source: &koyeb.GitSource{
				Buildpack: &koyeb.BuildpackBuilder{},
			},
			expected: nil,
		},
		"source_buildpack_set_buildpack_args": {
			args: []string{"--git-buildpack-run-command", "run", "--git-buildpack-build-command", "build", "--privileged"},
			source: &koyeb.GitSource{
				Buildpack: &koyeb.BuildpackBuilder{
					Privileged: koyeb.PtrBool(false),
				},
			},
			expected: &koyeb.GitSource{
				Branch:       koyeb.PtrString("main"),
				RunCommand:   koyeb.PtrString("run"),
				BuildCommand: koyeb.PtrString("build"),
				Buildpack: &koyeb.BuildpackBuilder{
					RunCommand:   koyeb.PtrString("run"),
					BuildCommand: koyeb.PtrString("build"),
					Privileged:   koyeb.PtrBool(true),
				},
			},
		},
		"source_docker_set_buildpack_args": {
			args: []string{"--git-buildpack-run-command", "run", "--git-buildpack-build-command", "build", "--privileged"},
			source: &koyeb.GitSource{
				Docker: &koyeb.DockerBuilder{},
			},
			expected: nil,
		},
		"source_docker_set_docker_args": {
			args: []string{
				"--git-docker-command", "cmd",
				"--git-docker-args", "arg1", "--git-docker-args", "arg2",
				"--git-docker-entrypoint", "entrypoint.sh", "--git-docker-entrypoint", "entrypoint-arg",
				"--git-docker-dockerfile", "Dockerfile.dev",
				"--git-docker-target", "dev",
				"--privileged",
			},
			source: &koyeb.GitSource{
				Docker: &koyeb.DockerBuilder{},
			},
			expected: &koyeb.GitSource{
				Branch: koyeb.PtrString("main"),
				Docker: &koyeb.DockerBuilder{
					Command:    koyeb.PtrString("cmd"),
					Args:       []string{"arg1", "arg2"},
					Entrypoint: []string{"entrypoint.sh", "entrypoint-arg"},
					Dockerfile: koyeb.PtrString("Dockerfile.dev"),
					Target:     koyeb.PtrString("dev"),
					Privileged: koyeb.PtrBool(true),
				},
			},
		},
		"source_buildpack_override_privileged_without_other_args": {
			args: []string{"--privileged=false"},
			source: &koyeb.GitSource{
				Buildpack: &koyeb.BuildpackBuilder{
					Privileged: koyeb.PtrBool(true),
				},
			},
			expected: &koyeb.GitSource{
				Branch: koyeb.PtrString("main"),
				Buildpack: &koyeb.BuildpackBuilder{
					Privileged: koyeb.PtrBool(false),
				},
			},
		},
		"source_docker_override_privileged_without_other_args": {
			args: []string{"--privileged"},
			source: &koyeb.GitSource{
				Docker: &koyeb.DockerBuilder{
					Privileged: koyeb.PtrBool(false),
				},
			},
			expected: &koyeb.GitSource{
				Branch: koyeb.PtrString("main"),
				Docker: &koyeb.DockerBuilder{
					Privileged: koyeb.PtrBool(true),
				},
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			cmd := &cobra.Command{}

			h := NewServiceHandler()
			h.addServiceDefinitionFlags(cmd.Flags())

			cmd.SetArgs(tc.args)
			if err := cmd.ParseFlags(tc.args); err != nil {
				t.Fatal()
			}

			ret, err := h.parseGitSource(cmd.Flags(), tc.source)

			if tc.expected != nil {
				assert.Nil(t, err)
				assert.Equal(t, tc.expected, ret)
			} else {
				assert.NotNil(t, err)
				assert.Nil(t, ret)
			}
		})
	}
}

func TestParseRegions(t *testing.T) {
	tests := map[string]struct {
		cliFlags       []string
		currentRegions []string
		expected       []string
	}{
		"set default": {
			cliFlags:       []string{},
			currentRegions: []string{},
			expected:       []string{"was"},
		},
		"replace default": {
			cliFlags:       []string{"--region", "tyo"},
			currentRegions: []string{},
			expected:       []string{"tyo"},
		},
		"two regions": {
			cliFlags:       []string{"--region", "tyo", "--region", "sin"},
			currentRegions: []string{},
			expected:       []string{"tyo", "sin"},
		},
		"override one of the two regions": {
			cliFlags:       []string{"--region", "tyo", "--region", "sin"},
			currentRegions: []string{"tyo"},
			expected:       []string{"tyo", "sin"},
		},
		"remove one region": {
			cliFlags:       []string{"--region", "!tyo"},
			currentRegions: []string{"sin", "tyo"},
			expected:       []string{"sin"},
		},
		"remove non existing region (noop)": {
			cliFlags:       []string{"--region", "!tyo"},
			currentRegions: []string{"sin"},
			expected:       []string{"sin"},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			cmd := &cobra.Command{}
			h := NewServiceHandler()
			h.addServiceDefinitionFlags(cmd.Flags())

			err := cmd.ParseFlags(test.cliFlags)
			assert.NoError(t, err)

			regions, err := h.parseRegions(cmd.Flags(), test.currentRegions)
			assert.Equal(t, test.expected, regions)
			assert.NoError(t, err)
		})
	}
}

func TestSetRegions(t *testing.T) {
	tests := map[string]struct {
		definition koyeb.DeploymentDefinition
		regions    []string
		expected   koyeb.DeploymentDefinition
	}{
		"empty definition": {
			definition: koyeb.DeploymentDefinition{},
			regions:    []string{"par", "fra"},
			expected: koyeb.DeploymentDefinition{
				Regions: []string{"par", "fra"},
			},
		},
		"remove region": {
			definition: koyeb.DeploymentDefinition{
				Env: []koyeb.DeploymentEnv{
					{Key: koyeb.PtrString("Env"), Value: koyeb.PtrString("value"), Scopes: []string{"region:par", "region:fra"}},
					{Key: koyeb.PtrString("Env2"), Value: koyeb.PtrString("value2"), Scopes: []string{"region:par", "region:fra"}},
				},
				Scalings: []koyeb.DeploymentScaling{
					{Min: koyeb.PtrInt64(1), Max: koyeb.PtrInt64(1), Scopes: []string{"region:par", "region:fra"}},
				},
				InstanceTypes: []koyeb.DeploymentInstanceType{
					{Type: koyeb.PtrString("nano"), Scopes: []string{"region:par", "region:fra"}},
				},
				Regions: []string{"par", "fra"},
			},
			regions: []string{"par"},
			expected: koyeb.DeploymentDefinition{
				Env: []koyeb.DeploymentEnv{
					{Key: koyeb.PtrString("Env"), Value: koyeb.PtrString("value"), Scopes: []string{"region:par"}},
					{Key: koyeb.PtrString("Env2"), Value: koyeb.PtrString("value2"), Scopes: []string{"region:par"}},
				},
				Scalings: []koyeb.DeploymentScaling{
					{Min: koyeb.PtrInt64(1), Max: koyeb.PtrInt64(1), Scopes: []string{"region:par"}},
				},
				InstanceTypes: []koyeb.DeploymentInstanceType{
					{Type: koyeb.PtrString("nano"), Scopes: []string{"region:par"}},
				},
				Regions: []string{"par"},
			},
		},
		"add region": {
			definition: koyeb.DeploymentDefinition{
				Env: []koyeb.DeploymentEnv{
					{Key: koyeb.PtrString("Env"), Value: koyeb.PtrString("value"), Scopes: []string{"region:par"}},
					{Key: koyeb.PtrString("Env2"), Value: koyeb.PtrString("value2"), Scopes: []string{"region:par"}},
				},
				Scalings: []koyeb.DeploymentScaling{
					{Min: koyeb.PtrInt64(1), Max: koyeb.PtrInt64(1), Scopes: []string{"region:par"}},
				},
				InstanceTypes: []koyeb.DeploymentInstanceType{
					{Type: koyeb.PtrString("nano"), Scopes: []string{"region:par"}},
				},
				Regions: []string{"par"},
			},
			regions: []string{"par", "fra"},
			expected: koyeb.DeploymentDefinition{
				Env: []koyeb.DeploymentEnv{
					{Key: koyeb.PtrString("Env"), Value: koyeb.PtrString("value"), Scopes: []string{"region:par", "region:fra"}},
					{Key: koyeb.PtrString("Env2"), Value: koyeb.PtrString("value2"), Scopes: []string{"region:par", "region:fra"}},
				},
				Scalings: []koyeb.DeploymentScaling{
					{Min: koyeb.PtrInt64(1), Max: koyeb.PtrInt64(1), Scopes: []string{"region:par", "region:fra"}},
				},
				InstanceTypes: []koyeb.DeploymentInstanceType{
					{Type: koyeb.PtrString("nano"), Scopes: []string{"region:par", "region:fra"}},
				},
				Regions: []string{"par", "fra"},
			},
		},
		// scopes other than "region:xxx" should be ignored. They don't exist yet but we don't want to break if they are added in the future.
		"ignore extra scopes": {
			definition: koyeb.DeploymentDefinition{
				Env: []koyeb.DeploymentEnv{
					{Key: koyeb.PtrString("Env"), Value: koyeb.PtrString("value"), Scopes: []string{"whatever", "region:par"}},
				},
				Scalings: []koyeb.DeploymentScaling{
					{Min: koyeb.PtrInt64(1), Max: koyeb.PtrInt64(1), Scopes: []string{"whatever", "region:par"}},
				},
				InstanceTypes: []koyeb.DeploymentInstanceType{
					{Type: koyeb.PtrString("nano"), Scopes: []string{"whatever", "region:par"}},
				},
				Regions: []string{"par"},
			},
			regions: []string{"par", "fra"},
			expected: koyeb.DeploymentDefinition{
				Env: []koyeb.DeploymentEnv{
					{Key: koyeb.PtrString("Env"), Value: koyeb.PtrString("value"), Scopes: []string{"whatever", "region:par", "region:fra"}},
				},
				Scalings: []koyeb.DeploymentScaling{
					{Min: koyeb.PtrInt64(1), Max: koyeb.PtrInt64(1), Scopes: []string{"whatever", "region:par", "region:fra"}},
				},
				InstanceTypes: []koyeb.DeploymentInstanceType{
					{Type: koyeb.PtrString("nano"), Scopes: []string{"whatever", "region:par", "region:fra"}},
				},
				Regions: []string{"par", "fra"},
			},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			h := NewServiceHandler()
			h.setRegions(&test.definition, test.regions)
			assert.Equal(t, test.expected, test.definition)
		})
	}
}

func TestSetDefaultPortsAndRoutes(t *testing.T) {
	tests := map[string]struct {
		definition  koyeb.DeploymentDefinition
		expected    koyeb.DeploymentDefinition
		expectedErr bool
	}{
		"No port set, no route set": {
			definition: koyeb.DeploymentDefinition{},
			expected: koyeb.DeploymentDefinition{
				Ports: []koyeb.DeploymentPort{
					{Port: koyeb.PtrInt64(8000), Protocol: koyeb.PtrString("http")},
				},
				Routes: []koyeb.DeploymentRoute{
					{Port: koyeb.PtrInt64(8000), Path: koyeb.PtrString("/")},
				},
			},
			expectedErr: false,
		},
		"Two routes set, no port": {
			definition: koyeb.DeploymentDefinition{
				Routes: []koyeb.DeploymentRoute{
					{Port: koyeb.PtrInt64(5555), Path: koyeb.PtrString("/")},
					{Port: koyeb.PtrInt64(9999), Path: koyeb.PtrString("/api")},
				},
			},
			expected:    koyeb.DeploymentDefinition{},
			expectedErr: true,
		},
		"Two ports set, no route": {
			definition: koyeb.DeploymentDefinition{
				Ports: []koyeb.DeploymentPort{
					{Port: koyeb.PtrInt64(5555), Protocol: koyeb.PtrString("http")},
					{Port: koyeb.PtrInt64(9999), Protocol: koyeb.PtrString("http")},
				},
			},
			expected:    koyeb.DeploymentDefinition{},
			expectedErr: true,
		},
		"One port set, no route": {
			definition: koyeb.DeploymentDefinition{
				Ports: []koyeb.DeploymentPort{
					{Port: koyeb.PtrInt64(90), Protocol: koyeb.PtrString("http")},
				},
			},
			expected: koyeb.DeploymentDefinition{
				Ports: []koyeb.DeploymentPort{
					{Port: koyeb.PtrInt64(90), Protocol: koyeb.PtrString("http")},
				},
				Routes: []koyeb.DeploymentRoute{
					{Port: koyeb.PtrInt64(90), Path: koyeb.PtrString("/")},
				},
			},
			expectedErr: false,
		},
		"One route set, no port": {
			definition: koyeb.DeploymentDefinition{
				Routes: []koyeb.DeploymentRoute{
					{Port: koyeb.PtrInt64(90), Path: koyeb.PtrString("/")},
				},
			},
			expected: koyeb.DeploymentDefinition{
				Ports: []koyeb.DeploymentPort{
					{Port: koyeb.PtrInt64(90), Protocol: koyeb.PtrString("http")},
				},
				Routes: []koyeb.DeploymentRoute{
					{Port: koyeb.PtrInt64(90), Path: koyeb.PtrString("/")},
				},
			},
			expectedErr: false,
		},
		"Two routes set, one port set": {
			definition: koyeb.DeploymentDefinition{
				Routes: []koyeb.DeploymentRoute{
					{Port: koyeb.PtrInt64(90), Path: koyeb.PtrString("/")},
					{Port: koyeb.PtrInt64(5555), Path: koyeb.PtrString("/api")},
				},
				Ports: []koyeb.DeploymentPort{
					{Port: koyeb.PtrInt64(90), Protocol: koyeb.PtrString("http")},
				},
			},
			expected: koyeb.DeploymentDefinition{
				Routes: []koyeb.DeploymentRoute{
					{Port: koyeb.PtrInt64(90), Path: koyeb.PtrString("/")},
					{Port: koyeb.PtrInt64(5555), Path: koyeb.PtrString("/api")},
				},
				Ports: []koyeb.DeploymentPort{
					{Port: koyeb.PtrInt64(90), Protocol: koyeb.PtrString("http")},
				},
			},
			expectedErr: false,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			h := NewServiceHandler()
			err := h.setDefaultPortsAndRoutes(&test.definition, test.definition.Ports, test.definition.Routes)

			if test.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.expected, test.definition)
			}
		})
	}
}

func TestSetSleepDelayFlags(t *testing.T) {
	tests := map[string]struct {
		args           []string
		minScale       int64
		currentTargets []koyeb.DeploymentScalingTarget
		expected       []koyeb.DeploymentScalingTarget
		expectedErr    bool
	}{
		"set light sleep delay on new service": {
			args:           []string{"--light-sleep-delay", "5m"},
			minScale:       0,
			currentTargets: nil,
			expected: []koyeb.DeploymentScalingTarget{
				{
					SleepIdleDelay: &koyeb.DeploymentScalingTargetSleepIdleDelay{
						LightSleepValue: koyeb.PtrInt64(300),
					},
				},
			},
		},
		"set deep sleep delay on new service": {
			args:           []string{"--deep-sleep-delay", "30m"},
			minScale:       0,
			currentTargets: nil,
			expected: []koyeb.DeploymentScalingTarget{
				{
					SleepIdleDelay: &koyeb.DeploymentScalingTargetSleepIdleDelay{
						DeepSleepValue: koyeb.PtrInt64(1800),
					},
				},
			},
		},
		"set both sleep delays on new service": {
			args:           []string{"--light-sleep-delay", "1m", "--deep-sleep-delay", "10m"},
			minScale:       0,
			currentTargets: nil,
			expected: []koyeb.DeploymentScalingTarget{
				{
					SleepIdleDelay: &koyeb.DeploymentScalingTargetSleepIdleDelay{
						LightSleepValue: koyeb.PtrInt64(60),
						DeepSleepValue:  koyeb.PtrInt64(600),
					},
				},
			},
		},
		"update light sleep delay on existing service": {
			args:     []string{"--light-sleep-delay", "10m"},
			minScale: 0,
			currentTargets: []koyeb.DeploymentScalingTarget{
				{
					SleepIdleDelay: &koyeb.DeploymentScalingTargetSleepIdleDelay{
						LightSleepValue: koyeb.PtrInt64(300),
						DeepSleepValue:  koyeb.PtrInt64(1800),
					},
				},
			},
			expected: []koyeb.DeploymentScalingTarget{
				{
					SleepIdleDelay: &koyeb.DeploymentScalingTargetSleepIdleDelay{
						LightSleepValue: koyeb.PtrInt64(600),
						DeepSleepValue:  koyeb.PtrInt64(1800),
					},
				},
			},
		},
		"disable light sleep delay": {
			args:     []string{"--light-sleep-delay", "0"},
			minScale: 0,
			currentTargets: []koyeb.DeploymentScalingTarget{
				{
					SleepIdleDelay: &koyeb.DeploymentScalingTargetSleepIdleDelay{
						LightSleepValue: koyeb.PtrInt64(300),
						DeepSleepValue:  koyeb.PtrInt64(1800),
					},
				},
			},
			expected: []koyeb.DeploymentScalingTarget{
				{
					SleepIdleDelay: &koyeb.DeploymentScalingTargetSleepIdleDelay{
						DeepSleepValue: koyeb.PtrInt64(1800),
					},
				},
			},
		},
		"disable both sleep delays removes target": {
			args:     []string{"--light-sleep-delay", "0", "--deep-sleep-delay", "0"},
			minScale: 0,
			currentTargets: []koyeb.DeploymentScalingTarget{
				{
					SleepIdleDelay: &koyeb.DeploymentScalingTargetSleepIdleDelay{
						LightSleepValue: koyeb.PtrInt64(300),
						DeepSleepValue:  koyeb.PtrInt64(1800),
					},
				},
			},
			expected: []koyeb.DeploymentScalingTarget{},
		},
		"preserve other targets when setting sleep delay": {
			args:     []string{"--light-sleep-delay", "5m"},
			minScale: 0,
			currentTargets: []koyeb.DeploymentScalingTarget{
				{
					AverageCpu: &koyeb.DeploymentScalingTargetAverageCPU{
						Value: koyeb.PtrInt64(80),
					},
				},
			},
			expected: []koyeb.DeploymentScalingTarget{
				{
					AverageCpu: &koyeb.DeploymentScalingTargetAverageCPU{
						Value: koyeb.PtrInt64(80),
					},
				},
				{
					SleepIdleDelay: &koyeb.DeploymentScalingTargetSleepIdleDelay{
						LightSleepValue: koyeb.PtrInt64(300),
					},
				},
			},
		},
		"no flags does not modify targets": {
			args:     []string{},
			minScale: 0,
			currentTargets: []koyeb.DeploymentScalingTarget{
				{
					SleepIdleDelay: &koyeb.DeploymentScalingTargetSleepIdleDelay{
						LightSleepValue: koyeb.PtrInt64(300),
					},
				},
			},
			expected: []koyeb.DeploymentScalingTarget{
				{
					SleepIdleDelay: &koyeb.DeploymentScalingTargetSleepIdleDelay{
						LightSleepValue: koyeb.PtrInt64(300),
					},
				},
			},
		},
		"error when min-scale is not zero with light-sleep-delay": {
			args:        []string{"--light-sleep-delay", "5m"},
			minScale:    1,
			expectedErr: true,
		},
		"error when min-scale is not zero with deep-sleep-delay": {
			args:        []string{"--deep-sleep-delay", "30m"},
			minScale:    1,
			expectedErr: true,
		},
		"error when min-scale is not zero with both sleep delays": {
			args:        []string{"--light-sleep-delay", "5m", "--deep-sleep-delay", "30m"},
			minScale:    2,
			expectedErr: true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			cmd := &cobra.Command{}
			h := NewServiceHandler()
			h.addServiceDefinitionFlags(cmd.Flags())

			err := cmd.ParseFlags(tc.args)
			assert.NoError(t, err)

			scaling := koyeb.NewDeploymentScalingWithDefaults()
			scaling.SetMin(tc.minScale)
			scaling.SetMax(1)
			scaling.Targets = tc.currentTargets

			err = h.setScalingsTargets(cmd.Flags(), scaling)
			if tc.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, scaling.Targets)
			}
		})
	}
}
