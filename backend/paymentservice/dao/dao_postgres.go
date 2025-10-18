package dao

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// PostgresDAO implements the DAO interface using PostgreSQL and sqlx
type PostgresDAO struct {
	db *sqlx.DB
}

// NewPostgresDAO creates a new PostgreSQL DAO instance
func NewPostgresDAO(config *PostgresConfig) (*PostgresDAO, error) {
	dsn := config.GetPostgresDSN()

	db, err := sqlx.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open PostgreSQL connection: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(config.Pool.MaxOpenConnections)
	db.SetMaxIdleConns(config.Pool.MaxIdleConnections)
	db.SetConnMaxLifetime(config.Pool.ConnectionMaxLifetime)

	// Test the connection
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping PostgreSQL database: %w", err)
	}

	slog.Info("PostgreSQL DAO created successfully",
		"host", config.Host,
		"port", config.Port,
		"database", config.Database,
		"max_open_conns", config.Pool.MaxOpenConnections)

	return &PostgresDAO{db: db}, nil
}

// NewPostgresDAOWithDB creates a new PostgreSQL DAO instance using a shared database connection
func NewPostgresDAOWithDB(db *sqlx.DB) (*PostgresDAO, error) {
	return &PostgresDAO{db: db}, nil
}

func (d *PostgresDAO) Infer(dummy string) error {
	return nil
}

func (d *PostgresDAO) CreateProduct(stripeProductID string, razorpayProductID string, userID string, name string, description string, amountInSmallestUnit int64, currency string, isRecurring bool, intervalCount int64, intervalPeriod string) (string, error) {
	id := uuid.New().String()
	slog.Info("paymentservice:dao_postgres:CreateProduct", "userID", userID, "name", name, "description", description, "cost", amountInSmallestUnit, "currency", currency, "isRecurring", isRecurring, "intervalCount", intervalCount, "intervalPeriod", intervalPeriod)

	query := `INSERT INTO products (id, stripe_product_id, razorpay_product_id, user_id, name, description, price, currency, is_recurring, interval_count, interval_period, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`
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
		slog.Error("paymentservice:dao_postgres:CreateProduct", "error", err)
		return "", err
	}

	return id, nil
}

func (d *PostgresDAO) ListProducts() ([]*Product, error) {
	slog.Info("paymentservice:dao_postgres:ListProducts")

	query := `SELECT * FROM products`
	productList, err := d.db.Queryx(query)
	if err != nil {
		slog.Error("paymentservice:dao_postgres:ListProducts", "error", err)
		return nil, err
	}
	defer productList.Close()

	products := []*Product{}
	for productList.Next() {
		product := &Product{}
		err := productList.StructScan(product)
		if err != nil {
			slog.Error("paymentservice:dao_postgres:ListProducts", "error", err)
			return nil, err
		}
		products = append(products, product)
	}
	return products, nil
}

func (d *PostgresDAO) CreateUserPurchase(sessionID string, userID string, productID string, transaction_metadata string, is_success bool, provider string) (string, error) {
	id := uuid.New().String()
	now := time.Now().Format(time.RFC3339)
	slog.Info("paymentservice:dao_postgres:CreateUserPurchase", "sessionID", sessionID, "userID", userID, "productID", productID, "is_success", is_success, "provider", provider)

	// Use INSERT ... ON CONFLICT for upsert functionality in PostgreSQL
	query := `INSERT INTO user_purchases (id, session_id, user_id, product_id, transaction_metadata, is_success, provider, created_at, updated_at) 
			  VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
			  ON CONFLICT (provider, session_id) 
			  DO UPDATE SET 
				  user_id = EXCLUDED.user_id,
				  product_id = EXCLUDED.product_id,
				  transaction_metadata = EXCLUDED.transaction_metadata,
				  is_success = EXCLUDED.is_success,
				  updated_at = EXCLUDED.updated_at
			  RETURNING id`

	var actualID string
	err := d.db.Get(&actualID, query, id, sessionID, userID, productID, transaction_metadata, is_success, provider, now, now)
	if err != nil {
		slog.Error("paymentservice:dao_postgres:CreateUserPurchase", "error", err)
		return "", err
	}

	return actualID, nil
}

func (d *PostgresDAO) GetProductById(productID string) (*Product, error) {
	slog.Info("paymentservice:dao_postgres:GetProductById", "productID", productID)

	query := `SELECT * FROM products WHERE id = $1`
	product := &Product{}
	err := d.db.Get(product, query, productID)
	if err != nil {
		slog.Error("paymentservice:dao_postgres:GetProductById", "error", err)
		return nil, err
	}
	return product, nil
}

// Subscription methods
func (d *PostgresDAO) CreateSubscription(userID, productID, provider string) (string, error) {
	id := uuid.New().String()
	now := time.Now().Format(time.RFC3339)

	// For one-time payments, set far future end date to indicate lifetime access
	currentPeriodStart := now
	currentPeriodEnd := "9999-12-31T23:59:59Z" // Far future for one-time payments

	query := `INSERT INTO subscriptions (id, user_id, product_id, provider, status, current_period_start, current_period_end, cancel_at_period_end, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`
	_, err := d.db.Exec(query, id, userID, productID, provider, "pending", currentPeriodStart, currentPeriodEnd, false, now, now)
	if err != nil {
		slog.Error("paymentservice:dao_postgres:CreateSubscription", "error", err)
		return "", err
	}

	slog.Info("paymentservice:dao_postgres:CreateSubscription", "subscriptionID", id, "userID", userID, "productID", productID)
	return id, nil
}

func (d *PostgresDAO) UpdateSubscription(subscriptionID, providerSubscriptionID, providerSubscriptionStatus, status, currentPeriodStart, currentPeriodEnd string, cancelAtPeriodEnd bool) error {
	now := time.Now().Format(time.RFC3339)

	query := `UPDATE subscriptions SET provider_subscription_id = $1, provider_subscription_status = $2, status = $3, current_period_start = $4, current_period_end = $5, cancel_at_period_end = $6, updated_at = $7 WHERE id = $8`
	_, err := d.db.Exec(query, providerSubscriptionID, providerSubscriptionStatus, status, currentPeriodStart, currentPeriodEnd, cancelAtPeriodEnd, now, subscriptionID)
	if err != nil {
		slog.Error("paymentservice:dao_postgres:UpdateSubscription", "error", err)
		return err
	}

	slog.Info("paymentservice:dao_postgres:UpdateSubscription", "subscriptionID", subscriptionID, "status", status)
	return nil
}

func (d *PostgresDAO) GetSubscriptionByID(subscriptionID string) (*Subscription, error) {
	subscription := &Subscription{}
	query := `SELECT * FROM subscriptions WHERE id = $1`

	err := d.db.Get(subscription, query, subscriptionID)
	if err != nil {
		slog.Error("paymentservice:dao_postgres:GetSubscriptionByID", "error", err)
		return nil, err
	}

	return subscription, nil
}

func (d *PostgresDAO) CheckUserProductAccess(userID, productID string) (*Subscription, error) {
	subscription := &Subscription{}
	query := `SELECT * FROM subscriptions WHERE user_id = $1 AND product_id = $2 AND status = 'active' AND current_period_end > NOW() ORDER BY created_at DESC LIMIT 1`

	err := d.db.Get(subscription, query, userID, productID)
	if err != nil {
		slog.Error("paymentservice:dao_postgres:CheckUserProductAccess", "error", err)
		return nil, err
	}

	return subscription, nil
}

// User payment methods
func (d *PostgresDAO) CreateUserPayment(userID, productID, subscriptionID, paymentID string) (string, error) {
	id := uuid.New().String()
	now := time.Now().Format(time.RFC3339)

	query := `INSERT INTO user_payments (id, user_id, product_id, subscription_id, payment_id, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err := d.db.Exec(query, id, userID, productID, subscriptionID, paymentID, now, now)
	if err != nil {
		slog.Error("paymentservice:dao_postgres:CreateUserPayment", "error", err)
		return "", err
	}

	slog.Info("paymentservice:dao_postgres:CreateUserPayment", "paymentID", id, "userID", userID, "productID", productID)
	return id, nil
}
