package app

import (
	"context"
	"testing"

	"github.com/joelmccoy/trail-hunk/internal/ai"
	"github.com/joelmccoy/trail-hunk/internal/git"
	"github.com/joelmccoy/trail-hunk/internal/github"
)

func TestRunReviewBuildsSessionFromPullRequestContext(t *testing.T) {
	deps := ReviewDependencies{
		Repo: fakeRepoDiscoverer{
			repo: git.Repository{
				Root:   "/repo",
				Branch: "feature",
				Ref:    git.RepoRef{Owner: "joelmccoy", Name: "trail-hunk"},
			},
		},
		GitHub: fakeGitHubContext{
			prs: []github.PullRequest{{Number: 12, Title: "Add review flow", Body: "Adds guided review."}},
			diff: `diff --git a/app.go b/app.go
--- a/app.go
+++ b/app.go
@@ -1,2 +1,2 @@
 package main
-func oldName() {}
+func newName() {}
`,
		},
		AI: fakeAIProvider{
			response: ai.ReviewResponse{
				Overview: "Adds a guided review flow.",
				ReviewOrder: []ai.ReviewStep{
					{
						ID:       "step-1",
						FilePath: "app.go",
						HunkID:   "app.go:1",
						Title:    "Review rename",
						Summary:  "A function was renamed.",
						Why:      "Callers may need updates.",
						Suggestions: []ai.SuggestedComment{
							{
								FilePath: "app.go",
								Side:     github.SideRight,
								Line:     2,
								Body:     "Confirm callers were updated.",
								Priority: "medium",
								Category: "correctness",
							},
						},
					},
				},
			},
		},
	}

	session, err := RunReview(context.Background(), "/repo", deps)
	if err != nil {
		t.Fatal(err)
	}
	if session.Repo.Owner != "joelmccoy" || session.PR.Number != 12 {
		t.Fatalf("session repo/pr = %+v %+v", session.Repo, session.PR)
	}
	if session.Plan.Overview != "Adds a guided review flow." {
		t.Fatalf("Overview = %q", session.Plan.Overview)
	}
	if len(session.Comments) != 1 {
		t.Fatalf("len(Comments) = %d, want 1", len(session.Comments))
	}
}

type fakeRepoDiscoverer struct {
	repo git.Repository
}

func (f fakeRepoDiscoverer) Discover(context.Context, string) (git.Repository, error) {
	return f.repo, nil
}

type fakeGitHubContext struct {
	prs  []github.PullRequest
	diff string
}

func (f fakeGitHubContext) FindPullRequestsForBranch(context.Context, string, string, string) ([]github.PullRequest, error) {
	return f.prs, nil
}

func (f fakeGitHubContext) PullRequestDiff(context.Context, string, string, int) (string, error) {
	return f.diff, nil
}

type fakeAIProvider struct {
	response ai.ReviewResponse
}

func (f fakeAIProvider) Review(context.Context, ai.ReviewRequest) (ai.ReviewResponse, error) {
	return f.response, nil
}
