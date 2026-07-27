package kraft

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/ijaidev/kraftui/internal/api"
)

// ListMachines returns typed local machine data based on OpenAPI ListMachinesParams.
func (c *Client) ListMachines(ctx context.Context, params api.ListMachinesParams) ([]api.Machine, error) {
	args := []string{"ps", "--output", "json"}
	if params.All == nil || *params.All {
		args = append(args, "--all")
	}
	if params.Platform != nil && *params.Platform != "" {
		args = append(args, "--plat", *params.Platform)
	}
	if params.Architecture != nil && *params.Architecture != "" {
		args = append(args, "--arch", *params.Architecture)
	}
	if params.Long != nil && *params.Long {
		args = append(args, "--long")
	}
	rows, err := c.list(ctx, args...)
	if err != nil {
		return nil, err
	}
	items := make([]api.Machine, 0, len(rows))
	for _, row := range rows {
		platform, architecture := machinePlatform(row)
		if row["name"] == "" || row["status"] == "" || platform == "" || architecture == "" {
			return nil, fmt.Errorf("Kraft machine output is missing required fields")
		}
		items = append(items, api.Machine{
			Id:           optional(row["machine_id"]),
			Name:         row["name"],
			Kernel:       optional(row["kernel"]),
			Args:         optional(row["args"]),
			Created:      optional(row["created"]),
			Status:       row["status"],
			Memory:       optional(row["mem"]),
			Ports:        optional(row["ports"]),
			Ip:           optional(row["ip"]),
			Pid:          optional(row["pid"]),
			Platform:     platform,
			Architecture: architecture,
		})
	}
	return items, nil
}

// ListPackages returns typed locally available package data based on OpenAPI ListPackagesParams.
func (c *Client) ListPackages(ctx context.Context, params api.ListPackagesParams) ([]api.Package, error) {
	limit := 50
	if params.Limit != nil {
		limit = *params.Limit
	}
	if limit < 1 || limit > 100 {
		return nil, fmt.Errorf("package limit must be between 1 and 100")
	}
	args := []string{"pkg", "list", "--local", "--output", "json", "--limit", strconv.Itoa(limit)}
	if params.Kind != nil {
		switch *params.Kind {
		case api.ListPackagesParamsKindAll, "":
		case api.ListPackagesParamsKindApp:
			args = append(args, "--apps")
		case api.ListPackagesParamsKindLib:
			args = append(args, "--libs")
		case api.ListPackagesParamsKindCore:
			args = append(args, "--core")
		default:
			return nil, fmt.Errorf("unsupported package kind %q", *params.Kind)
		}
	}
	if params.Platform != nil && *params.Platform != "" {
		args = append(args, "--plat", *params.Platform)
	}
	if params.Architecture != nil && *params.Architecture != "" {
		args = append(args, "--arch", *params.Architecture)
	}
	rows, err := c.list(ctx, args...)
	if err != nil {
		return nil, err
	}
	items := make([]api.Package, 0, len(rows))
	for _, row := range rows {
		if row["name"] == "" || row["version"] == "" || row["type"] == "" {
			return nil, fmt.Errorf("Kraft package output is missing required fields")
		}
		items = append(items, api.Package{
			Name:         row["name"],
			Version:      row["version"],
			Type:         row["type"],
			Format:       optional(row["format"]),
			Created:      optional(row["created"]),
			Updated:      optional(row["updated"]),
			Pulled:       optional(row["pulled"]),
			Architecture: optional(row["arch"]),
			Platform:     optional(row["plat"]),
			Size:         optional(row["size"]),
			Description:  optional(row["description"]),
			Origin:       optional(row["origin"]),
			Index:        optional(row["index"]),
			Manifest:     optional(row["manifest"]),
		})
	}
	return items, nil
}

// ListNetworks returns typed local network data based on OpenAPI ListNetworksParams.
func (c *Client) ListNetworks(ctx context.Context, params api.ListNetworksParams) ([]api.Network, error) {
	args := []string{"net", "list", "--output", "json"}
	if params.Long != nil && *params.Long {
		args = append(args, "--long")
	}
	rows, err := c.list(ctx, args...)
	if err != nil {
		return nil, err
	}
	items := make([]api.Network, 0, len(rows))
	for _, row := range rows {
		if row["name"] == "" || row["driver"] == "" || row["status"] == "" {
			return nil, fmt.Errorf("Kraft network output is missing required fields")
		}
		items = append(items, api.Network{
			MachineId: optional(row["machine_id"]),
			Name:      row["name"],
			Network:   optional(row["network"]),
			Driver:    row["driver"],
			Status:    row["status"],
		})
	}
	return items, nil
}

// ListVolumes returns typed local volume data based on OpenAPI ListVolumesParams.
func (c *Client) ListVolumes(ctx context.Context, params api.ListVolumesParams) ([]api.Volume, error) {
	args := []string{"vol", "ls", "--output", "json"}
	if params.Long != nil && *params.Long {
		args = append(args, "--long")
	}
	rows, err := c.list(ctx, args...)
	if err != nil {
		return nil, err
	}
	items := make([]api.Volume, 0, len(rows))
	for _, row := range rows {
		if row["volume_name"] == "" || row["driver"] == "" || row["status"] == "" {
			return nil, fmt.Errorf("Kraft volume output is missing required fields")
		}
		items = append(items, api.Volume{
			Id:     optional(row["volume_id"]),
			Name:   row["volume_name"],
			Driver: row["driver"],
			Status: row["status"],
			Source: optional(row["source"]),
		})
	}
	return items, nil
}

func (c *Client) list(ctx context.Context, args ...string) ([]map[string]string, error) {
	result, err := c.run(ctx, args...)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(string(result.Stdout)) == "" {
		return []map[string]string{}, nil
	}
	decoder := json.NewDecoder(bytes.NewReader(result.Stdout))
	var rows []map[string]string
	for {
		var batch []map[string]string
		if err := decoder.Decode(&batch); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, fmt.Errorf("decoding Kraft JSON output: %w", err)
		}
		rows = append(rows, batch...)
	}
	return rows, nil
}

func machinePlatform(row map[string]string) (string, string) {
	if row["arch"] != "" {
		return row["plat"], row["arch"]
	}
	platform, architecture, found := strings.Cut(row["plat"], "/")
	if !found {
		return row["plat"], "unknown"
	}
	return platform, architecture
}

func optional(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}
