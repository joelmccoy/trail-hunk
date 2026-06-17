package dummypr

import "testing"

func TestNormalizeDisplayNameUnicode(t *testing.T) {
	got := NormalizeDisplayName("José 🚀 customer account")
	if got == "" {
		t.Fatal("expected display name")
	}
}
