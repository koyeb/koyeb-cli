package koyeb

import (
	"context"
	"net/http"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/idmapper"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/renderer"
)

// fakeAPI is the in-memory adapter for the koyebAPI port used by tests.
type fakeAPI struct {
	apps            []koyeb.App
	createdAppID    string
	createdApps     []string
	deletedApps     []string
	deletedServices []string

	servicesFetched []string
	service         *koyeb.Service

	createServiceReq   *koyeb.CreateService
	createServiceReply *koyeb.Service
	createServiceErr   error

	snapshot      *koyeb.InstanceSnapshot
	snapshotErr   error
	snapshots     []koyeb.InstanceSnapshot
	snapshotCalls int

	deployments map[string]koyeb.Deployment

	claimReply *koyeb.PoolClaimReply
	claimErr   error
}

func (f *fakeAPI) GetService(_ context.Context, serviceID string) (*koyeb.GetServiceReply, *http.Response, error) {
	f.servicesFetched = append(f.servicesFetched, serviceID)
	if f.service == nil {
		return &koyeb.GetServiceReply{}, nil, nil
	}
	return &koyeb.GetServiceReply{Service: f.service}, nil, nil
}

func (f *fakeAPI) CreateService(_ context.Context, req koyeb.CreateService) (*koyeb.CreateServiceReply, *http.Response, error) {
	stored := req
	f.createServiceReq = &stored
	if f.createServiceErr != nil {
		return nil, nil, f.createServiceErr
	}
	service := f.createServiceReply
	if service == nil {
		id := "323e4567-e89b-42d3-a456-426614174000"
		service = &koyeb.Service{Id: &id}
	}
	return &koyeb.CreateServiceReply{Service: service}, nil, nil
}

func (f *fakeAPI) GetApp(_ context.Context, _ string) (*koyeb.GetAppReply, *http.Response, error) {
	return &koyeb.GetAppReply{}, nil, nil
}

func (f *fakeAPI) CreateApp(_ context.Context, req koyeb.CreateApp) (*koyeb.CreateAppReply, *http.Response, error) {
	f.createdApps = append(f.createdApps, req.GetName())
	id := f.createdAppID
	return &koyeb.CreateAppReply{App: &koyeb.App{Id: &id}}, nil, nil
}

func (f *fakeAPI) ListApps(_ context.Context, _, _, _ string) (*koyeb.ListAppsReply, *http.Response, error) {
	apps := make([]koyeb.AppListItem, 0, len(f.apps))
	for _, app := range f.apps {
		apps = append(apps, koyeb.AppListItem{Id: app.Id, Name: app.Name})
	}
	return &koyeb.ListAppsReply{Apps: apps, Count: koyeb.PtrInt64(int64(len(apps)))}, nil, nil
}

func (f *fakeAPI) DeleteApp(_ context.Context, appID string) (*http.Response, error) {
	f.deletedApps = append(f.deletedApps, appID)
	return nil, nil
}

func (f *fakeAPI) DeleteService(_ context.Context, serviceID string) (*http.Response, error) {
	f.deletedServices = append(f.deletedServices, serviceID)
	return nil, nil
}

func (f *fakeAPI) GetDeployment(_ context.Context, deploymentID string) (*koyeb.GetDeploymentReply, *http.Response, error) {
	if d, ok := f.deployments[deploymentID]; ok {
		return &koyeb.GetDeploymentReply{Deployment: &d}, nil, nil
	}
	return &koyeb.GetDeploymentReply{}, nil, nil
}

func (f *fakeAPI) ListDeploymentsByService(_ context.Context, _, _ string) (*koyeb.ListDeploymentsReply, *http.Response, error) {
	return &koyeb.ListDeploymentsReply{}, nil, nil
}

func (f *fakeAPI) GetInstanceSnapshot(_ context.Context, _ string) (*koyeb.GetInstanceSnapshotReply, *http.Response, error) {
	f.snapshotCalls++
	if f.snapshotErr != nil {
		return nil, nil, f.snapshotErr
	}
	if f.snapshot == nil {
		return &koyeb.GetInstanceSnapshotReply{}, nil, nil
	}
	return &koyeb.GetInstanceSnapshotReply{InstanceSnapshot: f.snapshot}, nil, nil
}

func (f *fakeAPI) ListInstanceSnapshotsByName(_ context.Context, _ string) (*koyeb.ListInstanceSnapshotsReply, *http.Response, error) {
	return &koyeb.ListInstanceSnapshotsReply{InstanceSnapshots: f.snapshots}, nil, nil
}

func (f *fakeAPI) Claim(_ context.Context, _ koyeb.PoolClaimRequest) (*koyeb.PoolClaimReply, *http.Response, error) {
	if f.claimErr != nil {
		return nil, nil, f.claimErr
	}
	if f.claimReply == nil {
		return koyeb.NewPoolClaimReply(), nil, nil
	}
	return f.claimReply, nil, nil
}

func sandboxTestContext(fake *fakeAPI) *CLIContext {
	return &CLIContext{
		Context:  context.Background(),
		API:      fake,
		Mapper:   idmapper.NewMapper(context.Background(), nil),
		Renderer: renderer.NewRenderer(renderer.JSONFormat),
	}
}
