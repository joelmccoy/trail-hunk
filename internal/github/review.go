package github

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/joelmccoy/trail-hunk/internal/review"
)

const ReviewEventComment = "COMMENT"

type PullRequestReviewRequest struct {
	Body     string                     `json:"body,omitempty"`
	Event    string                     `json:"event"`
	Comments []PullRequestReviewComment `json:"comments"`
}

type PullRequestReviewComment struct {
	Path      string `json:"path"`
	Body      string `json:"body"`
	Line      int    `json:"line"`
	Side      string `json:"side"`
	StartLine *int   `json:"start_line,omitempty"`
	StartSide string `json:"start_side,omitempty"`
}

type PullRequestReview struct {
	ID int `json:"id"`
}

func BuildPullRequestReviewRequest(body string, comments []review.ReviewComment) PullRequestReviewRequest {
	req := PullRequestReviewRequest{
		Body:  body,
		Event: ReviewEventComment,
	}

	for _, comment := range comments {
		if comment.Status != review.StatusApproved && comment.Status != review.StatusEdited {
			continue
		}

		reviewComment := PullRequestReviewComment{
			Path:      comment.FilePath,
			Body:      comment.Body,
			Line:      comment.Line,
			Side:      comment.Side,
			StartLine: comment.StartLine,
		}
		if comment.StartLine != nil {
			reviewComment.StartSide = comment.Side
		}
		req.Comments = append(req.Comments, reviewComment)
	}

	return req
}

func (c *Client) SubmitReview(ctx context.Context, owner string, repo string, number int, req PullRequestReviewRequest) (PullRequestReview, error) {
	path := fmt.Sprintf("/repos/%s/%s/pulls/%d/reviews", url.PathEscape(owner), url.PathEscape(repo), number)

	var out PullRequestReview
	if err := c.DoJSON(ctx, http.MethodPost, path, req, &out); err != nil {
		return PullRequestReview{}, err
	}
	return out, nil
}
