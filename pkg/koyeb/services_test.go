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
		"source_buildpack_no_arg": {
			args: []string{},
			source: &koyeb.GitSource{
				Buildpack: &koyeb.BuildpackBuilder{},
			},
			expected: &koyeb.GitSource{
				Buildpack: &koyeb.BuildpackBuilder{},
			},
		},
		"source_docker_no_arg": {
			args: []string{},
			source: &koyeb.GitSource{
				Docker: &koyeb.DockerBuilder{},
			},
			expected: &koyeb.GitSource{
				Docker: &koyeb.DockerBuilder{},
			},
		},
		"source_buildpack_set_builder_to_buildpack": {
			args: []string{"--git-builder", "buildpack"},
			source: &koyeb.GitSource{
				Buildpack: &koyeb.BuildpackBuilder{},
			},
			expected: &koyeb.GitSource{
				Buildpack: &koyeb.BuildpackBuilder{},
			},
		},
		"source_buildpack_set_builder_to_docker": {
			args: []string{"--git-builder", "docker"},
			source: &koyeb.GitSource{
				Buildpack: &koyeb.BuildpackBuilder{},
			},
			expected: &koyeb.GitSource{
				Docker: &koyeb.DockerBuilder{},
			},
		},
		"source_docker_set_builder_to_buildpack": {
			args: []string{"--git-builder", "buildpack"},
			source: &koyeb.GitSource{
				Docker: &koyeb.DockerBuilder{},
			},
			expected: &koyeb.GitSource{
				Buildpack: &koyeb.BuildpackBuilder{},
			},
		},
		"source_docker_set_builder_to_docker": {
			args: []string{"--git-builder", "docker"},
			source: &koyeb.GitSource{
				Docker: &koyeb.DockerBuilder{},
			},
			expected: &koyeb.GitSource{
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
				Docker: &koyeb.DockerBuilder{
					Privileged: koyeb.PtrBool(true),
				},
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			cmd := &cobra.Command{}

			addServiceDefinitionFlags(cmd.Flags())

			cmd.SetArgs(tc.args)
			if err := cmd.ParseFlags(tc.args); err != nil {
				t.Fatal()
			}

			ret, err := setGitSourceBuilder(cmd.Flags(), tc.source)

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
			expected:       []string{"fra"},
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
			addServiceDefinitionFlags(cmd.Flags())

			err := cmd.ParseFlags(test.cliFlags)
			assert.NoError(t, err)

			regions, err := parseRegions(cmd.Flags(), test.currentRegions)
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
			setRegions(&test.definition, test.regions)
			assert.Equal(t, test.expected, test.definition)
		})
	}
}

func TestDynamicDefaults(t *testing.T) {
	tests := map[string]struct {
		cliFlags       []string
		expectedRoutes []string
		expectedPorts  []string
		expectedErr    bool
	}{
		"One port set, one route set": {
			cliFlags:       []string{"--ports", "80", "--routes", "/:80"},
			expectedRoutes: []string{"/:80"},
			expectedPorts:  []string{"80"},
			expectedErr:    false,
		},
		"Two routes set, no port": {
			cliFlags:       []string{"--routes", "/:9999,/api:5555"},
			expectedRoutes: []string{""},
			expectedPorts:  []string{""},
			expectedErr:    true,
		},
		"Two ports set, no routes": {
			cliFlags:       []string{"--ports", "9999,5555"},
			expectedRoutes: []string{""},
			expectedPorts:  []string{""},
			expectedErr:    true,
		},
		"One route set, no port": {
			cliFlags:       []string{"--routes", "/:90"},
			expectedRoutes: []string{"/:90"},
			expectedPorts:  []string{"90"},
			expectedErr:    false,
		},
		"One port set, no route": {
			cliFlags:       []string{"--ports", "90:http"},
			expectedRoutes: []string{"/:90"},
			expectedPorts:  []string{"90:http"},
			expectedErr:    false,
		},
		"One port set (no protocol), no route": {
			cliFlags:       []string{"--ports", "90"},
			expectedRoutes: []string{"/:90"},
			expectedPorts:  []string{"90"},
			expectedErr:    false,
		},
		"No port, no route": {
			cliFlags:       []string{},
			expectedRoutes: []string{},
			expectedPorts:  []string{},
			expectedErr:    false,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			cmd := &cobra.Command{}
			addServiceDefinitionFlags(cmd.Flags())

			err := cmd.ParseFlags(test.cliFlags)
			assert.NoError(t, err)

			err = dynamicDefaults(koyeb.DEPLOYMENTDEFINITIONTYPE_WEB, cmd.Flags())

			if test.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)

				actualRoutes, err := cmd.Flags().GetStringSlice("routes")
				if err != nil {
					t.Fatalf("Failed to get 'routes': %v", err)
				}
				assert.Equal(t, test.expectedRoutes, actualRoutes)

				actualPorts, err := cmd.Flags().GetStringSlice("ports")
				if err != nil {
					t.Fatalf("Failed to get 'ports': %v", err)
				}
				assert.Equal(t, test.expectedPorts, actualPorts)
			}
		})
	}
}
