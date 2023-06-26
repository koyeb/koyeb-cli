package koyeb

import (
	"github.com/spf13/cobra"
)

func NewDomainCmd() *cobra.Command {
	h := NewDomainHandler()

	domainCmd := &cobra.Command{
		Use:     "domains ACTION",
		Aliases: []string{"dom", "domain"},
		Short:   "Domains",
	}

	getDomainCmd := &cobra.Command{
		Use:   "get NAME",
		Short: "Get domain",
		Args:  cobra.ExactArgs(1),
		RunE:  WithCLIContext(h.Get),
	}
	domainCmd.AddCommand(getDomainCmd)

	createDomainCmd := &cobra.Command{
		Use:   "create NAME",
		Short: "Create domain",
		Args:  cobra.ExactArgs(1),
		RunE:  WithCLIContext(h.Create),
	}
	createDomainCmd.Flags().String("attach-to", "", "Upon creation, assign to given app")
	domainCmd.AddCommand(createDomainCmd)

	describeDomainCmd := &cobra.Command{
		Use:   "describe",
		Short: "Describe domain",
		Args:  cobra.ExactArgs(1),
		RunE:  WithCLIContext(h.Describe),
	}
	domainCmd.AddCommand(describeDomainCmd)

	listDomainCmd := &cobra.Command{
		Use:   "list",
		Short: "List domains",
		RunE:  WithCLIContext(h.List),
	}
	domainCmd.AddCommand(listDomainCmd)

	deleteDomainCmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete domain",
		RunE:  WithCLIContext(h.Delete),
	}
	domainCmd.AddCommand(deleteDomainCmd)

	refreshDomainCmd := &cobra.Command{
		Use:   "refresh NAME",
		Short: "Refresh a custom domain verification status",
		Args:  cobra.ExactArgs(1),
		RunE:  WithCLIContext(h.Refresh),
	}
	domainCmd.AddCommand(refreshDomainCmd)

	attachDomainCmd := &cobra.Command{
		Use:   "attach NAME APP",
		Short: "Attach a custom domain to an existing app",
		Args:  cobra.ExactArgs(2),
		RunE:  WithCLIContext(h.Attach),
	}
	domainCmd.AddCommand(attachDomainCmd)

	detachDomainCmd := &cobra.Command{
		Use:   "detach NAME",
		Short: "Detach a custom domain from the app it is currently attached to",
		Args:  cobra.ExactArgs(1),
		RunE:  WithCLIContext(h.Detach),
	}
	domainCmd.AddCommand(detachDomainCmd)

	return domainCmd
}

type DomainHandler struct {
}

func NewDomainHandler() *DomainHandler {
	return &DomainHandler{}
}

func (h *DomainHandler) ResolveDomainArgs(ctx *CLIContext, val string) string {
	domainMapper := ctx.Mapper.Domain()
	id, err := domainMapper.ResolveID(val)
	if err != nil {
		fatalApiError(err, nil)
	}

	return id
}
