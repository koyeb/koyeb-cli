package koyeb

import (
	"strconv"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/errors"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/idmapper"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/renderer"
	"github.com/spf13/cobra"
)

// List lists all sandbox services
func (h *SandboxHandler) List(ctx *CLIContext, cmd *cobra.Command, args []string) error {
	list := []koyeb.ServiceListItem{}

	// Parse flags once before the loop
	appID := GetStringFlags(cmd, "app")
	name := GetStringFlags(cmd, "name")

	var resolvedAppID string
	if appID != "" {
		appMapper := ctx.Mapper.App()
		var err error
		resolvedAppID, err = appMapper.ResolveID(appID)
		if err != nil {
			return &errors.CLIError{
				What:       "Error while listing sandboxes",
				Why:        "could not resolve application name",
				Additional: nil,
				Orig:       err,
				Solution:   "Make sure the application exists. Use 'koyeb app list' to see available applications.",
			}
		}
	}

	page := int64(0)
	offset := int64(0)
	limit := int64(100)
	for {
		req := ctx.Client.ServicesApi.ListServices(ctx.Context).
			Types([]string{"SANDBOX"})

		if resolvedAppID != "" {
			req = req.AppId(resolvedAppID)
		}
		if name != "" {
			req = req.Name(name)
		}

		res, resp, err := req.Limit(strconv.FormatInt(limit, 10)).Offset(strconv.FormatInt(offset, 10)).Execute()
		if err != nil {
			return errors.NewCLIErrorFromAPIError("Error while listing sandboxes", err, resp)
		}

		list = append(list, res.GetServices()...)

		page++
		offset = page * limit
		if offset >= res.GetCount() {
			break
		}
	}

	full := GetBoolFlags(cmd, "full")
	reply := NewListSandboxesReply(ctx.Mapper, list, full)
	ctx.Renderer.Render(reply)
	return nil
}

// ListSandboxesReply implements the renderer.ApiResources interface for sandbox listing
type ListSandboxesReply struct {
	mapper    *idmapper.Mapper
	sandboxes []koyeb.ServiceListItem
	full      bool
}

func NewListSandboxesReply(mapper *idmapper.Mapper, sandboxes []koyeb.ServiceListItem, full bool) *ListSandboxesReply {
	return &ListSandboxesReply{
		mapper:    mapper,
		sandboxes: sandboxes,
		full:      full,
	}
}

func (ListSandboxesReply) Title() string {
	return "Sandboxes"
}

func (r *ListSandboxesReply) MarshalBinary() ([]byte, error) {
	return koyeb.NewNullableListServicesReply(&koyeb.ListServicesReply{Services: r.sandboxes}).MarshalJSON()
}

func (r *ListSandboxesReply) Headers() []string {
	return []string{"id", "app", "name", "status", "created_at"}
}

func (r *ListSandboxesReply) Fields() []map[string]string {
	resp := make([]map[string]string, 0, len(r.sandboxes))

	for _, item := range r.sandboxes {
		fields := map[string]string{
			"id":         renderer.FormatID(item.GetId(), r.full),
			"app":        renderer.FormatAppName(r.mapper, item.GetAppId(), r.full),
			"name":       item.GetName(),
			"status":     formatServiceStatus(item.GetStatus()),
			"created_at": renderer.FormatTime(item.GetCreatedAt()),
		}
		resp = append(resp, fields)
	}

	return resp
}
