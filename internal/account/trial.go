package account

import "time"

// TrialDuration is how long a new account has full access before it must
// subscribe. Kept here for reference — the window is set in SQL at account
// creation (CreatePersonalAccount) and on the migration backfill.
const TrialDuration = 14 * 24 * time.Hour

// TrialActive reports whether a trial-plan account is still inside its
// full-access window. A trial account with no end date is treated as active
// so missing data never locks anyone out.
func TrialActive(a Account) bool {
	return a.Plan == PlanTrial && (a.TrialEndsAt == nil || time.Now().Before(*a.TrialEndsAt))
}

// TrialLapsed reports whether a trial account's window has closed without it
// subscribing. Writes should be blocked for a lapsed account.
func TrialLapsed(a Account) bool {
	return a.Plan == PlanTrial && a.TrialEndsAt != nil && time.Now().After(*a.TrialEndsAt)
}

// TrialDaysLeft is the whole days remaining in the trial, rounded up and
// never negative. Returns 0 for paid plans and lapsed trials.
func TrialDaysLeft(a Account) int {
	if a.Plan != PlanTrial || a.TrialEndsAt == nil {
		return 0
	}
	remaining := time.Until(*a.TrialEndsAt)
	if remaining <= 0 {
		return 0
	}
	days := int(remaining / (24 * time.Hour))
	if remaining%(24*time.Hour) > 0 {
		days++
	}
	return days
}
