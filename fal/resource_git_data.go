package fal

import (
	"context"
	"fmt"
	"net/url"

	"github.com/fal-ai/terraform-provider-fal/internal/git"
	"github.com/go-git/go-git/v6/plumbing/transport/http"
	"github.com/go-git/go-git/v6/plumbing/transport/ssh"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

type gitData struct {
	git *Git
}

func gitFromResourceModel(ctx context.Context, data *AppResourceModel) *gitData {
	var _git Git
	if !data.Git.IsNull() {
		data.Git.As(ctx, &_git, basetypes.ObjectAsOptions{})
	}
	return &gitData{
		git: &_git,
	}
}

func (ard *gitData) Client() (*git.Client, error) {
	authOpts, err := getAuthOpts(ard.git)
	if err != nil {
		return nil, err
	}

	if ard.git.HTTP != nil && ard.git.HTTP.InsecureHTTPAllowed.ValueBool() {
		authOpts.InsecureSkipTLS = true
	}

	client := git.New(authOpts)
	return client, nil
}

func (ard *gitData) RepositoryURL() *url.URL {
	repositoryURL, err := url.Parse(ard.git.URL.ValueString())
	if err != nil {
		panic(err)
	}
	if ard.git.HTTP != nil {
		repositoryURL.User = nil
	}
	if ard.git.SSH != nil {
		if repositoryURL.User == nil || ard.git.SSH.Username.ValueString() != "git" {
			repositoryURL.User = url.User(ard.git.SSH.Username.ValueString())
		}
	}
	return repositoryURL
}

func getAuthOpts(g *Git) (*git.AuthOpts, error) {
	u, err := url.Parse(g.URL.ValueString())
	if err != nil {
		panic(err)
	}
	switch u.Scheme {
	case "http":
		if g.HTTP == nil {
			return nil, fmt.Errorf("Git URL scheme is http but http configuration is empty")
		}
		return &git.AuthOpts{
			AuthMethod: &http.BasicAuth{
				Username: g.HTTP.Username.ValueString(),
				Password: g.HTTP.Password.ValueString(),
			},
		}, nil
	case "https":
		if g.HTTP == nil {
			return nil, fmt.Errorf("Git URL scheme is https but http configuration is empty")
		}
		return &git.AuthOpts{
			AuthMethod: &http.BasicAuth{
				Username: g.HTTP.Username.ValueString(),
				Password: g.HTTP.Password.ValueString(),
			},
			CABundle: []byte(g.HTTP.CertificateAuthority.ValueString()),
		}, nil
	case "ssh":
		if g.SSH == nil {
			return nil, fmt.Errorf("Git URL scheme is ssh but ssh configuration is empty")
		}
		if g.SSH.PrivateKey.ValueString() != "" {
			sshKey, err := ssh.NewPublicKeys(g.SSH.Username.ValueString(), []byte(g.SSH.PrivateKey.ValueString()), g.SSH.Password.ValueString())
			if err != nil {
				return nil, fmt.Errorf("could not handle ssh key auth: %w", err)
			}
			return &git.AuthOpts{
				AuthMethod: sshKey,
			}, nil
		}
		return nil, fmt.Errorf("ssh scheme cannot be used without private key")
	default:
		return nil, fmt.Errorf("scheme %q is not supported", u.Scheme)
	}
}
