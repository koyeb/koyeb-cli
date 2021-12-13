package idmapper2

import (
	"context"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
)

type Mapper struct {
	app        *AppMapper
	service    *ServiceMapper
	deployment *DeploymentMapper
	instance   *InstanceMapper
	secret     *SecretMapper
}

func NewMapper(ctx context.Context, client *koyeb.APIClient) *Mapper {
	appMapper := NewAppMapper(ctx, client)
	serviceMapper := NewServiceMapper(ctx, client, appMapper)
	deploymentMapper := NewDeploymentMapper(ctx, client)
	instanceMapper := NewInstanceMapper(ctx, client)
	secretMapper := NewSecretMapper(ctx, client)

	return &Mapper{
		app:        appMapper,
		service:    serviceMapper,
		deployment: deploymentMapper,
		instance:   instanceMapper,
		secret:     secretMapper,
	}
}

func (mapper *Mapper) App() *AppMapper {
	return mapper.app
}

func (mapper *Mapper) Service() *ServiceMapper {
	return mapper.service
}

func (mapper *Mapper) Deployment() *DeploymentMapper {
	return mapper.deployment
}

func (mapper *Mapper) Instance() *InstanceMapper {
	return mapper.instance
}

func (mapper *Mapper) Secret() *SecretMapper {
	return mapper.secret
}
