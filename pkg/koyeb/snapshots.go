package koyeb

import (
	"github.com/spf13/cobra"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
)

func NewSnapshotCmd() *cobra.Command {
	h := NewSnapshotHandler()
	_ = h

	snapshotCmd := &cobra.Command{
		Use:     "snapshots ACTION",
		Aliases: []string{"vol", "snapshot"},
		Short:   "Manage snapshots",
	}

	createSnapshotCmd := &cobra.Command{
		Use:   "create NAME",
		Short: "Create a new snapshot",
		Args:  cobra.ExactArgs(2),
		RunE: WithCLIContext(func(ctx *CLIContext, cmd *cobra.Command, args []string) error {
			req := koyeb.NewCreateSnapshotRequestWithDefaults()

			parentVolumeID, err := ctx.Mapper.Volume().ResolveID(args[1])
			if err != nil {
				return err
			}

			req.SetName(args[0])
			req.SetParentVolumeId(parentVolumeID)

			return h.Create(ctx, cmd, args, req)
		}),
	}
	createSnapshotCmd.Flags().String("region", "was", "Region of the snapshot")
	createSnapshotCmd.Flags().Int64("size", 10, "Size of the snapshot in GB")
	createSnapshotCmd.Flags().Bool("read-only", false, "Force the snapshot to be read-only")
	snapshotCmd.AddCommand(createSnapshotCmd)

	getSnapshotCmd := &cobra.Command{
		Use:   "get NAME",
		Short: "Get a snapshot",
		Args:  cobra.ExactArgs(1),
		RunE: WithCLIContext(func(ctx *CLIContext, cmd *cobra.Command, args []string) error {
			return h.Get(ctx, cmd, args)
		}),
	}
	snapshotCmd.AddCommand(getSnapshotCmd)

	listSnapshotCmd := &cobra.Command{
		Use:   "list",
		Short: "List snapshots",
		RunE: WithCLIContext(func(ctx *CLIContext, cmd *cobra.Command, args []string) error {
			return h.List(ctx, cmd, args)
		}),
	}
	snapshotCmd.AddCommand(listSnapshotCmd)

	updateSnapshotCmd := &cobra.Command{
		Use:   "update NAME",
		Short: "Update a snapshot",
		Args:  cobra.ExactArgs(1),
		RunE: WithCLIContext(func(ctx *CLIContext, cmd *cobra.Command, args []string) error {
			req := koyeb.NewUpdateSnapshotRequestWithDefaults()

			name, _ := cmd.Flags().GetString("name")
			if name != "" {
				req.SetName(name)
			}

			return h.Update(ctx, cmd, args, req)
		}),
	}
	updateSnapshotCmd.Flags().String("name", "", "Change the snapshot name")
	updateSnapshotCmd.Flags().Int64("size", -1, "Increase the snapshot size")
	snapshotCmd.AddCommand(updateSnapshotCmd)

	deleteSnapshotCmd := &cobra.Command{
		Use:   "delete NAME",
		Short: "Delete a snapshot",
		Args:  cobra.ExactArgs(1),
		RunE: WithCLIContext(func(ctx *CLIContext, cmd *cobra.Command, args []string) error {
			return h.Delete(ctx, cmd, args)
		}),
	}
	snapshotCmd.AddCommand(deleteSnapshotCmd)

	return snapshotCmd
}

func NewSnapshotHandler() *SnapshotHandler {
	return &SnapshotHandler{}
}

type SnapshotHandler struct {
}

func ResolveSnapshotArgs(ctx *CLIContext, val string) (string, error) {
	snapshotMapper := ctx.Mapper.Snapshot()
	id, err := snapshotMapper.ResolveID(val)
	if err != nil {
		return "", err
	}
	return id, nil
}
