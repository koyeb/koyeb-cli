package koyeb

import (
	"context"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/idmapper"
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

	page := int64(0)
	offset := int64(0)
	limit := int64(100)
	for {
		resp, _, err := query.Limit(strconv.FormatInt(limit, 10)).Offset(strconv.FormatInt(offset, 10)).Execute()
		if err != nil {
			fatalApiError(err)
		}
		render("table", NewListInstancesReply(resp, appMapper, serviceMapper))
		page += 1
		offset = page * limit
		if offset >= resp.GetCount() {
			break
		}
	}

	return nil
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
	serviceID := ""

	query := client.InstancesApi.ListInstances(ctx).Statuses([]string{
		string(koyeb.INSTANCESTATUS_ALLOCATING),
		string(koyeb.INSTANCESTATUS_STARTING),
		string(koyeb.INSTANCESTATUS_HEALTHY),
		string(koyeb.INSTANCESTATUS_UNHEALTHY),
		string(koyeb.INSTANCESTATUS_STOPPING),
	})

	if appFilter != "" {
		id, err := appMapper.GetID(appFilter)
		if err != nil {
			fatalApiError(err)
		}
		appID = id
		query = query.AppId(appID)
	}

	if serviceFilter != "" {
		if appID == "" {
			log.Fatalf("Cannot use service filter without an application filter")
		}
		id, err := serviceMapper.GetID(appID, serviceFilter)
		if err != nil {
			fatalApiError(err)
		}
		serviceID = id
		query = query.ServiceId(serviceID)
	}

	return query
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
}

func NewListInstancesReply(items koyeb.ListInstancesReply, appMapper *idmapper.AppMapper, serviceMapper *idmapper.ServiceMapper) *ListInstancesReply {
	return &ListInstancesReply{
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
	headers := reply.Headers()
	resp := make([]map[string]string, 0, len(items))

	for _, item := range items {
		fields := make(map[string]string, len(headers))
		for _, field := range headers {
			fields[field] = reply.getField(field, item)
		}
		resp = append(resp, fields)
	}

	return resp
}

func (reply *ListInstancesReply) getField(field string, item koyeb.InstanceListItem) string {
	switch field {
	case "app_name":
		appName, err := reply.appMapper.GetName(*item.AppId)
		if err != nil {
			fatalApiError(err)
		}

		return appName

	case "service_name":
		serviceName, err := reply.serviceMapper.GetName(*item.AppId, *item.ServiceId)
		if err != nil {
			fatalApiError(err)
		}

		return serviceName

	default:
		return GetField(item, field)
	}
}
