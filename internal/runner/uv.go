package runner

import (
	"context"

	"github.com/fal-ai/terraform-provider-fal/internal/command"
)

const (
	commandUv = "uv"
)

type uv struct {
	path string
}

func FromUv(path string) *uv {
	return &uv{path: path}
}

func (u *uv) Init(ctx context.Context) (<-chan []byte, error) {
	return command.Exec(ctx, commandUv, command.WithArgs("init", "--no-workspace", "--bare"), command.WithDirectory(u.path))
}

func (u *uv) Venv(ctx context.Context) (<-chan []byte, error) {
	return command.Exec(ctx, commandUv, command.WithArgs("venv"), command.WithDirectory(u.path))
}

func (u *uv) Add(ctx context.Context, packages ...string) (<-chan []byte, error) {
	return command.Exec(ctx, commandUv, command.WithArgs("add", packages...), command.WithDirectory(u.path))
}

func (u *uv) Run(ctx context.Context, environment map[string]string, args ...string) (<-chan []byte, error) {
	return command.Exec(ctx, commandUv, command.WithEnvironmentVariables(environment), command.WithArgs("run", args...), command.WithDirectory(u.path))
}

func (u *uv) Sync(ctx context.Context) (<-chan []byte, error) {
	return command.Exec(ctx, commandUv, command.WithArgs("sync"), command.WithDirectory(u.path))
}
