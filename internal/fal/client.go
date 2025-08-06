package fal

import (
	"bytes"
	"os"
)

type DeployStrategy string

const (
	DeployStrategyRecreate DeployStrategy = "recreate"
	DeployStrategyRolling  DeployStrategy = "rolling"
)

type AuthMode string

const (
	AuthModePublic  AuthMode = "public"
	AuthModePrivate AuthMode = "private"
)

type Client struct {
	key string
	dir string
}

func NewWithTemp(key string) (*Client, error) {
	dir, err := os.MkdirTemp("", "fal-*")
	if err != nil {
		return nil, err
	}
	return &Client{
		key: key,
		dir: dir,
	}, nil
}

func (f *Client) Directory() string {
	return f.dir
}

func (f *Client) sharedEnvironmentVariables() map[string]string {
	return map[string]string{
		"FAL_KEY": f.key,
	}
}

func readAll(c <-chan []byte) string {
	var output bytes.Buffer

wait:
	for {
		select {
		case v, ok := <-c:
			if !ok {
				break wait
			}
			output.WriteString(string(v))
		}
	}

	return output.String()
}
