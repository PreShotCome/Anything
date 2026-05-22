-- +goose Up

-- +goose StatementBegin
-- Subscription state synced from Stripe webhooks. plan already exists; these
-- record the Stripe subscription it derives from and its lifecycle status
-- (active, trialing, past_due, canceled, ...).
ALTER TABLE accounts
    ADD COLUMN stripe_subscription_id TEXT,
    ADD COLUMN subscription_status    TEXT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE accounts
    DROP COLUMN stripe_subscription_id,
    DROP COLUMN subscription_status;
-- +goose StatementEnd
