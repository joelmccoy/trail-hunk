package tui

import (
	"github.com/charmbracelet/bubbles/help"
	bubblekey "github.com/charmbracelet/bubbles/key"
)

type keyMap struct {
	StartReview        bubblekey.Binding
	FocusNext          bubblekey.Binding
	MoveDown           bubblekey.Binding
	MoveUp             bubblekey.Binding
	NextStep           bubblekey.Binding
	PreviousStep       bubblekey.Binding
	NextSuggestion     bubblekey.Binding
	PreviousSuggestion bubblekey.Binding
	NextFile           bubblekey.Binding
	PreviousFile       bubblekey.Binding
	FocusMode          bubblekey.Binding
	ToggleViewed       bubblekey.Binding
	Accept             bubblekey.Binding
	Dismiss            bubblekey.Binding
	Queue              bubblekey.Binding
	Submit             bubblekey.Binding
	ToggleFiles        bubblekey.Binding
	ToggleAsk          bubblekey.Binding
	Search             bubblekey.Binding
	MoreHelp           bubblekey.Binding
	Quit               bubblekey.Binding
}

func defaultKeyMap() keyMap {
	return keyMap{
		StartReview: bubblekey.NewBinding(
			bubblekey.WithKeys(keyStartReview),
			bubblekey.WithHelp("R", "review"),
		),
		FocusNext: bubblekey.NewBinding(
			bubblekey.WithKeys("tab"),
			bubblekey.WithHelp("tab", "focus"),
		),
		MoveDown: bubblekey.NewBinding(
			bubblekey.WithKeys(keyMoveDown, "down"),
			bubblekey.WithHelp("j", "down"),
		),
		MoveUp: bubblekey.NewBinding(
			bubblekey.WithKeys(keyMoveUp, "up"),
			bubblekey.WithHelp("k", "up"),
		),
		NextStep: bubblekey.NewBinding(
			bubblekey.WithKeys(keyNextStep),
			bubblekey.WithHelp("n", "next"),
		),
		PreviousStep: bubblekey.NewBinding(
			bubblekey.WithKeys(keyPreviousStep),
			bubblekey.WithHelp("p", "prev"),
		),
		NextSuggestion: bubblekey.NewBinding(
			bubblekey.WithKeys(keyNextSuggestion),
			bubblekey.WithHelp("J", "note"),
		),
		PreviousSuggestion: bubblekey.NewBinding(
			bubblekey.WithKeys(keyPreviousSuggestion),
			bubblekey.WithHelp("K", "note"),
		),
		NextFile: bubblekey.NewBinding(
			bubblekey.WithKeys(keyNextFile),
			bubblekey.WithHelp("]", "file"),
		),
		PreviousFile: bubblekey.NewBinding(
			bubblekey.WithKeys(keyPreviousFile),
			bubblekey.WithHelp("[", "file"),
		),
		FocusMode: bubblekey.NewBinding(
			bubblekey.WithKeys(keyFocusMode),
			bubblekey.WithHelp("z", "focus"),
		),
		ToggleViewed: bubblekey.NewBinding(
			bubblekey.WithKeys(keyToggleViewed),
			bubblekey.WithHelp("v", "viewed"),
		),
		Accept: bubblekey.NewBinding(
			bubblekey.WithKeys(keyAcceptComment),
			bubblekey.WithHelp("a", "approve"),
		),
		Dismiss: bubblekey.NewBinding(
			bubblekey.WithKeys(keyDismissComment),
			bubblekey.WithHelp("d", "dismiss"),
		),
		Queue: bubblekey.NewBinding(
			bubblekey.WithKeys(keyComments),
			bubblekey.WithHelp("C", "queue"),
		),
		Submit: bubblekey.NewBinding(
			bubblekey.WithKeys(keySubmitReview),
			bubblekey.WithHelp("S", "submit"),
		),
		ToggleFiles: bubblekey.NewBinding(
			bubblekey.WithKeys(keyToggleFiles),
			bubblekey.WithHelp("f", "files"),
		),
		ToggleAsk: bubblekey.NewBinding(
			bubblekey.WithKeys(keyToggleAskPane),
			bubblekey.WithHelp("t", "ask"),
		),
		Search: bubblekey.NewBinding(
			bubblekey.WithKeys("/"),
			bubblekey.WithHelp("/", "search"),
		),
		MoreHelp: bubblekey.NewBinding(
			bubblekey.WithKeys("?"),
			bubblekey.WithHelp("?", "help"),
		),
		Quit: bubblekey.NewBinding(
			bubblekey.WithKeys(keyQuit, "ctrl+c"),
			bubblekey.WithHelp("q", "quit"),
		),
	}
}

func (k keyMap) ShortHelp() []bubblekey.Binding {
	return []bubblekey.Binding{
		k.StartReview,
		k.FocusNext,
		k.MoveDown,
		k.MoveUp,
		k.NextStep,
		k.PreviousStep,
		k.NextFile,
		k.PreviousFile,
		k.FocusMode,
		k.ToggleViewed,
		k.Accept,
		k.Dismiss,
		k.Queue,
		k.Submit,
		k.ToggleFiles,
		k.ToggleAsk,
		k.Quit,
	}
}

func (k keyMap) FullHelp() [][]bubblekey.Binding {
	return [][]bubblekey.Binding{
		{k.StartReview, k.FocusNext, k.MoveDown, k.MoveUp, k.NextStep, k.PreviousStep},
		{k.NextFile, k.PreviousFile, k.NextSuggestion, k.PreviousSuggestion, k.FocusMode, k.ToggleViewed},
		{k.Accept, k.Dismiss, k.Queue, k.Submit},
		{k.ToggleFiles, k.ToggleAsk, k.Search, k.MoreHelp, k.Quit},
	}
}

func newHelpModel() help.Model {
	h := help.New()
	h.ShowAll = false
	return h
}
