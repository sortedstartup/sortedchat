package dao

import pb "sortedstartup/paymentservice/proto"

type Product struct {
	ID                string `db:"id"`
	StripeProductID   string `db:"stripe_product_id"`
	RazorpayProductID string `db:"razorpay_product_id"`
	UserID            string `db:"user_id"`
	Name              string `db:"name"`
	Description       string `db:"description"`
	Price             int64  `db:"price"`
	Currency          string `db:"currency"`
	IsRecurring       bool   `db:"is_recurring"`
	IntervalCount     int64  `db:"interval_count"`
	IntervalPeriod    string `db:"interval_period"`
	CreatedAt         string `db:"created_at"`
	UpdatedAt         string `db:"updated_at"`
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
