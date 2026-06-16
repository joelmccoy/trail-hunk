package app

import (
	"context"
	"fmt"

	"github.com/joelmccoy/trail-hunk/internal/github"
	"github.com/joelmccoy/trail-hunk/internal/review"
)

type StartupDependencies struct {
	Repo   RepoDiscoverer
	GitHub GitHubPullRequestFinder
}

type GitHubPullRequestFinder interface {
	FindPullRequestsForBranch(ctx context.Context, owner, repo, branch string) ([]github.PullRequest, error)
}

func ResolveStartupContext(ctx context.Context, dir string, deps StartupDependencies) (review.StartupContext, error) {
	if deps.Repo == nil {
		return review.StartupContext{}, fmt.Errorf("repo discoverer is required")
	}
	if deps.GitHub == nil {
		return review.StartupContext{}, fmt.Errorf("GitHub context client is required")
	}

	repo, err := deps.Repo.Discover(ctx, dir)
	if err != nil {
		return review.StartupContext{}, err
	}

	info := review.StartupContext{
		Repo: review.RepoRef{
			Owner:  repo.Ref.Owner,
			Name:   repo.Ref.Name,
			Root:   repo.Root,
			Branch: repo.Branch,
		},
	}

	prs, err := deps.GitHub.FindPullRequestsForBranch(ctx, repo.Ref.Owner, repo.Ref.Name, repo.Branch)
	if err != nil {
		return review.StartupContext{}, err
	}
	if len(prs) == 0 {
		info.Message = fmt.Sprintf("No open GitHub pull request found for branch %q.", repo.Branch)
		return info, nil
	}
	if len(prs) > 1 {
		return review.StartupContext{}, fmt.Errorf("multiple open GitHub pull requests found for branch %q", repo.Branch)
	}

	pr := prs[0]
	info.PR = &review.PullRequest{
		Number: pr.Number,
		Title:  pr.Title,
		Body:   pr.Body,
		State:  pr.State,
		URL:    pr.HTMLURL,
	}
	info.Message = "Current pull request detected."
	return info, nil
}
