-- +goose Up
CREATE TABLE user_sessions (
    token_hash BYTEA PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX user_sessions_expires_at_idx ON user_sessions (expires_at);
CREATE INDEX user_sessions_user_id_idx ON user_sessions (user_id);

ALTER TABLE bills ADD COLUMN user_id UUID REFERENCES users(id) ON DELETE CASCADE;
UPDATE bills b SET user_id = a.user_id
FROM accounts a
WHERE b.account_id = a.id;
UPDATE bills
SET user_id = '2b8ba95c-4f2f-422e-b9bb-ac0521eb6330'
WHERE user_id IS NULL
  AND EXISTS (SELECT 1 FROM users WHERE id = '2b8ba95c-4f2f-422e-b9bb-ac0521eb6330');
CREATE INDEX bills_user_id_due_date_idx ON bills (user_id, due_date);

-- +goose Down
ALTER TABLE bills DROP COLUMN user_id;
DROP TABLE user_sessions;
