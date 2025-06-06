package koyeb

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/errors"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/idmapper"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/renderer"
	"github.com/spf13/cobra"
)

// DatabaseListItemInfo wraps a service returned by the services API and it's latest deployment.
type DatabaseInfo struct {
	Service           koyeb.Service    `json:"service"`
	Deployment        koyeb.Deployment `json:"deployment"`
	ConnectionStrings []string
}

// Given a koyeb.DatabaseSource (stored in a koyeb.DeploymentDefinition object), return the list of roles.
func getConnectionStrings(ctx *CLIContext, dbSource koyeb.DatabaseSource, dbInfo koyeb.DeploymentDatabaseInfo) ([]string, error) {
	// At the moment, we only support neon postgres so the if statement is always true.
	neon, hasNeon := dbInfo.GetNeonPostgresOk()
	if !hasNeon {
		return nil, nil
	}

	connectionStrings := []string{}
	for _, role := range neon.Roles {
		body := make(map[string]interface{})
		res, resp, err := ctx.Client.SecretsApi.RevealSecret(ctx.Context, role.GetSecretId()).Body(body).Execute()
		if err != nil {
			return nil, errors.NewCLIErrorFromAPIError(
				fmt.Sprintf("Error while revealing the secret `%s` for the database role `%s`", role.GetSecretId(), role.GetName()),
				err,
				resp,
			)
		}

		formatError := &errors.CLIError{
			What:     fmt.Sprintf("Unable to get the password for the database role %s", role.GetName()),
			Why:      "the response format cannot be understood by the CLI",
			Orig:     nil,
			Solution: "Try to update the CLI to the latest version. If the problem persists, please create an issue on https://github.com/koyeb/koyeb-cli/issues/new",
		}

		username, ok := res.Value["username"]
		if !ok {
			return nil, formatError
		}
		password, ok := res.Value["password"]
		if !ok {
			return nil, formatError
		}

		for _, db := range dbSource.GetNeonPostgres().Databases {
			// host is empty when the database is not yet fully provisioned
			if neon.GetServerHost() != "" {
				s := fmt.Sprintf(
					"postgres://%s:%s@%s:%d/%s",
					username,
					password,
					neon.GetServerHost(),
					neon.GetServerPort(),
					*db.Name,
				)
				connectionStrings = append(connectionStrings, s)
			}
		}
	}
	return connectionStrings, nil
}

func (h *DatabaseHandler) Get(ctx *CLIContext, cmd *cobra.Command, args []string) error {
	serviceHandler := NewServiceHandler()

	serviceName, err := serviceHandler.parseServiceName(cmd, args[0])
	if err != nil {
		return err
	}

	database, err := h.ResolveDatabaseArgs(ctx, serviceName)
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

	connectionStrings, err := getConnectionStrings(
		ctx,
		resDeployment.Deployment.Definition.GetDatabase(),
		resDeployment.Deployment.GetDatabaseInfo(),
	)
	if err != nil {
		return err
	}

	full := GetBoolFlags(cmd, "full")
	getDatabaseReply := NewGetDatabaseReply(
		ctx.Mapper,
		DatabaseInfo{
			Service:           resService.GetService(),
			Deployment:        resDeployment.GetDeployment(),
			ConnectionStrings: connectionStrings,
		},
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
	return []string{"id", "name", "region", "engine", "status", "compute_time", "instance", "used_storage", "created_at", "connection_strings"}
}

func (r *GetDatabaseReply) Fields() []map[string]string {
	var region, engine, computeTime, instanceType, usedStorage string

	// At the moment, we only support neon postgres so the if statement is
	// always true. If we add support for other providers in the future, the statement
	// will prevent a nil pointer dereference.
	if r.value.Deployment.DatabaseInfo.HasNeonPostgres() {
		region = r.value.Deployment.Definition.GetDatabase().NeonPostgres.GetRegion()
		engine = fmt.Sprintf("Postgres %d", r.value.Deployment.Definition.GetDatabase().NeonPostgres.GetPgVersion())

		instanceType = *r.value.Deployment.Definition.Database.NeonPostgres.InstanceType

		size, _ := strconv.Atoi(r.value.Deployment.DatabaseInfo.NeonPostgres.GetDefaultBranchLogicalSize())
		// Convert to MB
		size = size / 1024 / 1024

		computeTimeValue, _ := strconv.ParseFloat(r.value.Deployment.DatabaseInfo.NeonPostgres.GetComputeTimeSeconds(), 32)

		// Free instances have a maximum compute time of 12.5h and a maximum
		// size of 1Gb, which is not configurable. Other types of instances
		// don't have limits.
		if instanceType == "free" {
			computeTime = fmt.Sprintf("%.1fh/12.5h", computeTimeValue/60/60)
			usedStorage = fmt.Sprintf("%dMB/1GB", size)
		} else {
			computeTime = fmt.Sprintf("%.1fh", computeTimeValue/60/60)
			usedStorage = fmt.Sprintf("%dMB", size)
		}
	}

	fields := map[string]string{
		"id":                 renderer.FormatID(r.value.Service.GetId(), r.full),
		"name":               r.value.Service.GetName(),
		"region":             region,
		"engine":             engine,
		"status":             formatServiceStatus(r.value.Service.GetStatus()),
		"compute_time":       computeTime,
		"instance":           instanceType,
		"used_storage":       usedStorage,
		"created_at":         renderer.FormatTime(r.value.Service.GetCreatedAt()),
		"connection_strings": strings.Join(r.value.ConnectionStrings, "\n"),
	}
	return []map[string]string{fields}
}
