-- +goose Up
-- Time-limited trial: an account on the trial plan has full access until
-- trial_ends_at, after which writes are blocked until it subscribes.
ALTER TABLE accounts ADD COLUMN trial_ends_at timestamptz;

-- Backfill: existing trial accounts get a fresh 14-day window starting
-- from when this migration ran, NOT from when they were created. Using
-- created_at would retroactively lapse every dev/staging account older
-- than two weeks the moment the migration applied.
UPDATE accounts
   SET trial_ends_at = now() + interval '14 days'
 WHERE plan = 'trial'
   AND trial_ends_at IS NULL;

-- +goose Down
ALTER TABLE accounts DROP COLUMN trial_ends_at;
