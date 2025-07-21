-- CREATE TABLE IF NOT EXISTS users (id SERIAL PRIMARY KEY, username TEXT UNIQUE, password_hash TEXT);
-- CREATE TABLE IF NOT EXISTS accounts (id SERIAL PRIMARY KEY, user_id INT REFERENCES users(id), balance DECIMAL);

SELECT 'CREATE DATABASE transactions' WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'transactions')\gexec

CREATE TABLE transactions.transaction (
    id UUID PRIMARY KEY,
    kafka_id UUID NOT NULL,
    account_id UUID NOT NULL,
    amount INT NOT NULL,
    created_at TIMESTAMP DEFAULT now()
);
