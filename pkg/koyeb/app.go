package koyeb

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/logrusorgru/aurora"
	log "github.com/sirupsen/logrus"
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

func (h *AppHandler) Update(cmd *cobra.Command, args []string, updateApp *koyeb.UpdateApp) error {
	format := getFormat("table")
	client := getApiClient()
	ctx := getAuth(context.Background())
	_, _, err := client.AppsApi.UpdateApp2(ctx, args[0]).Body(*updateApp).Execute()
	if err != nil {
		fatalApiError(err)
	}
	return h.getFormat(cmd, args, format)
}

func (h *AppHandler) Delete(cmd *cobra.Command, args []string) error {
	client := getApiClient()
	ctx := getAuth(context.Background())

	force, _ := cmd.Flags().GetBool("force")
	for _, arg := range args {
		if force {
			for {
				res, _, err := client.ServicesApi.DeprecatedListServices(ctx, arg).Limit("100").Execute()
				if err != nil {
					fatalApiError(err)
				}
				if res.GetCount() == 0 {
					break
				}
				for _, svc := range res.GetServices() {
					if svc.State.GetStatus() == "STOPPING" || svc.State.GetStatus() == "STOPPED" {
						continue
					}
					_, _, err := client.ServicesApi.DeleteService(ctx, arg, svc.GetId()).Execute()
					if err != nil {
						fatalApiError(err)
					}
				}
				time.Sleep(2 * time.Second)
			}
		}
		_, _, err := client.AppsApi.DeleteApp(ctx, arg).Execute()
		if err != nil {
			fatalApiError(err)
		}
	}

	log.Infof("Apps %s deleted.", strings.Join(args, ", "))
	return nil
}

func (h *AppHandler) getFormat(cmd *cobra.Command, args []string, format string) error {
	client := getApiClient()
	ctx := getAuth(context.Background())

	for _, arg := range args {
		res, _, err := client.AppsApi.GetApp(ctx, arg).Execute()
		if err != nil {
			fatalApiError(err)
		}
		render(format, &GetAppReply{res: &res})
		if format == "detail" {
			res, _, err := client.ServicesApi.ListServices(ctx).AppId(res.App.GetId()).Limit("100").Execute()
			if err != nil {
				fatalApiError(err)
			}
			rend := &ListServicesReply{res}
			fmt.Printf("\n%s\n", aurora.Bold(rend.Title()))
			render(getFormat("table"), rend)
		}
	}

	return nil
}
