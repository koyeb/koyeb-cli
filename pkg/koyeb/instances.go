package koyeb

import (
	"github.com/spf13/cobra"
)

func NewInstanceCmd() *cobra.Command {
	instanceHandler := NewInstanceHandler()

	instanceCmd := &cobra.Command{
		Use:     "instances ACTION",
		Aliases: []string{"i", "inst", "instance"},
		Short:   "Instances",
	}

	listInstanceCmd := &cobra.Command{
		Use:   "list",
		Short: "List instances",
		RunE:  WithCLIContext(instanceHandler.List),
	}
	listInstanceCmd.Flags().String("app", "", "Filter on App id or name")
	listInstanceCmd.Flags().String("service", "", "Filter on Service id or name")
	instanceCmd.AddCommand(listInstanceCmd)

	getInstanceCmd := &cobra.Command{
		Use:   "get NAME",
		Short: "Get instance",
		Args:  cobra.ExactArgs(1),
		RunE:  WithCLIContext(instanceHandler.Get),
	}
	instanceCmd.AddCommand(getInstanceCmd)

	describeInstanceCmd := &cobra.Command{
		Use:   "describe NAME",
		Short: "Describe instance",
		Args:  cobra.ExactArgs(1),
		RunE:  WithCLIContext(instanceHandler.Describe),
	}
	instanceCmd.AddCommand(describeInstanceCmd)

	execInstanceCmd := &cobra.Command{
		Use:     "exec NAME CMD -- [args...]",
		Short:   "Run a command in the context of an instance",
		Aliases: []string{"run", "attach"},
		Args:    cobra.MinimumNArgs(2),
		RunE:    WithCLIContext(instanceHandler.Exec),
	}
	instanceCmd.AddCommand(execInstanceCmd)

	logInstanceCmd := &cobra.Command{
		Use:     "logs NAME",
		Aliases: []string{"l", "log"},
		Short:   "Get instance logs",
		Args:    cobra.ExactArgs(1),
		RunE:    WithCLIContext(instanceHandler.Logs),
	}
	instanceCmd.AddCommand(logInstanceCmd)

	return instanceCmd
}

func NewInstanceHandler() *InstanceHandler {
	return &InstanceHandler{}
}

type InstanceHandler struct {
}

func (h *InstanceHandler) ResolveInstanceArgs(ctx *CLIContext, val string) string {
	instanceMapper := ctx.mapper.Instance()
	id, err := instanceMapper.ResolveID(val)
	if err != nil {
		fatalApiError(err, nil)
	}

	return id
}

func (h *InstanceHandler) ResolveServiceArgs(ctx *CLIContext, val string) string {
	svcMapper := ctx.mapper.Service()
	id, err := svcMapper.ResolveID(val)
	if err != nil {
		fatalApiError(err, nil)
	}

	return id
}
