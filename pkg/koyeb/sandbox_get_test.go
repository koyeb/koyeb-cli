package koyeb

import (
	"net/http"
	"strings"
	"testing"

	"github.com/koyeb/koyeb-api-client-go/api/v1/koyeb"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// sandboxFixtureServiceID is the service UUID wired by sandboxFixtureAPI.
const sandboxFixtureServiceID = "323e4567-e89b-42d3-a456-426614174000"

// sandboxSubCmd resolves a sandbox subcommand against the sandbox command
// tree the way cobra does at execution time.
func sandboxSubCmd(t *testing.T, name string) *cobra.Command {
	t.Helper()
	cmd, _, err := NewSandboxCmd().Find([]string{name})
	require.NoError(t, err, "sandbox must register the %s subcommand", name)
	require.Equal(t, name, cmd.Name(), "sandbox must register the %s subcommand", name)
	return cmd
}

func TestSandboxGetCmdRegistered(t *testing.T) {
	cmd := sandboxSubCmd(t, "get")

	assert.Equal(t, "get NAME", cmd.Use)
	assert.NotEmpty(t, cmd.Short, "sandbox get must document what it prints")
	assert.NotNil(t, cmd.ValidArgsFunction,
		"sandbox get must wire dynamic completion like the neighboring commands")
	require.NotNil(t, cmd.Args, "sandbox get must validate its positional arguments")
	assert.Error(t, cmd.Args(cmd, nil), "sandbox get takes exactly one argument")
	assert.NoError(t, cmd.Args(cmd, []string{"myapp/mysandbox"}))
}

func TestSandboxGetFlow(t *testing.T) {
	t.Run("renders the sandbox through the renderer", func(t *testing.T) {
		fake := sandboxFixtureAPI()
		status := koyeb.SERVICESTATUS_HEALTHY
		fake.service.Status = &status
		ctx := sandboxTestContext(fake) // JSON renderer

		stdout, _ := captureStd(t, func() {
			require.NoError(t, NewSandboxHandler().Get(ctx, sandboxSubCmd(t, "get"), []string{sandboxFixtureServiceID}))
		})

		// The reply honors --output: the JSON renderer prints the marshaled
		// service, not a hand-rolled format.
		assert.Contains(t, stdout, `"id":"`+sandboxFixtureServiceID+`"`)
		assert.Contains(t, stdout, `"status":"HEALTHY"`)
	})

	t.Run("rejects non-sandbox services", func(t *testing.T) {
		fake := sandboxFixtureAPI()
		webType := koyeb.SERVICETYPE_WEB
		fake.service.Type = &webType
		ctx := sandboxTestContext(fake)

		err := NewSandboxHandler().Get(ctx, sandboxSubCmd(t, "get"), []string{sandboxFixtureServiceID})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not a sandbox")
	})

	t.Run("reports why the connection info is unavailable", func(t *testing.T) {
		fake := sandboxFixtureAPI()
		fake.service.ActiveDeploymentId = nil
		fake.service.LatestDeploymentId = nil
		ctx := sandboxTestContext(fake)

		err := NewSandboxHandler().Get(ctx, sandboxSubCmd(t, "get"), []string{sandboxFixtureServiceID})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "no deployment")
	})
}

// TestSandboxGetReplyTable pins the default table row of `sandbox get`: the
// service status together with the executor connection information.
func TestSandboxGetReplyTable(t *testing.T) {
	t.Run("prints the executor connection info from the deployment metadata", func(t *testing.T) {
		fake := sandboxFixtureAPI()
		status := koyeb.SERVICESTATUS_HEALTHY
		fake.service.Status = &status
		fake.service.Name = koyeb.PtrString("mysandbox")
		ctx := sandboxTestContext(fake)

		info, err := fetchSandboxInfo(ctx, sandboxFixtureServiceID)
		require.NoError(t, err)
		// full=true keeps the ID and app untouched, so the reply needs no
		// mapper lookups here.
		reply := NewGetSandboxReply(ctx.Mapper, *fake.service, info, true)

		assert.Equal(t, "Sandbox", reply.Title())
		headers := reply.Headers()
		assert.Contains(t, headers, "status")
		assert.Contains(t, headers, "url")
		assert.Contains(t, headers, "routing_key")

		fields := reply.Fields()
		require.Len(t, fields, 1)
		assert.Equal(t, "HEALTHY", fields[0]["status"])
		assert.Equal(t, "https://myapp-myapp.koyeb.app/koyeb-sandbox", fields[0]["url"])
		assert.Equal(t, "routing-key", fields[0]["routing_key"])
		assert.Equal(t, sandboxFixtureServiceID, fields[0]["id"], "--full keeps the full service ID")
		assert.Equal(t, "mysandbox", fields[0]["name"])
	})

	t.Run("falls back to the legacy domain URL without deployment metadata", func(t *testing.T) {
		fake := sandboxFixtureAPI()
		deployment := fake.deployments["123e4567-e89b-42d3-a456-426614174000"]
		deployment.Metadata = nil
		fake.deployments["123e4567-e89b-42d3-a456-426614174000"] = deployment
		ctx := sandboxTestContext(fake)

		info, err := fetchSandboxInfo(ctx, sandboxFixtureServiceID)
		require.NoError(t, err)
		reply := NewGetSandboxReply(ctx.Mapper, *fake.service, info, true)

		fields := reply.Fields()
		require.Len(t, fields, 1)
		assert.Equal(t, "https://myapp-myapp.koyeb.app/koyeb-sandbox", fields[0]["url"])
		assert.Empty(t, fields[0]["routing_key"], "no routing key must be shown when the API provides none")
	})
}

// TestSandboxInfoConnectionURLMatchesNewClient pins the invariant that the
// URL the get command prints is the URL the executor client dials: both
// derive their URL selection from the same SandboxInfo state.
func TestSandboxInfoConnectionURLMatchesNewClient(t *testing.T) {
	fromMetadata := &SandboxInfo{
		ServiceID:     "svc",
		AppID:         "app",
		Domain:        "myapp-myapp.koyeb.app",
		SandboxSecret: "secret",
		BaseURL:       "https://myapp-myapp.koyeb.app/koyeb-sandbox",
		RoutingKey:    "routing-key",
	}
	assert.Equal(t, fromMetadata.BaseURL, fromMetadata.ConnectionURL())
	assert.Equal(t, strings.TrimRight(fromMetadata.ConnectionURL(), "/"), fromMetadata.NewClient().baseURL)

	legacy := &SandboxInfo{
		ServiceID:     "svc",
		AppID:         "app",
		Domain:        "myapp-myapp.koyeb.app",
		SandboxSecret: "secret",
	}
	assert.Equal(t, "https://myapp-myapp.koyeb.app/koyeb-sandbox", legacy.ConnectionURL())
	assert.Equal(t, strings.TrimRight(legacy.ConnectionURL(), "/"), legacy.NewClient().baseURL)
}

// TestSandboxGetResolvesNameAndShortID pins the resolution contract end to
// end: both identifiers resolve through the real id mapper against a fake
// API server, the way the binary runs in a shell, and the default table
// output shows the status together with the executor connection info.
func TestSandboxGetResolvesNameAndShortID(t *testing.T) {
	for _, identifier := range []string{"myapp/mysandbox", "323e4567"} {
		t.Run("resolves "+identifier, func(t *testing.T) {
			server := newKoyebAPITestServer(t, sandboxAPIChainHandler(t, nil))
			args := append(koyebAPITestConfigArgs(t, server), "sandbox", "get", identifier)

			var exitCode int
			stdout, _ := captureStd(t, func() { exitCode = Run(args) })

			require.Equal(t, 0, exitCode, "sandbox get %s must succeed", identifier)
			assert.Contains(t, stdout, "HEALTHY")
			assert.Contains(t, stdout, "https://myapp-myapp.koyeb.app/koyeb-sandbox")
			assert.Contains(t, stdout, "routing-key")
		})
	}
}

// sandboxAPIChainHandler serves the API chain backing one sandbox: the
// service listing (id mapper), the app listing (app name resolution), the
// sandbox service itself, its app and deployment, and the service deletion.
// When deleted is non-nil it records that the deletion was requested.
func sandboxAPIChainHandler(t *testing.T, deleted *bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/v1/services":
			writeJSON(t, w, http.StatusOK, map[string]any{
				"services": []any{map[string]any{
					"id":     sandboxFixtureServiceID,
					"name":   "mysandbox",
					"app_id": "223e4567-e89b-42d3-a456-426614174000",
					"type":   "SANDBOX",
				}},
				"count": 1,
			})
		case r.Method == http.MethodGet && r.URL.Path == "/v1/apps":
			writeJSON(t, w, http.StatusOK, map[string]any{
				"apps": []any{map[string]any{
					"id":   "223e4567-e89b-42d3-a456-426614174000",
					"name": "myapp",
				}},
				"count": 1,
			})
		case r.Method == http.MethodGet && r.URL.Path == "/v1/services/"+sandboxFixtureServiceID:
			writeJSON(t, w, http.StatusOK, map[string]any{
				"service": map[string]any{
					"id":                   sandboxFixtureServiceID,
					"name":                 "mysandbox",
					"app_id":               "223e4567-e89b-42d3-a456-426614174000",
					"type":                 "SANDBOX",
					"status":               "HEALTHY",
					"active_deployment_id": "123e4567-e89b-42d3-a456-426614174000",
				},
			})
		case r.Method == http.MethodGet && r.URL.Path == "/v1/apps/223e4567-e89b-42d3-a456-426614174000":
			writeJSON(t, w, http.StatusOK, map[string]any{
				"app": map[string]any{
					"id":   "223e4567-e89b-42d3-a456-426614174000",
					"name": "myapp",
					"domains": []any{map[string]any{
						"name": "myapp-myapp.koyeb.app",
					}},
				},
			})
		case r.Method == http.MethodGet && r.URL.Path == "/v1/deployments/123e4567-e89b-42d3-a456-426614174000":
			writeJSON(t, w, http.StatusOK, map[string]any{
				"deployment": map[string]any{
					"id": "123e4567-e89b-42d3-a456-426614174000",
					"definition": map[string]any{
						"env": []any{map[string]any{
							"key":   SandboxSecretKey,
							"value": "sandbox-secret",
						}},
					},
					"metadata": map[string]any{
						"sandbox": map[string]any{
							"public_url":  "https://myapp-myapp.koyeb.app",
							"routing_key": "routing-key",
						},
					},
				},
			})
		case r.Method == http.MethodDelete && r.URL.Path == "/v1/services/"+sandboxFixtureServiceID:
			if deleted != nil {
				*deleted = true
			}
			writeJSON(t, w, http.StatusOK, map[string]any{})
		default:
			t.Errorf("unexpected API call: %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}
}
