package kraft

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/ijaidev/kraftui/config"
	"github.com/ijaidev/kraftui/internal/api"
)

type fakeRunner struct {
	result Result
	err    error
	binary string
	args   []string
}

func (r *fakeRunner) Run(_ context.Context, binary string, args []string) (Result, error) {
	r.binary = binary
	r.args = append([]string(nil), args...)
	return r.result, r.err
}

func testClient(t *testing.T, runner *fakeRunner) *Client {
	t.Helper()
	client, err := NewWithRunner(config.KraftConfig{
		Binary:          "kraft",
		ExpectedVersion: "0.12.14",
		CommandTimeout:  time.Second,
	}, "/test/kraft", runner)
	if err != nil {
		t.Fatalf("NewWithRunner() error = %v", err)
	}
	return client
}

func TestVerifyChecksExpectedVersion(t *testing.T) {
	runner := &fakeRunner{result: Result{Stdout: []byte("kraft 0.12.14 (abc)\n")}}
	version, err := testClient(t, runner).Verify(context.Background())
	if err != nil || version != "0.12.14" {
		t.Fatalf("Verify() = %q, %v", version, err)
	}
	wantArgs := append(append([]string{}, baseArgs...), "version")
	if runner.binary != "/test/kraft" || !reflect.DeepEqual(runner.args, wantArgs) {
		t.Fatalf("command = %q %q, want %q %q", runner.binary, runner.args, "/test/kraft", wantArgs)
	}
}

func TestVerifyRejectsUnsupportedVersion(t *testing.T) {
	runner := &fakeRunner{result: Result{Stdout: []byte("kraft 0.13.0\n")}}
	if _, err := testClient(t, runner).Verify(context.Background()); err == nil {
		t.Fatal("Verify() error = nil, want unsupported version")
	}
}

func TestListMachinesNormalizesTableJSON(t *testing.T) {
	runner := &fakeRunner{result: Result{Stdout: []byte(`[{"name":"web","kernel":"kernel","args":"-v","created":"now","status":"running","mem":"64 MiB","ports":"8080:80","plat":"qemu/x86_64"}]`)}}
	all := true
	items, err := testClient(t, runner).ListMachines(context.Background(), api.ListMachinesParams{All: &all})
	if err != nil {
		t.Fatalf("ListMachines() error = %v", err)
	}
	if len(items) != 1 || items[0].Name != "web" || items[0].Platform != "qemu" || items[0].Architecture != "x86_64" || items[0].Id != nil {
		t.Fatalf("ListMachines() = %#v", items)
	}
	wantArgs := append(append([]string{}, baseArgs...), "ps", "--output", "json", "--all")
	if !reflect.DeepEqual(runner.args, wantArgs) {
		t.Fatalf("args = %q, want %q", runner.args, wantArgs)
	}
}

func TestListVolumesUsesLongIDAndParsesOutput(t *testing.T) {
	runner := &fakeRunner{result: Result{Stdout: []byte(`[{"driver":"9pfs","volume_name":"data","volume_id":"v1","status":"up","source":"/tmp/data"}]`)}}
	long := true
	items, err := testClient(t, runner).ListVolumes(context.Background(), api.ListVolumesParams{Long: &long})
	if err != nil || len(items) != 1 || items[0].Id == nil || *items[0].Id != "v1" {
		t.Fatalf("ListVolumes() = %#v, %v", items, err)
	}
}

func TestListPackagesAcceptsMultipleJSONBatches(t *testing.T) {
	runner := &fakeRunner{result: Result{Stdout: []byte(`[{"type":"app","name":"first","version":"1"}]
[{"type":"lib","name":"second","version":"2"}]`)}}
	items, err := testClient(t, runner).ListPackages(context.Background(), api.ListPackagesParams{})
	if err != nil || len(items) != 2 || items[0].Name != "first" || items[1].Name != "second" {
		t.Fatalf("ListPackages() = %#v, %v", items, err)
	}
}

func TestCommandErrorKeepsStderrOutOfJSON(t *testing.T) {
	runner := &fakeRunner{result: Result{Stderr: []byte("permission denied\n")}, err: errors.New("exit status 1")}
	_, err := testClient(t, runner).ListNetworks(context.Background(), api.ListNetworksParams{})
	var commandErr *CommandError
	if !errors.As(err, &commandErr) || commandErr.Stderr != "permission denied" {
		t.Fatalf("ListNetworks() error = %v, want CommandError with stderr", err)
	}
}
