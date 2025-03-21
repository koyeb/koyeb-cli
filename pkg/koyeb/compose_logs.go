package koyeb

import (
	"context"
	"sync"

	"github.com/spf13/cobra"
)

func NewComposeLogsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "logs",
		Short: "l",
		Args:  cobra.ExactArgs(1),
		RunE: WithCLIContext(func(ctx *CLIContext, cmd *cobra.Command, args []string) error {
			composeFile, err := parseComposeFile(args[0])
			if err != nil {
				return err
			}

			if composeFile == nil {
				return nil
			}

			appList, _, err := ctx.Client.AppsApi.ListApps(ctx.Context).Name(*composeFile.Name).Execute()
			if err != nil {
				return err
			}
			if !appList.HasApps() {
				return nil
			}

			serviceList, _, err := ctx.Client.ServicesApi.ListServices(ctx.Context).AppId(*appList.Apps[0].Id).Execute()
			if err != nil {
				return err
			}

			var cancel context.CancelFunc
			ctx.Context, cancel = context.WithCancel(ctx.Context)
			defer cancel()

			wg := sync.WaitGroup{}
			wg.Add(1)

			for _, svc := range serviceList.Services {
				lq := LogsQuery{
					ServiceId: svc.GetId(),
					Order:     "asc",
					Tail:      true,
				}
				go ctx.LogsClient.PrintLogs(ctx, lq)
			}

			wg.Wait()

			return nil
		}),
	}

	return cmd
}
