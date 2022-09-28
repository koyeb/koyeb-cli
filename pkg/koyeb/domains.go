package koyeb

import (
	"context"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/idmapper"
	"github.com/spf13/cobra"
)

func NewDomainCmd() *cobra.Command {
	h := NewDomainHandler()

	domainCmd := &cobra.Command{
		Use:               "domains ACTION",
		Aliases:           []string{"dom", "domain"},
		Short:             "Domains",
		PersistentPreRunE: h.InitHandler,
	}

	getDomainCmd := &cobra.Command{
		Use:   "get NAME",
		Short: "Get domain",
		Args:  cobra.ExactArgs(1),
		RunE:  h.Get,
	}
	domainCmd.AddCommand(getDomainCmd)

	createDomainCmd := &cobra.Command{
		Use:   "create NAME",
		Short: "Create domain",
		Args:  cobra.ExactArgs(1),
		RunE:  h.Create,
	}
	createDomainCmd.Flags().String("attach-to", "", "Upon creation, assign to given app")
	domainCmd.AddCommand(createDomainCmd)

	describeDomainCmd := &cobra.Command{
		Use:   "describe",
		Short: "Describe domain",
		RunE:  h.Describe,
	}
	domainCmd.AddCommand(describeDomainCmd)

	listDomainCmd := &cobra.Command{
		Use:   "list",
		Short: "List domains",
		RunE:  h.List,
	}
	domainCmd.AddCommand(listDomainCmd)

	deleteDomainCmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete domain",
		RunE:  h.Delete,
	}
	domainCmd.AddCommand(deleteDomainCmd)

	refreshDomainCmd := &cobra.Command{
		Use:   "refresh NAME",
		Short: "Refresh a custom domain verification status",
		Args:  cobra.ExactArgs(1),
		RunE:  h.Refresh,
	}
	domainCmd.AddCommand(refreshDomainCmd)

	attachDomainCmd := &cobra.Command{
		Use:   "attach NAME APP",
		Short: "Attach a custom domain to an existing app",
		Args:  cobra.ExactArgs(2),
		RunE:  h.Attach,
	}
	domainCmd.AddCommand(attachDomainCmd)

	detachDomainCmd := &cobra.Command{
		Use:   "detach NAME",
		Short: "Detach a custom domain from the app it is currently attached to",
		Args:  cobra.ExactArgs(1),
		RunE:  h.Detach,
	}
	domainCmd.AddCommand(detachDomainCmd)

	return domainCmd
}

type DomainHandler struct {
	ctx    context.Context
	client *koyeb.APIClient
	mapper *idmapper.Mapper
}

func NewDomainHandler() *DomainHandler {
	return &DomainHandler{}
}

func (h *DomainHandler) InitHandler(cmd *cobra.Command, args []string) error {
	h.ctx = getAuth(context.Background())
	h.client = getApiClient()
	h.mapper = idmapper.NewMapper(h.ctx, h.client)
	return nil
}

func (h *DomainHandler) ResolveDomainArgs(val string) string {
	domainMapper := h.mapper.Domain()
	id, err := domainMapper.ResolveID(val)
	if err != nil {
		fatalApiError(err, nil)
	}

	return id
}
