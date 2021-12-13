package idmapper2

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/pkg/errors"
)

type InstanceMapper struct {
	ctx     context.Context
	client  *koyeb.APIClient
	fetched bool
	sidMap  *IDMap
}

func NewInstanceMapper(ctx context.Context, client *koyeb.APIClient) *InstanceMapper {
	return &InstanceMapper{
		ctx:     ctx,
		client:  client,
		fetched: false,
		sidMap:  NewIDMap(),
	}
}

func (mapper *InstanceMapper) ResolveID(val string) (string, error) {
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

	return "", fmt.Errorf("id not found %q", val)
}

func (mapper *InstanceMapper) GetShortID(id string) (string, error) {
	if !mapper.fetched {
		err := mapper.fetch()
		if err != nil {
			return "", err
		}
	}

	sid, ok := mapper.sidMap.GetValue(id)
	if !ok {
		return "", fmt.Errorf("instance short id not found for %q", id)
	}

	return sid, nil
}

func (mapper *InstanceMapper) fetch() error {
	radix := NewRadixTree()

	page := int64(0)
	offset := int64(0)
	limit := int64(100)
	for {

		resp, _, err := mapper.client.InstancesApi.ListInstances(mapper.ctx).
			Limit(strconv.FormatInt(limit, 10)).
			Offset(strconv.FormatInt(offset, 10)).
			Execute()
		if err != nil {
			return errors.Wrap(err, "cannot list apps from API")
		}

		instances := resp.GetInstances()
		for i := range instances {
			instance := &instances[i]
			id := instance.GetId()
			radix.Insert(Key(strings.ReplaceAll(id, "-", "")), instance)
		}

		page++
		offset = page * limit
		if offset >= resp.GetCount() {
			break
		}
	}

	minLength := radix.MinimalLength(8)
	radix.ForEach(func(key Key, value Value) {
		app := value.(*koyeb.InstanceListItem)
		id := app.GetId()
		sid := strings.ReplaceAll(id, "-", "")[:minLength]

		mapper.sidMap.Set(id, sid)
	})

	mapper.fetched = true

	return nil
}
