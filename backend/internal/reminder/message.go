// Package reminder generates the neutral reminder text shown for
// planned categories that don't have a recorded transaction yet. It has
// no database dependency, which keeps the "never guilt-induce" wording
// rule directly unit-testable.
package reminder

import (
	"fmt"
	"strings"
)

// CategoryName is the minimal shape Message needs for one pending
// category.
type CategoryName struct {
	Name string
}

// Message builds the reminder text for a date's pending categories.
// Returns an empty string when there's nothing pending - a reminder is
// never shown just to say everything is fine, since that isn't a
// reminder at all. The wording is a neutral, factual statement ("belum
// ada catatan untuk...") - it never uses guilt, blame, or urgency
// language ("kamu lupa", "seharusnya", "jangan lupa", "!").
func Message(pending []CategoryName) string {
	if len(pending) == 0 {
		return ""
	}

	names := make([]string, len(pending))
	for i, c := range pending {
		names[i] = c.Name
	}

	return fmt.Sprintf(
		"Belum ada catatan pengeluaran untuk kategori: %s. Anda bisa mencatatnya kapan saja, atau melewati peninjauan hari ini.",
		strings.Join(names, ", "),
	)
}
