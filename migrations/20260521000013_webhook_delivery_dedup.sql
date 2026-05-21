-- +goose Up

-- +goose StatementBegin
-- event_key dedups deliveries: a retried fan-out of the same event to the
-- same endpoint reuses the existing delivery row instead of creating a
-- duplicate. NULL keys (e.g. dashboard replays) are never deduped.
ALTER TABLE webhook_deliveries ADD COLUMN event_key TEXT;
CREATE UNIQUE INDEX webhook_deliveries_event_key_idx
    ON webhook_deliveries (endpoint_id, event_key) WHERE event_key IS NOT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS webhook_deliveries_event_key_idx;
ALTER TABLE webhook_deliveries DROP COLUMN event_key;
-- +goose StatementEnd
