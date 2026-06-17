package app

import (
	"context"
	"testing"

	"github.com/joelmccoy/trail-hunk/internal/ai"
	"github.com/joelmccoy/trail-hunk/internal/git"
	"github.com/joelmccoy/trail-hunk/internal/github"
	"github.com/joelmccoy/trail-hunk/internal/review"
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
						ID:         "step-1",
						FilePath:   "app.go",
						HunkID:     "app.go:1",
						Title:      "Review rename",
						GroupID:    "rename",
						GroupTitle: "Rename flow",
						LayerIndex: 1,
						LayerTitle: "Function rename",
						Summary:    "A function was renamed.",
						Why:        "Callers may need updates.",
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
	step := session.Plan.ReviewOrder[0]
	if step.GroupTitle != "Rename flow" || step.LayerTitle != "Function rename" || step.LayerIndex != 1 {
		t.Fatalf("change stack metadata = %+v", step)
	}
}

func TestRunReviewAddsDiffLinesToReviewSteps(t *testing.T) {
	deps := ReviewDependencies{
		Repo: fakeRepoDiscoverer{
			repo: git.Repository{
				Root:   "/repo",
				Branch: "feature",
				Ref:    git.RepoRef{Owner: "joelmccoy", Name: "trail-hunk"},
			},
		},
		GitHub: fakeGitHubContext{
			prs: []github.PullRequest{{Number: 12, Title: "Add review flow"}},
			diff: `diff --git a/app.go b/app.go
--- a/app.go
+++ b/app.go
@@ -1,2 +1,3 @@
 package main
-func oldName() {}
+func newName() {}
+func added() {}
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
						Title:    "Review hunk",
						Summary:  "A hunk changed.",
						Why:      "The reviewer needs the code.",
					},
				},
			},
		},
	}

	session, err := RunReview(context.Background(), "/repo", deps)
	if err != nil {
		t.Fatal(err)
	}
	step := session.Plan.ReviewOrder[0]
	if len(step.DiffLines) != 4 {
		t.Fatalf("len(DiffLines) = %d, want 4", len(step.DiffLines))
	}
	if step.DiffLines[1].Kind != review.DiffLineDeleted || step.DiffLines[1].OldLine == nil {
		t.Fatalf("deleted line = %+v", step.DiffLines[1])
	}
	if step.DiffLines[2].Kind != review.DiffLineAdded || step.DiffLines[2].NewLine == nil {
		t.Fatalf("added line = %+v", step.DiffLines[2])
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
	if len(session.Plan.ReviewOrder) != 4 {
		t.Fatalf("len(ReviewOrder) = %d, want 4", len(session.Plan.ReviewOrder))
	}
	if len(session.Comments) != 5 {
		t.Fatalf("len(Comments) = %d, want 5", len(session.Comments))
	}
	seenFiles := map[string]bool{}
	for _, comment := range session.Comments {
		seenFiles[comment.FilePath] = true
		if comment.Line == 0 {
			t.Fatalf("comment line was not mapped: %+v", comment)
		}
	}
	for _, file := range []string{
		"dev/fixtures/dummy-pr/review_target.go",
		"dev/fixtures/dummy-pr/billing_handler.go",
		"dev/fixtures/dummy-pr/review_target_test.go",
	} {
		if !seenFiles[file] {
			t.Fatalf("mapped comments missing file %q: %+v", file, session.Comments)
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
diff --git a/dev/fixtures/dummy-pr/billing_handler.go b/dev/fixtures/dummy-pr/billing_handler.go
new file mode 100644
index 0000000..2222222
--- /dev/null
+++ b/dev/fixtures/dummy-pr/billing_handler.go
@@ -0,0 +1,19 @@
+package dummypr
+
+import "errors"
+
+var ErrForbidden = errors.New("forbidden")
+
+type BillingRequest struct {
+	AccountID   string
+	AmountCents int
+}
+
+func HandleBillingRequest(account Account, request BillingRequest) error {
+	if request.AmountCents <= 0 {
+		return errors.New("amount must be positive")
+	}
+	if !CanAccessBilling(account, request.AccountID) {
+		return ErrForbidden
+	}
+	return nil
+}
diff --git a/dev/fixtures/dummy-pr/review_target_test.go b/dev/fixtures/dummy-pr/review_target_test.go
new file mode 100644
index 0000000..3333333
--- /dev/null
+++ b/dev/fixtures/dummy-pr/review_target_test.go
@@ -0,0 +1,10 @@
+package dummypr
+
+import "testing"
+
+func TestNormalizeDisplayNameUnicode(t *testing.T) {
+	got := NormalizeDisplayName("José 🚀 customer account")
+	if got == "" {
+		t.Fatal("expected display name")
+	}
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
