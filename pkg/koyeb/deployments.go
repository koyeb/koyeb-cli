package koyeb

import (
	"context"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/idmapper"
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
		RunE:    h.Logs,
	}
	deploymentCmd.AddCommand(logDeploymentCmd)
	logDeploymentCmd.Flags().StringP("type", "t", "", "Type of log (runtime,build)")

	return deploymentCmd
}

func NewDeploymentHandler() *DeploymentHandler {
	return &DeploymentHandler{}
}

type DeploymentHandler struct {
	ctx    context.Context
	client *koyeb.APIClient
	mapper *idmapper.Mapper
}

func (h *DeploymentHandler) InitHandler(cmd *cobra.Command, args []string) error {
	h.ctx = getAuth(context.Background())
	h.client = getApiClient()
	h.mapper = idmapper.NewMapper(h.ctx, h.client)
	return nil
}

func (h *DeploymentHandler) ResolveDeploymentArgs(val string) string {
	deploymentMapper := h.mapper.Deployment()
	id, err := deploymentMapper.ResolveID(val)
	if err != nil {
		fatalApiError(err, nil)
	}

	return id
}
