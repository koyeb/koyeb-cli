package koyeb

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKillAllProcessesKillsOnlyRunningAndContinuesOnFailure(t *testing.T) {
	client := &fakeSandboxClient{
		processes: []ProcessInfo{
			{ID: "proc-1", Status: "running"},
			{ID: "proc-2", Status: "stopped"},
			{ID: "proc-3", Status: "running"},
			{ID: "", Status: "running"},
		},
		killErrs: map[string]error{"proc-3": fmt.Errorf("executor down")},
	}

	killed, err := killAllProcesses(context.Background(), client)

	require.NoError(t, err)
	assert.Equal(t, 1, killed, "only the successful kill counts")
	assert.Equal(t, []string{"proc-1", "proc-3"}, client.killed,
		"a failed kill must not stop the remaining ones")
}

func TestKillAllProcessesListFailureAborts(t *testing.T) {
	client := &fakeSandboxClient{listErr: fmt.Errorf("list failed")}

	_, err := killAllProcesses(context.Background(), client)

	require.Error(t, err)
	assert.Empty(t, client.killed, "no kill may be attempted when listing fails")
}

func TestSandboxKillAllSubcommand(t *testing.T) {
	cmd, _, err := NewSandboxCmd().Find([]string{"kill-all"})
	require.NoError(t, err, "sandbox must register the kill-all subcommand")
	assert.Equal(t, "kill-all NAME", cmd.Use)
	assert.NoError(t, cmd.Args(nil, []string{"sb"}), "kill-all takes exactly one sandbox name")
	assert.Error(t, cmd.Args(nil, nil), "kill-all requires a sandbox name")
}

func TestSandboxKillAllHandler(t *testing.T) {
	t.Run("kills every running process and reports the count", func(t *testing.T) {
		var killIDs []string
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/koyeb-sandbox/list_processes":
				writeChecked(w, `{"processes":[
					{"id":"a","status":"running"},
					{"id":"b","status":"stopped"},
					{"id":"c","status":"running"}]}`)
			case "/koyeb-sandbox/kill_process":
				var req struct {
					ID string `json:"id"`
				}
				body, _ := io.ReadAll(r.Body)
				require.NoError(t, json.Unmarshal(body, &req))
				killIDs = append(killIDs, req.ID)
				writeChecked(w, `{"success":true}`)
			default:
				http.NotFound(w, r)
			}
		}))
		t.Cleanup(server.Close)

		ctx := sandboxTestContext(sandboxFixtureAPIWithURL(server.URL))
		cmd, _, err := NewSandboxCmd().Find([]string{"kill-all"})
		require.NoError(t, err)

		var handlerErr error
		out, _ := captureStd(t, func() {
			handlerErr = NewSandboxHandler().KillAllProcesses(ctx, cmd, []string{"323e4567-e89b-42d3-a456-426614174000"})
		})

		require.NoError(t, handlerErr)
		assert.Equal(t, []string{"a", "c"}, killIDs, "only running processes are killed")
		assert.Contains(t, out, "Killed 2 processes")
	})

	t.Run("reports when nothing is running", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/koyeb-sandbox/list_processes" {
				writeChecked(w, `{"processes":[]}`)
				return
			}
			http.NotFound(w, r)
		}))
		t.Cleanup(server.Close)

		ctx := sandboxTestContext(sandboxFixtureAPIWithURL(server.URL))
		cmd, _, err := NewSandboxCmd().Find([]string{"kill-all"})
		require.NoError(t, err)

		var handlerErr error
		out, _ := captureStd(t, func() {
			handlerErr = NewSandboxHandler().KillAllProcesses(ctx, cmd, []string{"323e4567-e89b-42d3-a456-426614174000"})
		})

		require.NoError(t, handlerErr)
		assert.Contains(t, out, "No running background processes")
	})
}
