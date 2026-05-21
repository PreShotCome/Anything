-- +goose Up

-- +goose StatementBegin
-- Per-key scopes gate what a /v1 API key may do. Existing keys are
-- backfilled with the full set so behaviour is unchanged; new keys are
-- created with an explicit scope set chosen on the Account page.
ALTER TABLE api_keys
    ADD COLUMN scopes TEXT[] NOT NULL
    DEFAULT '{"databases:read","databases:write","drills:read","drills:write"}';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE api_keys DROP COLUMN scopes;
-- +goose StatementEnd
