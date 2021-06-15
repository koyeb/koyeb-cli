package koyeb

import (
	"context"
	"fmt"
	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/logrusorgru/aurora"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"strconv"
	"strings"
)

func addServiceDefinitionFlags(flags *pflag.FlagSet) {
	flags.String("docker", "koyeb/demo", "Docker image")
	flags.String("docker-command", "", "Docker command")
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
		createDockerSource.SetImage(image)
		definition.SetDocker(*createDockerSource)
	}
	return nil
}

func NewServiceCmd() *cobra.Command {
	h := NewServiceHandler()

	serviceCmd := &cobra.Command{
		Use:     "services [action]",
		Aliases: []string{"s", "service"},
		Short:   "Services",
	}

	createServiceCmd := &cobra.Command{
		Use:   "create [name]",
		Short: "Create services",
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
	serviceCmd.AddCommand(createServiceCmd)

	getServiceCmd := &cobra.Command{
		Use:   "get [name]",
		Short: "Get service",
		RunE:  h.Get,
	}
	serviceCmd.AddCommand(getServiceCmd)

	listServiceCmd := &cobra.Command{
		Use:   "list",
		Short: "List services",
		RunE:  h.List,
	}
	serviceCmd.AddCommand(listServiceCmd)

	describeServiceCmd := &cobra.Command{
		Use:   "describe [name]",
		Short: "Describe services",
		RunE:  h.Describe,
	}
	serviceCmd.AddCommand(describeServiceCmd)

	updateServiceCmd := &cobra.Command{
		Use:   "update [name]",
		Short: "Update services",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			updateService := koyeb.NewUpdateServiceWithDefaults()

			client := getApiClient()
			ctx := getAuth(context.Background())
			app := getSelectedApp()
			revDetail, _, err := client.ServicesApi.GetRevision(ctx, app, args[0], "_latest").Execute()
			if err != nil {
				fatalApiError(err)
			}
			updateDef := revDetail.GetRevision().Definition
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
		Use:   "redeploy [name]",
		Short: "Redeploy services",
		Args:  cobra.MinimumNArgs(1),
		RunE:  h.ReDeploy,
	}
	serviceCmd.AddCommand(redeployServiceCmd)

	deleteServiceCmd := &cobra.Command{
		Use:   "delete [name]",
		Short: "Delete services",
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

func (h *ServiceHandler) Create(cmd *cobra.Command, args []string, createService *koyeb.CreateService) error {
	client := getApiClient()
	ctx := getAuth(context.Background())

	app := getSelectedApp()
	_, _, err := client.ServicesApi.CreateService(ctx, app).Body(*createService).Execute()
	if err != nil {
		fatalApiError(err)
	}
	return nil
}

func (h *ServiceHandler) Update(cmd *cobra.Command, args []string, updateService *koyeb.UpdateService) error {
	client := getApiClient()
	ctx := getAuth(context.Background())

	app := getSelectedApp()
	_, _, err := client.ServicesApi.UpdateService(ctx, app, args[0]).Body(*updateService).Execute()
	if err != nil {
		fatalApiError(err)
	}
	return nil
}

func (h *ServiceHandler) Get(cmd *cobra.Command, args []string) error {
	format := "table"
	if len(args) == 0 {
		return h.listFormat(cmd, args, format)
	}
	return h.getFormat(cmd, args, format)
}

func (h *ServiceHandler) Describe(cmd *cobra.Command, args []string) error {
	format := "detail"
	if len(args) == 0 {
		return h.listFormat(cmd, args, format)
	}
	return h.getFormat(cmd, args, format)
}

func (h *ServiceHandler) ReDeploy(cmd *cobra.Command, args []string) error {
	client := getApiClient()
	ctx := getAuth(context.Background())

	app := getSelectedApp()
	for _, arg := range args {
		redeployRequest := koyeb.NewRedeployRequestInfoWithDefaults()
		_, _, err := client.ServicesApi.ReDeploy(ctx, app, arg).Body(*redeployRequest).Execute()
		if err != nil {
			fatalApiError(err)
		}
	}
	return nil
}

func (h *ServiceHandler) Delete(cmd *cobra.Command, args []string) error {
	client := getApiClient()
	ctx := getAuth(context.Background())

	app := getSelectedApp()
	for _, arg := range args {
		_, _, err := client.ServicesApi.DeleteService(ctx, app, arg).Execute()
		if err != nil {
			fatalApiError(err)
		}
	}
	return nil
}

func (h *ServiceHandler) List(cmd *cobra.Command, args []string) error {
	format := "table"
	return h.listFormat(cmd, args, format)
}

func (h *ServiceHandler) getFormat(cmd *cobra.Command, args []string, format string) error {
	client := getApiClient()
	ctx := getAuth(context.Background())
	app := getSelectedApp()

	for _, arg := range args {
		res, _, err := client.ServicesApi.GetService(ctx, app, arg).Execute()
		if err != nil {
			fatalApiError(err)
		}
		render(format, &GetServiceReply{res})
		if format == "detail" {
			res, _, err := client.ServicesApi.ListRevisions(ctx, app, arg).Limit("100").Execute()
			if err != nil {
				fatalApiError(err)
			}

			revDetail, _, err := client.ServicesApi.GetRevision(ctx, app, arg, "_latest").Execute()
			if err != nil {
				fatalApiError(err)
			}
			rendDetail := &GetServiceRevisionReply{revDetail}
			fmt.Printf("\n")
			render("detail", rendDetail)

			rend := &ListServiceRevisionsReply{res}
			fmt.Printf("\n%s history\n", aurora.Bold(rend.Title()))
			render("table", rend)
		}
	}

	return nil
}

func (h *ServiceHandler) listFormat(cmd *cobra.Command, args []string, format string) error {
	client := getApiClient()
	ctx := getAuth(context.Background())

	app := getSelectedApp()

	page := 0
	offset := 0
	limit := 100
	for {
		res, _, err := client.ServicesApi.ListServices(ctx, app).Limit(fmt.Sprintf("%d", limit)).Offset(fmt.Sprintf("%d", offset)).Execute()
		if err != nil {
			fatalApiError(err)
		}
		render(format, &ListServicesReply{res})
		page += 1
		offset = page * limit
		if int64(offset) >= res.GetCount() {
			break
		}
	}

	return nil
}

type GetServiceReply struct {
	koyeb.GetServiceReply
}

func (a *GetServiceReply) MarshalBinary() ([]byte, error) {
	return a.GetServiceReply.GetService().MarshalJSON()
}

func (a *GetServiceReply) Title() string {
	return "Service"
}

func (a *GetServiceReply) Headers() []string {
	return []string{"id", "name", "version", "status", "updated_at"}
}

func (a *GetServiceReply) Fields() []map[string]string {
	res := []map[string]string{}
	item := a.GetService()
	fields := map[string]string{}
	for _, field := range a.Headers() {
		switch field {
		case "status":
			fields[field] = GetField(item, "state.status")
		default:
			fields[field] = GetField(item, field)
		}
	}
	res = append(res, fields)
	return res
}

type ListServicesReply struct {
	koyeb.ListServicesReply
}

func (a *ListServicesReply) Title() string {
	return "Services"
}

func (a *ListServicesReply) MarshalBinary() ([]byte, error) {
	return a.ListServicesReply.MarshalJSON()
}

func (a *ListServicesReply) Headers() []string {
	return []string{"id", "name", "status", "updated_at"}
}

func (a *ListServicesReply) Fields() []map[string]string {
	res := []map[string]string{}
	for _, item := range a.GetServices() {
		fields := map[string]string{}
		for _, field := range a.Headers() {
			switch field {
			case "status":
				fields[field] = GetField(item, "state.status")
			default:
				fields[field] = GetField(item, field)
			}
		}
		res = append(res, fields)
	}
	return res
}

type ListServiceRevisionsReply struct {
	koyeb.ListServiceRevisionsReply
}

func (a *ListServiceRevisionsReply) Title() string {
	return "Revisions"
}

func (a *ListServiceRevisionsReply) MarshalBinary() ([]byte, error) {
	return a.ListServiceRevisionsReply.MarshalJSON()
}

func (a *ListServiceRevisionsReply) Headers() []string {
	return []string{"id", "status", "updated_at"}
}

func (a *ListServiceRevisionsReply) Fields() []map[string]string {
	res := []map[string]string{}
	for _, item := range a.GetRevisions() {
		fields := map[string]string{}
		for _, field := range a.Headers() {
			switch field {
			default:
				fields[field] = GetField(item, field)
			}
		}
		res = append(res, fields)
	}
	return res
}

type GetServiceRevisionReply struct {
	koyeb.GetServiceRevisionReply
}

func (a *GetServiceRevisionReply) Title() string {
	return "Revision Detail"
}

func (a *GetServiceRevisionReply) MarshalBinary() ([]byte, error) {
	return a.GetServiceRevisionReply.MarshalJSON()
}

func (a *GetServiceRevisionReply) Headers() []string {
	return []string{"id", "version", "status", "status_message", "instances", "definition", "updated_at"}
}

func (a *GetServiceRevisionReply) Fields() []map[string]string {
	res := []map[string]string{}
	item := a.GetRevision()
	fields := map[string]string{}
	for _, field := range a.Headers() {
		switch field {
		case "status":
			fields[field] = GetField(item, "state.status")
		case "status_message":
			fields[field] = GetField(item, "state.status_message")
		case "definition":
			b, err := item.Definition.MarshalJSON()
			if err == nil {
				fields[field] = string(b)
			}
		case "instances":
			var instances []string
			for _, inst := range item.State.GetInstances() {
				instances = append(instances, fmt.Sprintf("%s:%s", inst.GetId(), inst.GetStatus()))
			}
			fields[field] = strings.Join(instances, "\n")
		default:
			fields[field] = GetField(item, field)
		}
	}
	res = append(res, fields)
	return res
}
