package app

import (
	"context"
	"errors"
	"fmt"

	"github.com/joelmccoy/trail-hunk/internal/ai"
	"github.com/joelmccoy/trail-hunk/internal/git"
	"github.com/joelmccoy/trail-hunk/internal/github"
	"github.com/joelmccoy/trail-hunk/internal/review"
)

type RepoDiscoverer interface {
	Discover(ctx context.Context, dir string) (git.Repository, error)
}

type GitHubContext interface {
	FindPullRequestsForBranch(ctx context.Context, owner, repo, branch string) ([]github.PullRequest, error)
	PullRequestDiff(ctx context.Context, owner, repo string, number int) (string, error)
}

type ReviewAI interface {
	Review(ctx context.Context, req ai.ReviewRequest) (ai.ReviewResponse, error)
}

type ReviewDependencies struct {
	Repo   RepoDiscoverer
	GitHub GitHubContext
	AI     ReviewAI
}

func RunReview(ctx context.Context, dir string, deps ReviewDependencies) (review.ReviewSession, error) {
	if deps.Repo == nil {
		return review.ReviewSession{}, errors.New("repo discoverer is required")
	}
	if deps.GitHub == nil {
		return review.ReviewSession{}, errors.New("GitHub context client is required")
	}
	if deps.AI == nil {
		return review.ReviewSession{}, errors.New("AI reviewer is required")
	}

	repo, err := deps.Repo.Discover(ctx, dir)
	if err != nil {
		return review.ReviewSession{}, err
	}

	prs, err := deps.GitHub.FindPullRequestsForBranch(ctx, repo.Ref.Owner, repo.Ref.Name, repo.Branch)
	if err != nil {
		return review.ReviewSession{}, err
	}
	if len(prs) == 0 {
		return review.ReviewSession{}, fmt.Errorf("no open GitHub pull request found for branch %q", repo.Branch)
	}
	if len(prs) > 1 {
		return review.ReviewSession{}, fmt.Errorf("multiple open GitHub pull requests found for branch %q", repo.Branch)
	}
	pr := prs[0]

	rawDiff, err := deps.GitHub.PullRequestDiff(ctx, repo.Ref.Owner, repo.Ref.Name, pr.Number)
	if err != nil {
		return review.ReviewSession{}, err
	}
	parsedDiff, err := github.ParsePullRequestDiff(rawDiff)
	if err != nil {
		return review.ReviewSession{}, err
	}

	aiReview, err := deps.AI.Review(ctx, ai.ReviewRequest{
		PRTitle: pr.Title,
		PRBody:  pr.Body,
		Diff:    rawDiff,
	})
	if err != nil {
		return review.ReviewSession{}, err
	}

	return buildReviewSession(repo, pr, parsedDiff, aiReview), nil
}

func buildReviewSession(repo git.Repository, pr github.PullRequest, diff github.PullRequestDiff, aiReview ai.ReviewResponse) review.ReviewSession {
	session := review.ReviewSession{
		Repo: review.RepoRef{
			Owner:  repo.Ref.Owner,
			Name:   repo.Ref.Name,
			Root:   repo.Root,
			Branch: repo.Branch,
		},
		PR: review.PullRequest{
			Number: pr.Number,
			Title:  pr.Title,
			Body:   pr.Body,
		},
		Plan: review.WalkthroughPlan{
			Overview: aiReview.Overview,
		},
	}

	for _, risk := range aiReview.Risks {
		session.Plan.Risks = append(session.Plan.Risks, review.Risk{
			Priority: review.Priority(risk.Priority),
			Category: review.CommentCategory(risk.Category),
			Summary:  risk.Summary,
		})
	}

	for _, step := range aiReview.ReviewOrder {
		reviewStep := review.ReviewStep{
			ID:         step.ID,
			FilePath:   step.FilePath,
			HunkID:     step.HunkID,
			Title:      step.Title,
			GroupID:    step.GroupID,
			GroupTitle: step.GroupTitle,
			LayerIndex: step.LayerIndex,
			LayerTitle: step.LayerTitle,
			Summary:    step.Summary,
			Why:        step.Why,
			Focus:      step.Focus,
			DiffLines:  diffLinesForStep(diff, step.FilePath, step.HunkID),
		}

		for _, suggestion := range step.Suggestions {
			target, err := diff.FindTarget(suggestion.FilePath, suggestion.Side, suggestion.Line)
			if err != nil {
				continue
			}
			comment := review.ReviewComment{
				ID:        fmt.Sprintf("ai-%d", len(session.Comments)+1),
				FilePath:  target.Path,
				Side:      target.Side,
				Line:      target.Line,
				StartLine: suggestion.StartLine,
				Body:      suggestion.Body,
				Priority:  review.Priority(suggestion.Priority),
				Category:  review.CommentCategory(suggestion.Category),
				Status:    review.StatusSuggested,
				Source:    review.SourceAI,
			}
			session.Comments = append(session.Comments, comment)
			reviewStep.Suggestions = append(reviewStep.Suggestions, comment)
		}

		session.Plan.ReviewOrder = append(session.Plan.ReviewOrder, reviewStep)
	}

	return session
}

func diffLinesForStep(diff github.PullRequestDiff, filePath string, hunkID string) []review.DiffLine {
	for _, file := range diff.Files {
		if file.Path != filePath {
			continue
		}
		for _, hunk := range file.Hunks {
			if hunk.ID != hunkID {
				continue
			}
			lines := make([]review.DiffLine, 0, len(hunk.Lines))
			for _, line := range hunk.Lines {
				lines = append(lines, review.DiffLine{
					Kind:    reviewDiffLineKind(line.Kind),
					OldLine: line.OldLine,
					NewLine: line.NewLine,
					Text:    line.Text,
				})
			}
			return lines
		}
	}
	return nil
}

func reviewDiffLineKind(kind github.DiffLineKind) review.DiffLineKind {
	switch kind {
	case github.DiffLineAdded:
		return review.DiffLineAdded
	case github.DiffLineDeleted:
		return review.DiffLineDeleted
	default:
		return review.DiffLineContext
	}
}
