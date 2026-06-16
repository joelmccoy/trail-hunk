package ai

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

const (
	DefaultReviewTimeout = 5 * time.Minute
	DefaultQuickTimeout  = time.Minute
)

type Runner struct {
	Command string
	Args    []string
	Timeout time.Duration
}

func (r Runner) Run(ctx context.Context, stdin string) ([]byte, error) {
	timeout := r.Timeout
	if timeout == 0 {
		timeout = DefaultQuickTimeout
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, r.Command, r.Args...)
	cmd.Stdin = strings.NewReader(stdin)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		if ctx.Err() != nil {
			return nil, fmt.Errorf("run %s timed out after %s", r.Command, timeout)
		}
		return nil, fmt.Errorf("run %s: %w: %s", r.Command, err, strings.TrimSpace(stderr.String()))
	}

	return stdout.Bytes(), nil
}
