package app

import (
	"context"
	"testing"

	"github.com/joelmccoy/trail-hunk/internal/git"
	"github.com/joelmccoy/trail-hunk/internal/github"
)

func TestResolveStartupContextFindsCurrentPullRequest(t *testing.T) {
	info, err := ResolveStartupContext(context.Background(), "/repo", StartupDependencies{
		Repo: fakeRepoDiscoverer{
			repo: git.Repository{
				Root:   "/repo",
				Branch: "feature",
				Ref:    git.RepoRef{Owner: "joelmccoy", Name: "trail-hunk"},
			},
		},
		GitHub: fakeGitHubContext{
			prs: []github.PullRequest{{Number: 12, Title: "Add guided review", State: "open"}},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if info.Repo.Branch != "feature" {
		t.Fatalf("Branch = %q, want feature", info.Repo.Branch)
	}
	if info.PR == nil {
		t.Fatal("expected PR")
	}
	if info.PR.Number != 12 || info.PR.Title != "Add guided review" {
		t.Fatalf("PR = %+v", info.PR)
	}
}

func TestResolveStartupContextAllowsNoPullRequest(t *testing.T) {
	info, err := ResolveStartupContext(context.Background(), "/repo", StartupDependencies{
		Repo: fakeRepoDiscoverer{
			repo: git.Repository{
				Root:   "/repo",
				Branch: "main",
				Ref:    git.RepoRef{Owner: "joelmccoy", Name: "trail-hunk"},
			},
		},
		GitHub: fakeGitHubContext{},
	})
	if err != nil {
		t.Fatal(err)
	}
	if info.PR != nil {
		t.Fatalf("PR = %+v, want nil", info.PR)
	}
	if info.Message != `No open GitHub pull request found for branch "main".` {
		t.Fatalf("Message = %q", info.Message)
	}
}
