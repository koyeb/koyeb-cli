package koyeb

import (
	"context"
	"fmt"
	"os"
	"os/signal"
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
		Use:     "instances ACTION",
		Aliases: []string{"i", "inst", "instance"},
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

	getInstanceCmd := &cobra.Command{
		Use:   "get NAME",
		Short: "Get instance",
		Args:  cobra.ExactArgs(1),
		RunE:  instanceHandler.Get,
	}
	instanceCmd.AddCommand(getInstanceCmd)

	describeInstanceCmd := &cobra.Command{
		Use:   "describe NAME",
		Short: "Describe instance",
		Args:  cobra.ExactArgs(1),
		RunE:  instanceHandler.Describe,
	}
	instanceCmd.AddCommand(describeInstanceCmd)

	execInstanceCmd := &cobra.Command{
		Use:   "exec NAME CMD -- [args...]",
		Short: "Run a command in the context of an instance",
		Args:  cobra.MinimumNArgs(2),
		RunE:  instanceHandler.Exec,
	}
	instanceCmd.AddCommand(execInstanceCmd)

	logInstanceCmd := &cobra.Command{
		Use:     "logs NAME",
		Aliases: []string{"l", "log"},
		Short:   "Get instance logs",
		Args:    cobra.ExactArgs(1),
		RunE:    instanceHandler.Log,
	}
	instanceCmd.AddCommand(logInstanceCmd)

	return instanceCmd
}

func NewInstanceHandler() *InstanceHandler {
	return &InstanceHandler{
		client:      getApiClient(),
		ctxWithAuth: getAuth(context.Background()),
	}
}

type InstanceHandler struct {
	client      *koyeb.APIClient
	ctxWithAuth context.Context
}

func (d *InstanceHandler) ResolveInstanceShortID(id string) string {
	return ResolveSecretShortID(d.ctxWithAuth, d.client, id)
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
	instanceId, userCmd := h.ResolveInstanceShortID(args[0]), args[1:]

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

func formatInstanceStatus(status koyeb.InstanceStatus) string {
	return fmt.Sprintf("%s", status)
}
