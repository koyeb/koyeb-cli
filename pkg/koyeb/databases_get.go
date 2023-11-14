package koyeb

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/errors"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/idmapper"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/renderer"
	"github.com/spf13/cobra"
)

// DatabaseListItemInfo wraps a service returned by the services API and it's latest deployment.
type DatabaseInfo struct {
	Service    koyeb.Service    `json:"service"`
	Deployment koyeb.Deployment `json:"deployment"`
}

func (h *DatabaseHandler) Get(ctx *CLIContext, cmd *cobra.Command, args []string) error {
	database, err := h.ResolveDatabaseArgs(ctx, args[0])
	if err != nil {
		return err
	}

	resService, resp, err := ctx.Client.ServicesApi.GetService(ctx.Context, database).Execute()
	if err != nil {
		return errors.NewCLIErrorFromAPIError(
			fmt.Sprintf("Error while retrieving the database `%s`", args[0]),
			err,
			resp,
		)
	}

	resDeployment, resp, err := ctx.Client.DeploymentsApi.GetDeployment(ctx.Context, resService.Service.GetLatestDeploymentId()).Execute()
	if err != nil {
		return errors.NewCLIErrorFromAPIError(
			fmt.Sprintf("Error while fetching the deployment for the database service `%s`", resService.Service.GetId()),
			err,
			resp,
		)
	}

	full := GetBoolFlags(cmd, "full")
	getDatabaseReply := NewGetDatabaseReply(
		ctx.Mapper,
		DatabaseInfo{Service: resService.GetService(), Deployment: resDeployment.GetDeployment()},
		full,
	)
	ctx.Renderer.Render(getDatabaseReply)
	return nil
}

type GetDatabaseReply struct {
	mapper *idmapper.Mapper
	value  DatabaseInfo
	full   bool
}

func NewGetDatabaseReply(mapper *idmapper.Mapper, value DatabaseInfo, full bool) *GetDatabaseReply {
	return &GetDatabaseReply{
		mapper: mapper,
		value:  value,
		full:   full,
	}
}

func (GetDatabaseReply) Title() string {
	return "Database"
}

func (r *GetDatabaseReply) MarshalBinary() ([]byte, error) {
	return json.Marshal(r.value)
}

func (r *GetDatabaseReply) Headers() []string {
	return []string{"id", "name", "region", "engine", "status", "active_time", "used_storage", "created_at"}
}

func (r *GetDatabaseReply) Fields() []map[string]string {
	var region, engine, activeTime, usedStorage string

	// At the moment, we only support neon postgres so the if statement is
	// always true. If we add support for other providers in the future, the statement
	// will prevent a nil pointer dereference.
	if r.value.Deployment.DatabaseInfo.HasNeonPostgres() {
		region = r.value.Deployment.Definition.GetDatabase().NeonPostgres.GetRegion()
		engine = fmt.Sprintf("Postgres %d", r.value.Deployment.Definition.GetDatabase().NeonPostgres.GetPgVersion())

		size, _ := strconv.Atoi(r.value.Deployment.DatabaseInfo.NeonPostgres.GetDefaultBranchLogicalSize())
		// Convert to MB
		size = size / 1024 / 1024
		// The maximum size is 3GB and is not configurable yet.
		maxSize := 3 * 1024
		activeTimeValue, _ := strconv.ParseFloat(r.value.Deployment.DatabaseInfo.NeonPostgres.GetActiveTimeSeconds(), 32)
		// The maximum active time is 100h and is not configurable yet.
		activeTime = fmt.Sprintf("%.1fh/100h", activeTimeValue/60/60)
		usedStorage = fmt.Sprintf("%dMB/%dMB (%d%%)", size, maxSize, size*100/maxSize)
	}

	fields := map[string]string{
		"id":           renderer.FormatID(r.value.Service.GetId(), r.full),
		"name":         r.value.Service.GetName(),
		"region":       region,
		"engine":       engine,
		"status":       formatServiceStatus(r.value.Service.GetStatus()),
		"active_time":  activeTime,
		"used_storage": usedStorage,
		"created_at":   renderer.FormatTime(r.value.Service.GetCreatedAt()),
	}
	return []map[string]string{fields}
}
