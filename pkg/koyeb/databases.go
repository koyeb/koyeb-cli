package koyeb

import (
	"fmt"

	"github.com/gofrs/uuid"
	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// DatabaseAppName is the name of the database app that is created when the user
// creates a database. This name is hard-coded, and has the same value as the
// one used in the console.
const DatabaseAppName = "koyeb-db-preview-app"

func NewDatabaseCmd() *cobra.Command {
	h := NewDatabaseHandler()

	databaseCmd := &cobra.Command{
		Use:     "databases ACTION",
		Aliases: []string{"db", "database"},
		Short:   "Databases",
	}

	listDbCmd := &cobra.Command{
		Use:   "list",
		Short: "List databases",
		RunE:  WithCLIContext(h.List),
	}
	databaseCmd.AddCommand(listDbCmd)

	getDbCmd := &cobra.Command{
		Use:   "get NAME",
		Short: "Get database",
		Args:  cobra.ExactArgs(1),
		RunE:  WithCLIContext(h.Get),
	}
	databaseCmd.AddCommand(getDbCmd)

	createDbCmd := &cobra.Command{
		Use:   "create NAME",
		Short: "Create database",
		Args:  cobra.ExactArgs(1),
		RunE: WithCLIContext(func(ctx *CLIContext, cmd *cobra.Command, args []string) error {
			createService := koyeb.NewCreateServiceWithDefaults()
			createDefinition := koyeb.NewDeploymentDefinitionWithDefaults()
			serviceName := args[0]

			if err := parseDbServiceDefinitionFlags(cmd.Flags(), serviceName, createDefinition); err != nil {
				return err
			}

			createDefinition.Name = koyeb.PtrString(serviceName)
			createService.SetDefinition(*createDefinition)
			return h.Create(ctx, cmd, args, createService)
		}),
	}
	addCreateDbServiceDefinitionFlags(createDbCmd.Flags())
	databaseCmd.AddCommand(createDbCmd)

	updateDbCmd := &cobra.Command{
		Use:   "update NAME",
		Short: "Update database",
		Args:  cobra.ExactArgs(1),
		RunE: WithCLIContext(func(ctx *CLIContext, cmd *cobra.Command, args []string) error {
			service, err := h.ResolveDatabaseArgs(ctx, args[0])
			if err != nil {
				return err
			}

			latestDeployment, resp, err := ctx.Client.DeploymentsApi.
				ListDeployments(ctx.Context).
				Limit("1").
				ServiceId(service).
				Execute()
			if err != nil {
				return errors.NewCLIErrorFromAPIError(
					fmt.Sprintf("Error while updating the service `%s`", args[0]),
					err,
					resp,
				)
			}
			if len(latestDeployment.GetDeployments()) == 0 {
				return &errors.CLIError{
					What: "Error while updating the database",
					Why:  "we couldn't find the latest deployment of your database",
					Additional: []string{
						"When you create a database for the first time, it can take a few seconds for the first deployment to be created.",
						"We need to fetch the configuration of this latest deployment to update your database.",
					},
					Orig:     nil,
					Solution: "Try again in a few seconds. If the problem persists, please create an issue on https://github.com/koyeb/koyeb-cli/issues/new",
				}
			}

			updateDef := latestDeployment.GetDeployments()[0].GetDefinition()

			if err := parseDbServiceDefinitionFlags(cmd.Flags(), updateDef.GetName(), &updateDef); err != nil {
				return err
			}

			updateService := koyeb.NewUpdateServiceWithDefaults()
			updateService.SetDefinition(updateDef)

			return h.Update(ctx, cmd, args, service, updateService)
		}),
	}
	addUpdateDbServiceDefinitionFlags(updateDbCmd.Flags())
	databaseCmd.AddCommand(updateDbCmd)

	deleteDbCmd := &cobra.Command{
		Use:   "delete NAME",
		Short: "Delete database",
		Args:  cobra.ExactArgs(1),
		RunE:  WithCLIContext(h.Delete),
	}
	databaseCmd.AddCommand(deleteDbCmd)

	return databaseCmd
}

func addUpdateDbServiceDefinitionFlags(flags *pflag.FlagSet) {
	flags.String("instance-type", "free", "Instance type (free, small, medium or large)")
}

// All the flags to create a database include the flags to update a database, plus more.
func addCreateDbServiceDefinitionFlags(flags *pflag.FlagSet) {
	addUpdateDbServiceDefinitionFlags(flags)

	flags.Int64("pg-version", 16, "PostgreSQL version")
	flags.String("region", "fra", "Region where the database is deployed")
	flags.String("db-name", "koyebdb", "Database name")
	flags.String("db-owner", "koyeb-adm", "Database owner")
}

// parseDbServiceDefinitionFlags parses the flags to update the deployment definition.
// The deployment definition can be represented as:
//
//	{
//	  "definition": {
//	    "type": "DATABASE",
//	    "name": "<service name>",
//	    "database": {
//	      "neon_postgres": {
//	        "pg_version": "<postgres version>",
//	        "region": "<region>",
//	        "instance_type": "<instance type>",
//	        "roles": [
//	          {
//	            "name": "<role name>",
//	            "secret": "<secret name>"
//	          }
//	        ],
//	        "databases": [
//	          {
//	            "name": "<database name>",
//	            "owner": "<role name>"
//	          }
//	        ]
//	      }
//	    }
//	  }
//	}
func parseDbServiceDefinitionFlags(flags *pflag.FlagSet, serviceName string, definition *koyeb.DeploymentDefinition) error {
	definition.SetType(koyeb.DEPLOYMENTDEFINITIONTYPE_DATABASE)
	definition.SetName(serviceName)

	flagChanged := func(name string) bool {
		f := flags.Lookup(name)
		if f == nil {
			return false
		}
		return f.Changed
	}

	if !definition.HasDatabase() {
		definition.SetDatabase(*koyeb.NewDatabaseSourceWithDefaults())
	}
	if !definition.Database.HasNeonPostgres() {
		definition.Database.SetNeonPostgres(*koyeb.NewNeonPostgresDatabaseWithDefaults())
	}

	if definition.Database.NeonPostgres.PgVersion == nil || flagChanged("pg-version") {
		version, _ := flags.GetInt64("pg-version")
		definition.Database.NeonPostgres.SetPgVersion(version)
	}
	if definition.Database.NeonPostgres.Region == nil || flagChanged("region") {
		region, _ := flags.GetString("region")
		definition.Database.NeonPostgres.SetRegion(region)
	}
	if definition.Database.NeonPostgres.InstanceType == nil || flagChanged("instance-type") {
		instanceType, _ := flags.GetString("instance-type")
		definition.Database.NeonPostgres.SetInstanceType(instanceType)
	}

	if len(definition.Database.NeonPostgres.Roles) == 0 || flagChanged("db-owner") {
		if len(definition.Database.NeonPostgres.Roles) > 1 {
			return &errors.CLIError{
				What: "Error while updating the database service definition",
				Why:  "the CLI does not support updating the database service definition with multiple roles",
				Additional: []string{
					"The Koyeb API supports multiple roles for a database, but the CLI does not support it yet.",
				},
				Orig:     nil,
				Solution: "Try upgrading the CLI to the latest version. If you are already using the latest version, please create an issue on https://github.com/koyeb/koyeb-cli/issues/new and request support for multiple roles.",
			}
		}

		if len(definition.Database.NeonPostgres.Roles) == 0 {
			role := koyeb.NewNeonPostgresDatabaseNeonRoleWithDefaults()
			secretName, err := getRoleSecretName(serviceName)
			if err != nil {
				return err
			}
			role.SetSecret(secretName)
			definition.Database.NeonPostgres.Roles = append(definition.Database.NeonPostgres.Roles, *role)
		}
		owner, _ := flags.GetString("db-owner")
		definition.Database.NeonPostgres.Roles[0].Name = &owner
	}

	if len(definition.Database.NeonPostgres.Databases) == 0 || flagChanged("db-name") {
		if len(definition.Database.NeonPostgres.Databases) > 1 {
			return &errors.CLIError{
				What: "Error while updating the database service definition",
				Why:  "the CLI does not support updating the database service definition with multiple databases",
				Additional: []string{
					"The Koyeb API supports multiple databases, but the CLI does not support it yet.",
				},
				Orig:     nil,
				Solution: "Try upgrading the CLI to the latest version. If you are already using the latest version, please create an issue on https://github.com/koyeb/koyeb-cli/issues/new and request support for multiple databases.",
			}
		}

		if len(definition.Database.NeonPostgres.Databases) == 0 {
			definition.Database.NeonPostgres.Databases = append(definition.Database.NeonPostgres.Databases, *koyeb.NewNeonPostgresDatabaseNeonDatabaseWithDefaults())
		}
		dbName, _ := flags.GetString("db-name")
		definition.Database.NeonPostgres.Databases[0].SetName(dbName)
		definition.Database.NeonPostgres.Databases[0].SetOwner(*definition.Database.NeonPostgres.Roles[0].Name)
	}
	return nil
}

// The Koyeb API requires to provide a secret name that will be used to store
// the database credentials. This function creates a random string based on the
// service name.
func getRoleSecretName(serviceName string) (string, error) {
	id, err := uuid.NewV4()
	if err != nil {
		return "", &errors.CLIError{
			What:       "Error while generating a random secret name",
			Why:        "the CLI failed to generate a UUID to use as a secret name",
			Additional: nil,
			Orig:       nil,
			Solution:   "Try upgrading the CLI to the latest version. If you are already using the latest version, please create an issue on https://github.com/koyeb/koyeb-cli/issues/new",
		}
	}
	return fmt.Sprintf("%s-%s", serviceName, id.String()[0:8]), nil
}

func NewDatabaseHandler() *DatabaseHandler {
	return &DatabaseHandler{}
}

type DatabaseHandler struct {
}

func (h *DatabaseHandler) ResolveAppArgs(ctx *CLIContext, val string) (string, error) {
	appMapper := ctx.Mapper.App()
	id, err := appMapper.ResolveID(val)
	if err != nil {
		return "", err
	}
	return id, nil
}

func (h *DatabaseHandler) ResolveDatabaseArgs(ctx *CLIContext, val string) (string, error) {
	databaseMapper := ctx.Mapper.Database()
	id, err := databaseMapper.ResolveID(val)
	if err != nil {
		return "", err
	}
	return id, nil
}
