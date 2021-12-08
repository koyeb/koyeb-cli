package koyeb

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
)

func NewDeploymentCmd() *cobra.Command {
	h := NewDeploymentHandler()

	deploymentCmd := &cobra.Command{
		Use:     "deployments ACTION",
		Aliases: []string{"d", "dep", "depl", "deploy", "deployment"},
		Short:   "Deployments",
	}

	listDeploymentCmd := &cobra.Command{
		Use:   "list",
		Short: "List deployments",
		RunE:  h.List,
	}
	deploymentCmd.AddCommand(listDeploymentCmd)

	getDeploymentCmd := &cobra.Command{
		Use:   "get NAME",
		Short: "Get deployment",
		Args:  cobra.ExactArgs(1),
		RunE:  h.Get,
	}
	deploymentCmd.AddCommand(getDeploymentCmd)

	describeDeploymentCmd := &cobra.Command{
		Use:   "describe NAME",
		Short: "Describe deployment",
		Args:  cobra.ExactArgs(1),
		RunE:  h.Describe,
	}
	deploymentCmd.AddCommand(describeDeploymentCmd)

	logDeploymentCmd := &cobra.Command{
		Use:     "logs NAME",
		Aliases: []string{"l", "log"},
		Short:   "Get deployment logs",
		Args:    cobra.ExactArgs(1),
		RunE:    h.Log,
	}
	deploymentCmd.AddCommand(logDeploymentCmd)
	logDeploymentCmd.Flags().StringP("type", "t", "", "Type of log (runtime,build)")

	return deploymentCmd
}

func NewDeploymentHandler() *DeploymentHandler {
	return &DeploymentHandler{}
}

type DeploymentHandler struct {
}

func buildDeploymentShortIDCache() map[string][]string {
	c := make(map[string][]string)
	client := getApiClient()
	ctx := getAuth(context.Background())

	page := 0
	offset := 0
	limit := 100
	for {
		res, _, err := client.DeploymentsApi.ListDeployments(ctx).Limit(fmt.Sprintf("%d", limit)).Offset(fmt.Sprintf("%d", offset)).Execute()
		if err != nil {
			fatalApiError(err)
		}
		for _, a := range *res.Deployments {
			id := a.GetId()[:8]
			c[id] = append(c[id], a.GetId())

		}

		page += 1
		offset = page * limit
		if int64(offset) >= res.GetCount() {
			break
		}
	}

	return c
}

func ResolveDeploymentShortID(id string) string {
	if len(id) == 8 {
		// TODO do a real cache
		cache := buildDeploymentShortIDCache()
		nlid, ok := cache[id]
		if ok {
			if len(nlid) == 1 {
				return nlid[0]
			} else {
				return "local-short-id-conflict"
			}
		}
		return id
	} else {
		return id
	}
}
