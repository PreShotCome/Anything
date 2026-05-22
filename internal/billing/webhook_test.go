package billing

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"testing"
	"time"
)

// signEvent builds a valid Stripe-Signature header for a payload + secret.
func signEvent(secret string, ts int64, payload []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	fmt.Fprintf(mac, "%d.", ts)
	mac.Write(payload)
	return fmt.Sprintf("t=%d,v1=%s", ts, hex.EncodeToString(mac.Sum(nil)))
}

func TestVerifyWebhook(t *testing.T) {
	s := &stripeService{cfg: Config{WebhookSecret: "whsec_test_secret"}}
	payload := []byte(`{"type":"customer.subscription.updated"}`)
	now := time.Now().Unix()

	if err := s.VerifyWebhook(payload, signEvent("whsec_test_secret", now, payload)); err != nil {
		t.Fatalf("a valid signature should verify: %v", err)
	}
	// Wrong secret.
	if err := s.VerifyWebhook(payload, signEvent("whsec_wrong", now, payload)); err == nil {
		t.Fatal("a signature under the wrong secret must not verify")
	}
	// Tampered payload.
	if err := s.VerifyWebhook([]byte(`{"type":"evil"}`), signEvent("whsec_test_secret", now, payload)); err == nil {
		t.Fatal("a tampered payload must not verify")
	}
	// Stale timestamp (outside tolerance).
	old := time.Now().Add(-30 * time.Minute).Unix()
	if err := s.VerifyWebhook(payload, signEvent("whsec_test_secret", old, payload)); err == nil {
		t.Fatal("a stale webhook timestamp must be rejected")
	}
	// Malformed header.
	if err := s.VerifyWebhook(payload, "not-a-signature"); err == nil {
		t.Fatal("a malformed signature header must be rejected")
	}
}

func TestParseWebhook(t *testing.T) {
	body := []byte(`{
		"type": "customer.subscription.updated",
		"data": {"object": {
			"id": "sub_123", "customer": "cus_456", "status": "active",
			"items": {"data": [{"price": {"id": "price_pro"}}]}
		}}
	}`)
	ev, err := ParseWebhook(body)
	if err != nil {
		t.Fatalf("ParseWebhook: %v", err)
	}
	if ev.Type != "customer.subscription.updated" || ev.CustomerID != "cus_456" ||
		ev.SubscriptionID != "sub_123" || ev.Status != "active" || ev.PriceID != "price_pro" {
		t.Fatalf("ParseWebhook returned %+v", ev)
	}
	if !IsSubscriptionEvent(ev.Type) {
		t.Fatal("subscription.updated should be a subscription event")
	}
	if IsSubscriptionEvent("invoice.paid") {
		t.Fatal("invoice.paid is not a subscription event")
	}
}
