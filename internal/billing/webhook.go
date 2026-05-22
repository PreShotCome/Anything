package billing

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// webhookTolerance is the clock skew a webhook timestamp may have before it
// is rejected as a possible replay.
const webhookTolerance = 5 * time.Minute

// VerifyWebhook validates a Stripe-Signature header against the raw request
// body, following Stripe's signed-payload scheme: the signed content is
// "<timestamp>.<body>", HMAC-SHA256 under the webhook signing secret.
func (s *stripeService) VerifyWebhook(payload []byte, sigHeader string) error {
	ts, sigs := parseSignatureHeader(sigHeader)
	if ts == "" || len(sigs) == 0 {
		return errors.New("stripe: malformed Stripe-Signature header")
	}
	tsInt, err := strconv.ParseInt(ts, 10, 64)
	if err != nil {
		return errors.New("stripe: bad timestamp in signature")
	}
	if d := time.Since(time.Unix(tsInt, 0)); d > webhookTolerance || d < -webhookTolerance {
		return errors.New("stripe: webhook timestamp outside tolerance")
	}

	mac := hmac.New(sha256.New, []byte(s.cfg.WebhookSecret))
	mac.Write([]byte(ts))
	mac.Write([]byte("."))
	mac.Write(payload)
	want := mac.Sum(nil)

	for _, sig := range sigs {
		if got, err := hex.DecodeString(sig); err == nil && hmac.Equal(got, want) {
			return nil
		}
	}
	return errors.New("stripe: webhook signature does not match")
}

// parseSignatureHeader splits "t=123,v1=abc,v1=def" into its timestamp and
// the one-or-more v1 signatures.
func parseSignatureHeader(h string) (ts string, v1 []string) {
	for _, part := range strings.Split(h, ",") {
		k, v, ok := strings.Cut(strings.TrimSpace(part), "=")
		if !ok {
			continue
		}
		switch k {
		case "t":
			ts = v
		case "v1":
			v1 = append(v1, v)
		}
	}
	return ts, v1
}

// WebhookEvent is the slice of a Stripe event the billing webhook handler
// acts on.
type WebhookEvent struct {
	Type           string
	CustomerID     string
	SubscriptionID string
	Status         string // Stripe subscription status
	PriceID        string
}

// ParseWebhook extracts the fields the app needs from a Stripe event body.
// It understands customer.subscription.* events; other event types parse to
// a WebhookEvent with just Type set, which the handler then ignores.
func ParseWebhook(payload []byte) (WebhookEvent, error) {
	var env struct {
		Type string `json:"type"`
		Data struct {
			Object struct {
				ID       string `json:"id"`
				Customer string `json:"customer"`
				Status   string `json:"status"`
				Items    struct {
					Data []struct {
						Price struct {
							ID string `json:"id"`
						} `json:"price"`
					} `json:"data"`
				} `json:"items"`
			} `json:"object"`
		} `json:"data"`
	}
	if err := json.Unmarshal(payload, &env); err != nil {
		return WebhookEvent{}, fmt.Errorf("stripe: decode webhook: %w", err)
	}
	ev := WebhookEvent{
		Type:           env.Type,
		CustomerID:     env.Data.Object.Customer,
		SubscriptionID: env.Data.Object.ID,
		Status:         env.Data.Object.Status,
	}
	if len(env.Data.Object.Items.Data) > 0 {
		ev.PriceID = env.Data.Object.Items.Data[0].Price.ID
	}
	return ev, nil
}

// IsSubscriptionEvent reports whether an event type is one the handler syncs
// account state from.
func IsSubscriptionEvent(eventType string) bool {
	switch eventType {
	case "customer.subscription.created",
		"customer.subscription.updated",
		"customer.subscription.deleted":
		return true
	default:
		return false
	}
}
