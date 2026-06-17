package ai

import "context"

type Provider interface {
	Name() string
	Models(ctx context.Context) ([]ModelRef, error)
	Review(ctx context.Context, req ReviewRequest) (ReviewResponse, error)
	Ask(ctx context.Context, req AskRequest) (AskResponse, error)
	Reword(ctx context.Context, req RewordRequest) (RewordResponse, error)
}

type ModelRef struct {
	Name string
}

type ReviewRequest struct {
	PRTitle          string
	PRBody           string
	Diff             string
	ExistingComments []string
}

type AskRequest struct {
	Question string
	Context  string
}

type RewordRequest struct {
	Body        string
	Instruction string
}

type ReviewResponse struct {
	Overview    string       `json:"overview"`
	Risks       []Risk       `json:"risks"`
	ReviewOrder []ReviewStep `json:"review_order"`
}

type Risk struct {
	Priority string `json:"priority"`
	Category string `json:"category"`
	Summary  string `json:"summary"`
}

type ReviewStep struct {
	ID          string             `json:"id"`
	FilePath    string             `json:"file_path"`
	HunkID      string             `json:"hunk_id"`
	Title       string             `json:"title"`
	GroupID     string             `json:"group_id,omitempty"`
	GroupTitle  string             `json:"group_title,omitempty"`
	LayerIndex  int                `json:"layer_index,omitempty"`
	LayerTitle  string             `json:"layer_title,omitempty"`
	Summary     string             `json:"summary"`
	Why         string             `json:"why"`
	Focus       []string           `json:"focus"`
	Suggestions []SuggestedComment `json:"suggestions"`
}

type SuggestedComment struct {
	FilePath  string `json:"file_path"`
	Side      string `json:"side"`
	Line      int    `json:"line"`
	StartLine *int   `json:"start_line,omitempty"`
	Body      string `json:"body"`
	Priority  string `json:"priority"`
	Category  string `json:"category"`
}

type AskResponse struct {
	Answer string `json:"answer"`
}

type RewordResponse struct {
	Body string `json:"body"`
}
