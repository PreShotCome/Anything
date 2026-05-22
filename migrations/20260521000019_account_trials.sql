-- +goose Up
-- Time-limited trial: an account on the trial plan has full access until
-- trial_ends_at, after which writes are blocked until it subscribes.
ALTER TABLE accounts ADD COLUMN trial_ends_at timestamptz;

-- Existing trial accounts get a 14-day window from when they were created.
UPDATE accounts
   SET trial_ends_at = created_at + interval '14 days'
 WHERE plan = 'trial';

-- +goose Down
ALTER TABLE accounts DROP COLUMN trial_ends_at;
