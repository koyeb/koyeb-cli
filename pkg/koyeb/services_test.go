package koyeb

import (
	"testing"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

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
