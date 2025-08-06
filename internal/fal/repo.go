package fal

import (
	"strings"
)

type GitURL struct {
	Host  string
	Owner string
	Repo  string
}

func parseGitURL(gitURL string) *GitURL {
	clean := strings.TrimSuffix(gitURL, ".git")

	var parts []string

	// SSH: git@github.com:owner/repo
	if strings.Contains(clean, "@") && strings.Contains(clean, ":") {
		clean = strings.TrimPrefix(clean, "git@")
		parts = strings.FieldsFunc(clean, func(r rune) bool {
			return r == ':' || r == '/'
		})
	} else {
		// HTTP: https://github.com/owner/repo
		clean = strings.TrimPrefix(clean, "https://")
		clean = strings.TrimPrefix(clean, "http://")
		parts = strings.Split(clean, "/")
	}

	if len(parts) < 3 {
		return nil
	}

	return &GitURL{
		Host:  parts[0],
		Owner: parts[1],
		Repo:  parts[2],
	}
}
