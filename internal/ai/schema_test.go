package ai

import "testing"

func TestDecodeReviewResponseValidatesStructuredJSON(t *testing.T) {
	raw := []byte(`{
		"overview": "Adds a guided review workflow.",
		"risks": [
			{"priority": "high", "category": "bug", "summary": "Diff targets must map to GitHub lines."}
		],
		"review_order": [
			{
				"id": "step-1",
				"file_path": "app.go",
				"hunk_id": "app.go:1",
				"title": "Review app startup",
				"summary": "Config is loaded before the TUI starts.",
				"why": "Provider choice affects the rest of the flow.",
				"focus": ["startup", "configuration"],
				"suggestions": [
					{
						"file_path": "app.go",
						"side": "RIGHT",
						"line": 12,
						"body": "Consider surfacing this error in the TUI.",
						"priority": "medium",
						"category": "maintainability"
					}
				]
			}
		]
	}`)

	response, err := DecodeReviewResponse(raw)
	if err != nil {
		t.Fatal(err)
	}
	if response.Overview == "" {
		t.Fatal("expected overview")
	}
	if len(response.ReviewOrder) != 1 {
		t.Fatalf("len(ReviewOrder) = %d, want 1", len(response.ReviewOrder))
	}
}

func TestDecodeReviewResponseRequiresOverview(t *testing.T) {
	_, err := DecodeReviewResponse([]byte(`{"review_order":[]}`))
	if err == nil {
		t.Fatal("expected missing overview error")
	}
}

func TestDecodeReviewResponseRejectsInvalidPriority(t *testing.T) {
	_, err := DecodeReviewResponse([]byte(`{
		"overview": "Adds a guided review workflow.",
		"review_order": [
			{
				"id": "step-1",
				"file_path": "app.go",
				"hunk_id": "app.go:1",
				"title": "Review app startup",
				"summary": "Config is loaded before the TUI starts.",
				"why": "Provider choice affects the rest of the flow.",
				"suggestions": [
					{
						"file_path": "app.go",
						"side": "RIGHT",
						"line": 12,
						"body": "Consider surfacing this error in the TUI.",
						"priority": "urgent",
						"category": "maintainability"
					}
				]
			}
		]
	}`))
	if err == nil {
		t.Fatal("expected invalid priority error")
	}
}
