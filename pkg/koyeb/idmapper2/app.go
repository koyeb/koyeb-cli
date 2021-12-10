package idmapper2

import (
	"context"
	"strconv"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/pkg/errors"
)

type AppMapper struct {
	ctx    context.Context
	client *koyeb.APIClient
}

func NewAppMapper(ctx context.Context, client *koyeb.APIClient) *AppMapper {
	return &AppMapper{
		ctx:    ctx,
		client: client,
	}
}

func (mapper *AppMapper) list() error {
	shortIDCache := map[string][]string{}

	page := int64(0)
	offset := int64(0)
	limit := int64(100)
	for {

		resp, _, err := mapper.client.AppsApi.ListApps(mapper.ctx).
			Limit(strconv.FormatInt(limit, 10)).
			Offset(strconv.FormatInt(offset, 10)).
			Execute()
		if err != nil {
			return errors.Wrap(err, "cannot list apps from API")
		}

		for _, app := range resp.GetApps() {
			id := app.GetId()
			shortID := id[:8]
			shortIDCache[shortID] = append(shortIDCache[shortID], id)
		}

		page++
		offset = page * limit
		if offset >= resp.GetCount() {
			break
		}
	}

	return nil
}
