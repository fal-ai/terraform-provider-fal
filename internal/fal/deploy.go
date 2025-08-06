package fal

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/fal-ai/terraform-provider-fal/internal/git"
	"github.com/fal-ai/terraform-provider-fal/internal/runner"
)

type DeployOpts struct {
	Entrypoint string
	Strategy   DeployStrategy
	AuthMode   AuthMode
}

func (f *Client) Deploy(ctx context.Context, git *git.Client, repo string, opts *DeployOpts) (*DeployResult, error) {
	u := parseGitURL(repo)
	path := f.dir + "/" + u.Repo

	if err := git.Clone(ctx, path, repo); err != nil {
		return nil, fmt.Errorf("error cloning git repo: %w", err)
	}

	uv := runner.FromUv(path)

	if c, err := uv.Sync(ctx); err != nil {
		return nil, fmt.Errorf("error syncing virtual environment: %w: %s", err, readAll(c))
	}

	strategyFlag := fmt.Sprintf("--strategy=%s", opts.Strategy)
	authModeFlag := fmt.Sprintf("--auth=%s", opts.AuthMode)
	env := f.sharedEnvironmentVariables()

	c, err := uv.Run(ctx, env, "fal", "deploy", strategyFlag, authModeFlag, opts.Entrypoint)
	if err != nil {
		return nil, fmt.Errorf("error running fal deploy: %w: %s", err, readAll(c))
	}

	r := parseDeployResult(c)
	if r.FunctionName == "" && r.Revision == "" {
		return nil, fmt.Errorf("deployment failed: %s", r.Output)
	}

	return r, nil
}

var (
	functionRe = regexp.MustCompile(`function '([^']+)'`)
	revisionRe = regexp.MustCompile(`revision='([^']+)'`)

	completedResultLine = "Registered a new revision for function"
)

type DeployResult struct {
	FunctionName string
	Revision     string

	Output string
}

func parseDeployResult(bytes <-chan []byte) *DeployResult {
	output := strings.Builder{}
	for b := range bytes {
		line := string(b)
		output.WriteString("\n")
		output.WriteString(line)

		if !strings.Contains(line, completedResultLine) {
			continue
		}

		matches := functionRe.FindStringSubmatch(line)
		if len(matches) != 2 {
			continue
		}

		nextBytes, ok := <-bytes
		if !ok {
			return nil
		}

		nextLine := string(nextBytes)
		revMatches := revisionRe.FindStringSubmatch(nextLine)
		if len(revMatches) != 2 {
			continue
		}

		return &DeployResult{
			FunctionName: matches[1],
			Revision:     revMatches[1],
			Output:       output.String(),
		}
	}

	return &DeployResult{
		Output: output.String(),
	}
}
