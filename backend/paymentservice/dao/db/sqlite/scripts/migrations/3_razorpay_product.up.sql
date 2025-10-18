ALTER TABLE products ADD COLUMN razorpay_product_id TEXT;
ALTER TABLE products ADD COLUMN stripe_product_id TEXT;
ALTER TABLE user_purchases ADD COLUMN provider TEXT;
ALTER TABLE user_purchases ADD COLUMN session_id TEXT;
CREATE UNIQUE INDEX idx_user_purchases_provider_session ON user_purchases(provider, session_id);