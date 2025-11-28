# Payment Service

## Proto API
- `CreateProduct` - Create product on Stripe & Razorpay (supports both one-time and recurring)
- `ListProducts` - Get all products with user access status
- `CreateStripeCheckoutSession` - Generate Stripe one-time payment URL
- `CreateRazorpayCheckoutSession` - Generate Razorpay one-time order details
- `CreateStripeSubscriptionCheckoutSession` - Generate Stripe subscription payment URL
- `CreateRazorpaySubscriptionCheckoutSession` - Generate Razorpay subscription details
- `CheckUserProductAccess` - Verify if user has active access to a product

## Request Workflow
1. **Create Product** → Registers product with both payment providers
2. **List Products** → Shows available products with provider IDs
3. **Create Checkout** → Generates payment session/order
4. **Handle Webhook** → Processes payment completion/failure

## Webhook Endpoints

### Stripe Webhooks (`/stripe-webhook`)
- **`charge.succeeded`** - One-time payment completed successfully
- **`charge.failed`** - One-time payment failed
- **`customer.subscription.created`** - New subscription created and activated
- **`customer.subscription.updated`** - Subscription details modified (periods, status, etc.)
- **`invoice.paid`** - Recurring subscription payment successful
- **`invoice.payment_failed`** - Recurring subscription payment failed

### Razorpay Webhooks (`/razorpay-webhook`)
- **`payment.captured`** - One-time payment completed successfully
- **`payment.failed`** - One-time payment failed
- **`subscription.authenticated`** - New subscription created and customer authorized
- **`subscription.charged`** - Recurring subscription payment successful
- **`subscription.pending`** - Recurring subscription payment failed/pending

## DAO & Tables

### Products Table (`paymentservice_products`)
- Stores product info with Stripe & Razorpay IDs
- Fields: id, user_id, name, description, price, currency, is_recurring, interval_count, interval_period
- Supports both one-time and recurring products
- Provider-specific IDs: razorpay_product_id, stripe_product_id

### Subscriptions Table (`paymentservice_subscriptions`)
- Tracks user subscriptions and access periods
- Fields: id, user_id, product_id, provider, provider_subscription_id, provider_customer_id
- Status management: provider_subscription_status, status, cancel_at_period_end
- Period tracking: current_period_start, current_period_end, canceled_at
- Deduplication: event_id (unique) prevents duplicate subscriptions
- Supports both recurring subscriptions and one-time access grants

### User Payments Table (`paymentservice_user_payments`)
- Tracks individual payment transactions
- Fields: id, user_id, product_id, subscription_id, payment_id, transaction_metadata, is_success
- Links payments to products and subscriptions
- Stores full webhook payload for audit trail

## Payment Types
- **One-time payments** - Single purchase with immediate access
- **Recurring subscriptions** - Periodic billing with ongoing access
- **Supported intervals**: Weekly, Monthly, Quarterly (3 months), Yearly
- **Providers**: Stripe & Razorpay for both payment types
- **Currency support**: USD, INR

## One-Time Payment Flow
1. **User selects one-time product** → Chooses payment provider (Stripe/Razorpay)
2. **Checkout session creation**:
   - Stripe → `CreateStripeCheckoutSession` returns session URL
   - Razorpay → `CreateRazorpayCheckoutSession` returns order details
3. **User redirected** → To provider's payment page
4. **Payment completion** → User completes payment on provider
5. **Webhook received** → Provider notifies backend of payment status
6. **Subscription created** → Entry in `paymentservice_subscriptions` with access period
7. **Payment recorded** → Transaction saved in `paymentservice_user_payments`
8. **Access granted** → User gains immediate access to product

## Recurring Subscription Flow
1. **User selects recurring product** → Chooses payment provider (Stripe/Razorpay)
2. **Subscription checkout creation**:
   - Stripe → `CreateStripeSubscriptionCheckoutSession` returns session URL
   - Razorpay → `CreateRazorpaySubscriptionCheckoutSession` returns subscription details
3. **User completes setup** → Authorizes recurring payments
4. **Subscription activated** → Provider creates recurring billing schedule
5. **Webhook received** → Provider notifies of subscription activation
6. **Subscription tracked** → Entry in `paymentservice_subscriptions` with billing cycle
7. **Initial payment recorded** → First payment saved in `paymentservice_user_payments`
8. **Access granted** → User gains access for current billing period

## Access Control Logic

### Unified Subscription Table
- Both one-time and recurring subscriptions stored in same `paymentservice_subscriptions` table
- `is_recurring` field differentiates between payment types

### One-Time Payment Access Check
1. **Entry exists** → Check if subscription record exists for user + product
2. **Payment type** → Verify `is_recurring = false`
3. **Payment success** → Confirm associated payment record has `is_success = true`
4. **Access granted** → If all conditions met, user has permanent access

### Recurring Subscription Access Check
1. **Entry exists** → Check if subscription record exists for user + product
2. **Payment type** → Verify `is_recurring = true`
3. **Status validation** → Both `provider_subscription_status` and `status` must be active
4. **Period validation** → Current time within `current_period_start` and `current_period_end`
5. **Access granted** → If all conditions met, user has access for current period

