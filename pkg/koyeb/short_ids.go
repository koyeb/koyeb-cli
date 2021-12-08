package koyeb

import (
	"context"
	"fmt"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
)

func buildAppShortIDCache(ctx context.Context, client *koyeb.APIClient) map[string][]string {
	c := make(map[string][]string)

	page := 0
	offset := 0
	limit := 100
	for {
		res, _, err := client.AppsApi.ListApps(ctx).Limit(fmt.Sprintf("%d", limit)).Offset(fmt.Sprintf("%d", offset)).Execute()
		if err != nil {
			fatalApiError(err)
		}
		for _, a := range *res.Apps {
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

func ResolveAppShortID(ctx context.Context, client *koyeb.APIClient, id string) string {
	if len(id) == 8 {
		// TODO do a real cache
		cache := buildAppShortIDCache(ctx, client)
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

func buildDeploymentShortIDCache(ctx context.Context, client *koyeb.APIClient) map[string][]string {
	c := make(map[string][]string)

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

func ResolveDeploymentShortID(ctx context.Context, client *koyeb.APIClient, id string) string {
	if len(id) == 8 {
		// TODO do a real cache
		cache := buildDeploymentShortIDCache(ctx, client)
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

func buildInstanceShortIDCache(ctx context.Context, client *koyeb.APIClient) map[string][]string {
	c := make(map[string][]string)

	page := 0
	offset := 0
	limit := 100
	for {
		res, _, err := client.InstancesApi.ListInstances(ctx).Limit(fmt.Sprintf("%d", limit)).Offset(fmt.Sprintf("%d", offset)).Execute()
		if err != nil {
			fatalApiError(err)
		}
		for _, a := range *res.Instances {
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

func ResolveInstanceShortID(ctx context.Context, client *koyeb.APIClient, id string) string {
	if len(id) == 8 {
		// TODO do a real cache
		cache := buildInstanceShortIDCache(ctx, client)
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

func buildSecretShortIDCache(ctx context.Context, client *koyeb.APIClient) map[string][]string {
	c := make(map[string][]string)

	page := 0
	offset := 0
	limit := 100
	for {
		res, _, err := client.SecretsApi.ListSecrets(ctx).Limit(fmt.Sprintf("%d", limit)).Offset(fmt.Sprintf("%d", offset)).Execute()
		if err != nil {
			fatalApiError(err)
		}
		for _, a := range *res.Secrets {
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

func ResolveSecretShortID(ctx context.Context, client *koyeb.APIClient, id string) string {
	if len(id) == 8 {
		// TODO do a real cache
		cache := buildSecretShortIDCache(ctx, client)
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

func buildServiceShortIDCache(ctx context.Context, client *koyeb.APIClient) map[string][]string {
	c := make(map[string][]string)

	page := 0
	offset := 0
	limit := 100
	for {
		res, _, err := client.ServicesApi.ListServices(ctx).Limit(fmt.Sprintf("%d", limit)).Offset(fmt.Sprintf("%d", offset)).Execute()
		if err != nil {
			fatalApiError(err)
		}
		for _, a := range *res.Services {
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

func ResolveServiceShortID(ctx context.Context, client *koyeb.APIClient, id string) string {
	if len(id) == 8 {
		// TODO do a real cache
		cache := buildServiceShortIDCache(ctx, client)
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
