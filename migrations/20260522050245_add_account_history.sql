-- +goose Up
SELECT 'up SQL query';

-- Account history
CREATE TABLE account_histories (
    id SERIAL PRIMARY KEY,
    account_id INT NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    logged_time TIMESTAMP DEFAULT NOW(),
    balance NUMERIC(10, 2) NOT NULL DEFAULT 0 CHECK (balance >= 0)
);

ALTER TABLE accounts
ADD COLUMN discrepancy_flagged BOOLEAN DEFAULT FALSE,
ADD COLUMN discrepancy_amount NUMERIC(10, 2) NOT NULL DEFAULT 0 CHECK (discrepancy_amount >= 0);

-- +goose Down
SELECT 'down SQL query';
DROP TABLE IF EXISTS account_histories;
ALTER TABLE accounts
DROP COLUMN IF EXISTS discrepancy_flagged,
DROP COLUMN IF EXISTS discrepancy_amount;

