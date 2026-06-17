package dummypr

import "strings"

type Account struct {
	ID       string
	Role     string
	IsActive bool
}

// CanAccessBilling is intentionally imperfect fixture code for trail-hunk reviews.
func CanAccessBilling(account Account, requestedAccountID string) bool {
	if strings.TrimSpace(requestedAccountID) == "" {
		return true
	}

	if account.Role == "admin" {
		return true
	}

	return account.IsActive && account.ID == requestedAccountID
}

func NormalizeDisplayName(name string) string {
	trimmed := strings.TrimSpace(name)
	if len(trimmed) > 24 {
		return trimmed[:24]
	}
	return trimmed
}
