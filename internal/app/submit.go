package app

import (
	"context"
	"errors"

	"github.com/joelmccoy/trail-hunk/internal/github"
	"github.com/joelmccoy/trail-hunk/internal/review"
)

type GitHubReviewSubmitter interface {
	SubmitReview(ctx context.Context, owner string, repo string, number int, req github.PullRequestReviewRequest) (github.PullRequestReview, error)
}

func SubmitApprovedReview(ctx context.Context, session review.ReviewSession, submitter GitHubReviewSubmitter) error {
	approved := session.ApprovedComments()
	if len(approved) == 0 {
		return errors.New("no approved comments to submit")
	}

	req := github.BuildPullRequestReviewRequest("trail-hunk review", approved)
	_, err := submitter.SubmitReview(ctx, session.Repo.Owner, session.Repo.Name, session.PR.Number, req)
	return err
}
