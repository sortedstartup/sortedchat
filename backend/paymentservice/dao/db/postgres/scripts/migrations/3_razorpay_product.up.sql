ALTER TABLE products ADD COLUMN razorpay_product_id TEXT;
ALTER TABLE products ADD COLUMN stripe_product_id TEXT;
ALTER TABLE user_purchases ADD COLUMN provider TEXT;