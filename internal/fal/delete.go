package fal

import (
	"context"
	"fmt"

	"github.com/fal-ai/terraform-provider-fal/internal/runner"
)

func (f *Client) Delete(ctx context.Context, app string) error {
	uv := runner.FromUv(f.dir)

	if c, err := uv.Init(ctx); err != nil {
		return fmt.Errorf("error calling uv init: %w: %s", err, readAll(c))
	}

	if c, err := uv.Add(ctx, "fal"); err != nil {
		return fmt.Errorf("error adding fal client into new env: %w: %s", err, readAll(c))
	}

	env := f.sharedEnvironmentVariables()

	_, err := uv.Run(ctx, env, "fal", "apps", "delete", app)
	if err != nil {
		return fmt.Errorf("error running fal delete in uv environment: %w", err)
	}
	return nil
}
