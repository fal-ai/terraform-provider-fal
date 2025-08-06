package command

import (
	"fmt"
	"os/exec"
)

type ExecutorInterface interface {
	Exec(*exec.Cmd) (<-chan []byte, error)
}

type readableExecutor struct{}

func (readableExecutor) Exec(cmd *exec.Cmd) (<-chan []byte, error) {
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("could not open stdout pipe: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("could not open stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("could not execute command: %w", err)
	}

	c := make(chan []byte, 256)
	go func() {
		reader := multiReader(stdout, stderr)
		_ = scan(reader, func(line []byte) {
			c <- line
		})
		close(c)
	}()

	return c, nil
}
