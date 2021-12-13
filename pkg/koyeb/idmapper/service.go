package idmapper

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/pkg/errors"
)

type ServiceMapper struct {
	ctx       context.Context
	client    *koyeb.APIClient
	appMapper *AppMapper
	fetched   bool
	sidMap    *IDMap
	nameMap   *IDMap
	slugMap   *IDMap
}

func NewServiceMapper(ctx context.Context, client *koyeb.APIClient, appMapper *AppMapper) *ServiceMapper {
	return &ServiceMapper{
		ctx:       ctx,
		client:    client,
		appMapper: appMapper,
		fetched:   false,
		sidMap:    NewIDMap(),
		nameMap:   NewIDMap(),
		slugMap:   NewIDMap(),
	}
}

func (mapper *ServiceMapper) ResolveID(val string) (string, error) {
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

	id, ok = mapper.slugMap.GetID(val)
	if ok {
		return id, nil
	}

	return "", fmt.Errorf("id not found %q", val)
}

func (mapper *ServiceMapper) GetShortID(id string) (string, error) {
	if !mapper.fetched {
		err := mapper.fetch()
		if err != nil {
			return "", err
		}
	}

	sid, ok := mapper.sidMap.GetValue(id)
	if !ok {
		return "", fmt.Errorf("service short id not found for %q", id)
	}

	return sid, nil
}

func (mapper *ServiceMapper) GetSlug(id string) (string, error) {
	if !mapper.fetched {
		err := mapper.fetch()
		if err != nil {
			return "", err
		}
	}

	slug, ok := mapper.slugMap.GetValue(id)
	if !ok {
		return "", fmt.Errorf("service slug not found for %q", id)
	}

	return slug, nil
}

func (mapper *ServiceMapper) fetch() error {
	radix := NewRadixTree()

	page := int64(0)
	offset := int64(0)
	limit := int64(100)
	for {

		resp, _, err := mapper.client.ServicesApi.ListServices(mapper.ctx).
			Limit(strconv.FormatInt(limit, 10)).
			Offset(strconv.FormatInt(offset, 10)).
			Execute()
		if err != nil {
			return errors.Wrap(err, "cannot list apps from API")
		}

		services := resp.GetServices()
		for i := range services {
			service := &services[i]
			id := service.GetId()
			radix.Insert(Key(strings.ReplaceAll(id, "-", "")), service)
		}

		page++
		offset = page * limit
		if offset >= resp.GetCount() {
			break
		}
	}

	minLength := radix.MinimalLength(8)
	radix.ForEach(func(key Key, value Value) {
		service := value.(*koyeb.ServiceListItem)
		id := service.GetId()
		name := service.GetName()
		sid := strings.ReplaceAll(id, "-", "")[:minLength]

		mapper.sidMap.Set(id, sid)
		mapper.nameMap.Set(id, name)

		appName, err := mapper.appMapper.GetName(service.GetAppId())
		if err == nil {
			slug := fmt.Sprint(appName, "/", name)
			mapper.slugMap.Set(id, slug)
		}
	})

	mapper.fetched = true

	return nil
}
