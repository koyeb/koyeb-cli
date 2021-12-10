package idmapper2

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/pkg/errors"
)

type AppMapper struct {
	ctx        context.Context
	client     *koyeb.APIClient
	cache      map[string]*koyeb.AppListItem
	idCache    map[string]string
	shortCache map[string]string
	nameCache  map[string]string
}

func NewAppMapper(ctx context.Context, client *koyeb.APIClient) *AppMapper {
	return &AppMapper{
		ctx:        ctx,
		client:     client,
		cache:      map[string]*koyeb.AppListItem{},
		idCache:    map[string]string{},
		shortCache: map[string]string{},
		nameCache:  map[string]string{},
	}
}

func (mapper *AppMapper) ResolveID(val string) (string, error) {
	if IsUUIDv4(val) {
		return val, nil
	}

	if len(mapper.cache) == 0 {
		err := mapper.list()
		if err != nil {
			return "", err
		}
	}

	id, ok := mapper.shortCache[val]
	if ok {
		return id, nil
	}

	id, ok = mapper.nameCache[val]
	if ok {
		return id, nil
	}

	return "", fmt.Errorf("id not found %q", val)
}

func (mapper *AppMapper) GetShortID(id string) (string, error) {
	if len(mapper.cache) == 0 {
		err := mapper.list()
		if err != nil {
			return "", err
		}
	}

	sid, ok := mapper.idCache[id]
	if !ok {
		return "", fmt.Errorf("id not found %q", id)
	}

	return sid, nil
}

func (mapper *AppMapper) list() error {
	cache := map[string]*koyeb.AppListItem{}
	radix := NewRadixTree()

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

		apps := resp.GetApps()
		for i := range apps {
			app := &apps[i]
			id := app.GetId()
			radix.Insert(Key(strings.ReplaceAll(id, "-", "")))
			cache[id] = app
		}

		page++
		offset = page * limit
		if offset >= resp.GetCount() {
			break
		}
	}

	shortIDLength := radix.MinimalLength(8) + 3
	for id, app := range cache {
		sid := strings.ReplaceAll(id, "-", "")[:shortIDLength]
		name := app.GetName()

		mapper.cache[id] = app
		mapper.idCache[id] = sid
		mapper.shortCache[sid] = id
		mapper.nameCache[name] = id
	}

	return nil
}
