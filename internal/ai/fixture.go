package ai

import (
	"context"
	"fmt"
)

const (
	fixtureReviewFile  = "dev/fixtures/dummy-pr/review_target.go"
	fixtureHandlerFile = "dev/fixtures/dummy-pr/billing_handler.go"
	fixtureTestFile    = "dev/fixtures/dummy-pr/review_target_test.go"
)

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
			{
				ID:         "fixture-handler-callsite",
				FilePath:   fixtureHandlerFile,
				HunkID:     fixtureHandlerFile + ":1",
				Title:      "Review permission check callsite",
				GroupID:    "fixture-billing-flow",
				GroupTitle: "Billing request flow",
				LayerIndex: 3,
				LayerTitle: "Permission check callsite",
				Summary:    "The handler delegates billing access to the new helper.",
				Why:        "The helper now controls a request path, so boundary validation and auditing matter.",
				Focus: []string{
					"Verify empty account IDs cannot reach the helper.",
					"Check whether denied access is logged or observable.",
				},
				Suggestions: []SuggestedComment{
					{
						FilePath: fixtureHandlerFile,
						Side:     "RIGHT",
						Line:     16,
						Body:     "Validate request.AccountID before calling the helper so malformed identifiers fail closed at the boundary.",
						Priority: "high",
						Category: "security",
					},
				},
			},
			{
				ID:         "fixture-display-test",
				FilePath:   fixtureTestFile,
				HunkID:     fixtureTestFile + ":1",
				Title:      "Review Unicode display-name test",
				GroupID:    "fixture-billing-flow",
				GroupTitle: "Billing request flow",
				LayerIndex: 4,
				LayerTitle: "Unicode display-name test",
				Summary:    "A regression test documents display-name behavior.",
				Why:        "Tests should make the intended truncation semantics explicit before reviewers approve the helper.",
				Focus: []string{
					"Confirm the expected output remains valid UTF-8.",
					"Add a boundary case for exactly 24 display cells.",
				},
				Suggestions: []SuggestedComment{
					{
						FilePath: fixtureTestFile,
						Side:     "RIGHT",
						Line:     8,
						Body:     "This assertion only checks for a non-empty value. Assert the exact expected Unicode-safe result so the test catches byte-slicing regressions.",
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
