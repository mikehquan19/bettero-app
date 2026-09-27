-- +goose Up

CREATE EXTENSION IF NOT EXISTS pg_trgm;

SELECT 'up SQL query';

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    first_name VARCHAR(50) NOT NULL,
    last_name VARCHAR(50) NOT NULL,

    username VARCHAR(100) UNIQUE,
    email VARCHAR(100) UNIQUE,
    password_hash VARCHAR(100) NOT NULL,

    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE accounts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    user_id UUID NOT NULL
        REFERENCES users(id)
        ON DELETE CASCADE,

    account_number INT UNIQUE,
    name VARCHAR(50) NOT NULL,
    institution VARCHAR(50) NOT NULL
        CHECK (
            institution IN (
                'JP Morgan Chase',
                'Bank of America',
                'Wells Fargo',
                'Citi Bank',
                'Capital One',
                'Discover',
                'Sofi Bank',
                'Ally Bank'
            )
        ),

    type VARCHAR(20) NOT NULL
        CHECK (
            type IN ('Debit', 'Credit')
        ),

    balance_cents BIGINT NOT NULL DEFAULT 0 
        CHECK (balance_cents >= 0),

    credit_limit_cents BIGINT,
    next_due TIMESTAMP,
    discrepancy_flagged BOOLEAN DEFAULT FALSE,

    discrepancy_amount_cents BIGINT NOT NULL DEFAULT 0
        CHECK (discrepancy_amount_cents >= 0),

    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    account_id UUID NOT NULL
        REFERENCES accounts(id)
        ON DELETE CASCADE,

    merchant VARCHAR(50) NOT NULL,
    description VARCHAR(200) NOT NULL,
    category VARCHAR(30) NOT NULL
        CHECK (
            category IN (
                'income',
                'housing',
                'automobile',
                'medical',
                'subscription',
                'grocery',
                'dining',
                'shopping',
                'gas',
                'others'
            )
        ),

    -- Stored in cents.
    -- No DEFAULT 0 because that would violate this constraint.
    amount_cents BIGINT NOT NULL CHECK (amount_cents >= 1),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);


-- =========================================================
-- Budget plans
-- =========================================================

CREATE TABLE budget_plans (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    user_id UUID NOT NULL
        REFERENCES users(id)
        ON DELETE CASCADE,

    interval_type VARCHAR(20) NOT NULL
        CHECK (
            interval_type IN (
                'month',
                'bi_week',
                'week'
            )
        ),

    recurring_income_cents BIGINT NOT NULL
        CHECK (recurring_income_cents >= 1),

    expense_portion NUMERIC(5, 2) NOT NULL
        CHECK (
            expense_portion >= 0
            AND expense_portion <= 100
        ),

    category_portion JSONB NOT NULL DEFAULT '{
        "housing": 10,
        "automobile": 10,
        "medical": 10,
        "subscription": 10,
        "grocery": 10,
        "dining": 10,
        "shopping": 10,
        "gas": 10,
        "others": 20
    }'::JSONB,

    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);


-- =========================================================
-- Bills
-- =========================================================

CREATE TABLE bills (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    account_id UUID
        REFERENCES accounts(id)
        ON DELETE SET NULL,

    merchant VARCHAR(50) NOT NULL,
    description VARCHAR(200) NOT NULL,
    category VARCHAR(30) NOT NULL
        CHECK (
            category IN (
                'housing',
                'automobile',
                'medical',
                'subscription',
                'grocery',
                'dining',
                'shopping',
                'gas',
                'others'
            )
        ),

    amount_cents BIGINT NOT NULL
        CHECK (amount_cents >= 100),

    due_date DATE NOT NULL
);

CREATE TABLE account_histories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    account_id UUID NOT NULL
        REFERENCES accounts(id)
        ON DELETE CASCADE,

    logged_time TIMESTAMP DEFAULT NOW(),
    balance_cents BIGINT NOT NULL DEFAULT 0
        CHECK (balance_cents >= 0)
);

-- +goose Down

SELECT 'down SQL query';

DROP TABLE IF EXISTS account_histories;
DROP TABLE IF EXISTS bills;
DROP TABLE IF EXISTS budget_plans;
DROP TABLE IF EXISTS transactions;
DROP TABLE IF EXISTS accounts;
DROP TABLE IF EXISTS users;