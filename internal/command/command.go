package command

import (
	"context"
	"fmt"
	"os/exec"
)

type (
	Opt  func(*opts)
	opts struct {
		args                 []string
		dir                  string
		environmentVariables map[string]string
	}
)

func Exec(ctx context.Context, name string, o ...Opt) (<-chan []byte, error) {
	var options opts
	for _, opt := range o {
		opt(&options)
	}
	cmd := exec.CommandContext(ctx, name, options.args...)
	if options.dir != "" {
		cmd.Dir = options.dir
	}
	if options.environmentVariables != nil {
		env := make([]string, 0, len(options.environmentVariables))
		for k, v := range options.environmentVariables {
			env = append(env, fmt.Sprintf("%s=%s", k, v))
		}
		cmd.Env = append(cmd.Environ(), env...)
	}

	executor := &readableExecutor{}
	return executor.Exec(cmd)
}

func WithArgs(arg string, args ...string) Opt {
	return func(o *opts) {
		o.args = append([]string{arg}, args...)
	}
}

func WithDirectory(directory string) Opt {
	return func(o *opts) {
		o.dir = directory
	}
}

func WithEnvironmentVariables(environmentVariables map[string]string) Opt {
	return func(o *opts) {
		o.environmentVariables = environmentVariables
	}
}
