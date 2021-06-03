package koyeb

import (
	"context"
	"fmt"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/spf13/cobra"
)

func NewAppCmd() *cobra.Command {
	h := NewAppHandler()

	appCmd := &cobra.Command{
		Use:     "apps [action]",
		Aliases: []string{"s", "app"},
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

	getAppCmd := &cobra.Command{
		Use:   "get [name]",
		Short: "Get app",
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

	deleteAppCmd := &cobra.Command{
		Use:   "delete [name]",
		Short: "Delete apps",
		Args:  cobra.MinimumNArgs(1),
		RunE:  h.Delete,
	}
	appCmd.AddCommand(deleteAppCmd)

	return appCmd
}

func NewAppHandler() *AppHandler {
	return &AppHandler{}
}

type AppHandler struct {
}

func (h *AppHandler) Create(cmd *cobra.Command, args []string, createApp *koyeb.CreateApp) error {
	client := getApiClient()
	ctx := getAuth(context.Background())

	createApp.SetName(args[0])
	_, _, err := client.AppsApi.CreateApp(ctx).Body(*createApp).Execute()
	if err != nil {
		fatalApiError(err)
	}
	return nil
}

func (h *AppHandler) Update(cmd *cobra.Command, args []string, updateApp *koyeb.UpdateApp) error {
	client := getApiClient()
	ctx := getAuth(context.Background())
	_, _, err := client.AppsApi.UpdateApp2(ctx, args[0]).Body(*updateApp).Execute()
	if err != nil {
		fatalApiError(err)
	}
	return nil
}

func (h *AppHandler) Get(cmd *cobra.Command, args []string) error {
	format := "table"
	if len(args) == 0 {
		return h.listFormat(cmd, args, format)
	}
	return h.getFormat(cmd, args, format)
}

func (h *AppHandler) Describe(cmd *cobra.Command, args []string) error {
	format := "detail"
	if len(args) == 0 {
		return h.listFormat(cmd, args, format)
	}
	return h.getFormat(cmd, args, format)
}

func (h *AppHandler) Delete(cmd *cobra.Command, args []string) error {
	client := getApiClient()
	ctx := getAuth(context.Background())

	for _, arg := range args {
		_, _, err := client.AppsApi.DeleteApp(ctx, arg).Execute()
		if err != nil {
			fatalApiError(err)
		}
	}
	return nil
}

func (h *AppHandler) List(cmd *cobra.Command, args []string) error {
	format := "table"
	return h.listFormat(cmd, args, format)
}

func (h *AppHandler) getFormat(cmd *cobra.Command, args []string, format string) error {
	client := getApiClient()
	ctx := getAuth(context.Background())

	var items []ApiResources
	for _, arg := range args {
		res, _, err := client.AppsApi.GetApp(ctx, arg).Execute()
		if err != nil {
			fatalApiError(err)
		}
		items = append(items, &GetAppReply{res})
	}

	render(format, items)

	return nil
}

func (h *AppHandler) listFormat(cmd *cobra.Command, args []string, format string) error {
	client := getApiClient()
	ctx := getAuth(context.Background())

	var items []ApiResources

	page := 0
	offset := 0
	limit := 10
	for {
		res, _, err := client.AppsApi.ListApps(ctx).Limit(fmt.Sprintf("%d", limit)).Offset(fmt.Sprintf("%d", offset)).Execute()
		if err != nil {
			fatalApiError(err)
		}
		items = append(items, &ListAppsReply{res})
		page += 1
		offset = page * limit
		if int64(offset) >= res.GetCount() {
			break
		}
	}

	render(format, items)

	return nil
}

type GetAppReply struct {
	koyeb.GetAppReply
}

func (a *GetAppReply) MarshalBinary() ([]byte, error) {
	return a.GetAppReply.GetApp().MarshalJSON()
}

func (a *GetAppReply) Title() string {
	return "App"
}

func (a *GetAppReply) Headers() []string {
	return []string{"id", "name", "domains", "updated_at"}
}

func (a *GetAppReply) Fields() []map[string]string {
	res := []map[string]string{}
	item := a.GetApp()
	fields := map[string]string{}
	for _, field := range a.Headers() {
		fields[field] = GetField(item, field)
	}
	res = append(res, fields)
	return res
}

type ListAppsReply struct {
	koyeb.ListAppsReply
}

func (a *ListAppsReply) MarshalBinary() ([]byte, error) {
	return a.ListAppsReply.MarshalJSON()
}

func (a *ListAppsReply) Headers() []string {
	return []string{"id", "name", "domains", "updated_at"}
}

func (a *ListAppsReply) Fields() []map[string]string {
	res := []map[string]string{}
	for _, item := range a.GetApps() {
		fields := map[string]string{}
		for _, field := range a.Headers() {
			fields[field] = GetField(item, field)
		}
		res = append(res, fields)
	}
	return res
}
