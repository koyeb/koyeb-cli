package idmapper

import (
	"context"
	"fmt"
	"strconv"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/errors"
)

type ServiceMapper struct {
	ctx       context.Context
	client    *koyeb.APIClient
	appMapper *AppMapper
	fetched   bool
	sidMap    *IDMap
	slugMap   *IDMap
}

func NewServiceMapper(ctx context.Context, client *koyeb.APIClient, appMapper *AppMapper) *ServiceMapper {
	return &ServiceMapper{
		ctx:       ctx,
		client:    client,
		appMapper: appMapper,
		fetched:   false,
		sidMap:    NewIDMap(),
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

	id, ok = mapper.slugMap.GetValue(val)
	if ok {
		return id, nil
	}
	return "", errors.NewCLIErrorForMapperResolve(
		"service",
		val,
		[]string{"service full UUID", "service short ID (8 characters)", "the service name prefixed by the application name and a slash (e.g. my-app/my-service)"},
	)
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
		return "", errors.NewCLIErrorForMapperResolve(
			"service",
			id,
			[]string{"service full UUID", "service short ID (8 characters)", "the service name prefixed by the application name and a slash (e.g. my-app/my-service)"},
		)
	}
	return slug, nil
}

func (mapper *ServiceMapper) fetch() error {
	radix := NewRadixTree()

	page := int64(0)
	offset := int64(0)
	limit := int64(100)
	for {

		res, resp, err := mapper.client.ServicesApi.ListServices(mapper.ctx).
			Limit(strconv.FormatInt(limit, 10)).
			Offset(strconv.FormatInt(offset, 10)).
			Execute()
		if err != nil {
			return errors.NewCLIErrorFromAPIError(
				"Error listing services to resolve the provided identifier to an object ID",
				err,
				resp,
			)
		}

		services := res.GetServices()
		for i := range services {
			service := &services[i]
			radix.Insert(getKey(service.GetId()), service)
		}

		page++
		offset = page * limit
		if offset >= res.GetCount() {
			break
		}
	}

	minLength := radix.MinimalLength(8)
	err := radix.ForEach(func(key Key, value Value) error {
		service := value.(*koyeb.ServiceListItem)

		appName, err := mapper.appMapper.GetName(service.GetAppId())
		if err != nil {
			return err
		}

		mapper.sidMap.Set(service.GetId(), getShortID(service.GetId(), minLength))

		// Possible values:
		// <app_name>/<service_id>
		// <app_id>/<service_id>
		// <app_short_id>/<service_id>
		//
		// <app_name>/<short_service_id>
		// <app_id>/<short_service_id>
		// <app_short_id>/<short_service_id>
		//
		// <app_name>/<service_name>
		// <app_id>/<service_name>
		// <app_short_id>/<service_name>
		for _, key := range []string{
			fmt.Sprint(appName, "/", service.GetId()),
			fmt.Sprint(service.GetAppId(), "/", service.GetId()),
			fmt.Sprint(service.GetAppId()[:8], "/", service.GetId()),

			fmt.Sprint(appName, "/", getShortID(service.GetId(), minLength)),
			fmt.Sprint(service.GetAppId(), "/", getShortID(service.GetId(), minLength)),
			fmt.Sprint(service.GetAppId()[:8], "/", getShortID(service.GetId(), minLength)),

			fmt.Sprint(appName, "/", service.GetName()),
			fmt.Sprint(service.GetAppId(), "/", service.GetName()),
			fmt.Sprint(service.GetAppId()[:8], "/", service.GetName()),
		} {
			mapper.slugMap.Set(key, service.GetId())
		}
		return nil
	})
	if err != nil {
		return err
	}

	mapper.fetched = true

	return nil
}
