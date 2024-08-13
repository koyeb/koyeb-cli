package idmapper

import (
	"context"
	"fmt"
	"strconv"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/errors"
)

type DatabaseMapper struct {
	ctx       context.Context
	client    *koyeb.APIClient
	appMapper *AppMapper
	fetched   bool
	sidMap    *IDMap
	nameMap   *IDMap
}

func NewDatabaseMapper(ctx context.Context, client *koyeb.APIClient, appMapper *AppMapper) *DatabaseMapper {
	return &DatabaseMapper{
		ctx:       ctx,
		client:    client,
		appMapper: appMapper,
		fetched:   false,
		sidMap:    NewIDMap(),
		nameMap:   NewIDMap(),
	}
}

func (mapper *DatabaseMapper) ResolveID(val string) (string, error) {
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

	return "", errors.NewCLIErrorForMapperResolve(
		"database",
		val,
		[]string{"database full UUID", "service short ID (8 characters)", "the database name prefixed by the application name and a slash (e.g. my-app/my-database)"},
	)
}

func (mapper *DatabaseMapper) fetch() error {
	page := int64(0)
	offset := int64(0)
	limit := int64(100)
	for {
		res, resp, err := mapper.client.ServicesApi.ListServices(mapper.ctx).
			Types([]string{"DATABASE"}).
			Limit(strconv.FormatInt(limit, 10)).
			Offset(strconv.FormatInt(offset, 10)).
			Execute()
		if err != nil {
			return errors.NewCLIErrorFromAPIError(
				"Error listing databases to resolve the provided identifier to an object ID",
				err,
				resp,
			)
		}

		for _, service := range res.GetServices() {
			mapper.sidMap.Set(service.GetId(), getShortID(service.GetId(), 8))

			appName, err := mapper.appMapper.GetName(service.GetAppId())
			if err != nil {
				return err
			}

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

				fmt.Sprint(appName, "/", getShortID(service.GetId(), 8)),
				fmt.Sprint(service.GetAppId(), "/", getShortID(service.GetId(), 8)),
				fmt.Sprint(service.GetAppId()[:8], "/", getShortID(service.GetId(), 8)),

				fmt.Sprint(appName, "/", service.GetName()),
				fmt.Sprint(service.GetAppId(), "/", service.GetName()),
				fmt.Sprint(service.GetAppId()[:8], "/", service.GetName()),
			} {
				mapper.nameMap.Set(service.GetId(), key)
			}
		}

		page++
		offset = page * limit
		if offset >= res.GetCount() {
			break
		}
	}

	mapper.fetched = true
	return nil
}
