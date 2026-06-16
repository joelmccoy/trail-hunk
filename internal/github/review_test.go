package github

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/joelmccoy/trail-hunk/internal/review"
)

func TestBuildPullRequestReviewRequest(t *testing.T) {
	req := BuildPullRequestReviewRequest("Review body", []review.ReviewComment{
		{
			FilePath: "app.go",
			Side:     SideRight,
			Line:     12,
			Body:     "Please add error handling here.",
			Status:   review.StatusApproved,
		},
		{
			FilePath: "old.go",
			Side:     SideLeft,
			Line:     4,
			Body:     "This deletion changes behavior.",
			Status:   review.StatusEdited,
		},
		{
			FilePath: "dismissed.go",
			Side:     SideRight,
			Line:     8,
			Body:     "Dismissed comment.",
			Status:   review.StatusDismissed,
		},
	})

	if req.Event != ReviewEventComment {
		t.Fatalf("Event = %q, want COMMENT", req.Event)
	}
	if req.Body != "Review body" {
		t.Fatalf("Body = %q, want Review body", req.Body)
	}
	if len(req.Comments) != 2 {
		t.Fatalf("len(Comments) = %d, want 2", len(req.Comments))
	}
	if req.Comments[0].Path != "app.go" || req.Comments[0].Line != 12 || req.Comments[0].Side != SideRight {
		t.Fatalf("first comment = %+v", req.Comments[0])
	}
}

func TestSubmitReviewPostsPullRequestReview(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %q, want POST", r.Method)
		}
		if r.URL.Path != "/repos/joelmccoy/trail-hunk/pulls/12/reviews" {
			t.Fatalf("path = %q", r.URL.Path)
		}

		var req PullRequestReviewRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatal(err)
		}
		if req.Event != ReviewEventComment {
			t.Fatalf("Event = %q, want COMMENT", req.Event)
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(PullRequestReview{ID: 99})
	}))
	t.Cleanup(server.Close)

	client := NewClient(server.URL, "abc123")
	resp, err := client.SubmitReview(context.Background(), "joelmccoy", "trail-hunk", 12, PullRequestReviewRequest{
		Event: ReviewEventComment,
		Comments: []PullRequestReviewComment{
			{Path: "app.go", Line: 12, Side: SideRight, Body: "Looks good."},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.ID != 99 {
		t.Fatalf("ID = %d, want 99", resp.ID)
	}
}
