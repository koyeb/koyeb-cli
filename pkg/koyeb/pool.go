package koyeb

import (
	"strconv"

	"github.com/spf13/cobra"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/idmapper"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/renderer"
)

func NewPoolCmd() *cobra.Command {
	poolCmd := &cobra.Command{
		Use:     "pool ACTION",
		Aliases: []string{"pools", "sp"},
		Short:   "Manage service pools",
	}
	poolCmd.PersistentFlags().StringP("project", "p", "", "Workspace ID or name")
	poolCmd.PersistentFlags().StringP("workspace", "w", "", "Workspace ID or name (alias for --project)")

	poolCmd.AddCommand(newPoolCreateCmd())
	poolCmd.AddCommand(newPoolListCmd())
	poolCmd.AddCommand(newPoolGetCmd())
	poolCmd.AddCommand(newPoolDescribeCmd())
	poolCmd.AddCommand(newPoolDeleteCmd())
	poolCmd.AddCommand(newPoolClaimCmd())
	poolCmd.AddCommand(newPoolClaimsCmd())

	return poolCmd
}

func NewPoolHandler() *PoolHandler {
	return &PoolHandler{}
}

type PoolHandler struct{}

// ResolvePoolArgs resolves a pool name, short ID, or full UUID to a pool ID.
func ResolvePoolArgs(ctx *CLIContext, val string) (string, error) {
	return ctx.Mapper.Pool().ResolveID(val)
}

func newPoolListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List service pools",
		RunE: WithCLIContext(func(ctx *CLIContext, cmd *cobra.Command, args []string) error {
			return NewPoolHandler().List(ctx, cmd, args)
		}),
	}
}

func newPoolGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get NAME",
		Short: "Get a service pool",
		Args:  cobra.ExactArgs(1),
		RunE: WithCLIContext(func(ctx *CLIContext, cmd *cobra.Command, args []string) error {
			return NewPoolHandler().Get(ctx, cmd, args)
		}),
	}
}

func newPoolDescribeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "describe NAME",
		Short: "Describe a service pool",
		Args:  cobra.ExactArgs(1),
		RunE: WithCLIContext(func(ctx *CLIContext, cmd *cobra.Command, args []string) error {
			return NewPoolHandler().Describe(ctx, cmd, args)
		}),
	}
}

func newPoolDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete NAME",
		Short: "Delete a service pool",
		Args:  cobra.ExactArgs(1),
		RunE: WithCLIContext(func(ctx *CLIContext, cmd *cobra.Command, args []string) error {
			return NewPoolHandler().Delete(ctx, cmd, args)
		}),
	}
}

type PoolReply struct {
	mapper *idmapper.Mapper
	value  koyeb.ServicePool
	full   bool
}

func NewPoolReply(mapper *idmapper.Mapper, value koyeb.ServicePool, full bool) *PoolReply {
	return &PoolReply{
		mapper: mapper,
		value:  value,
		full:   full,
	}
}

func (PoolReply) Title() string {
	return "Pool"
}

func (r *PoolReply) MarshalBinary() ([]byte, error) {
	return r.value.MarshalJSON()
}

func (r *PoolReply) Headers() []string {
	return []string{"id", "name", "status", "size", "ready_count", "generation", "created_at"}
}

func (r *PoolReply) Fields() []map[string]string {
	fields := map[string]string{
		"id":          renderer.FormatID(r.value.GetId(), r.full),
		"name":        r.value.GetName(),
		"status":      string(r.value.GetStatus()),
		"size":        strconv.FormatInt(r.value.GetSize(), 10),
		"ready_count": strconv.FormatInt(r.value.GetReadyCount(), 10),
		"generation":  r.value.GetGeneration(),
		"created_at":  renderer.FormatTime(r.value.GetCreatedAt()),
	}

	return []map[string]string{fields}
}
