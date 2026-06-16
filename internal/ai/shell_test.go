package ai

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestRunnerPassesStdinAndReturnsStdout(t *testing.T) {
	runner := Runner{
		Command: "sh",
		Args:    []string{"-c", "cat"},
		Timeout: time.Second,
	}

	out, err := runner.Run(context.Background(), "hello")
	if err != nil {
		t.Fatal(err)
	}
	if string(out) != "hello" {
		t.Fatalf("stdout = %q, want hello", string(out))
	}
}

func TestRunnerIncludesStderrOnFailure(t *testing.T) {
	runner := Runner{
		Command: "sh",
		Args:    []string{"-c", "echo failed >&2; exit 7"},
		Timeout: time.Second,
	}

	_, err := runner.Run(context.Background(), "")
	if err == nil {
		t.Fatal("expected command failure")
	}
	if !strings.Contains(err.Error(), "failed") {
		t.Fatalf("error = %q, want stderr", err.Error())
	}
}
