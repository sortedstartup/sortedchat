## Coupons/Discount Integration

### Implementation Approach

**Challenge**: Managing coupons across both Stripe and Razorpay through a unified API is difficult due to:
- Fundamental differences between both providers
- Razorpay does not provide API to create coupons programmatically

**Solution**: Create coupons at application level and generate checkout sessions with discounted prices based on user's coupon.

### Payment Provider Differences

**Razorpay**:
- Checkout session requires only `amount` and `currency`
- No product ID required
- Purely payment-focused, not product-focused
- Supports dynamic pricing by default

**Stripe**:
- Previously required passing `price_id` associated with product
- Now supports dynamic pricing using `price_data`:

```
LineItems: []*stripe.CheckoutSessionLineItemParams{
{
PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
Currency: stripe.String("inr"),
ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
Name: stripe.String(product.Name),
Description: stripe.String(product.Description),
},
UnitAmount: stripe.Int64(int64(finalAmount)),
},
Quantity: stripe.Int64(1),
},
}
```

- Stripe handles price-to-product mapping internally

### Database Implications

**Subscriptions Table**:
- Must maintain a `price` column for each user
- Required because single product can be sold at different prices based on coupons

**Coupons Table** (for advanced functionality):
- Required for features like:
- Coupon expiration dates
- User-specific coupons
- Maximum usage limits per coupon
- Discount percentage/amount
