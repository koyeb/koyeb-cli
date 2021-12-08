package koyeb

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func addServiceDefinitionFlags(flags *pflag.FlagSet) {
	flags.String("docker", "koyeb/demo", "Docker image")
	flags.String("docker-private-registry-secret", "", "Docker private registry secret")
	flags.String("docker-command", "", "Docker command")
	flags.StringSlice("docker-args", []string{}, "Docker args")
	flags.StringSlice("regions", []string{"par"}, "Regions")
	flags.StringSlice("env", []string{}, "Env")
	flags.StringSlice("routes", []string{"/:80"}, "Ports")
	flags.StringSlice("ports", []string{"80:http"}, "Ports")
	flags.String("instance-type", "nano", "Instance type")
	flags.Int64("min-scale", 1, "Min scale")
	flags.Int64("max-scale", 1, "Max scale")
}

func parseServiceDefinitionFlags(flags *pflag.FlagSet, definition *koyeb.ServiceDefinition, useDefault bool) error {

	if useDefault || flags.Lookup("env").Changed {
		env, _ := flags.GetStringSlice("env")
		var envs []koyeb.Env
		for _, e := range env {
			newEnv := koyeb.NewEnvWithDefaults()

			spli := strings.Split(e, "=")
			if len(spli) < 2 {
				return errors.New("Unable to parse env")
			}
			newEnv.Key = koyeb.PtrString(spli[0])
			if spli[1][0] == '@' {
				newEnv.ValueFromSecret = koyeb.PtrString(spli[1][1:])
			} else {
				newEnv.Value = koyeb.PtrString(spli[1])
			}
			envs = append(envs, *newEnv)
		}
		definition.SetEnv(envs)
	}

	if useDefault || flags.Lookup("instance-type").Changed {
		instanceType, _ := flags.GetString("instance-type")
		definition.SetInstanceType(instanceType)
	}
	if useDefault || flags.Lookup("regions").Changed {
		regions, _ := flags.GetStringSlice("regions")
		definition.SetRegions(regions)
	}

	if useDefault || flags.Lookup("ports").Changed {
		port, _ := flags.GetStringSlice("ports")
		var ports []koyeb.Port
		for _, p := range port {
			newPort := koyeb.NewPortWithDefaults()

			spli := strings.Split(p, ":")
			if len(spli) < 1 {
				return errors.New("Unable to parse port")
			}
			portNum, err := strconv.Atoi(spli[0])
			if err != nil {
				errors.Wrap(err, "Invalid port number")
			}
			newPort.Port = koyeb.PtrInt64(int64(portNum))
			newPort.Protocol = koyeb.PtrString("http")
			if len(spli) > 1 {
				newPort.Protocol = koyeb.PtrString(spli[1])
			}
			ports = append(ports, *newPort)

		}
		definition.SetPorts(ports)
	}

	if useDefault || flags.Lookup("routes").Changed {
		route, _ := flags.GetStringSlice("routes")
		var routes []koyeb.Route
		for _, p := range route {
			newRoute := koyeb.NewRouteWithDefaults()

			spli := strings.Split(p, ":")
			if len(spli) < 1 {
				return errors.New("Unable to parse route")
			}
			newRoute.Path = koyeb.PtrString(spli[0])
			newRoute.Port = koyeb.PtrInt64(80)
			if len(spli) > 1 {
				portNum, err := strconv.Atoi(spli[1])
				if err != nil {
					errors.Wrap(err, "Invalid route number")
				}
				newRoute.Port = koyeb.PtrInt64(int64(portNum))
			}
			routes = append(routes, *newRoute)

		}
		definition.SetRoutes(routes)
	}

	if useDefault || flags.Lookup("min-scale").Changed || flags.Lookup("max-scale").Changed {
		scaling := koyeb.NewScalingWithDefaults()
		minScale, _ := flags.GetInt64("min-scale")
		maxScale, _ := flags.GetInt64("max-scale")
		scaling.SetMin(minScale)
		scaling.SetMax(maxScale)
		definition.SetScaling(*scaling)
	}

	// Docker
	if useDefault || flags.Lookup("docker").Changed {
		createDockerSource := koyeb.NewDockerSourceWithDefaults()
		image, _ := flags.GetString("docker")
		args, _ := flags.GetStringSlice("docker-args")
		command, _ := flags.GetString("docker-command")
		image_registry_secret, _ := flags.GetString("docker-private-registry-secret")
		createDockerSource.SetImage(image)
		if command != "" {
			createDockerSource.SetCommand(command)
		}
		if image_registry_secret != "" {
			createDockerSource.SetImageRegistrySecret(image_registry_secret)
		}
		if len(args) > 0 {
			createDockerSource.SetArgs(args)
		}
		definition.SetDocker(*createDockerSource)
	}
	return nil
}

func NewServiceCmd() *cobra.Command {
	h := NewServiceHandler()

	serviceCmd := &cobra.Command{
		Use:     "services ACTION",
		Aliases: []string{"s", "svc", "service"},
		Short:   "Services",
	}

	createServiceCmd := &cobra.Command{
		Use:   "create NAME",
		Short: "Create service",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			createService := koyeb.NewCreateServiceWithDefaults()
			createDef := koyeb.NewServiceDefinitionWithDefaults()

			err := parseServiceDefinitionFlags(cmd.Flags(), createDef, true)
			if err != nil {
				return err
			}
			createDef.Name = koyeb.PtrString(args[0])

			createService.SetDefinition(*createDef)
			return h.Create(cmd, args, createService)
		},
	}
	addServiceDefinitionFlags(createServiceCmd.Flags())
	createServiceCmd.Flags().StringP("app", "a", "", "App")
	createServiceCmd.MarkFlagRequired("app")
	serviceCmd.AddCommand(createServiceCmd)

	getServiceCmd := &cobra.Command{
		Use:   "get NAME",
		Short: "Get service",
		Args:  cobra.ExactArgs(1),
		RunE:  h.Get,
	}
	serviceCmd.AddCommand(getServiceCmd)

	logsServiceCmd := &cobra.Command{
		Use:     "logs NAME",
		Aliases: []string{"l", "log"},
		Short:   "Get the service logs",
		Args:    cobra.ExactArgs(1),
		RunE:    h.Log,
	}
	serviceCmd.AddCommand(logsServiceCmd)
	logsServiceCmd.Flags().String("instance", "", "Instance")

	listServiceCmd := &cobra.Command{
		Use:   "list",
		Short: "List services",
		RunE:  h.List,
	}
	serviceCmd.AddCommand(listServiceCmd)
	listServiceCmd.Flags().StringP("app", "a", "", "App")
	listServiceCmd.Flags().StringP("name", "n", "", "Service name")

	describeServiceCmd := &cobra.Command{
		Use:   "describe NAME",
		Short: "Describe service",
		Args:  cobra.ExactArgs(1),
		RunE:  h.Describe,
	}
	serviceCmd.AddCommand(describeServiceCmd)

	updateServiceCmd := &cobra.Command{
		Use:   "update NAME",
		Short: "Update service",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			updateService := koyeb.NewUpdateServiceWithDefaults()

			client := getApiClient()
			ctx := getAuth(context.Background())
			latestDeploy, _, err := client.DeploymentsApi.ListDeployments(ctx).Limit(fmt.Sprintf("%d", 1)).ServiceId(ResolveServiceShortID(args[0])).Execute()
			if err != nil {
				fatalApiError(err)
			}
			if len(latestDeploy.GetDeployments()) == 0 {
				return errors.New("Unable to load latest deployment")
			}
			updateDef := latestDeploy.GetDeployments()[0].Definition
			err = parseServiceDefinitionFlags(cmd.Flags(), updateDef, false)
			if err != nil {
				return err
			}
			updateService.SetDefinition(*updateDef)
			return h.Update(cmd, args, updateService)
		},
	}
	addServiceDefinitionFlags(updateServiceCmd.Flags())
	serviceCmd.AddCommand(updateServiceCmd)

	redeployServiceCmd := &cobra.Command{
		Use:   "redeploy NAME",
		Short: "Redeploy service",
		Args:  cobra.MinimumNArgs(1),
		RunE:  h.ReDeploy,
	}
	serviceCmd.AddCommand(redeployServiceCmd)

	deleteServiceCmd := &cobra.Command{
		Use:   "delete NAME",
		Short: "Delete service",
		Args:  cobra.MinimumNArgs(1),
		RunE:  h.Delete,
	}
	serviceCmd.AddCommand(deleteServiceCmd)

	return serviceCmd
}

func NewServiceHandler() *ServiceHandler {
	return &ServiceHandler{}
}

type ServiceHandler struct {
}

func buildServiceShortIDCache() map[string][]string {
	c := make(map[string][]string)
	client := getApiClient()
	ctx := getAuth(context.Background())

	page := 0
	offset := 0
	limit := 100
	for {
		res, _, err := client.ServicesApi.ListServices(ctx).Limit(fmt.Sprintf("%d", limit)).Offset(fmt.Sprintf("%d", offset)).Execute()
		if err != nil {
			fatalApiError(err)
		}
		for _, a := range *res.Services {
			id := a.GetId()[:8]
			c[id] = append(c[id], a.GetId())

		}

		page += 1
		offset = page * limit
		if int64(offset) >= res.GetCount() {
			break
		}
	}

	return c
}

func ResolveServiceShortID(id string) string {
	if len(id) == 8 {
		// TODO do a real cache
		cache := buildServiceShortIDCache()
		nlid, ok := cache[id]
		if ok {
			if len(nlid) == 1 {
				return nlid[0]
			} else {
				return "local-short-id-conflict"
			}
		}
		return id
	} else {
		return id
	}
}
