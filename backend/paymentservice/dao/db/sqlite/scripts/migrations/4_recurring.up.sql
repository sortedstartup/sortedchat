ALTER TABLE products ADD COLUMN is_recurring BOOLEAN DEFAULT FALSE;
ALTER TABLE products ADD COLUMN interval_count INTEGER DEFAULT 1;
ALTER TABLE products ADD COLUMN interval_period TEXT DEFAULT 'month';
ALTER TABLE products ADD COLUMN plan_id TEXT; -- Razorpay plan ID or Stripe PRICE ID


CREATE TABLE IF NOT EXISTS subscriptions (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    plan_id TEXT NOT NULL,
    transaction_metadata TEXT NOT NULL,
    status TEXT NOT NULL,
    current_period_start TEXT NOT NULL,
    current_period_end TEXT NOT NULL,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    canceled_at TEXT,
    ended_at TEXT,

    FOREIGN KEY (plan_id) REFERENCES products(plan_id),
);
