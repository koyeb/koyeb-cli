package koyeb

import "github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"

type cogAddon struct{}

func (c *cogAddon) PreDeploy(ctx *CLIContext, definition *koyeb.DeploymentDefinition) error {
	return nil
}

func (c *cogAddon) PostDeploy(ctx *CLIContext, definition *koyeb.DeploymentDefinition) error {
	return nil
}

func (c *cogAddon) Setup(ctx *CLIContext) error {
	return nil
}

func (c *cogAddon) Cleanup(ctx *CLIContext) error {
	return nil
}
