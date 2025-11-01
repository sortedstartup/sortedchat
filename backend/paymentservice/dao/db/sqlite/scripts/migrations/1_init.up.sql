CREATE TABLE IF NOT EXISTS paymentservice_products (
    id TEXT PRIMARY KEY,
    razorpay_product_id TEXT, -- Razorpay product ID or plan ID
    stripe_product_id TEXT, -- Stripe product ID
    user_id TEXT NOT NULL,
    name TEXT NOT NULL,
    description TEXT NOT NULL,
    price INTEGER NOT NULL,
    currency TEXT NOT NULL,
    is_recurring BOOLEAN DEFAULT FALSE,
    interval_count INTEGER, -- Only for recurring (NULL for one-time)
    interval_period TEXT, -- Only for recurring (NULL for one-time)
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS paymentservice_subscriptions (
    id TEXT PRIMARY KEY, --uuid 
    user_id TEXT NOT NULL, 
    product_id TEXT NOT NULL, -- product ID
    provider TEXT NOT NULL, -- Stripe or Razorpay
    provider_subscription_id TEXT, -- Razorpay subscription ID or Stripe subscription ID or null if one-time payment
    provider_subscription_status TEXT, -- Provider's subscription lifecycle  status (active, canceled, past_due, etc.)
    provider_customer_id TEXT, -- Provider's customer ID
    status TEXT NOT NULL, -- User access status (active, inactive, expired)
    current_period_start INTEGER NOT NULL, --period cycle start date (Unix timestamp)
    current_period_end INTEGER NOT NULL, --period cycle end date (Unix timestamp)
    cancel_at_period_end BOOLEAN DEFAULT FALSE, -- whether the subscription will be canceled at the end of the current period
    created_at TEXT NOT NULL, -- timestamp of when the subscription was created
    updated_at TEXT NOT NULL, -- timestamp of when the subscription was last updated
    canceled_at INTEGER, -- timestamp of when the subscription was canceled
    is_recurring BOOLEAN DEFAULT FALSE, -- whether the subscription is a one-time payment
    event_id TEXT UNIQUE NOT NULL, -- this is to avoid duplicate subscriptions
    FOREIGN KEY (product_id) REFERENCES paymentservice_products(id)
);


-- Fixed: Added product_id column
CREATE TABLE IF NOT EXISTS paymentservice_user_payments (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    product_id TEXT NOT NULL,
    subscription_id TEXT,
    transaction_metadata TEXT NOT NULL,
    payment_id TEXT UNIQUE NOT NULL,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    is_success BOOLEAN DEFAULT FALSE,
    FOREIGN KEY (product_id) REFERENCES paymentservice_products(id),
    FOREIGN KEY (subscription_id) REFERENCES paymentservice_subscriptions(id)
);