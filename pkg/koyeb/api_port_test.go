package koyeb

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/idmapper"
	"github.com/koyeb/koyeb-cli/pkg/koyeb/renderer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeAPI is the in-memory adapter for the koyebAPI port used by tests.
type fakeAPI struct {
	apps            []koyeb.App
	createdAppID    string
	createdApps     []string
	deletedApps     []string
	deletedServices []string
	app             *koyeb.App

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

func (f *fakeAPI) CreateService(_ context.Context, req koyeb.CreateService) (
	*koyeb.CreateServiceReply, *http.Response, error) {
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
	if f.app == nil {
		return &koyeb.GetAppReply{}, nil, nil
	}
	return &koyeb.GetAppReply{App: f.app}, nil, nil
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

func (f *fakeAPI) GetDeployment(_ context.Context, deploymentID string) (
	*koyeb.GetDeploymentReply, *http.Response, error) {
	if d, ok := f.deployments[deploymentID]; ok {
		return &koyeb.GetDeploymentReply{Deployment: &d}, nil, nil
	}
	return &koyeb.GetDeploymentReply{}, nil, nil
}

func (f *fakeAPI) ListDeploymentsByService(_ context.Context, _, _ string) (
	*koyeb.ListDeploymentsReply, *http.Response, error) {
	return &koyeb.ListDeploymentsReply{}, nil, nil
}

func (f *fakeAPI) GetInstanceSnapshot(_ context.Context, _ string) (
	*koyeb.GetInstanceSnapshotReply, *http.Response, error) {
	f.snapshotCalls++
	if f.snapshotErr != nil {
		return nil, nil, f.snapshotErr
	}
	if f.snapshot == nil {
		return &koyeb.GetInstanceSnapshotReply{}, nil, nil
	}
	return &koyeb.GetInstanceSnapshotReply{InstanceSnapshot: f.snapshot}, nil, nil
}

func (f *fakeAPI) ListInstanceSnapshotsByName(_ context.Context, _ string) (
	*koyeb.ListInstanceSnapshotsReply, *http.Response, error) {
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

// sandboxServiceFixture wires a resolvable sandbox through the fake port:
// mapper resolution passes full UUIDs through without fetching, so the
// service/app/deployment chain below is all the resolution needs.
func sandboxFixtureAPI() *fakeAPI {
	serviceID := "323e4567-e89b-42d3-a456-426614174000"
	appID := "223e4567-e89b-42d3-a456-426614174000"
	deploymentID := "123e4567-e89b-42d3-a456-426614174000"
	secret := "sandbox-secret"
	domain := "myapp-myapp.koyeb.app"
	publicURL := "https://myapp-myapp.koyeb.app"
	routingKey := "routing-key"
	serviceType := koyeb.SERVICETYPE_SANDBOX
	publicURLField := publicURL

	app := koyeb.App{
		Id:      &appID,
		Domains: []koyeb.Domain{{Id: &domain, Name: &domain}},
	}
	deployment := koyeb.Deployment{
		Id: &deploymentID,
		Definition: &koyeb.DeploymentDefinition{
			Env: []koyeb.DeploymentEnv{{Key: koyeb.PtrString(SandboxSecretKey), Value: &secret}},
		},
		Metadata: &koyeb.DeploymentMetadata{
			Sandbox: &koyeb.SandboxMetadata{PublicUrl: &publicURLField, RoutingKey: &routingKey},
		},
	}
	fake := &fakeAPI{app: &app}
	fake.service = &koyeb.Service{
		Id:                 &serviceID,
		Type:               &serviceType,
		AppId:              &appID,
		ActiveDeploymentId: &deploymentID,
	}
	fake.apps = []koyeb.App{app}
	fake.deployments = map[string]koyeb.Deployment{deploymentID: deployment}
	return fake
}

func TestFetchSandboxInfo(t *testing.T) {
	t.Run("resolves the connection through the port", func(t *testing.T) {
		ctx := sandboxTestContext(sandboxFixtureAPI())

		info, err := fetchSandboxInfo(ctx, "323e4567-e89b-42d3-a456-426614174000")
		require.NoError(t, err)
		assert.Equal(t, "sandbox-secret", info.SandboxSecret)
		assert.Equal(t, "myapp-myapp.koyeb.app", info.Domain)
		assert.Equal(t, "https://myapp-myapp.koyeb.app/koyeb-sandbox", info.BaseURL)
		assert.Equal(t, "routing-key", info.RoutingKey)
	})

	t.Run("rejects non-sandbox services", func(t *testing.T) {
		fake := sandboxFixtureAPI()
		webType := koyeb.SERVICETYPE_WEB
		fake.service.Type = &webType
		ctx := sandboxTestContext(fake)

		_, err := fetchSandboxInfo(ctx, "323e4567-e89b-42d3-a456-426614174000")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not a sandbox")
	})

	t.Run("rejects missing sandbox secret", func(t *testing.T) {
		fake := sandboxFixtureAPI()
		deployment := fake.deployments["123e4567-e89b-42d3-a456-426614174000"]
		deployment.Definition.Env = nil
		fake.deployments["123e4567-e89b-42d3-a456-426614174000"] = deployment
		ctx := sandboxTestContext(fake)

		_, err := fetchSandboxInfo(ctx, "323e4567-e89b-42d3-a456-426614174000")
		require.Error(t, err)
		assert.Contains(t, err.Error(), SandboxSecretKey)
	})
}

func TestWithSandboxClient(t *testing.T) {
	t.Run("runs the operation against the resolved connection", func(t *testing.T) {
		ctx := sandboxTestContext(sandboxFixtureAPI())

		var gotInfo *SandboxInfo
		var clientType SandboxClientInterface
		err := withSandboxClient(ctx, "323e4567-e89b-42d3-a456-426614174000",
			func(client SandboxClientInterface, info *SandboxInfo) error {
				clientType = client
				gotInfo = info
				return nil
			})
		require.NoError(t, err)
		assert.NotNil(t, clientType, "the operation receives an executor client")
		assert.Equal(t, "https://myapp-myapp.koyeb.app/koyeb-sandbox", gotInfo.BaseURL)
	})

	t.Run("resolution errors propagate", func(t *testing.T) {
		fake := sandboxFixtureAPI()
		webType := koyeb.SERVICETYPE_WEB
		fake.service.Type = &webType
		ctx := sandboxTestContext(fake)

		err := withSandboxClient(ctx, "323e4567-e89b-42d3-a456-426614174000",
			func(SandboxClientInterface, *SandboxInfo) error {
				return nil
			})
		require.Error(t, err)
	})

	t.Run("operation errors propagate", func(t *testing.T) {
		ctx := sandboxTestContext(sandboxFixtureAPI())

		err := withSandboxClient(ctx, "323e4567-e89b-42d3-a456-426614174000",
			func(SandboxClientInterface, *SandboxInfo) error {
				return fmt.Errorf("op failed")
			})
		require.EqualError(t, err, "op failed")
	})
}
