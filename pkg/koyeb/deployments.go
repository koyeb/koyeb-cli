package koyeb

import (
	"github.com/spf13/cobra"
)

func NewDeploymentCmd() *cobra.Command {
	h := NewDeploymentHandler()

	deploymentCmd := &cobra.Command{
		Use:     "deployments ACTION",
		Aliases: []string{"d", "dep", "depl", "deploy", "deployment"},
		Short:   "Deployments",
	}

	listDeploymentCmd := &cobra.Command{
		Use:   "list",
		Short: "List deployments",
		RunE:  WithCLIContext(h.List),
	}
	deploymentCmd.AddCommand(listDeploymentCmd)

	getDeploymentCmd := &cobra.Command{
		Use:   "get NAME",
		Short: "Get deployment",
		Args:  cobra.ExactArgs(1),
		RunE:  WithCLIContext(h.Get),
	}
	deploymentCmd.AddCommand(getDeploymentCmd)

	describeDeploymentCmd := &cobra.Command{
		Use:   "describe NAME",
		Short: "Describe deployment",
		Args:  cobra.ExactArgs(1),
		RunE:  WithCLIContext(h.Describe),
	}
	deploymentCmd.AddCommand(describeDeploymentCmd)

	cancelDeploymentCmd := &cobra.Command{
		Use:   "cancel NAME",
		Short: "Cancel deployment",
		Args:  cobra.ExactArgs(1),
		RunE:  WithCLIContext(h.Cancel),
	}
	deploymentCmd.AddCommand(cancelDeploymentCmd)

	logDeploymentCmd := &cobra.Command{
		Use:     "logs NAME",
		Aliases: []string{"l", "log"},
		Short:   "Get deployment logs",
		Args:    cobra.ExactArgs(1),
		RunE:    WithCLIContext(h.Logs),
	}
	deploymentCmd.AddCommand(logDeploymentCmd)
	logDeploymentCmd.Flags().StringP("type", "t", "", "Type of log (runtime,build)")

	return deploymentCmd
}

func NewDeploymentHandler() *DeploymentHandler {
	return &DeploymentHandler{}
}

type DeploymentHandler struct {
}

func (h *DeploymentHandler) ResolveDeploymentArgs(ctx *CLIContext, val string) (string, error) {
	deploymentMapper := ctx.Mapper.Deployment()
	id, err := deploymentMapper.ResolveID(val)
	if err != nil {
		return "", err
	}
	return id, nil
}
