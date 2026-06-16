package review

import "testing"

func TestCommentWorkflow(t *testing.T) {
	session := ReviewSession{
		Comments: []ReviewComment{
			{
				ID:       "c1",
				FilePath: "app.go",
				Side:     "RIGHT",
				Line:     12,
				Body:     "Consider surfacing this error.",
				Priority: PriorityMedium,
				Category: CategoryMaintainability,
				Status:   StatusSuggested,
				Source:   SourceAI,
			},
			{
				ID:       "c2",
				FilePath: "app.go",
				Side:     "RIGHT",
				Line:     20,
				Body:     "This is too noisy.",
				Priority: PriorityLow,
				Category: CategoryQuestion,
				Status:   StatusSuggested,
				Source:   SourceAI,
			},
		},
	}

	if err := session.AcceptSuggestion("c1"); err != nil {
		t.Fatal(err)
	}
	if session.Comments[0].Status != StatusApproved {
		t.Fatalf("Status = %q, want approved", session.Comments[0].Status)
	}

	if err := session.EditComment("c1", "Please surface this error in the TUI."); err != nil {
		t.Fatal(err)
	}
	if session.Comments[0].Status != StatusEdited {
		t.Fatalf("Status = %q, want edited", session.Comments[0].Status)
	}
	if session.Comments[0].Body != "Please surface this error in the TUI." {
		t.Fatalf("Body = %q", session.Comments[0].Body)
	}

	if err := session.DismissSuggestion("c2"); err != nil {
		t.Fatal(err)
	}
	if session.Comments[1].Status != StatusDismissed {
		t.Fatalf("Status = %q, want dismissed", session.Comments[1].Status)
	}

	manual := session.AddManualComment(ReviewComment{
		FilePath: "main.go",
		Side:     "RIGHT",
		Line:     7,
		Body:     "Can we add a test for this path?",
		Priority: PriorityMedium,
		Category: CategoryTest,
	})
	if manual.Source != SourceUser {
		t.Fatalf("Source = %q, want user", manual.Source)
	}
	if manual.Status != StatusApproved {
		t.Fatalf("Status = %q, want approved", manual.Status)
	}

	approved := session.ApprovedComments()
	if len(approved) != 2 {
		t.Fatalf("len(approved) = %d, want 2", len(approved))
	}
}

func TestReviewCursorNavigation(t *testing.T) {
	session := ReviewSession{
		Plan: WalkthroughPlan{
			ReviewOrder: []ReviewStep{
				{ID: "step-1"},
				{ID: "step-2"},
			},
		},
	}

	if session.Cursor.StepIndex != 0 {
		t.Fatalf("StepIndex = %d, want 0", session.Cursor.StepIndex)
	}
	if moved := session.NextStep(); !moved {
		t.Fatal("expected next step")
	}
	if session.Cursor.StepIndex != 1 {
		t.Fatalf("StepIndex = %d, want 1", session.Cursor.StepIndex)
	}
	if moved := session.NextStep(); moved {
		t.Fatal("did not expect next step at end")
	}
	if moved := session.PreviousStep(); !moved {
		t.Fatal("expected previous step")
	}
	if session.Cursor.StepIndex != 0 {
		t.Fatalf("StepIndex = %d, want 0", session.Cursor.StepIndex)
	}
}
