package koyeb

import (
	"fmt"
	"time"

	"github.com/briandowns/spinner"
	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/errors"
	"github.com/spf13/cobra"
)

func NewComposeDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "d",
		Args:  cobra.ExactArgs(1),
		RunE: WithCLIContext(func(ctx *CLIContext, cmd *cobra.Command, args []string) error {
			composeFile, err := parseComposeFile(args[0])
			if err != nil {
				return err
			}

			if composeFile == nil {
				return nil
			}

			appList, _, err := ctx.Client.AppsApi.ListApps(ctx.Context).Name(*composeFile.App.Name).Execute()
			if err != nil {
				return err
			}
			if !appList.HasApps() || len(appList.Apps) == 0 {
				return nil
			}

			if _, _, err := ctx.Client.AppsApi.DeleteApp(ctx.Context, appList.Apps[0].GetId()).Execute(); err != nil {
				return err
			}

			return monitorAppDelete(ctx, appList.Apps[0].GetId())
		}),
	}

	return cmd
}

func monitorAppDelete(ctx *CLIContext, appId string) error {
	s := spinner.New(spinner.CharSets[21], 100*time.Millisecond, spinner.WithColor("red"))
	s.Start()
	defer s.Stop()

	previousStatus := koyeb.AppStatus("")
	for {
		resApp, resp, err := ctx.Client.AppsApi.GetApp(ctx.Context, appId).Execute()
		if err != nil {
			return errors.NewCLIErrorFromAPIError(
				"Error while fetching deployment status",
				err,
				resp,
			)
		}

		currentStatus := resApp.App.GetStatus()
		if previousStatus != currentStatus {
			previousStatus = currentStatus
			s.Suffix = fmt.Sprintf(" Deleting app %s: %s", *resApp.App.Name, currentStatus)
		}

		if isAppMonitoringEndState(currentStatus) {
			break
		}

		time.Sleep(5 * time.Second)
	}

	fmt.Printf("\n")
	if previousStatus == koyeb.APPSTATUS_DELETED {
		fmt.Println("Succcessfully deleted ✅♻️")
	}

	return nil
}
