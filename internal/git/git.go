package git

import (
	"context"

	"github.com/go-git/go-git/v6"
	"github.com/go-git/go-git/v6/plumbing/transport"
)

type AuthOpts struct {
	AuthMethod      transport.AuthMethod
	InsecureSkipTLS bool
	CABundle        []byte
}

type Client struct {
	auth *AuthOpts
}

func New(authOpts *AuthOpts) *Client {
	return &Client{
		auth: authOpts,
	}
}

func (c *Client) Clone(ctx context.Context, path, repoURL string) (err error) {
	_, err = git.PlainCloneContext(ctx, path, &git.CloneOptions{
		URL:             repoURL,
		Auth:            c.auth.AuthMethod,
		Depth:           1,
		InsecureSkipTLS: c.auth.InsecureSkipTLS,
		CABundle:        c.auth.CABundle,
	})
	return
}
