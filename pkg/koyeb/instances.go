package koyeb

import (
	"context"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/idmapper"
	"github.com/spf13/cobra"
)

func NewInstanceCmd() *cobra.Command {
	instanceHandler := NewInstanceHandler()

	instanceCmd := &cobra.Command{
		Use:               "instances ACTION",
		Aliases:           []string{"i", "inst", "instance"},
		Short:             "Instances",
		PersistentPreRunE: instanceHandler.InitHandler,
	}

	listInstanceCmd := &cobra.Command{
		Use:   "list",
		Short: "List instances",
		RunE:  instanceHandler.List,
	}
	listInstanceCmd.Flags().String("app", "", "Filter on App id or name")
	listInstanceCmd.Flags().String("service", "", "Filter on Service id or name")
	instanceCmd.AddCommand(listInstanceCmd)

	getInstanceCmd := &cobra.Command{
		Use:   "get NAME",
		Short: "Get instance",
		Args:  cobra.ExactArgs(1),
		RunE:  instanceHandler.Get,
	}
	instanceCmd.AddCommand(getInstanceCmd)

	describeInstanceCmd := &cobra.Command{
		Use:   "describe NAME",
		Short: "Describe instance",
		Args:  cobra.ExactArgs(1),
		RunE:  instanceHandler.Describe,
	}
	instanceCmd.AddCommand(describeInstanceCmd)

	execInstanceCmd := &cobra.Command{
		Use:   "exec NAME CMD -- [args...]",
		Short: "Run a command in the context of an instance",
		Args:  cobra.MinimumNArgs(2),
		RunE:  instanceHandler.Exec,
	}
	instanceCmd.AddCommand(execInstanceCmd)

	logInstanceCmd := &cobra.Command{
		Use:     "logs NAME",
		Aliases: []string{"l", "log"},
		Short:   "Get instance logs",
		Args:    cobra.ExactArgs(1),
		RunE:    instanceHandler.Log,
	}
	instanceCmd.AddCommand(logInstanceCmd)

	return instanceCmd
}

func NewInstanceHandler() *InstanceHandler {
	return &InstanceHandler{}
}

type InstanceHandler struct {
	ctx    context.Context
	client *koyeb.APIClient
	mapper *idmapper.Mapper
}

func (h *InstanceHandler) ResolveInstanceArgs(val string) string {
	instanceMapper := h.mapper.Instance()
	id, err := instanceMapper.ResolveID(val)
	if err != nil {
		fatalApiError(err)
	}

	return id
}

func (h *InstanceHandler) InitHandler(cmd *cobra.Command, args []string) error {
	h.ctx = getAuth(context.Background())
	h.client = getApiClient()
	h.mapper = idmapper.NewMapper(h.ctx, h.client)
	return nil
}
