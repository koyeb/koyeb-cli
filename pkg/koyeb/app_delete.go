package koyeb

import (
	"context"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func (h *AppHandler) Delete(cmd *cobra.Command, args []string) error {
	client := getApiClient()
	ctx := getAuth(context.Background())

	force, _ := cmd.Flags().GetBool("force")
	for _, arg := range args {
		if force {
			app, _, err := client.AppsApi.GetApp(ctx, ResolveAppShortID(args[0])).Execute()
			if err != nil {
				fatalApiError(err)
			}

			log.Infof("Deleting app %s...", app.App.GetName())
			for {
				res, _, err := client.ServicesApi.ListServices(ctx).AppId(app.App.GetId()).Limit("100").Execute()
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
					_, _, err := client.ServicesApi.DeleteService(ctx, app.App.GetId(), svc.GetId()).Execute()
					if err != nil {
						fatalApiError(err)
					}
				}
				time.Sleep(2 * time.Second)
			}
		}
		_, _, err := client.AppsApi.DeleteApp(ctx, arg).Execute()
		if err != nil {
			fatalApiError(err)
		}
	}

	log.Infof("Apps %s deleted.", strings.Join(args, ", "))
	return nil
}
