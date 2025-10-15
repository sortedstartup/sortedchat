package dao

type Product struct {
	ID                string `db:"id"`
	StripeProductID   string `db:"stripe_product_id"`
	RazorpayProductID string `db:"razorpay_product_id"`
	UserID            string `db:"user_id"`
	Name              string `db:"name"`
	Description       string `db:"description"`
	Price             string `db:"price"`
	Currency          string `db:"currency"`
	CreatedAt         string `db:"created_at"`
	UpdatedAt         string `db:"updated_at"`
}
