package idmapper

import (
	"context"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
)

type Mapper struct {
	app          *AppMapper
	domain       *DomainMapper
	service      *ServiceMapper
	deployment   *DeploymentMapper
	regional     *RegionalDeploymentMapper
	instance     *InstanceMapper
	secret       *SecretMapper
	organization *OrganizationMapper
	database     *DatabaseMapper
	volume       *VolumeMapper
	snapshot     *SnapshotMapper
}

func NewMapper(ctx context.Context, client *koyeb.APIClient) *Mapper {
	appMapper := NewAppMapper(ctx, client)
	domainMapper := NewDomainMapper(ctx, client)
	serviceMapper := NewServiceMapper(ctx, client, appMapper)
	deploymentMapper := NewDeploymentMapper(ctx, client)
	regionalMapper := NewRegionalDeploymentMapper(ctx, client)
	instanceMapper := NewInstanceMapper(ctx, client)
	secretMapper := NewSecretMapper(ctx, client)
	organizationMapper := NewOrganizationMapper(ctx, client)
	databaseMapper := NewDatabaseMapper(ctx, client, appMapper)
	volumeMapper := NewVolumeMapper(ctx, client)
	snapshotMapper := NewSnapshotMapper(ctx, client)

	return &Mapper{
		app:          appMapper,
		domain:       domainMapper,
		service:      serviceMapper,
		deployment:   deploymentMapper,
		regional:     regionalMapper,
		instance:     instanceMapper,
		secret:       secretMapper,
		organization: organizationMapper,
		database:     databaseMapper,
		volume:       volumeMapper,
		snapshot:     snapshotMapper,
	}
}

func (mapper *Mapper) App() *AppMapper {
	return mapper.app
}

func (mapper *Mapper) Domain() *DomainMapper {
	return mapper.domain
}

func (mapper *Mapper) Service() *ServiceMapper {
	return mapper.service
}

func (mapper *Mapper) Deployment() *DeploymentMapper {
	return mapper.deployment
}

func (mapper *Mapper) RegionalDeployment() *RegionalDeploymentMapper {
	return mapper.regional
}

func (mapper *Mapper) Instance() *InstanceMapper {
	return mapper.instance
}

func (mapper *Mapper) Secret() *SecretMapper {
	return mapper.secret
}

func (mapper *Mapper) Organization() *OrganizationMapper {
	return mapper.organization
}

func (mapper *Mapper) Database() *DatabaseMapper {
	return mapper.database
}

func (mapper *Mapper) Volume() *VolumeMapper {
	return mapper.volume
}

func (mapper *Mapper) Snapshot() *SnapshotMapper {
	return mapper.snapshot
}
