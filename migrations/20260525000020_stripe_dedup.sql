-- +goose Up

-- +goose StatementBegin
-- Tracks every Stripe event we've already processed so retries and replays
-- become no-ops (Stripe retries 5xx for 3+ days and may also replay older
-- events out of order). Primary key on event.id is the dedup gate.
CREATE TABLE stripe_events (
    event_id    TEXT PRIMARY KEY,
    received_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
-- +goose StatementEnd

-- +goose StatementBegin
-- Records the Stripe event.created timestamp of the most recent subscription
-- event we applied to an account. The webhook handler uses this to reject
-- older events that arrive after a newer one (Stripe does not guarantee
-- order). A NULL here means no subscription event has been applied yet.
ALTER TABLE accounts ADD COLUMN subscription_status_updated_at TIMESTAMPTZ;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE accounts DROP COLUMN subscription_status_updated_at;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE stripe_events;
-- +goose StatementEnd
