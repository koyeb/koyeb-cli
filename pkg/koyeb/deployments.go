package koyeb

import (
	"github.com/koyeb/koyeb-cli/pkg/koyeb/dates"
	"github.com/spf13/cobra"
)

func NewDeploymentCmd() *cobra.Command {
	h := NewDeploymentHandler()

	deploymentCmd := &cobra.Command{
		Use:     "deployments ACTION",
		Aliases: []string{"d", "dep", "depl", "deployment"},
		Short:   "Deployments",
	}

	listDeploymentCmd := &cobra.Command{
		Use:   "list",
		Short: "List deployments",
		RunE:  WithCLIContext(h.List),
	}
	listDeploymentCmd.Flags().String("app", "", "Limit the list to deployments of a specific app")
	listDeploymentCmd.Flags().String("service", "", "Limit the list to deployments of a specific service")
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

	var since dates.HumanFriendlyDate
	logDeploymentCmd := &cobra.Command{
		Use:     "logs NAME",
		Aliases: []string{"l", "log"},
		Short:   "Get deployment logs",
		Args:    cobra.ExactArgs(1),
		RunE: WithCLIContext(func(ctx *CLIContext, cmd *cobra.Command, args []string) error {
			return h.Logs(ctx, cmd, since.Time, args)
		}),
	}
	deploymentCmd.AddCommand(logDeploymentCmd)
	logDeploymentCmd.Flags().StringP("type", "t", "", "Type of log (runtime, build)")
	logDeploymentCmd.Flags().Var(&since, "since", "DEPRECATED. DO NOT USE. Tail logs after this specific date")
	logDeploymentCmd.Flags().Bool("tail", false, "Tail logs if no `--end-time` is provided.")
	logDeploymentCmd.Flags().StringP("start-time", "s", "", "Return logs after this date")
	logDeploymentCmd.Flags().StringP("end-time", "e", "", "Return logs before this date")
	logDeploymentCmd.Flags().String("regex-search", "", "Filter logs returned with this regex")
	logDeploymentCmd.Flags().String("text-search", "", "Filter logs returned with this text")
	logDeploymentCmd.Flags().String("order", "asc", "Order logs by `asc` or `desc`")
	return deploymentCmd
}

func NewDeploymentHandler() *DeploymentHandler {
	return &DeploymentHandler{}
}

type DeploymentHandler struct {
}

func (h *DeploymentHandler) ResolveAppArgs(ctx *CLIContext, val string) (string, error) {
	appMapper := ctx.Mapper.App()
	id, err := appMapper.ResolveID(val)
	if err != nil {
		return "", err
	}
	return id, nil
}

func (h *DeploymentHandler) ResolveServiceArgs(ctx *CLIContext, val string) (string, error) {
	serviceMapper := ctx.Mapper.Service()
	id, err := serviceMapper.ResolveID(val)
	if err != nil {
		return "", err
	}
	return id, nil
}

func (h *DeploymentHandler) ResolveDeploymentArgs(ctx *CLIContext, val string) (string, error) {
	deploymentMapper := ctx.Mapper.Deployment()
	id, err := deploymentMapper.ResolveID(val)
	if err != nil {
		return "", err
	}
	return id, nil
}
