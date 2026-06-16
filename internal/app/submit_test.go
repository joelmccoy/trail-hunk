package app

import (
	"context"
	"testing"

	"github.com/joelmccoy/trail-hunk/internal/github"
	"github.com/joelmccoy/trail-hunk/internal/review"
)

func TestSubmitApprovedReviewUsesApprovedComments(t *testing.T) {
	submitter := fakeReviewSubmitter{}
	session := review.ReviewSession{
		Repo: review.RepoRef{Owner: "joelmccoy", Name: "trail-hunk"},
		PR:   review.PullRequest{Number: 12},
		Comments: []review.ReviewComment{
			{
				FilePath: "app.go",
				Side:     github.SideRight,
				Line:     7,
				Body:     "Please add a test.",
				Status:   review.StatusApproved,
			},
			{
				FilePath: "old.go",
				Side:     github.SideLeft,
				Line:     2,
				Body:     "Dismissed comment.",
				Status:   review.StatusDismissed,
			},
		},
	}

	if err := SubmitApprovedReview(context.Background(), session, &submitter); err != nil {
		t.Fatal(err)
	}
	if submitter.owner != "joelmccoy" || submitter.repo != "trail-hunk" || submitter.number != 12 {
		t.Fatalf("target = %s/%s#%d", submitter.owner, submitter.repo, submitter.number)
	}
	if len(submitter.req.Comments) != 1 {
		t.Fatalf("len(Comments) = %d, want 1", len(submitter.req.Comments))
	}
}

func TestSubmitApprovedReviewRejectsEmptyQueue(t *testing.T) {
	err := SubmitApprovedReview(context.Background(), review.ReviewSession{}, &fakeReviewSubmitter{})
	if err == nil {
		t.Fatal("expected empty queue error")
	}
}

type fakeReviewSubmitter struct {
	owner  string
	repo   string
	number int
	req    github.PullRequestReviewRequest
}

func (f *fakeReviewSubmitter) SubmitReview(ctx context.Context, owner string, repo string, number int, req github.PullRequestReviewRequest) (github.PullRequestReview, error) {
	f.owner = owner
	f.repo = repo
	f.number = number
	f.req = req
	return github.PullRequestReview{ID: 99}, nil
}
