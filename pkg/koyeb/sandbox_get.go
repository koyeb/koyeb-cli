package koyeb

import (
	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/idmapper"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/renderer"
	"github.com/spf13/cobra"
)

// Get retrieves a sandbox by name, short ID, or full UUID and renders its
// status together with its executor connection information.
func (h *SandboxHandler) Get(ctx *CLIContext, cmd *cobra.Command, args []string) error {
	name := args[0]

	service, err := resolveSandboxService(ctx, name)
	if err != nil {
		return err
	}

	info, err := sandboxConnectionInfo(ctx, name, service)
	if err != nil {
		return err
	}

	full := GetBoolFlags(cmd, "full")
	ctx.Renderer.Render(NewGetSandboxReply(ctx.Mapper, *service, info, full))
	return nil
}

// GetSandboxReply implements the renderer.ApiResources interface for a
// single sandbox: the service fields plus the executor connection info.
type GetSandboxReply struct {
	mapper  *idmapper.Mapper
	service koyeb.Service
	info    *SandboxInfo
	full    bool
}

func NewGetSandboxReply(mapper *idmapper.Mapper, service koyeb.Service, info *SandboxInfo, full bool) *GetSandboxReply {
	return &GetSandboxReply{
		mapper:  mapper,
		service: service,
		info:    info,
		full:    full,
	}
}

func (GetSandboxReply) Title() string {
	return "Sandbox"
}

func (r *GetSandboxReply) MarshalBinary() ([]byte, error) {
	return r.service.MarshalJSON()
}

func (r *GetSandboxReply) Headers() []string {
	return []string{"id", "app", "name", "status", "url", "routing_key", "created_at"}
}

func (r *GetSandboxReply) Fields() []map[string]string {
	fields := map[string]string{
		"id":          renderer.FormatID(r.service.GetId(), r.full),
		"app":         renderer.FormatAppName(r.mapper, r.service.GetAppId(), r.full),
		"name":        r.service.GetName(),
		"status":      formatServiceStatus(r.service.GetStatus()),
		"url":         r.info.ConnectionURL(),
		"routing_key": r.info.RoutingKey,
		"created_at":  renderer.FormatTime(r.service.GetCreatedAt()),
	}

	return []map[string]string{fields}
}
