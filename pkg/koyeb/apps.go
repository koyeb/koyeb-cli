package koyeb

import (
	"fmt"
	"time"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func NewAppCmd() *cobra.Command {
	h := NewAppHandler()
	serviceHandler := NewServiceHandler()

	appCmd := &cobra.Command{
		Use:     "apps ACTION",
		Aliases: []string{"a", "app"},
		Short:   "Apps",
	}

	createAppCmd := &cobra.Command{
		Use:   "create NAME",
		Short: "Create app",
		Args:  cobra.ExactArgs(1),
		RunE: WithCLIContext(func(ctx *CLIContext, cmd *cobra.Command, args []string) error {
			createApp := koyeb.NewCreateAppWithDefaults()
			SyncFlags(cmd, args, createApp)

			// Parse and set lifecycle
			lifecycle := h.parseAppLifeCycle(cmd.Flags(), nil)
			if lifecycle != nil {
				createApp.SetLifeCycle(*lifecycle)
			}

			return h.Create(ctx, cmd, args, createApp)
		}),
	}
	createAppCmd.Flags().Bool("delete-when-empty", false, "Automatically delete the app after the last service is deleted. Empty apps created without services are not deleted.")
	appCmd.AddCommand(createAppCmd)

	initAppCmd := &cobra.Command{
		Use:     "init NAME",
		Short:   "Create app and service",
		Example: "See examples of koyeb service create --help",
		Args:    cobra.ExactArgs(1),
		RunE: WithCLIContext(func(ctx *CLIContext, cmd *cobra.Command, args []string) error {
			createApp := koyeb.NewCreateAppWithDefaults()

			createService := koyeb.NewCreateServiceWithDefaults()
			createDefinition := koyeb.NewDeploymentDefinitionWithDefaults()

			err := serviceHandler.parseServiceDefinitionFlags(ctx, cmd.Flags(), createDefinition)
			if err != nil {
				return err
			}
			createDefinition.Name = koyeb.PtrString(args[0])

			createService.SetDefinition(*createDefinition)

			return h.Init(ctx, cmd, args, createApp, createService)
		}),
	}
	initAppCmd.Flags().Bool("wait", false, "Waits until app deployment is done")
	initAppCmd.Flags().Duration("wait-timeout", 5*time.Minute, "Duration the wait will last until timeout")
	appCmd.AddCommand(initAppCmd)
	serviceHandler.addServiceDefinitionFlags(initAppCmd.Flags())

	getAppCmd := &cobra.Command{
		Use:   "get NAME",
		Short: "Get app",
		Args:  cobra.ExactArgs(1),
		RunE:  WithCLIContext(h.Get),
	}
	appCmd.AddCommand(getAppCmd)

	listAppCmd := &cobra.Command{
		Use:   "list",
		Short: "List apps",
		RunE:  WithCLIContext(h.List),
	}
	appCmd.AddCommand(listAppCmd)

	describeAppCmd := &cobra.Command{
		Use:   "describe NAME",
		Short: "Describe app",
		Args:  cobra.ExactArgs(1),
		RunE:  WithCLIContext(h.Describe),
	}
	appCmd.AddCommand(describeAppCmd)

	updateAppCmd := &cobra.Command{
		Use:   "update NAME",
		Short: "Update app",
		Args:  cobra.ExactArgs(1),
		RunE: WithCLIContext(func(ctx *CLIContext, cmd *cobra.Command, args []string) error {
			updateApp := koyeb.NewUpdateAppWithDefaults()
			SyncFlags(cmd, args, updateApp)

			// Get current app to access lifecycle
			appId, err := h.ResolveAppArgs(ctx, args[0])
			if err != nil {
				return err
			}

			currentApp, resp, err := ctx.Client.AppsApi.GetApp(ctx.Context, appId).Execute()
			if err != nil {
				return errors.NewCLIErrorFromAPIError(
					fmt.Sprintf("Error while fetching app `%s`", args[0]),
					err,
					resp,
				)
			}

			// Parse and set lifecycle
			var currentLifeCycle *koyeb.AppLifeCycle
			if currentApp.App.HasLifeCycle() {
				lc := currentApp.App.GetLifeCycle()
				currentLifeCycle = &lc
			}
			lifecycle := h.parseAppLifeCycle(cmd.Flags(), currentLifeCycle)
			if lifecycle != nil {
				updateApp.SetLifeCycle(*lifecycle)
			}

			return h.Update(ctx, cmd, args, updateApp)
		}),
	}
	updateAppCmd.Flags().StringP("name", "n", "", "Change the name of the app")
	updateAppCmd.Flags().StringP("domain", "D", "", "Change the subdomain of the app (only specify the subdomain, skipping \".koyeb.app\")")
	updateAppCmd.Flags().Bool("delete-when-empty", false, "Automatically delete the app after the last service is deleted. Empty apps created without services are not deleted.")
	appCmd.AddCommand(updateAppCmd)

	deleteAppCmd := &cobra.Command{
		Use:   "delete NAME",
		Short: "Delete app",
		Args:  cobra.ExactArgs(1),
		RunE:  WithCLIContext(h.Delete),
	}
	appCmd.AddCommand(deleteAppCmd)

	pauseServiceCmd := &cobra.Command{
		Use:   "pause NAME",
		Short: "Pause app",
		Args:  cobra.ExactArgs(1),
		RunE:  WithCLIContext(h.Pause),
	}
	appCmd.AddCommand(pauseServiceCmd)

	resumeServiceCmd := &cobra.Command{
		Use:   "resume NAME",
		Short: "Resume app",
		Args:  cobra.ExactArgs(1),
		RunE:  WithCLIContext(h.Resume),
	}
	appCmd.AddCommand(resumeServiceCmd)

	return appCmd
}

func NewAppHandler() *AppHandler {
	return &AppHandler{}
}

type AppHandler struct {
}

func (h *AppHandler) ResolveAppArgs(ctx *CLIContext, val string) (string, error) {
	appMapper := ctx.Mapper.App()
	id, err := appMapper.ResolveID(val)
	if err != nil {
		return "", err
	}
	return id, nil
}

// Parse --delete-when-empty flag
// Automatically delete the app after the last service is deleted.
// Empty apps created without services are not deleted.
func (h *AppHandler) parseAppLifeCycle(flags *pflag.FlagSet, currentLifeCycle *koyeb.AppLifeCycle) *koyeb.AppLifeCycle {
	var lifecycle *koyeb.AppLifeCycle

	if currentLifeCycle != nil {
		lifecycle = currentLifeCycle
	} else if flags.Lookup("delete-when-empty").Changed {
		lifecycle = koyeb.NewAppLifeCycleWithDefaults()
	} else {
		return nil
	}

	if flags.Lookup("delete-when-empty").Changed {
		deleteWhenEmpty, _ := flags.GetBool("delete-when-empty")
		lifecycle.SetDeleteWhenEmpty(deleteWhenEmpty)
	}

	return lifecycle
}
