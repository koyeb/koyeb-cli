package koyeb

import (
	"encoding/json"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/idmapper"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/renderer"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func (h *AppHandler) Get(cmd *cobra.Command, args []string) error {
	res, resp, err := h.client.AppsApi.GetApp(h.ctx, h.ResolveAppArgs(args[0])).Execute()
	if err != nil {
		fatalApiError(err, resp)
	}

	full := GetBoolFlags(cmd, "full")
	output := GetStringFlags(cmd, "output")
	getAppsReply := NewGetAppReply(h.mapper, &res, full)

	return renderer.NewItemRenderer(getAppsReply).Render(output)
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
		"id":         renderer.FormatAppID(r.mapper, item.GetId(), r.full),
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
	if err != nil {
		// We're supposed to marshal a list of strings. If this fails, we can fairly fatal
		log.Fatalf("Could not marshal domain names as json: %+v", err)
	}

	return string(data)
}

func formatAppStatus(status koyeb.AppStatus) string {
	return string(status)
}
