package ai

import "strings"

func buildReviewPrompt(req ReviewRequest) string {
	var b strings.Builder
	b.WriteString("You are reviewing a GitHub pull request for correctness, security, maintainability, tests, and user impact.\n")
	b.WriteString("Return only strict JSON matching this shape:\n")
	b.WriteString(`{"overview":"...","risks":[{"priority":"high","category":"bug","summary":"..."}],"review_order":[{"id":"step-1","file_path":"path","hunk_id":"path:1","title":"...","summary":"...","why":"...","focus":["..."],"suggestions":[{"file_path":"path","side":"RIGHT","line":1,"body":"...","priority":"medium","category":"maintainability"}]}]}`)
	b.WriteString("\nUse priorities blocker, high, medium, low, or note. Use categories bug, security, correctness, maintainability, performance, test, or question.\n")
	b.WriteString("\nPR title:\n")
	b.WriteString(req.PRTitle)
	b.WriteString("\n\nPR body:\n")
	b.WriteString(req.PRBody)
	if len(req.ExistingComments) > 0 {
		b.WriteString("\n\nExisting comments:\n")
		for _, comment := range req.ExistingComments {
			b.WriteString("- ")
			b.WriteString(comment)
			b.WriteByte('\n')
		}
	}
	b.WriteString("\n\nDiff:\n")
	b.WriteString(req.Diff)
	return b.String()
}
