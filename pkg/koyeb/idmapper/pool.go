package idmapper

import (
	"context"
	"strconv"
	"uuid"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/errors"
)

type PoolMapper struct {
	ctx     context.Context
	client  *koyeb.APIClient
	fetched bool
	sidMap  *IDMap
	nameMap *IDMap
}

func NewPoolMapper(ctx context.Context, client *koyeb.APIClient) *PoolMapper {
	return &PoolMapper{
		ctx:     ctx,
		client:  client,
		fetched: false,
		sidMap:  NewIDMap(),
		nameMap: NewIDMap(),
	}
}

func (mapper *PoolMapper) ResolveID(val string) (string, error) {
	if _, err := uuid.Parse(val); err == nil {
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
		"pool",
		val,
		[]string{"pool full UUID", "pool short ID (8 characters)", "pool name"},
	)
}

func (mapper *PoolMapper) fetch() error {
	radix := NewRadixTree()

	pools, err := FetchAllPages(func(offset, limit int64) ([]koyeb.ServicePool, int64, error) {
		res, resp, err := mapper.client.ServicePoolsApi.ListServicePools(mapper.ctx).
			Limit(strconv.FormatInt(limit, 10)).
			Offset(strconv.FormatInt(offset, 10)).
			Execute()
		if err != nil {
			return nil, 0, errors.NewCLIErrorFromAPIError(
				"Error listing service pools to resolve the provided identifier to an object ID",
				err,
				resp,
			)
		}
		return res.GetServicePools(), res.GetCount(), nil
	}, 100)
	if err != nil {
		return err
	}

	for i := range pools {
		pool := &pools[i]
		radix.Insert(getKey(pool.GetId()), pool)
	}

	minLength := radix.MinimalLength(8)
	err = radix.ForEach(func(key Key, value Value) error {
		pool := value.(*koyeb.ServicePool)
		id := pool.GetId()
		name := pool.GetName()
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
