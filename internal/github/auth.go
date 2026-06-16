package github

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func TokenFromEnv() (string, error) {
	token := strings.TrimSpace(os.Getenv("GITHUB_TOKEN"))
	if token == "" {
		return "", errors.New("GITHUB_TOKEN is not set")
	}
	return token, nil
}

func TokenFromGH(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, "gh", "auth", "token")
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("run gh auth token: %w", err)
	}

	token := strings.TrimSpace(string(out))
	if token == "" {
		return "", errors.New("gh auth token returned an empty token")
	}
	return token, nil
}

func DiscoverToken(ctx context.Context) (string, error) {
	if token, err := TokenFromEnv(); err == nil {
		return token, nil
	}

	if token, err := TokenFromGH(ctx); err == nil {
		return token, nil
	}

	return "", errors.New("GitHub authentication not found; set GITHUB_TOKEN or run gh auth login")
}
