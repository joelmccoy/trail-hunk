package tui

import (
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

func key(value string) tea.KeyMsg {
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(value)}
}
