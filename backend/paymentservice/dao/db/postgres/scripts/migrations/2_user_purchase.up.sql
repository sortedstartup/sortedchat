CREATE TABLE IF NOT EXISTS user_purchases (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    product_id TEXT NOT NULL,
    transaction_metadata TEXT NOT NULL,
    is_success BOOLEAN NOT NULL,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);