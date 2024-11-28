package koyeb

import "github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"

type cogAddon struct {
	directory        string
}

func (c *cogAddon) PreDeploy(ctx *CLIContext, definition *koyeb.DeploymentDefinition) error {
	return nil
}

func (c *cogAddon) PostDeploy(ctx *CLIContext, definition *koyeb.DeploymentDefinition) error {
	return nil
}

func (c *cogAddon) Setup(ctx *CLIContext, dir string) error {
	c.directory = dir
	return nil
}

func (c *cogAddon) Cleanup(ctx *CLIContext) error {
	return nil
}
