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
type DatabaseListItemInfo struct {
	Service    koyeb.ServiceListItem `json:"service"`
	Deployment koyeb.Deployment      `json:"deployment"`
}

func (h *DatabaseHandler) List(ctx *CLIContext, cmd *cobra.Command, args []string) error {
	list := []DatabaseListItemInfo{}

	page := int64(0)
	offset := int64(0)
	limit := int64(100)
	for {
		res, resp, err := ctx.Client.ServicesApi.ListServices(ctx.Context).
			Limit(strconv.FormatInt(limit, 10)).
			Offset(strconv.FormatInt(offset, 10)).
			Types([]string{"DATABASE"}).
			Execute()
		if err != nil {
			return errors.NewCLIErrorFromAPIError(
				"Error while listing database services",
				err,
				resp,
			)
		}

		for _, svc := range res.GetServices() {
			res, resp, err := ctx.Client.DeploymentsApi.GetDeployment(ctx.Context, svc.GetLatestDeploymentId()).Execute()
			if err != nil {
				return errors.NewCLIErrorFromAPIError(
					fmt.Sprintf("Error while fetching the deployment for the database service `%s`", svc.GetId()),
					err,
					resp,
				)
			}

			list = append(list, DatabaseListItemInfo{
				Service:    svc,
				Deployment: *res.Deployment,
			})
		}

		page++
		offset = page * limit
		if offset >= res.GetCount() {
			break
		}
	}

	full := GetBoolFlags(cmd, "full")
	listDatabasesReply := NewListDatabasesReply(ctx.Mapper, list, full)
	ctx.Renderer.Render(listDatabasesReply)
	return nil
}

type ListDatabasesReply struct {
	mapper    *idmapper.Mapper
	databases []DatabaseListItemInfo
	full      bool
}

func NewListDatabasesReply(mapper *idmapper.Mapper, databases []DatabaseListItemInfo, full bool) *ListDatabasesReply {
	return &ListDatabasesReply{
		mapper:    mapper,
		databases: databases,
		full:      full,
	}
}

func (ListDatabasesReply) Title() string {
	return "Databases"
}

func (r *ListDatabasesReply) MarshalBinary() ([]byte, error) {
	return json.Marshal(r.databases)
}

func (r *ListDatabasesReply) Headers() []string {
	return []string{"id", "name", "region", "engine", "status", "active_time", "instance", "used_storage", "created_at"}
}

func (r *ListDatabasesReply) Fields() []map[string]string {
	items := r.databases
	resp := make([]map[string]string, 0, len(items))

	for _, item := range items {
		var region, engine, activeTime, instanceType, usedStorage string

		// At the moment, we only support neon postgres so the if statement is
		// always true. This condition is to make sure we don't panic in the
		// future while accessing NeonPostgres if we support other database
		// engines.
		if item.Deployment.DatabaseInfo.HasNeonPostgres() {
			region = item.Deployment.Definition.GetDatabase().NeonPostgres.GetRegion()
			engine = fmt.Sprintf("Postgres %d", item.Deployment.Definition.GetDatabase().NeonPostgres.GetPgVersion())

			instanceType = *item.Deployment.Definition.Database.NeonPostgres.InstanceType

			size, _ := strconv.Atoi(item.Deployment.DatabaseInfo.NeonPostgres.GetDefaultBranchLogicalSize())
			// Convert to MB
			size = size / 1024 / 1024

			activeTimeValue, _ := strconv.ParseFloat(item.Deployment.DatabaseInfo.NeonPostgres.GetActiveTimeSeconds(), 32)

			// Free instances have a maximum active time of 50h and a maximum
			// size of 1Gb, which is not configurable. Other types of instances
			// don't have limits.
			if instanceType == "free" {
				activeTime = fmt.Sprintf("%.1fh/50h", activeTimeValue/60/60)
				usedStorage = fmt.Sprintf("%dMB/1GB", size)
			} else {
				activeTime = fmt.Sprintf("%.1fh", activeTimeValue/60/60)
				usedStorage = fmt.Sprintf("%dMB", size)
			}

		}

		fields := map[string]string{
			"id":           renderer.FormatID(item.Service.GetId(), r.full),
			"name":         item.Service.GetName(),
			"region":       region,
			"engine":       engine,
			"status":       formatServiceStatus(item.Service.GetStatus()),
			"active_time":  activeTime,
			"instance":     instanceType,
			"used_storage": usedStorage,
			"created_at":   renderer.FormatTime(item.Service.GetCreatedAt()),
		}
		resp = append(resp, fields)
	}

	return resp
}
