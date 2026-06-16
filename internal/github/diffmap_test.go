package github

import "testing"

const sampleDiff = `diff --git a/app.go b/app.go
index 2c26b46..81b637d 100644
--- a/app.go
+++ b/app.go
@@ -1,4 +1,4 @@
 package main
-func oldName() {}
+func newName() {}
 func keep() {}
`

func TestParsePullRequestDiffMapsLineTargets(t *testing.T) {
	diff, err := ParsePullRequestDiff(sampleDiff)
	if err != nil {
		t.Fatal(err)
	}

	if len(diff.Files) != 1 {
		t.Fatalf("len(Files) = %d, want 1", len(diff.Files))
	}
	if diff.Files[0].Path != "app.go" {
		t.Fatalf("Path = %q, want app.go", diff.Files[0].Path)
	}
	if len(diff.Files[0].Hunks) != 1 {
		t.Fatalf("len(Hunks) = %d, want 1", len(diff.Files[0].Hunks))
	}

	added, err := diff.FindTarget("app.go", SideRight, 2)
	if err != nil {
		t.Fatal(err)
	}
	if added.HunkID != "app.go:1" {
		t.Fatalf("added HunkID = %q, want app.go:1", added.HunkID)
	}

	deleted, err := diff.FindTarget("app.go", SideLeft, 2)
	if err != nil {
		t.Fatal(err)
	}
	if deleted.Line != 2 || deleted.Side != SideLeft {
		t.Fatalf("deleted target = %+v, want LEFT line 2", deleted)
	}

	context, err := diff.FindTarget("app.go", SideRight, 3)
	if err != nil {
		t.Fatal(err)
	}
	if context.Line != 3 || context.Side != SideRight {
		t.Fatalf("context target = %+v, want RIGHT line 3", context)
	}
}

func TestFindTargetRejectsMissingDiffLine(t *testing.T) {
	diff, err := ParsePullRequestDiff(sampleDiff)
	if err != nil {
		t.Fatal(err)
	}

	_, err = diff.FindTarget("app.go", SideRight, 99)
	if err == nil {
		t.Fatal("expected missing target error")
	}
}
