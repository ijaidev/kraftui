package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/ijaidev/kraftui/config"
	"github.com/ijaidev/kraftui/internal/api"
	"github.com/ijaidev/kraftui/internal/kraft"
)

type fakeKraftRunner struct {
	result kraft.Result
	args   []string
}

func (r *fakeKraftRunner) Run(_ context.Context, _ string, args []string) (kraft.Result, error) {
	r.args = append([]string(nil), args...)
	return r.result, nil
}

func newTestServer(t *testing.T, runner *fakeKraftRunner) *Server {
	t.Helper()
	client, err := kraft.NewWithRunner(config.KraftConfig{
		Binary: "kraft", ExpectedVersion: "0.12.14", CommandTimeout: time.Second,
	}, "/test/kraft", runner)
	if err != nil {
		t.Fatalf("NewWithRunner() error = %v", err)
	}
	return NewServer(0, "test", "0.12.14", client)
}

func newTestHandler(t *testing.T, s *Server) http.Handler {
	t.Helper()
	swagger, err := api.GetSpec()
	if err != nil {
		t.Fatalf("GetSpec() error = %v", err)
	}
	swagger.Servers = nil
	return buildAPIRouter(s, swagger)
}

func TestHandleMachinesUsesDocumentedFilters(t *testing.T) {
	runner := &fakeKraftRunner{result: kraft.Result{Stdout: []byte(`[{"machine_id":"m1","name":"web","status":"running","plat":"qemu","arch":"x86_64"}]`)}}
	server := newTestServer(t, runner)
	handler := newTestHandler(t, server)

	request := httptest.NewRequest(http.MethodGet, "/api/v1/machines?all=false&platform=qemu&architecture=x86_64&long=true", nil)
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK || !strings.Contains(recorder.Body.String(), `"name":"web"`) {
		t.Fatalf("response = %d %s", recorder.Code, recorder.Body.String())
	}
	for _, argument := range []string{"ps", "--output", "json", "--plat", "qemu", "--arch", "x86_64", "--long"} {
		if !slices.Contains(runner.args, argument) {
			t.Fatalf("command args %q missing %q", runner.args, argument)
		}
	}
	if slices.Contains(runner.args, "--all") {
		t.Fatalf("command args %q unexpectedly include --all", runner.args)
	}
}

func TestHandlePackagesRejectsUndocumentedInput(t *testing.T) {
	server := newTestServer(t, &fakeKraftRunner{})
	handler := newTestHandler(t, server)

	request := httptest.NewRequest(http.MethodGet, "/api/v1/packages?update=true", nil)
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest || !strings.Contains(recorder.Body.String(), "invalid_query") {
		t.Fatalf("response = %d %s", recorder.Code, recorder.Body.String())
	}
}

func TestHandleHealthReturnsContractModel(t *testing.T) {
	server := newTestServer(t, &fakeKraftRunner{})
	handler := newTestHandler(t, server)

	request := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK || !strings.Contains(recorder.Body.String(), `"kraftVersion":"0.12.14"`) {
		t.Fatalf("response = %d %s", recorder.Code, recorder.Body.String())
	}
}

func TestHandlePackagesRejectsOutOfRangeLimit(t *testing.T) {
	server := newTestServer(t, &fakeKraftRunner{})
	handler := newTestHandler(t, server)

	request := httptest.NewRequest(http.MethodGet, "/api/v1/packages?limit=999", nil)
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest || !strings.Contains(recorder.Body.String(), "invalid_query") {
		t.Fatalf("response = %d %s", recorder.Code, recorder.Body.String())
	}
}
