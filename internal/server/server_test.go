package server

import (
	"context"
	"errors"
	"fmt"
	"net"
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
	err    error
}

func (r *fakeKraftRunner) Run(_ context.Context, _ string, args []string) (kraft.Result, error) {
	r.args = append([]string(nil), args...)
	return r.result, r.err
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

func TestHandleMachinesKraftFailure(t *testing.T) {
	runner := &fakeKraftRunner{err: errors.New("exit 1"), result: kraft.Result{Stderr: []byte("kraft: failed")}}
	handler := newTestHandler(t, newTestServer(t, runner))

	request := httptest.NewRequest(http.MethodGet, "/api/v1/machines", nil)
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadGateway || !strings.Contains(recorder.Body.String(), "kraft_command_failed") {
		t.Fatalf("response = %d %s", recorder.Code, recorder.Body.String())
	}
}

func TestHandleMachinesUnavailableWithoutClient(t *testing.T) {
	handler := newTestHandler(t, NewServer(0, "test", "0.12.14", nil))

	request := httptest.NewRequest(http.MethodGet, "/api/v1/machines", nil)
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadGateway || !strings.Contains(recorder.Body.String(), "kraft_unavailable") {
		t.Fatalf("response = %d %s", recorder.Code, recorder.Body.String())
	}
}

func TestGetListenerBindsConfiguredPort(t *testing.T) {
	probe, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Listen() error = %v", err)
	}
	port := probe.Addr().(*net.TCPAddr).Port
	if err := probe.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	listener, err := (&Server{port: port}).getListener()
	if err != nil {
		t.Fatalf("getListener() error = %v", err)
	}
	defer listener.Close()

	if got := listener.Addr().(*net.TCPAddr).Port; got != port {
		t.Fatalf("port = %d, want %d", got, port)
	}
}

func TestListenShutsDownOnCancel(t *testing.T) {
	probe, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Listen() error = %v", err)
	}
	port := probe.Addr().(*net.TCPAddr).Port
	if err := probe.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	server := newTestServer(t, &fakeKraftRunner{})
	server.port = port

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	errCh := make(chan error, 1)
	go func() {
		errCh <- server.Listen(ctx)
	}()

	url := fmt.Sprintf("http://127.0.0.1:%d/api/health", port)
	deadline := time.Now().Add(2 * time.Second)
	for {
		resp, err := http.Get(url)
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				break
			}
		}
		if time.Now().After(deadline) {
			cancel()
			t.Fatalf("server did not become ready: %v", err)
		}
		time.Sleep(20 * time.Millisecond)
	}

	cancel()
	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("Listen() = %v, want nil", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Listen() did not return after cancel")
	}
}
