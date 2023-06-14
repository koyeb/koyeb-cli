package koyeb

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
)

func (h *InstanceHandler) Exec(ctx *CLIContext, cmd *cobra.Command, args []string) error {
	returnCode, err := exec(ExecId{
		Id:   h.ResolveInstanceArgs(ctx, args[0]),
		Type: koyeb.EXECCOMMANDREQUESTIDTYPE_INSTANCE_ID,
	}, args[1:])
	if err != nil {
		fatalApiError(err, nil)
	}
	if returnCode != 0 {
		os.Exit(returnCode)
	}
	return nil
}
