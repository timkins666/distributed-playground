\connect banking postgres
BEGIN;

-- Transactions schema
CREATE SCHEMA IF NOT EXISTS transactions;

CREATE TABLE IF NOT EXISTS transactions.transaction (
    id UUID PRIMARY KEY,
    kafka_id UUID NOT NULL,
    account_id UUID NOT NULL,
    amount INT NOT NULL,
    created_at TIMESTAMP DEFAULT now()
);

-- Accounts schema
CREATE SCHEMA IF NOT EXISTS accounts;

CREATE TABLE IF NOT EXISTS accounts."user" (
    id SERIAL PRIMARY KEY, 
    username TEXT UNIQUE NOT NULL,
    roles TEXT[] NOT NULL DEFAULT '{}'
);

CREATE TABLE IF NOT EXISTS accounts.account (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    user_id INT NOT NULL REFERENCES accounts."user"(id),
    balance BIGINT NOT NULL DEFAULT 0
);

COMMIT;
