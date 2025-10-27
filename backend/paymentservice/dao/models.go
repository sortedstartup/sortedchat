package dao

import (
	"database/sql"
	pb "sortedstartup/paymentservice/proto"
)

type Product struct {
	ID                string         `db:"id"`
	StripeProductID   string         `db:"stripe_product_id"`
	RazorpayProductID string         `db:"razorpay_product_id"`
	UserID            string         `db:"user_id"`
	Name              string         `db:"name"`
	Description       string         `db:"description"`
	Price             int64          `db:"price"`
	Currency          string         `db:"currency"`
	IsRecurring       bool           `db:"is_recurring"`
	IntervalCount     sql.NullInt64  `db:"interval_count"`
	IntervalPeriod    sql.NullString `db:"interval_period"`
	CreatedAt         string         `db:"created_at"`
	UpdatedAt         string         `db:"updated_at"`
}

// GetCurrencyEnum converts the string currency to protobuf Currency enum
func (p *Product) GetCurrencyEnum() pb.Currency {
	switch p.Currency {
	case "USD":
		return pb.Currency_USD
	case "INR":
		return pb.Currency_INR
	default:
		return pb.Currency_USD // default to USD
	}
}

type Subscription struct {
	ID                         string         `db:"id"`
	UserID                     string         `db:"user_id"`
	ProductID                  string         `db:"product_id"`
	Provider                   string         `db:"provider"`
	ProviderSubscriptionID     string         `db:"provider_subscription_id"`
	ProviderCustomerID         string         `db:"provider_customer_id"`
	ProviderSubscriptionStatus string         `db:"provider_subscription_status"`
	Status                     string         `db:"status"`
	CurrentPeriodStart         int64          `db:"current_period_start"`
	CurrentPeriodEnd           int64          `db:"current_period_end"`
	CancelAtPeriodEnd          bool           `db:"cancel_at_period_end"`
	CreatedAt                  string         `db:"created_at"`
	UpdatedAt                  string         `db:"updated_at"`
	CanceledAt                 sql.NullString `db:"canceled_at"`
}

type UserPayment struct {
	ID             string `db:"id"`
	UserID         string `db:"user_id"`
	ProductID      string `db:"product_id"`
	SubscriptionID string `db:"subscription_id"`
	PaymentID      string `db:"payment_id"`
	CreatedAt      string `db:"created_at"`
	UpdatedAt      string `db:"updated_at"`
}
