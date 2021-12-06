package koyeb

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/idmapper"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/renderer"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func NewInstanceCmd() *cobra.Command {
	instanceHandler := NewInstanceHandler()

	instanceCmd := &cobra.Command{
		Use:     "instances [action]",
		Aliases: []string{"i", "instance"},
		Short:   "Instances",
	}

	listInstanceCmd := &cobra.Command{
		Use:   "list",
		Short: "List instances",
		RunE:  instanceHandler.List,
	}
	listInstanceCmd.Flags().String("app", "", "Filter on App id or name")
	listInstanceCmd.Flags().String("service", "", "Filter on Service id or name")
	instanceCmd.AddCommand(listInstanceCmd)

	execInstanceCmd := &cobra.Command{
		Use:   "exec [name] [cmd] [cmd...]",
		Short: "Run a command in the context of an instance",
		Args:  cobra.MinimumNArgs(2),
		RunE:  instanceHandler.Exec,
	}
	instanceCmd.AddCommand(execInstanceCmd)

	return instanceCmd
}

func NewInstanceHandler() *InstanceHandler {
	return &InstanceHandler{}
}

type InstanceHandler struct{}

func (h *InstanceHandler) List(cmd *cobra.Command, args []string) error {
	ctx := getAuth(context.Background())
	client := getApiClient()

	appMapper := idmapper.NewAppMapper(ctx, client)
	serviceMapper := idmapper.NewServiceMapper(ctx, client)

	query := h.getListQuery(ctx, cmd, client, appMapper, serviceMapper)
	results := koyeb.ListInstancesReply{}

	page := int64(0)
	offset := int64(0)
	limit := int64(100)
	for {
		resp, _, err := query.Limit(strconv.FormatInt(limit, 10)).Offset(strconv.FormatInt(offset, 10)).Execute()
		if err != nil {
			fatalApiError(err)
		}
		if results.Instances == nil {
			results = resp
		} else {
			*results.Instances = append(*results.Instances, *resp.Instances...)
		}
		page += 1
		offset = page * limit
		if offset >= resp.GetCount() {
			break
		}
	}
	full, _ := cmd.Flags().GetBool("full")
	listInstancesReply := NewListInstancesReply(results, appMapper, serviceMapper, full)

	output, _ := cmd.Flags().GetString("output")
	return renderer.NewListRenderer(listInstancesReply).Render(output)
}

func (h *InstanceHandler) Exec(cmd *cobra.Command, args []string) error {
	returnCode, err := h.exec(cmd, args)
	if err != nil {
		fatalApiError(err)
	}
	if returnCode != 0 {
		os.Exit(returnCode)
	}
	return nil
}

func (h *InstanceHandler) getListQuery(ctx context.Context, cmd *cobra.Command, client *koyeb.APIClient, appMapper *idmapper.AppMapper, serviceMapper *idmapper.ServiceMapper) koyeb.ApiListInstancesRequest {
	appFilter, _ := cmd.Flags().GetString("app")
	serviceFilter, _ := cmd.Flags().GetString("service")
	appID := ""

	query := client.InstancesApi.ListInstances(ctx).Statuses([]string{
		string(koyeb.INSTANCESTATUS_ALLOCATING),
		string(koyeb.INSTANCESTATUS_STARTING),
		string(koyeb.INSTANCESTATUS_HEALTHY),
		string(koyeb.INSTANCESTATUS_UNHEALTHY),
		string(koyeb.INSTANCESTATUS_STOPPING),
	})

	query, appID = h.getAppIDForListQuery(query, appFilter, appMapper)
	query = h.getServiceIDForListQuery(query, appID, serviceFilter, serviceMapper)

	return query
}

func (h *InstanceHandler) getAppIDForListQuery(query koyeb.ApiListInstancesRequest, filter string, appMapper *idmapper.AppMapper) (koyeb.ApiListInstancesRequest, string) {
	if filter == "" {
		return query, ""
	}
	if idmapper.IsUUIDv4(filter) {
		query = query.AppId(filter)
		return query, filter
	}

	id, err := appMapper.GetID(filter)
	if err != nil {
		fatalApiError(err)
	}

	query = query.AppId(id)
	return query, id
}

func (h *InstanceHandler) getServiceIDForListQuery(query koyeb.ApiListInstancesRequest, appID string, filter string, serviceMapper *idmapper.ServiceMapper) koyeb.ApiListInstancesRequest {
	if filter == "" {
		return query
	}
	if idmapper.IsUUIDv4(filter) {
		return query.ServiceId(filter)
	}
	if appID == "" {
		log.Fatalf("Cannot use service filter without an application filter")
	}

	id, err := serviceMapper.GetID(appID, filter)
	if err != nil {
		fatalApiError(err)
	}

	return query.ServiceId(id)
}

func (h *InstanceHandler) exec(cmd *cobra.Command, args []string) (int, error) {
	// Cobra options ensure we have at least 2 arguments here, but still
	if len(args) < 2 {
		return 0, errors.New("exec needs at least 2 arguments")
	}
	instanceId, userCmd := args[0], args[1:]

	stdStreams, cleanup, err := GetStdStreams()
	if err != nil {
		return 0, errors.Wrap(err, "could not get standard streams")
	}
	defer cleanup()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	termResizeCh := watchTermSize(ctx, stdStreams)

	e := NewExecutor(stdStreams.Stdin, stdStreams.Stdout, stdStreams.Stderr, userCmd, instanceId, termResizeCh)
	return e.Run(ctx)
}

func watchTermSize(ctx context.Context, s *StdStreams) <-chan *TerminalSize {
	out := make(chan *TerminalSize)
	go func() {
		defer close(out)
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGWINCH)
		for {
			select {
			case <-ctx.Done():
				return
			case <-sigCh:
				termSize, err := GetTermSize(s.Stdout)
				if err != nil {
					continue
				}
				select {
				case <-ctx.Done():
					return
				case out <- termSize:
				}
			}
		}
	}()
	return out
}

type ListInstancesReply struct {
	items         koyeb.ListInstancesReply
	appMapper     *idmapper.AppMapper
	serviceMapper *idmapper.ServiceMapper
	full          bool
}

func NewListInstancesReply(items koyeb.ListInstancesReply, appMapper *idmapper.AppMapper, serviceMapper *idmapper.ServiceMapper, full bool) *ListInstancesReply {
	return &ListInstancesReply{
		full:          full,
		items:         items,
		appMapper:     appMapper,
		serviceMapper: serviceMapper,
	}
}

func (ListInstancesReply) Title() string {
	return "Instances"
}

func (reply *ListInstancesReply) MarshalBinary() ([]byte, error) {
	return reply.items.MarshalJSON()
}

func (reply *ListInstancesReply) Headers() []string {
	return []string{"id", "status", "app_name", "service_name", "deployment_id", "datacenter"}
}

func (reply *ListInstancesReply) Fields() []map[string]string {
	items := reply.items.GetInstances()
	resp := make([]map[string]string, 0, len(items))

	for _, item := range items {
		appName, err := reply.appMapper.GetName(*item.AppId)
		if err != nil {
			fatalApiError(err)
		}
		serviceName, err := reply.serviceMapper.GetName(*item.AppId, *item.ServiceId)
		if err != nil {
			fatalApiError(err)
		}

		fields := map[string]string{
			"id":            renderer.FormatID(item.GetId(), reply.full),
			"status":        formatInstanceStatus(item.GetStatus()),
			"app_name":      appName,
			"service_name":  serviceName,
			"deployment_id": renderer.FormatID(item.GetDeploymentId(), reply.full),
			"datacenter":    item.GetDatacenter(),
		}
		resp = append(resp, fields)
	}

	return resp
}

func formatInstanceStatus(status koyeb.InstanceStatus) string {
	return fmt.Sprintf("%s", status)
}
