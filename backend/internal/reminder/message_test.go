package reminder

import (
	"strings"
	"testing"
)

func TestMessage_EmptyWhenNothingPending(t *testing.T) {
	if got := Message(nil); got != "" {
		t.Fatalf("expected empty message with no pending categories, got: %q", got)
	}
	if got := Message([]CategoryName{}); got != "" {
		t.Fatalf("expected empty message with no pending categories, got: %q", got)
	}
}

func TestMessage_NamesPendingCategories(t *testing.T) {
	got := Message([]CategoryName{{Name: "Kopi"}, {Name: "Makanan"}})

	if !strings.Contains(got, "Kopi") || !strings.Contains(got, "Makanan") {
		t.Fatalf("expected both category names in the message, got: %q", got)
	}
}

// TestMessage_NeverUsesGuiltInducingLanguage is the direct proof of
// TASK-013's second acceptance criterion.
func TestMessage_NeverUsesGuiltInducingLanguage(t *testing.T) {
	got := Message([]CategoryName{{Name: "Kopi"}})

	guiltPhrases := []string{
		"kamu lupa", "Anda lupa", "seharusnya", "jangan lupa",
		"harus segera", "wajib", "terlambat", "gagal", "kelalaian",
	}
	for _, phrase := range guiltPhrases {
		if strings.Contains(strings.ToLower(got), strings.ToLower(phrase)) {
			t.Fatalf("reminder text must never use guilt-inducing language, found %q in: %q", phrase, got)
		}
	}
	if strings.Contains(got, "!") {
		t.Fatalf("reminder text must not use urgency punctuation, got: %q", got)
	}
}
