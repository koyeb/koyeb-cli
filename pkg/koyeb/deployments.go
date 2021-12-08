package koyeb

import (
	"context"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/spf13/cobra"
)

func NewDeploymentCmd() *cobra.Command {
	h := NewDeploymentHandler()

	deploymentCmd := &cobra.Command{
		Use:               "deployments ACTION",
		Aliases:           []string{"d", "dep", "depl", "deploy", "deployment"},
		Short:             "Deployments",
		PersistentPreRunE: h.InitHandler,
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

	logDeploymentCmd := &cobra.Command{
		Use:     "logs NAME",
		Aliases: []string{"l", "log"},
		Short:   "Get deployment logs",
		Args:    cobra.ExactArgs(1),
		RunE:    h.Log,
	}
	deploymentCmd.AddCommand(logDeploymentCmd)
	logDeploymentCmd.Flags().StringP("type", "t", "", "Type of log (runtime,build)")

	return deploymentCmd
}

func NewDeploymentHandler() *DeploymentHandler {
	return &DeploymentHandler{}
}

type DeploymentHandler struct {
	client      *koyeb.APIClient
	ctxWithAuth context.Context
}

func (d *DeploymentHandler) InitHandler(cmd *cobra.Command, args []string) error {
	d.client = getApiClient()
	d.ctxWithAuth = getAuth(context.Background())
	return nil
}

func (d *DeploymentHandler) ResolveDeploymentShortID(id string) string {
	return ResolveDeploymentShortID(d.ctxWithAuth, d.client, id)
}
