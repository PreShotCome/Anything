-- +goose Up

-- +goose StatementBegin
-- The time-step counter of the last TOTP code accepted at login. A code
-- whose counter is <= this value is a replay and is rejected (RFC 6238 §5.2).
ALTER TABLE users ADD COLUMN totp_last_used_counter BIGINT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE users DROP COLUMN totp_last_used_counter;
-- +goose StatementEnd
