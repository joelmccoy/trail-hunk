package ai

import (
	"context"
	"fmt"
)

const fixtureReviewFile = "dev/fixtures/dummy-pr/review_target.go"

type FixtureProvider struct{}

func NewFixtureProvider() FixtureProvider {
	return FixtureProvider{}
}

func (FixtureProvider) Name() string {
	return "fixture"
}

func (FixtureProvider) Models(context.Context) ([]ModelRef, error) {
	return []ModelRef{{Name: "dummy-review"}}, nil
}

func (FixtureProvider) Review(context.Context, ReviewRequest) (ReviewResponse, error) {
	return ReviewResponse{
		Overview: "This dummy PR adds a small account-access helper with intentionally reviewable behavior. Use it to exercise the local walkthrough, AI suggestions, comment approval, queueing, and GitHub review submission flow.",
		Risks: []Risk{
			{
				Priority: "high",
				Category: "security",
				Summary:  "Blank billing account IDs currently allow access.",
			},
			{
				Priority: "medium",
				Category: "correctness",
				Summary:  "Display-name truncation slices bytes and can split multi-byte characters.",
			},
		},
		ReviewOrder: []ReviewStep{
			{
				ID:         "fixture-access-control",
				FilePath:   fixtureReviewFile,
				HunkID:     fixtureReviewFile + ":1",
				Title:      "Review billing access guard",
				GroupID:    "fixture-account-helpers",
				GroupTitle: "Fixture account helpers",
				LayerIndex: 1,
				LayerTitle: "Billing access guard",
				Summary:    "The helper grants access when the requested account ID is blank.",
				Why:        "Empty or malformed resource identifiers should fail closed so callers cannot accidentally bypass authorization.",
				Focus: []string{
					"Check whether blank requestedAccountID should ever be valid.",
					"Confirm admin bypass behavior is intentional and audited.",
				},
				Suggestions: []SuggestedComment{
					{
						FilePath: fixtureReviewFile,
						Side:     "RIGHT",
						Line:     14,
						Body:     "This branch fails open for a blank requested account ID. Consider returning false here, or validating the request before calling this helper, so malformed input cannot grant billing access.",
						Priority: "high",
						Category: "security",
					},
					{
						FilePath: fixtureReviewFile,
						Side:     "RIGHT",
						Line:     18,
						Body:     "The admin bypass may be intended, but it would be safer to make that policy explicit in the function name, documentation, or a caller-side authorization check.",
						Priority: "medium",
						Category: "question",
					},
				},
			},
			{
				ID:         "fixture-display-name",
				FilePath:   fixtureReviewFile,
				HunkID:     fixtureReviewFile + ":1",
				Title:      "Review display-name normalization",
				GroupID:    "fixture-account-helpers",
				GroupTitle: "Fixture account helpers",
				LayerIndex: 2,
				LayerTitle: "Display-name normalization",
				Summary:    "The display-name helper trims whitespace and truncates long names.",
				Why:        "User-visible strings often contain multi-byte characters, so byte slicing can produce invalid UTF-8 or surprising output.",
				Focus: []string{
					"Check whether the length limit is bytes, runes, or display cells.",
					"Look for tests around Unicode display names.",
				},
				Suggestions: []SuggestedComment{
					{
						FilePath: fixtureReviewFile,
						Side:     "RIGHT",
						Line:     27,
						Body:     "This truncates by byte index, which can split multi-byte characters. Convert to runes or use a display-width-aware helper before slicing user-visible names.",
						Priority: "medium",
						Category: "correctness",
					},
				},
			},
		},
	}, nil
}

func (FixtureProvider) Ask(_ context.Context, req AskRequest) (AskResponse, error) {
	return AskResponse{Answer: fmt.Sprintf("fixture answer: current context is %q and the question was %q.", req.Context, req.Question)}, nil
}

func (FixtureProvider) Reword(_ context.Context, req RewordRequest) (RewordResponse, error) {
	if req.Instruction == "" {
		return RewordResponse{Body: req.Body}, nil
	}
	return RewordResponse{Body: fmt.Sprintf("%s\n\nRewording note: %s", req.Body, req.Instruction)}, nil
}
