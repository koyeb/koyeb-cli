package koyeb

import (
	"context"
	"net/http"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
)

// koyebAPI is the narrow port over the control-plane client used by the
// command flows. The real adapter wraps *koyeb.APIClient; tests use fakes —
// two adapters, one seam.
type koyebAPI interface {
	GetService(ctx context.Context, serviceID string) (*koyeb.GetServiceReply, *http.Response, error)
	CreateService(ctx context.Context, req koyeb.CreateService) (*koyeb.CreateServiceReply, *http.Response, error)
	GetApp(ctx context.Context, appID string) (*koyeb.GetAppReply, *http.Response, error)
	CreateApp(ctx context.Context, req koyeb.CreateApp) (*koyeb.CreateAppReply, *http.Response, error)
	ListApps(ctx context.Context, name, offset, limit string) (*koyeb.ListAppsReply, *http.Response, error)
	DeleteApp(ctx context.Context, appID string) (*http.Response, error)
	DeleteService(ctx context.Context, serviceID string) (*http.Response, error)
	GetDeployment(ctx context.Context, deploymentID string) (*koyeb.GetDeploymentReply, *http.Response, error)
	ListDeploymentsByService(ctx context.Context, serviceID, limit string) (*koyeb.ListDeploymentsReply, *http.Response, error)
	GetInstanceSnapshot(ctx context.Context, id string) (*koyeb.GetInstanceSnapshotReply, *http.Response, error)
	ListInstanceSnapshotsByName(ctx context.Context, name string) (*koyeb.ListInstanceSnapshotsReply, *http.Response, error)
	Claim(ctx context.Context, req koyeb.PoolClaimRequest) (*koyeb.PoolClaimReply, *http.Response, error)
}

// apiPort adapts the generated control-plane client to the koyebAPI port.
type apiPort struct {
	client *koyeb.APIClient
}

func (a apiPort) GetService(ctx context.Context, serviceID string) (*koyeb.GetServiceReply, *http.Response, error) {
	return a.client.ServicesApi.GetService(ctx, serviceID).Execute()
}

func (a apiPort) CreateService(ctx context.Context, req koyeb.CreateService) (*koyeb.CreateServiceReply, *http.Response, error) {
	return a.client.ServicesApi.CreateService(ctx).Service(req).Execute()
}

func (a apiPort) GetApp(ctx context.Context, appID string) (*koyeb.GetAppReply, *http.Response, error) {
	return a.client.AppsApi.GetApp(ctx, appID).Execute()
}

func (a apiPort) CreateApp(ctx context.Context, req koyeb.CreateApp) (*koyeb.CreateAppReply, *http.Response, error) {
	return a.client.AppsApi.CreateApp(ctx).App(req).Execute()
}

func (a apiPort) ListApps(ctx context.Context, name, offset, limit string) (*koyeb.ListAppsReply, *http.Response, error) {
	return a.client.AppsApi.ListApps(ctx).Name(name).Offset(offset).Limit(limit).Execute()
}

func (a apiPort) DeleteApp(ctx context.Context, appID string) (*http.Response, error) {
	_, resp, err := a.client.AppsApi.DeleteApp(ctx, appID).Execute()
	return resp, err
}

func (a apiPort) DeleteService(ctx context.Context, serviceID string) (*http.Response, error) {
	_, resp, err := a.client.ServicesApi.DeleteService(ctx, serviceID).Execute()
	return resp, err
}

func (a apiPort) GetDeployment(ctx context.Context, deploymentID string) (*koyeb.GetDeploymentReply, *http.Response, error) {
	return a.client.DeploymentsApi.GetDeployment(ctx, deploymentID).Execute()
}

func (a apiPort) ListDeploymentsByService(ctx context.Context, serviceID, limit string) (*koyeb.ListDeploymentsReply, *http.Response, error) {
	return a.client.DeploymentsApi.ListDeployments(ctx).Limit(limit).ServiceId(serviceID).Execute()
}

func (a apiPort) GetInstanceSnapshot(ctx context.Context, id string) (*koyeb.GetInstanceSnapshotReply, *http.Response, error) {
	return a.client.InstanceSnapshotsApi.GetInstanceSnapshot(ctx, id).Execute()
}

func (a apiPort) ListInstanceSnapshotsByName(ctx context.Context, name string) (*koyeb.ListInstanceSnapshotsReply, *http.Response, error) {
	return a.client.InstanceSnapshotsApi.ListInstanceSnapshots(ctx).Name(name).Execute()
}

func (a apiPort) Claim(ctx context.Context, req koyeb.PoolClaimRequest) (*koyeb.PoolClaimReply, *http.Response, error) {
	return a.client.PoolClaimsApi.Claim(ctx).Body(req).Execute()
}
