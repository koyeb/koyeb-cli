package koyeb

import (
	"encoding/json"
	"fmt"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/errors"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/idmapper"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/renderer"
	"github.com/spf13/cobra"
)

func (h *AppHandler) Get(ctx *CLIContext, cmd *cobra.Command, args []string) error {
	app, err := h.ResolveAppArgs(ctx, args[0])
	if err != nil {
		return err
	}

	res, resp, err := ctx.Client.AppsApi.GetApp(ctx.Context, app).Execute()
	if err != nil {
		return errors.NewCLIErrorFromAPIError(
			fmt.Sprintf("Error while retrieving the application `%s`", args[0]),
			err,
			resp,
		)
	}

	full := GetBoolFlags(cmd, "full")
	getAppsReply := NewGetAppReply(ctx.Mapper, res, full)
	ctx.Renderer.Render(getAppsReply)
	return nil
}

type GetAppReply struct {
	mapper *idmapper.Mapper
	value  *koyeb.GetAppReply
	full   bool
}

func NewGetAppReply(mapper *idmapper.Mapper, value *koyeb.GetAppReply, full bool) *GetAppReply {
	return &GetAppReply{
		mapper: mapper,
		value:  value,
		full:   full,
	}
}

func (GetAppReply) Title() string {
	return "App"
}

func (r *GetAppReply) MarshalBinary() ([]byte, error) {
	return r.value.GetApp().MarshalJSON()
}

func (r *GetAppReply) Headers() []string {
	return []string{"id", "name", "status", "domains", "created_at"}
}

func (r *GetAppReply) Fields() []map[string]string {
	item := r.value.GetApp()
	fields := map[string]string{
		"id":         renderer.FormatID(item.GetId(), r.full),
		"name":       item.GetName(),
		"status":     formatAppStatus(item.GetStatus()),
		"domains":    formatDomains(item.GetDomains(), 80),
		"created_at": renderer.FormatTime(item.GetCreatedAt()),
	}

	resp := []map[string]string{fields}
	return resp
}

func formatDomains(domains []koyeb.Domain, max int) string {
	domainNames := []string{}
	totalLen := 0
	for _, d := range domains {
		name := d.GetName()
		if max > 0 && totalLen+len(name) >= max {
			domainNames = append(domainNames, "...")
			break
		}

		domainNames = append(domainNames, name)
		totalLen += len(name)
	}

	data, err := json.Marshal(domainNames)
	// Should never happen, as we are marshalling a list of strings
	if err != nil {
		panic("failed to marshal domains")
	}

	return string(data)
}

func formatAppStatus(status koyeb.AppStatus) string {
	return string(status)
}
