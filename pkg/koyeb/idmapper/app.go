package idmapper

import (
	"context"
	"fmt"
	"strconv"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/pkg/errors"
)

type AppMapper struct {
	ctx     context.Context
	client  *koyeb.APIClient
	fetched bool
	sidMap  *IDMap
	nameMap *IDMap
}

func NewAppMapper(ctx context.Context, client *koyeb.APIClient) *AppMapper {
	return &AppMapper{
		ctx:     ctx,
		client:  client,
		fetched: false,
		sidMap:  NewIDMap(),
		nameMap: NewIDMap(),
	}
}

func (mapper *AppMapper) ResolveID(val string) (string, error) {
	if IsUUIDv4(val) {
		return val, nil
	}

	if !mapper.fetched {
		err := mapper.fetch()
		if err != nil {
			return "", err
		}
	}

	id, ok := mapper.sidMap.GetID(val)
	if ok {
		return id, nil
	}

	id, ok = mapper.nameMap.GetID(val)
	if ok {
		return id, nil
	}

	return "", fmt.Errorf("id not found %q", val)
}

func (mapper *AppMapper) GetShortID(id string) (string, error) {
	if !mapper.fetched {
		err := mapper.fetch()
		if err != nil {
			return "", err
		}
	}

	sid, ok := mapper.sidMap.GetValue(id)
	if !ok {
		return "", fmt.Errorf("app short id not found for %q", id)
	}

	return sid, nil
}

func (mapper *AppMapper) GetName(id string) (string, error) {
	if !mapper.fetched {
		err := mapper.fetch()
		if err != nil {
			return "", err
		}
	}

	name, ok := mapper.nameMap.GetValue(id)
	if !ok {
		return "", fmt.Errorf("app name not found for %q", id)
	}

	return name, nil
}

func (mapper *AppMapper) fetch() error {
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
			radix.Insert(getKey(app.GetId()), app)
		}

		page++
		offset = page * limit
		if offset >= resp.GetCount() {
			break
		}
	}

	minLength := radix.MinimalLength(8)
	err := radix.ForEach(func(key Key, value Value) error {
		app := value.(*koyeb.AppListItem)
		id := app.GetId()
		name := app.GetName()
		sid := getShortID(id, minLength)

		mapper.sidMap.Set(id, sid)
		mapper.nameMap.Set(id, name)

		return nil
	})
	if err != nil {
		return err
	}

	mapper.fetched = true

	return nil
}
