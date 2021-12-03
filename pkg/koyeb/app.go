package koyeb

import (
	"context"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/spf13/cobra"
)

func NewAppCmd() *cobra.Command {
	h := NewAppHandler()

	appCmd := &cobra.Command{
		Use:     "apps [action]",
		Aliases: []string{"a", "app"},
		Short:   "Apps",
	}

	createAppCmd := &cobra.Command{
		Use:   "create [name]",
		Short: "Create apps",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			createApp := koyeb.NewCreateAppWithDefaults()
			SyncFlags(cmd, args, createApp)
			return h.Create(cmd, args, createApp)
		},
	}
	appCmd.AddCommand(createAppCmd)

	initAppCmd := &cobra.Command{
		Use:   "init [name]",
		Short: "Create app and service",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			createApp := koyeb.NewCreateAppWithDefaults()

			createService := koyeb.NewCreateServiceWithDefaults()
			createDef := koyeb.NewServiceDefinitionWithDefaults()

			err := parseServiceDefinitionFlags(cmd.Flags(), createDef, true)
			if err != nil {
				return err
			}
			createDef.Name = koyeb.PtrString(args[0])

			createService.SetDefinition(*createDef)

			return h.Init(cmd, args, createApp, createService)
		},
	}
	appCmd.AddCommand(initAppCmd)
	addServiceDefinitionFlags(initAppCmd.Flags())

	getAppCmd := &cobra.Command{
		Use:   "get [name]",
		Short: "Get app",
		Args:  cobra.ExactArgs(1),
		RunE:  h.Get,
	}
	appCmd.AddCommand(getAppCmd)

	listAppCmd := &cobra.Command{
		Use:   "list",
		Short: "List apps",
		RunE:  h.List,
	}
	appCmd.AddCommand(listAppCmd)

	describeAppCmd := &cobra.Command{
		Use:   "describe [name]",
		Short: "Describe apps",
		Args:  cobra.ExactArgs(1),
		RunE:  h.Describe,
	}
	appCmd.AddCommand(describeAppCmd)

	updateAppCmd := &cobra.Command{
		Use:   "update [name]",
		Short: "Update apps",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			updateApp := koyeb.NewUpdateAppWithDefaults()
			SyncFlags(cmd, args, updateApp)
			return h.Update(cmd, args, updateApp)
		},
	}
	appCmd.AddCommand(updateAppCmd)
	updateAppCmd.Flags().StringP("name", "n", "", "Name of the app")

	deleteAppCmd := &cobra.Command{
		Use:   "delete [name]",
		Short: "Delete apps",
		Args:  cobra.MinimumNArgs(1),
		RunE:  h.Delete,
	}
	appCmd.AddCommand(deleteAppCmd)
	deleteAppCmd.Flags().BoolP("force", "f", false, "Force delete app and services")

	return appCmd
}

func NewAppHandler() *AppHandler {
	return &AppHandler{}
}

type AppHandler struct {
}

func (h *AppHandler) Init(cmd *cobra.Command, args []string, createApp *koyeb.CreateApp, createService *koyeb.CreateService) error {
	client := getApiClient()
	ctx := getAuth(context.Background())

	_, _, err := client.ServicesApi.CreateService(ctx, args[0]).DryRun(true).Body(*createService).Execute()
	if err != nil {
		fatalApiError(err)
	}

	createApp.SetName(args[0])
	_, _, err = client.AppsApi.CreateApp(ctx).Body(*createApp).Execute()
	if err != nil {
		fatalApiError(err)
	}

	_, _, err = client.ServicesApi.CreateService(ctx, args[0]).Body(*createService).Execute()
	if err != nil {
		fatalApiError(err)
	}

	return nil
}
