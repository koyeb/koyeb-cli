package idmapper

import (
	"context"
	"strconv"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/errors"
)

type DatabaseMapper struct {
	ctx     context.Context
	client  *koyeb.APIClient
	fetched bool
	sidMap  *IDMap
	nameMap *IDMap
}

func NewDatabaseMapper(ctx context.Context, client *koyeb.APIClient) *DatabaseMapper {
	return &DatabaseMapper{
		ctx:     ctx,
		client:  client,
		fetched: false,
		sidMap:  NewIDMap(),
		nameMap: NewIDMap(),
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
		[]string{"database full UUID", "database short ID (8 characters)", "database name"},
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
			mapper.nameMap.Set(service.GetId(), service.GetName())
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
