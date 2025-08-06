package fal

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/fal-ai/terraform-provider-fal/internal/runner"
)

type App struct {
	Alias             string   `json:"alias"`
	Revision          string   `json:"revision"`
	AuthMode          string   `json:"auth_mode"`
	KeepAlive         int      `json:"keep_alive"`
	MaxConcurrency    int      `json:"max_concurrency"`
	MaxMultiplexing   int      `json:"max_multiplexing"`
	ActiveRunners     int      `json:"active_runners"`
	MinConcurrency    int      `json:"min_concurrency"`
	ConcurrencyBuffer int      `json:"concurrency_buffer"`
	MachineTypes      []string `json:"machine_types"`
	RequestTimeout    int      `json:"request_timeout"`
	StartupTimeout    int      `json:"startup_timeout"`
	ValidRegions      []string `json:"valid_regions"`
}

func (f *Client) List(ctx context.Context) ([]*App, error) {
	uv := runner.FromUv(f.dir)

	if c, err := uv.Init(ctx); err != nil {
		return nil, fmt.Errorf("error calling uv init: %w: %s", err, readAll(c))
	}

	if c, err := uv.Add(ctx, "fal"); err != nil {
		return nil, fmt.Errorf("error adding fal client into new env: %w: %s", err, readAll(c))
	}

	env := f.sharedEnvironmentVariables()

	c, err := uv.Run(ctx, env, "fal", "apps", "list", "--output", "json")
	if err != nil {
		return nil, fmt.Errorf("error running fal apps list in uv environment: %w: %s", err, readAll(c))
	}

	output := readAll(c)

	var result []*App
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		return nil, err
	}
	return result, nil
}
