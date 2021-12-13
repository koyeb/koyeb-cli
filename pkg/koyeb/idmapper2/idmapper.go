package idmapper2

import (
	"context"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
)

type Mapper struct {
	app     *AppMapper
	service *ServiceMapper
	secret  *SecretMapper
}

func NewMapper(ctx context.Context, client *koyeb.APIClient) *Mapper {
	appMapper := NewAppMapper(ctx, client)
	serviceMapper := NewServiceMapper(ctx, client, appMapper)
	secretMapper := NewSecretMapper(ctx, client)

	return &Mapper{
		app:     appMapper,
		service: serviceMapper,
		secret:  secretMapper,
	}
}

func (mapper *Mapper) App() *AppMapper {
	return mapper.app
}

func (mapper *Mapper) Service() *ServiceMapper {
	return mapper.service
}

func (mapper *Mapper) Secret() *SecretMapper {
	return mapper.secret
}
