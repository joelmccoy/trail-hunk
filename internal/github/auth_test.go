package github

import "testing"

func TestTokenFromEnv(t *testing.T) {
	t.Setenv("GITHUB_TOKEN", "abc123")
	token, err := TokenFromEnv()
	if err != nil {
		t.Fatal(err)
	}
	if token != "abc123" {
		t.Fatalf("token = %q, want abc123", token)
	}
}

func TestTokenFromEnvReturnsErrorWhenUnset(t *testing.T) {
	t.Setenv("GITHUB_TOKEN", "")
	_, err := TokenFromEnv()
	if err == nil {
		t.Fatal("expected missing token error")
	}
}
