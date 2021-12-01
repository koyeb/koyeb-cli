package idmapper

import (
	"context"
	"fmt"
	"strconv"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
)

// ServiceMapper is a resolver that translate a service id to a name, and vice-versa.
type ServiceMapper struct {
	ctx    context.Context
	client *koyeb.APIClient
	cache  map[string]*Cache
}

// NewServiceMapper creates a new instance.
func NewServiceMapper(ctx context.Context, client *koyeb.APIClient) *ServiceMapper {
	return &ServiceMapper{
		ctx:    ctx,
		client: client,
		cache:  map[string]*Cache{},
	}
}

// GetID translates a name to an id.
func (mapper *ServiceMapper) GetID(appID string, name string) (string, error) {
	if IsUUIDv4(name) {
		return name, nil
	}

	cache, ok := mapper.cache[appID]
	if ok {
		id, ok := cache.GetID(name)
		if ok {
			return id, nil
		}
	}

	err := mapper.fetchServices(appID)
	if err != nil {
		return "", err
	}

	id, ok := mapper.cache[appID].GetID(name)
	if !ok {
		return "", fmt.Errorf("cannot find service information: %q not found", name)
	}

	return id, nil
}

// GetName translates an id to a name.
func (mapper *ServiceMapper) GetName(appID string, id string) (string, error) {
	if !IsUUIDv4(id) {
		return id, nil
	}

	cache, ok := mapper.cache[appID]
	if ok {
		name, ok := cache.GetName(id)
		if ok {
			return name, nil
		}
	}

	err := mapper.fetchServices(appID)
	if err != nil {
		return "", err
	}

	name, ok := mapper.cache[appID].GetName(id)
	if !ok {
		return "", fmt.Errorf("cannot find service information: %q not found", id)
	}

	return name, nil
}

// fetchServices traverses all services for given app id to populate cache.
func (mapper *ServiceMapper) fetchServices(appID string) error {
	cache := NewCache()
	page := int64(0)
	offset := int64(0)
	limit := int64(100)
	for {

		resp, _, err := mapper.client.ServicesApi.ListServices(mapper.ctx).
			AppId(appID).
			Limit(strconv.FormatInt(limit, 10)).
			Offset(strconv.FormatInt(offset, 10)).
			Execute()
		if err != nil {
			return fmt.Errorf("cannot fetch services information: %w", err)
		}

		for _, service := range resp.GetServices() {
			cache.Set(service.GetId(), service.GetName())
		}

		page += 1
		offset = page * limit
		if offset >= resp.GetCount() {
			mapper.cache[appID] = cache
			return nil
		}
	}
}
