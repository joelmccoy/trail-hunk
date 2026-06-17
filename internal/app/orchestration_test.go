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

func TestRunReviewMapsFixtureProviderCommentsToDummyPRDiff(t *testing.T) {
	provider, err := NewAIProvider(Config{Provider: "fixture"})
	if err != nil {
		t.Fatal(err)
	}
	deps := ReviewDependencies{
		Repo: fakeRepoDiscoverer{
			repo: git.Repository{
				Root:   "/repo",
				Branch: "trail-hunk-dev/dummy-pr",
				Ref:    git.RepoRef{Owner: "joelmccoy", Name: "trail-hunk"},
			},
		},
		GitHub: fakeGitHubContext{
			prs:  []github.PullRequest{{Number: 1, Title: "Dummy PR for trail-hunk workflow testing", Body: "Fixture PR."}},
			diff: dummyPRDiff(),
		},
		AI: provider,
	}

	session, err := RunReview(context.Background(), "/repo", deps)
	if err != nil {
		t.Fatal(err)
	}
	if len(session.Plan.ReviewOrder) != 2 {
		t.Fatalf("len(ReviewOrder) = %d, want 2", len(session.Plan.ReviewOrder))
	}
	if len(session.Comments) != 3 {
		t.Fatalf("len(Comments) = %d, want 3", len(session.Comments))
	}
	for _, comment := range session.Comments {
		if comment.FilePath != "dev/fixtures/dummy-pr/review_target.go" {
			t.Fatalf("comment file = %q", comment.FilePath)
		}
		if comment.Line == 0 {
			t.Fatalf("comment line was not mapped: %+v", comment)
		}
	}
}

func dummyPRDiff() string {
	return `diff --git a/dev/fixtures/dummy-pr/review_target.go b/dev/fixtures/dummy-pr/review_target.go
new file mode 100644
index 0000000..1111111
--- /dev/null
+++ b/dev/fixtures/dummy-pr/review_target.go
@@ -0,0 +1,30 @@
+package dummypr
+
+import "strings"
+
+type Account struct {
+	ID       string
+	Role     string
+	IsActive bool
+}
+
+// CanAccessBilling is intentionally imperfect fixture code for trail-hunk reviews.
+func CanAccessBilling(account Account, requestedAccountID string) bool {
+	if strings.TrimSpace(requestedAccountID) == "" {
+		return true
+	}
+
+	if account.Role == "admin" {
+		return true
+	}
+
+	return account.IsActive && account.ID == requestedAccountID
+}
+
+func NormalizeDisplayName(name string) string {
+	trimmed := strings.TrimSpace(name)
+	if len(trimmed) > 24 {
+		return trimmed[:24]
+	}
+	return trimmed
+}
`
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
