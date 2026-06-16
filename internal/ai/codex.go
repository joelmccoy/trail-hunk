package ai

import (
	"context"
	"strings"
	"time"
)

type CodexProvider struct {
	Command string
	Model   string
}

func NewCodexProvider(model string) CodexProvider {
	return CodexProvider{Command: "codex", Model: model}
}

func (p CodexProvider) Name() string {
	return "codex"
}

func (p CodexProvider) Models(context.Context) ([]ModelRef, error) {
	if p.Model == "" {
		return nil, nil
	}
	return []ModelRef{{Name: p.Model}}, nil
}

func (p CodexProvider) Review(ctx context.Context, req ReviewRequest) (ReviewResponse, error) {
	out, err := p.run(ctx, buildReviewPrompt(req), DefaultReviewTimeout)
	if err != nil {
		return ReviewResponse{}, err
	}
	return DecodeReviewResponse(out)
}

func (p CodexProvider) Ask(ctx context.Context, req AskRequest) (AskResponse, error) {
	out, err := p.run(ctx, req.Context+"\n\nQuestion: "+req.Question, DefaultQuickTimeout)
	if err != nil {
		return AskResponse{}, err
	}
	return AskResponse{Answer: strings.TrimSpace(string(out))}, nil
}

func (p CodexProvider) Reword(ctx context.Context, req RewordRequest) (RewordResponse, error) {
	prompt := "Reword this review comment. Return only the reworded comment.\n\nInstruction: " + req.Instruction + "\n\nComment:\n" + req.Body
	out, err := p.run(ctx, prompt, DefaultQuickTimeout)
	if err != nil {
		return RewordResponse{}, err
	}
	return RewordResponse{Body: strings.TrimSpace(string(out))}, nil
}

func (p CodexProvider) run(ctx context.Context, prompt string, timeout time.Duration) ([]byte, error) {
	command := p.Command
	if command == "" {
		command = "codex"
	}
	args := []string{"exec", "--color", "never"}
	if p.Model != "" {
		args = append(args, "--model", p.Model)
	}
	args = append(args, "-")

	return Runner{Command: command, Args: args, Timeout: timeout}.Run(ctx, prompt)
}
