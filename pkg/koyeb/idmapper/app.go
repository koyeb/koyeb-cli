package idmapper

import (
	"context"
	"fmt"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
)

// AppMapper is a resolver that translate an app id to a name, and vice-versa.
type AppMapper struct {
	ctx    context.Context
	client *koyeb.APIClient
	cache  *Cache
}

// NewAppMapper creates a new instance.
func NewAppMapper(ctx context.Context, client *koyeb.APIClient) *AppMapper {
	return &AppMapper{
		ctx:    ctx,
		client: client,
		cache:  NewCache(),
	}
}

// GetID translates a name to an id.
func (mapper *AppMapper) GetID(name string) (string, error) {
	if IsUUIDv4(name) {
		return name, nil
	}

	id, ok := mapper.cache.GetID(name)
	if ok {
		return id, nil
	}

	resp, _, err := mapper.client.AppsApi.GetApp(mapper.ctx, name).Execute()
	if err != nil {
		return "", fmt.Errorf("cannot fetch app information: %w", err)
	}

	id = resp.App.GetId()
	mapper.cache.Set(id, name)

	return id, nil
}

// GetName translates an id to a name.
func (mapper *AppMapper) GetName(id string) (string, error) {
	if !IsUUIDv4(id) {
		return id, nil
	}

	name, ok := mapper.cache.GetName(id)
	if ok {
		return name, nil
	}

	resp, _, err := mapper.client.AppsApi.GetApp(mapper.ctx, id).Execute()
	if err != nil {
		return "", fmt.Errorf("cannot fetch app information: %w", err)
	}

	name = resp.App.GetName()
	mapper.cache.Set(id, name)

	return name, nil
}
