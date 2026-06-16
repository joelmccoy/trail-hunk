package github

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDoJSONAddsAuthAndDecodesResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer abc123" {
			t.Fatalf("Authorization = %q, want Bearer abc123", got)
		}
		if got := r.Header.Get("Accept"); got != githubJSONMediaType {
			t.Fatalf("Accept = %q, want %s", got, githubJSONMediaType)
		}
		if r.URL.Path != "/test" {
			t.Fatalf("path = %q, want /test", r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{"name": "trail-hunk"})
	}))
	t.Cleanup(server.Close)

	client := NewClient(server.URL, "abc123")
	var out struct {
		Name string `json:"name"`
	}

	if err := client.DoJSON(context.Background(), http.MethodGet, "/test", nil, &out); err != nil {
		t.Fatal(err)
	}
	if out.Name != "trail-hunk" {
		t.Fatalf("Name = %q, want trail-hunk", out.Name)
	}
}

func TestFindPullRequestsForBranchQueriesOpenHead(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/repos/joelmccoy/trail-hunk/pulls" {
			t.Fatalf("path = %q", r.URL.Path)
		}
		if got := r.URL.Query().Get("state"); got != "open" {
			t.Fatalf("state = %q, want open", got)
		}
		if got := r.URL.Query().Get("head"); got != "joelmccoy:feature" {
			t.Fatalf("head = %q, want joelmccoy:feature", got)
		}

		_ = json.NewEncoder(w).Encode([]PullRequest{
			{Number: 12, Title: "Add guided review"},
		})
	}))
	t.Cleanup(server.Close)

	client := NewClient(server.URL, "abc123")
	prs, err := client.FindPullRequestsForBranch(context.Background(), "joelmccoy", "trail-hunk", "feature")
	if err != nil {
		t.Fatal(err)
	}
	if len(prs) != 1 {
		t.Fatalf("len(prs) = %d, want 1", len(prs))
	}
	if prs[0].Number != 12 {
		t.Fatalf("Number = %d, want 12", prs[0].Number)
	}
}

func TestPullRequestDiffRequestsDiffMediaType(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/repos/joelmccoy/trail-hunk/pulls/12" {
			t.Fatalf("path = %q", r.URL.Path)
		}
		if got := r.Header.Get("Accept"); got != githubDiffMediaType {
			t.Fatalf("Accept = %q, want %s", got, githubDiffMediaType)
		}

		_, _ = w.Write([]byte("diff --git a/app.go b/app.go\n"))
	}))
	t.Cleanup(server.Close)

	client := NewClient(server.URL, "abc123")
	diff, err := client.PullRequestDiff(context.Background(), "joelmccoy", "trail-hunk", 12)
	if err != nil {
		t.Fatal(err)
	}
	if diff != "diff --git a/app.go b/app.go\n" {
		t.Fatalf("diff = %q", diff)
	}
}
