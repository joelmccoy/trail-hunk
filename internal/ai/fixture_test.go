package ai

import (
	"context"
	"strings"
	"testing"
)

func TestFixtureProviderReturnsDeterministicReview(t *testing.T) {
	provider := NewFixtureProvider()

	response, err := provider.Review(context.Background(), ReviewRequest{
		PRTitle: "Dummy PR for trail-hunk workflow testing",
		Diff: `diff --git a/dev/fixtures/dummy-pr/review_target.go b/dev/fixtures/dummy-pr/review_target.go
--- /dev/null
+++ b/dev/fixtures/dummy-pr/review_target.go
@@ -0,0 +1,30 @@
+package dummypr
+`,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := ValidateReviewResponse(response); err != nil {
		t.Fatal(err)
	}
	if response.Overview == "" {
		t.Fatal("expected overview")
	}
	if len(response.ReviewOrder) != 4 {
		t.Fatalf("len(ReviewOrder) = %d, want 4", len(response.ReviewOrder))
	}
	if len(response.ReviewOrder[0].Suggestions) == 0 {
		t.Fatal("expected suggestions on first step")
	}
	if response.ReviewOrder[0].Suggestions[0].FilePath != "dev/fixtures/dummy-pr/review_target.go" {
		t.Fatalf("suggestion file = %q", response.ReviewOrder[0].Suggestions[0].FilePath)
	}
}

func TestFixtureProviderAskAndReword(t *testing.T) {
	provider := NewFixtureProvider()

	answer, err := provider.Ask(context.Background(), AskRequest{Question: "why?", Context: "current hunk"})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(answer.Answer, "fixture") {
		t.Fatalf("Answer = %q, want fixture marker", answer.Answer)
	}

	reworded, err := provider.Reword(context.Background(), RewordRequest{Body: "old", Instruction: "shorter"})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(reworded.Body, "old") {
		t.Fatalf("Body = %q, want original body included", reworded.Body)
	}
}
