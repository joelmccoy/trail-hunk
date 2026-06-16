package tui

import (
	"context"
	"errors"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/joelmccoy/trail-hunk/internal/review"
)

func TestKeyQQuits(t *testing.T) {
	model := NewModel(review.ReviewSession{})
	_, cmd := model.Update(key("q"))
	if cmd == nil {
		t.Fatal("expected quit command")
	}
	msg := cmd()
	if _, ok := msg.(tea.QuitMsg); !ok {
		t.Fatalf("command returned %T, want tea.QuitMsg", msg)
	}
}

func TestStepNavigationKeys(t *testing.T) {
	model := NewModel(review.ReviewSession{
		Plan: review.WalkthroughPlan{
			ReviewOrder: []review.ReviewStep{
				{ID: "step-1"},
				{ID: "step-2"},
			},
		},
	})

	updated, _ := model.Update(key("n"))
	model = updated.(Model)
	if model.Session.Cursor.StepIndex != 1 {
		t.Fatalf("StepIndex = %d, want 1", model.Session.Cursor.StepIndex)
	}

	updated, _ = model.Update(key("p"))
	model = updated.(Model)
	if model.Session.Cursor.StepIndex != 0 {
		t.Fatalf("StepIndex = %d, want 0", model.Session.Cursor.StepIndex)
	}
}

func TestToggleKeys(t *testing.T) {
	model := NewModel(review.ReviewSession{})

	updated, _ := model.Update(key("f"))
	model = updated.(Model)
	if !model.ShowFileTree {
		t.Fatal("expected file tree visible")
	}

	updated, _ = model.Update(key("t"))
	model = updated.(Model)
	if !model.ShowAskPane {
		t.Fatal("expected ask pane visible")
	}
}

func TestSuggestionSelectionAndActions(t *testing.T) {
	model := NewModel(review.ReviewSession{
		Plan: review.WalkthroughPlan{
			ReviewOrder: []review.ReviewStep{
				{
					ID:      "step-1",
					Title:   "Review startup",
					Summary: "Startup initializes a review.",
					Why:     "The user needs a first step.",
					Suggestions: []review.ReviewComment{
						{ID: "c1", Body: "First suggestion", Status: review.StatusSuggested},
						{ID: "c2", Body: "Second suggestion", Status: review.StatusSuggested},
					},
				},
			},
		},
		Comments: []review.ReviewComment{
			{ID: "c1", Body: "First suggestion", Status: review.StatusSuggested},
			{ID: "c2", Body: "Second suggestion", Status: review.StatusSuggested},
		},
	})
	model.Screen = ScreenWalkthrough

	updated, _ := model.Update(key("j"))
	model = updated.(Model)
	if model.SelectedSuggestion != 1 {
		t.Fatalf("SelectedSuggestion = %d, want 1", model.SelectedSuggestion)
	}

	updated, _ = model.Update(key("a"))
	model = updated.(Model)
	if model.Session.Comments[1].Status != review.StatusApproved {
		t.Fatalf("Status = %q, want approved", model.Session.Comments[1].Status)
	}
	if model.Session.Plan.ReviewOrder[0].Suggestions[1].Status != review.StatusApproved {
		t.Fatalf("step suggestion status = %q, want approved", model.Session.Plan.ReviewOrder[0].Suggestions[1].Status)
	}

	updated, _ = model.Update(key("k"))
	model = updated.(Model)
	if model.SelectedSuggestion != 0 {
		t.Fatalf("SelectedSuggestion = %d, want 0", model.SelectedSuggestion)
	}

	updated, _ = model.Update(key("d"))
	model = updated.(Model)
	if model.Session.Comments[0].Status != review.StatusDismissed {
		t.Fatalf("Status = %q, want dismissed", model.Session.Comments[0].Status)
	}
}

func TestStartReviewKeyRunsStarterAndLoadsWalkthrough(t *testing.T) {
	expected := review.ReviewSession{
		Plan: review.WalkthroughPlan{
			Overview: "Adds a guided review flow.",
			ReviewOrder: []review.ReviewStep{
				{
					ID:      "step-1",
					Title:   "Review startup",
					Summary: "Startup now initializes a review.",
					Why:     "The user needs a concrete first review step.",
				},
			},
		},
	}
	model := NewModelWithStarter(review.ReviewSession{}, func(context.Context) (review.ReviewSession, error) {
		return expected, nil
	})

	updated, cmd := model.Update(key("R"))
	model = updated.(Model)
	if !model.Loading {
		t.Fatal("expected loading after pressing R")
	}
	if model.Screen != ScreenStartup {
		t.Fatalf("Screen = %q, want startup while loading", model.Screen)
	}
	if cmd == nil {
		t.Fatal("expected start review command")
	}

	msg := cmd()
	updated, _ = model.Update(msg)
	model = updated.(Model)
	if model.Loading {
		t.Fatal("did not expect loading after success")
	}
	if model.Screen != ScreenWalkthrough {
		t.Fatalf("Screen = %q, want walkthrough", model.Screen)
	}
	if model.Session.Plan.Overview != expected.Plan.Overview {
		t.Fatalf("Overview = %q", model.Session.Plan.Overview)
	}
}

func TestStartReviewFailureShowsError(t *testing.T) {
	model := NewModelWithStarter(review.ReviewSession{}, func(context.Context) (review.ReviewSession, error) {
		return review.ReviewSession{}, errors.New("no open GitHub pull request found")
	})

	updated, cmd := model.Update(key("R"))
	model = updated.(Model)
	msg := cmd()
	updated, _ = model.Update(msg)
	model = updated.(Model)

	if model.Loading {
		t.Fatal("did not expect loading after failure")
	}
	if model.Err == nil {
		t.Fatal("expected error")
	}
	if model.Screen != ScreenStartup {
		t.Fatalf("Screen = %q, want startup", model.Screen)
	}
	if got := model.View(); !contains(got, "no open GitHub pull request found") {
		t.Fatalf("View() = %q, want error text", got)
	}
}

func key(value string) tea.KeyMsg {
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(value)}
}

func contains(s string, substr string) bool {
	return strings.Contains(s, substr)
}
