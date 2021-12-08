package koyeb

import (
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func (h *AppHandler) Delete(cmd *cobra.Command, args []string) error {
	force, _ := cmd.Flags().GetBool("force")
	if force {
		app, _, err := h.client.AppsApi.GetApp(h.ctxWithAuth, h.ResolveAppShortID(args[0])).Execute()
		if err != nil {
			fatalApiError(err)
		}

		log.Infof("Deleting app %s...", app.App.GetName())
		for {
			res, _, err := h.client.ServicesApi.ListServices(h.ctxWithAuth).AppId(app.App.GetId()).Limit("100").Execute()
			if err != nil {
				fatalApiError(err)
			}
			if res.GetCount() == 0 {
				break
			}
			for _, svc := range res.GetServices() {
				if svc.State.GetStatus() == "STOPPING" || svc.State.GetStatus() == "STOPPED" {
					continue
				}
				log.Infof("Deleting service %s", svc.GetName())
				_, _, err := h.client.ServicesApi.DeleteService(h.ctxWithAuth, svc.GetId()).Execute()
				if err != nil {
					fatalApiError(err)
				}
			}
			time.Sleep(2 * time.Second)
		}
	}
	_, _, err := h.client.AppsApi.DeleteApp(h.ctxWithAuth, h.ResolveAppShortID(args[0])).Execute()
	if err != nil {
		fatalApiError(err)
	}

	log.Infof("App %s deleted.", args[0])
	return nil
}
