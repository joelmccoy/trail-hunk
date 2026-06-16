package git

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os/exec"
	"strings"
)

type RepoRef struct {
	Owner string
	Name  string
}

type Repository struct {
	Root   string
	Branch string
	Remote string
	Ref    RepoRef
}

func ParseGitHubRemote(remote string) (RepoRef, error) {
	if strings.HasPrefix(remote, "git@github.com:") {
		path := strings.TrimPrefix(remote, "git@github.com:")
		return parseGitHubPath(path)
	}

	u, err := url.Parse(remote)
	if err == nil && u.Host == "github.com" {
		return parseGitHubPath(strings.TrimPrefix(u.Path, "/"))
	}

	return RepoRef{}, fmt.Errorf("unsupported GitHub remote %q", remote)
}

func Discover(ctx context.Context, dir string) (Repository, error) {
	root, err := runGit(ctx, dir, "rev-parse", "--show-toplevel")
	if err != nil {
		return Repository{}, fmt.Errorf("discover git root: %w", err)
	}

	branch, err := runGit(ctx, dir, "branch", "--show-current")
	if err != nil {
		return Repository{}, fmt.Errorf("discover git branch: %w", err)
	}
	if branch == "" {
		return Repository{}, errors.New("current git checkout is detached; trail-hunk requires a branch")
	}

	remote, err := runGit(ctx, dir, "remote", "get-url", "origin")
	if err != nil {
		return Repository{}, fmt.Errorf("discover origin remote: %w", err)
	}

	ref, err := ParseGitHubRemote(remote)
	if err != nil {
		return Repository{}, err
	}

	return Repository{
		Root:   root,
		Branch: branch,
		Remote: remote,
		Ref:    ref,
	}, nil
}

func parseGitHubPath(path string) (RepoRef, error) {
	path = strings.TrimSuffix(path, ".git")
	parts := strings.Split(path, "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return RepoRef{}, fmt.Errorf("unsupported GitHub repository path %q", path)
	}

	return RepoRef{Owner: parts[0], Name: parts[1]}, nil
}

func runGit(ctx context.Context, dir string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = dir

	out, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(out)), nil
}
