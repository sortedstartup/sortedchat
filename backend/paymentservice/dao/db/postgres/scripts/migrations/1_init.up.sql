CREATE TABLE IF NOT EXISTS products (
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

CREATE TABLE IF NOT EXISTS user_payments (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    product_id TEXT NOT NULL,
    subscription_id TEXT,
    payment_id TEXT NOT NULL,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    FOREIGN KEY (product_id) REFERENCES products(id),
    FOREIGN KEY (subscription_id) REFERENCES subscriptions(id)
);

CREATE TABLE IF NOT EXISTS subscriptions (
    id TEXT PRIMARY KEY, --uuid 
    user_id TEXT NOT NULL, 
    product_id TEXT NOT NULL, -- product ID
    provider TEXT NOT NULL, -- Stripe or Razorpay
    provider_subscription_id TEXT, -- Razorpay subscription ID or Stripe subscription ID or null if one-time payment
    provider_subscription_status TEXT, -- Provider's subscription status (active, canceled, past_due, etc.)
    status TEXT NOT NULL, -- User access status (active, inactive, expired)
    current_period_start TIMESTAMP NOT NULL, --period cycle start date
    current_period_end TIMESTAMP NOT NULL, --period cycle end date
    cancel_at_period_end BOOLEAN DEFAULT FALSE, -- whether the subscription will be canceled at the end of the current period
    created_at TIMESTAMP NOT NULL, -- timestamp of when the subscription was created
    updated_at TIMESTAMP NOT NULL, -- timestamp of when the subscription was last updated
    canceled_at TIMESTAMP, -- timestamp of when the subscription was canceled
    FOREIGN KEY (product_id) REFERENCES products(id)
);


--stripe webhooks

-- Webhook Event                                               
-- -------------------------------
-- customer.subscription.created                       
-- customer.subscription.updated                       
-- customer.subscription.deleted                       
-- invoice.paid                                          
-- invoice.payment_failed         
-- payment_intent.succeeded                              
-- charge.refunded                                       


--razorpay webhooks

-- Webhook Event                       
-- ----------------------------
-- subscription.authenticated  
-- subscription.activated      
-- subscription.charged          
-- subscription.completed      
-- subscription.cancelled      
-- subscription.paused         
-- subscription.halted         
-- payment.captured              
-- payment.failed                