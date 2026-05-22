package account

// Unlimited marks a resource as uncapped. It is the zero value, so a Limits
// field left unset means "no cap" — which is exactly the Pro tier.
const Unlimited = 0

// Limits is the per-tier cap on the countable resources an account owns. A
// field set to Unlimited is not enforced.
type Limits struct {
	Databases int
	Seats     int // members + pending invitations
	APIKeys   int // active (non-revoked) keys
	Webhooks  int
}

// LimitsFor returns the resource caps for a plan tier. Pro is uncapped; an
// unknown plan is treated as trial — the most restrictive — so a bad value
// can never widen access.
func LimitsFor(p Plan) Limits {
	switch p {
	case PlanPro:
		return Limits{} // all Unlimited
	case PlanStarter:
		return Limits{Databases: 5, Seats: 10, APIKeys: 5, Webhooks: 5}
	default:
		return Limits{Databases: 1, Seats: 2, APIKeys: 1, Webhooks: 1}
	}
}

// AtLimit reports whether an account already holding `count` of a resource
// has reached a cap of `limit` — i.e. creating one more is not allowed. An
// Unlimited cap never blocks.
func AtLimit(count, limit int) bool {
	return limit != Unlimited && count >= limit
}
