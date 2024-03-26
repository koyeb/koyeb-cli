package koyeb

import (
	"github.com/spf13/cobra"
)

func NewRegionalDeploymentCmd() *cobra.Command {
	h := NewRegionalDeploymentHandler()

	regionalDeploymentCmd := &cobra.Command{
		Use:     "regional-deployments ACTION",
		Aliases: []string{"rd", "rdep", "rdepl", "rdeploy", "rdeployment", "regional-deployment"},
		Short:   "Regional deployments",
	}

	listRegionalDeploymentCmd := &cobra.Command{
		Use:   "list",
		Short: "List regional deployments",
		RunE:  WithCLIContext(h.List),
	}
	listRegionalDeploymentCmd.Flags().String("deployment", "", "Limit the list to regional deployments of a specific deployment")
	regionalDeploymentCmd.AddCommand(listRegionalDeploymentCmd)

	getRegionalDeploymentCmd := &cobra.Command{
		Use:   "get NAME",
		Short: "Get regional deployment",
		Args:  cobra.ExactArgs(1),
		RunE:  WithCLIContext(h.Get),
	}
	regionalDeploymentCmd.AddCommand(getRegionalDeploymentCmd)

	return regionalDeploymentCmd
}

func NewRegionalDeploymentHandler() *RegionalDeploymentHandler {
	return &RegionalDeploymentHandler{}
}

type RegionalDeploymentHandler struct {
}

func (h *RegionalDeploymentHandler) ResolveDeploymentArgs(ctx *CLIContext, val string) (string, error) {
	deploymentMapper := ctx.Mapper.Deployment()
	id, err := deploymentMapper.ResolveID(val)
	if err != nil {
		return "", err
	}
	return id, nil
}

func (h *RegionalDeploymentHandler) ResolveRegionalDeploymentArgs(ctx *CLIContext, val string) (string, error) {
	regionalDeploymentMapper := ctx.Mapper.RegionalDeployment()
	id, err := regionalDeploymentMapper.ResolveID(val)
	if err != nil {
		return "", err
	}
	return id, nil
}
