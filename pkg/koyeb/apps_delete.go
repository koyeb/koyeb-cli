package koyeb

import (
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func (h *AppHandler) Delete(cmd *cobra.Command, args []string) error {
	force := GetBoolFlags(cmd, "force")
	if force {
		app, resp, err := h.client.AppsApi.GetApp(h.ctx, h.ResolveAppArgs(args[0])).Execute()
		if err != nil {
			fatalApiError(err, resp)
		}

		log.Infof("Deleting app %s...", app.App.GetName())
		for {
			res, resp, err := h.client.ServicesApi.ListServices(h.ctx).AppId(app.App.GetId()).Limit("100").Execute()
			if err != nil {
				fatalApiError(err, resp)
			}
			if res.GetCount() == 0 {
				break
			}
			for _, svc := range res.GetServices() {
				if svc.State.GetStatus() == "STOPPING" || svc.State.GetStatus() == "STOPPED" {
					continue
				}
				log.Infof("Deleting service %s", svc.GetName())
				_, resp, err := h.client.ServicesApi.DeleteService(h.ctx, svc.GetId()).Execute()
				if err != nil {
					fatalApiError(err, resp)
				}
			}
			time.Sleep(2 * time.Second)
		}
	}

	_, resp, err := h.client.AppsApi.DeleteApp(h.ctx, h.ResolveAppArgs(args[0])).Execute()
	if err != nil {
		fatalApiError(err, resp)
	}

	log.Infof("App %s deleted.", args[0])
	return nil
}
