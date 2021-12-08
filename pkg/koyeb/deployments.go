package koyeb

import (
	"github.com/spf13/cobra"
)

func NewDeploymentCmd() *cobra.Command {
	h := NewDeploymentHandler()

	deploymentCmd := &cobra.Command{
		Use:     "deployments ACTION",
		Aliases: []string{"d", "deployment"},
		Short:   "Deployments",
	}

	listDeploymentCmd := &cobra.Command{
		Use:   "list",
		Short: "List deployments",
		RunE:  h.List,
	}
	deploymentCmd.AddCommand(listDeploymentCmd)

	getDeploymentCmd := &cobra.Command{
		Use:   "get NAME",
		Short: "Get deployment",
		Args:  cobra.ExactArgs(1),
		RunE:  h.Get,
	}
	deploymentCmd.AddCommand(getDeploymentCmd)

	describeDeploymentCmd := &cobra.Command{
		Use:   "describe NAME",
		Short: "Describe deployment",
		Args:  cobra.ExactArgs(1),
		RunE:  h.Describe,
	}
	deploymentCmd.AddCommand(describeDeploymentCmd)

	return deploymentCmd
}

func NewDeploymentHandler() *DeploymentHandler {
	return &DeploymentHandler{}
}

type DeploymentHandler struct {
}
