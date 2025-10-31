# Payment Service

## Proto API
- `CreateProduct` - Create product on Stripe & Razorpay
- `ListProducts` - Get all products  
- `CreateStripeCheckoutSession` - Generate Stripe payment URL
- `CreateRazorpayCheckoutSession` - Generate Razorpay order details

## Request Workflow
1. **Create Product** → Registers product with both payment providers
2. **List Products** → Shows available products with provider IDs
3. **Create Checkout** → Generates payment session/order
4. **Handle Webhook** → Processes payment completion/failure

## Webhook Endpoints
- `/stripe-webhook` - Handles Stripe events (checkout.session.completed/expired)
- `/razorpay-webhook` - Handles Razorpay events (payment.captured/failed)

## DAO & Tables

### Products Table
- Stores product info with Stripe & Razorpay IDs
- Fields: id, user_id, name, description, price, currency

### User Purchases Table  
- Tracks payment transactions
- Fields: id, user_id, product_id, transaction_metadata, is_success
- Stores full webhook payload for audit trail

## Payment Type
- **One-time payments only** (no subscriptions)
- Supports Stripe & Razorpay providers

## User Flow
1. **Admin creates product** → `CreateProduct` registers on both Stripe & Razorpay
2. **User browses products** → `ListProducts` shows all available products (no auth required)
3. **User selects payment method** → Two buttons: "Pay with Stripe" or "Pay with Razorpay"
4. **Checkout initiation** → Separate RPC calls:
   - Stripe → Returns session URL for redirect
   - Razorpay → Returns order ID for frontend integration
5. **User completes payment** → On respective payment provider
6. **Webhook notification** → Provider notifies backend of payment status
7. **Transaction saved** → Success/failure recorded in `user_purchases` table
