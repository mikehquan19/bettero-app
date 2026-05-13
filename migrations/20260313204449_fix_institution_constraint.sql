-- +goose Up
SELECT 'up SQL query';
ALTER TABLE accounts
DROP CONSTRAINT IF EXISTS accounts_institution_check;

ALTER TABLE accounts
ADD CONSTRAINT accounts_institution_check
CHECK (institution IN (
    'JP Morgan Chase',
    'Bank of America',
    'Wells Fargo',
    'Citi Bank',
    'Capital One',
    'Discover',
    'Sofi Bank',
    'Ally Bank'
));

-- +goose Down
SELECT 'down SQL query';
ALTER TABLE accounts
DROP CONSTRAINT IF EXISTS accounts_institution_check;