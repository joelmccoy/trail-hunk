package git

import "testing"

func TestParseGitHubRemote(t *testing.T) {
	tests := map[string]struct {
		remote string
		owner  string
		repo   string
	}{
		"ssh":             {"git@github.com:joelmccoy/trail-hunk.git", "joelmccoy", "trail-hunk"},
		"ssh without git": {"git@github.com:joelmccoy/trail-hunk", "joelmccoy", "trail-hunk"},
		"https":           {"https://github.com/joelmccoy/trail-hunk.git", "joelmccoy", "trail-hunk"},
		"https no git":    {"https://github.com/joelmccoy/trail-hunk", "joelmccoy", "trail-hunk"},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := ParseGitHubRemote(tt.remote)
			if err != nil {
				t.Fatal(err)
			}
			if got.Owner != tt.owner || got.Name != tt.repo {
				t.Fatalf("got %s/%s, want %s/%s", got.Owner, got.Name, tt.owner, tt.repo)
			}
		})
	}
}

func TestParseGitHubRemoteRejectsUnsupportedRemote(t *testing.T) {
	_, err := ParseGitHubRemote("https://gitlab.com/joelmccoy/trail-hunk.git")
	if err == nil {
		t.Fatal("expected unsupported remote error")
	}
}
