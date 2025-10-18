package dao

import (

	// sqlite_vec "github.com/asg017/sqlite-vec-go-bindings/cgo"

	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

// SQLiteDAO implements the DAO interface using SQLite and sqlx
type SQLiteDAO struct {
	db *sqlx.DB
}

// NewSQLiteDAO creates a new SQLite DAO instance
func NewSQLiteDAO(db *sqlx.DB) (*SQLiteDAO, error) {
	return &SQLiteDAO{db: db}, nil
}

func (d *SQLiteDAO) Infer(dummy string) error {
	return nil
}

func (d *SQLiteDAO) CreateProduct(stripeProductID string, razorpayProductID string, userID string, name string, description string, amountInSmallestUnit int64, currency string, isRecurring bool, intervalCount int64, intervalPeriod string) (string, error) {
	id := uuid.New().String()
	slog.Info("paymentservice:dao_sqlite:CreateProduct", "userID", userID, "name", name, "description", description, "cost", amountInSmallestUnit, "currency", currency, "isRecurring", isRecurring, "intervalCount", intervalCount, "intervalPeriod", intervalPeriod)

	query := `INSERT INTO products (id, stripe_product_id, razorpay_product_id, user_id, name, description, price, currency, is_recurring, interval_count, interval_period, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	now := time.Now().Format(time.RFC3339)

	// Handle NULL values for one-time payments
	var intervalCountValue interface{} = intervalCount
	var intervalPeriodValue interface{} = intervalPeriod

	if !isRecurring {
		intervalCountValue = nil
		intervalPeriodValue = nil
	}

	_, err := d.db.Exec(query, id, stripeProductID, razorpayProductID, userID, name, description, amountInSmallestUnit, currency, isRecurring, intervalCountValue, intervalPeriodValue, now, now)
	if err != nil {
		slog.Error("paymentservice:dao_sqlite:CreateProduct", "error", err)
		return "", err
	}

	return id, nil
}

func (d *SQLiteDAO) ListProducts() ([]*Product, error) {
	slog.Info("paymentservice:dao_sqlite:ListProducts")

	query := `SELECT * FROM products`
	productList, err := d.db.Queryx(query)
	if err != nil {
		slog.Error("paymentservice:dao_sqlite:ListProducts", "error", err)
		return nil, err
	}
	defer productList.Close()

	products := []*Product{}
	for productList.Next() {
		product := &Product{}
		err := productList.StructScan(product)
		if err != nil {
			slog.Error("paymentservice:dao_sqlite:ListProducts", "error", err)
			return nil, err
		}
		products = append(products, product)
	}
	return products, nil
}

func (d *SQLiteDAO) CreateUserPurchase(sessionID string, userID string, productID string, transaction_metadata string, is_success bool, provider string) (string, error) {
	id := uuid.New().String()
	now := time.Now().Format(time.RFC3339)
	slog.Info("paymentservice:dao_sqlite:CreateUserPurchase", "sessionID", sessionID, "userID", userID, "productID", productID, "is_success", is_success, "provider", provider)

	// Use proper SQLite UPSERT syntax with ON CONFLICT - atomic operation
	query := `INSERT INTO user_purchases (id, session_id, user_id, product_id, transaction_metadata, is_success, provider, created_at, updated_at) 
			  VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
			  ON CONFLICT(provider, session_id) 
			  DO UPDATE SET 
				  user_id = excluded.user_id,
				  product_id = excluded.product_id,
				  transaction_metadata = excluded.transaction_metadata,
				  is_success = excluded.is_success,
				  updated_at = excluded.updated_at
			  RETURNING id`

	var actualID string
	err := d.db.Get(&actualID, query, id, sessionID, userID, productID, transaction_metadata, is_success, provider, now, now)
	if err != nil {
		slog.Error("paymentservice:dao_sqlite:CreateUserPurchase", "error", err)
		return "", err
	}

	return actualID, nil
}

func (d *SQLiteDAO) GetProductById(productID string) (*Product, error) {
	slog.Info("paymentservice:dao_sqlite:GetProductById", "productID", productID)

	query := `SELECT * FROM products WHERE id = ?`
	product := &Product{}
	err := d.db.Get(product, query, productID)
	if err != nil {
		slog.Error("paymentservice:dao_sqlite:GetProductById", "error", err)
		return nil, err
	}

	return product, nil
}

// Subscription methods
func (d *SQLiteDAO) CreateSubscription(userID, productID, provider string) (string, error) {
	id := uuid.New().String()
	now := time.Now().Format(time.RFC3339)

	// For one-time payments, set far future end date to indicate lifetime access
	currentPeriodStart := now
	currentPeriodEnd := "9999-12-31T23:59:59Z" // Far future for one-time payments

	query := `INSERT INTO subscriptions (id, user_id, product_id, provider, status, current_period_start, current_period_end, cancel_at_period_end, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := d.db.Exec(query, id, userID, productID, provider, "pending", currentPeriodStart, currentPeriodEnd, false, now, now)
	if err != nil {
		slog.Error("paymentservice:dao_sqlite:CreateSubscription", "error", err)
		return "", err
	}

	slog.Info("paymentservice:dao_sqlite:CreateSubscription", "subscriptionID", id, "userID", userID, "productID", productID)
	return id, nil
}

func (d *SQLiteDAO) UpdateSubscription(subscriptionID, providerSubscriptionID, providerSubscriptionStatus, status, currentPeriodStart, currentPeriodEnd string, cancelAtPeriodEnd bool) error {
	now := time.Now().Format(time.RFC3339)

	query := `UPDATE subscriptions SET provider_subscription_id = ?, provider_subscription_status = ?, status = ?, current_period_start = ?, current_period_end = ?, cancel_at_period_end = ?, updated_at = ? WHERE id = ?`
	_, err := d.db.Exec(query, providerSubscriptionID, providerSubscriptionStatus, status, currentPeriodStart, currentPeriodEnd, cancelAtPeriodEnd, now, subscriptionID)
	if err != nil {
		slog.Error("paymentservice:dao_sqlite:UpdateSubscription", "error", err)
		return err
	}

	slog.Info("paymentservice:dao_sqlite:UpdateSubscription", "subscriptionID", subscriptionID, "status", status)
	return nil
}

func (d *SQLiteDAO) GetSubscriptionByID(subscriptionID string) (*Subscription, error) {
	subscription := &Subscription{}
	query := `SELECT * FROM subscriptions WHERE id = ?`

	err := d.db.Get(subscription, query, subscriptionID)
	if err != nil {
		slog.Error("paymentservice:dao_sqlite:GetSubscriptionByID", "error", err)
		return nil, err
	}

	return subscription, nil
}

func (d *SQLiteDAO) CheckUserProductAccess(userID, productID string) (*Subscription, error) {
	subscription := &Subscription{}
	query := `SELECT * FROM subscriptions WHERE user_id = ? AND product_id = ? AND status = 'active' AND current_period_end > datetime('now') ORDER BY created_at DESC LIMIT 1`

	err := d.db.Get(subscription, query, userID, productID)
	if err != nil {
		slog.Error("paymentservice:dao_sqlite:CheckUserProductAccess", "error", err)
		return nil, err
	}

	return subscription, nil
}

// User payment methods
func (d *SQLiteDAO) CreateUserPayment(userID, productID, subscriptionID, paymentID string) (string, error) {
	id := uuid.New().String()
	now := time.Now().Format(time.RFC3339)

	query := `INSERT INTO user_payments (id, user_id, product_id, subscription_id, payment_id, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)`
	_, err := d.db.Exec(query, id, userID, productID, subscriptionID, paymentID, now, now)
	if err != nil {
		slog.Error("paymentservice:dao_sqlite:CreateUserPayment", "error", err)
		return "", err
	}

	slog.Info("paymentservice:dao_sqlite:CreateUserPayment", "paymentID", id, "userID", userID, "productID", productID)
	return id, nil
}
