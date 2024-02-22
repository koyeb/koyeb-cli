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

	cpInstanceCmd := &cobra.Command{
		Use:     "cp SRC DST",
		Short:   "Copy files and directories to and from instances.",
		Aliases: []string{"copy"},
		Args:    cobra.ExactArgs(2),
		Example: "\nTo copy a file called `hello.txt` from the current directory of your local machine to the `/tmp` directory of a remote Koyeb Instance, type:\n$> koyeb instance cp hello.txt <instance_id>:/tmp/\nTo copy a `spreadsheet.csv` file from the `/tmp/` directory of your Koyeb Instance to the current directory on your local machine, type:\n$> koyeb instance cp <instance_id>:/tmp/spreadsheet.csv .",
		RunE:    WithCLIContext(instanceHandler.Cp),
	}
	instanceCmd.AddCommand(cpInstanceCmd)

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

func (h *InstanceHandler) ResolveInstanceArgs(ctx *CLIContext, val string) (string, error) {
	instanceMapper := ctx.Mapper.Instance()
	id, err := instanceMapper.ResolveID(val)
	if err != nil {
		return "", err
	}
	return id, nil
}
