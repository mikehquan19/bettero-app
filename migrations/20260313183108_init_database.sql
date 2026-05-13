-- +goose Up
SELECT 'up SQL query';

CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    first_name VARCHAR(50) NOT NULL,
    last_name VARCHAR(50) NOT NULL,
    username VARCHAR(100) UNIQUE,
    email VARCHAR(100) UNIQUE,
    user_password VARCHAR(100) NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE accounts (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    acc_number INT UNIQUE,
    acc_name VARCHAR(50) NOT NULL,
    institution VARCHAR(50) NOT NULL CHECK (institution IN (
        'JP Morgan Chase'
        'Bank of America',
        'Wells Fargo',
        'Citi Bank',
        'Capital One',
        'Discover',
        'Sofi Bank',
        'Ally Bank'
    )),
    type VARCHAR(20) NOT NULL CHECK (type IN ('Debit', 'Credit')),
    balance NUMERIC(10, 2) NOT NULL DEFAULT 0 CHECK (balance >= 0),
    credit_limit NUMERIC(10, 2), -- null=true
    next_due TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE transactions (
    id SERIAL PRIMARY KEY,
    account_id INT NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    merchant VARCHAR(50) NOT NULL,
    tran_description VARCHAR(200) NOT NULL,
    category VARCHAR(30) NOT NULL CHECK(category IN (
        'Income',
        'Housing',
        'Automobile',
        'Medical',
        'Subscription',
        'Grocery',
        'Dining',
        'Shopping',
        'Gas',
        'Others'
    )),
    amount NUMERIC(8, 2) DEFAULT 0 CHECK(amount >= 0.01),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE budget_plans (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    interval_type VARCHAR(20) NOT NULL CHECK(interval_type IN ('month', 'bi_week', 'week')),
    recurring_income NUMERIC(10,2) NOT NULL CHECK (recurring_income >= 0.01),
    expense_portion NUMERIC(5,2) NOT NULL CHECK (
        expense_portion >= 0 AND expense_portion <= 100
    ),
    category_portion JSONB NOT NULL DEFAULT '{
        "Housing": 10,
        "Automobile": 10,
        "Medical": 10,
        "Subscription": 10,
        "Grocery": 10,
        "Dining": 10,
        "Shopping": 10,
        "Gas": 10,
        "Others": 20
    }'::jsonb,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE bills (
    id SERIAL PRIMARY KEY,
    account_id INT DEFAULT 1 REFERENCES accounts(id) ON DELETE SET NULL,
    merchant VARCHAR(50) NOT NULL,
    description VARCHAR(200) NOT NULL,
    category VARCHAR(30) NOT NULL CHECK (category IN (
        'Housing',
        'Automobile',
        'Medical',
        'Subscription',
        'Grocery',
        'Dining',
        'Shopping',
        'Gas',
        'Others'
    )),
    amount NUMERIC(10,2) NOT NULL DEFAULT 0 CHECK (amount >= 1.00),
    due_date DATE NOT NULL
);

-- +goose Down
SELECT 'down SQL query';
DROP TABLE IF EXISTS bills;
DROP TABLE IF EXISTS budget_plans;
DROP TABLE IF EXISTS transactions;
DROP TABLE IF EXISTS accounts;
DROP TABLE IF EXISTS users;