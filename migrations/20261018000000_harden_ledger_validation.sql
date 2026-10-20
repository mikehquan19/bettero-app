-- +goose Up

ALTER TABLE account_histories
ALTER COLUMN logged_time TYPE TIMESTAMPTZ
USING logged_time AT TIME ZONE 'UTC';

ALTER TABLE accounts
DROP CONSTRAINT IF EXISTS accounts_discrepancy_amount_cents_check;

-- +goose Down

UPDATE accounts
SET discrepancy_amount_cents = ABS(discrepancy_amount_cents);

ALTER TABLE accounts
ADD CONSTRAINT accounts_discrepancy_amount_cents_check
CHECK (discrepancy_amount_cents >= 0);

ALTER TABLE account_histories
ALTER COLUMN logged_time TYPE TIMESTAMP
USING logged_time AT TIME ZONE 'UTC';
