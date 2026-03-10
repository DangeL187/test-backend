CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username TEXT NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS balances (
    user_id INT NOT NULL REFERENCES users(id),
    balance NUMERIC(18, 6) NOT NULL DEFAULT 0 CHECK (balance >= 0),
    currency TEXT NOT NULL DEFAULT 'USDT',
    PRIMARY KEY(user_id, currency)
);

CREATE TABLE IF NOT EXISTS withdrawals (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL REFERENCES users(id),
    amount NUMERIC(18, 6) NOT NULL CHECK (amount > 0),
    currency TEXT NOT NULL DEFAULT 'USDT',
    destination TEXT NOT NULL CHECK (destination <> ''),
    status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'completed', 'failed', 'confirmed')),
    idempotency_key TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
    UNIQUE(user_id, idempotency_key)
);
CREATE INDEX IF NOT EXISTS idx_withdrawals_user_status ON withdrawals(user_id, idempotency_key);

CREATE TABLE IF NOT EXISTS ledger_entries (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL REFERENCES users(id),
    reference_type TEXT NOT NULL CHECK(reference_type IN ('withdrawal', 'deposit', 'trade')),
    reference_id INT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
    UNIQUE(reference_type, reference_id)
);
