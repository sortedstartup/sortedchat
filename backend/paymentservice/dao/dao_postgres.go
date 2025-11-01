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

func (d *PostgresDAO) ListProducts(userID string) ([]*Product, error) {
	slog.Info("paymentservice:dao_postgres:ListProducts", "userID", userID)

	// Optimized query that joins products with subscription access in a single query
	query := `
		SELECT 
			p.*,
			CASE 
				WHEN s.id IS NOT NULL THEN true 
				ELSE false 
			END as has_access
		FROM products p
		LEFT JOIN subscriptions s ON p.id = s.product_id 
			AND s.user_id = $1 
			AND (
				(s.is_recurring = true AND s.status = 'active' AND s.current_period_end > EXTRACT(epoch FROM NOW())) OR
				(s.is_recurring = false)
			)`

	rows, err := d.db.Queryx(query, userID)
	if err != nil {
		slog.Error("paymentservice:dao_postgres:ListProducts", "error", err)
		return nil, err
	}
	defer rows.Close()

	products := []*Product{}
	for rows.Next() {
		product := &Product{}
		var hasAccess bool

		// Create a map to scan both product fields and has_access
		rowMap := make(map[string]interface{})
		err := rows.MapScan(rowMap)
		if err != nil {
			slog.Error("paymentservice:dao_postgres:ListProducts", "error scanning row", err)
			return nil, err
		}

		// Manually map the fields to the product struct
		product.ID = rowMap["id"].(string)
		product.StripeProductID = rowMap["stripe_product_id"].(string)
		product.RazorpayProductID = rowMap["razorpay_product_id"].(string)
		product.UserID = rowMap["user_id"].(string)
		product.Name = rowMap["name"].(string)
		product.Description = rowMap["description"].(string)
		product.Price = rowMap["price"].(int64)
		product.Currency = rowMap["currency"].(string)
		product.IsRecurring = rowMap["is_recurring"].(bool)
		product.CreatedAt = rowMap["created_at"].(string)
		product.UpdatedAt = rowMap["updated_at"].(string)

		// Handle nullable fields
		if intervalCount := rowMap["interval_count"]; intervalCount != nil {
			product.IntervalCount.Valid = true
			product.IntervalCount.Int64 = intervalCount.(int64)
		}
		if intervalPeriod := rowMap["interval_period"]; intervalPeriod != nil {
			product.IntervalPeriod.Valid = true
			product.IntervalPeriod.String = intervalPeriod.(string)
		}

		// Set access status
		hasAccess = rowMap["has_access"].(bool)
		product.HasAccess = hasAccess

		products = append(products, product)
	}
	return products, nil
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
func (d *PostgresDAO) CreateSubscription(userID, productID, provider, providerSubscriptionID, providerCustomerID, providerSubscriptionStatus, status string, currentPeriodStart, currentPeriodEnd int64, cancelAtPeriodEnd bool, isRecurring bool) (string, error) {
	id := uuid.New().String()
	now := time.Now().Format(time.RFC3339)

	query := `INSERT INTO subscriptions (id, user_id, product_id, provider, provider_subscription_id, provider_customer_id, provider_subscription_status, status, current_period_start, current_period_end, cancel_at_period_end, created_at, updated_at, is_recurring) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)`
	_, err := d.db.Exec(query, id, userID, productID, provider, providerSubscriptionID, providerCustomerID, providerSubscriptionStatus, status, currentPeriodStart, currentPeriodEnd, cancelAtPeriodEnd, now, now, isRecurring)
	if err != nil {
		slog.Error("paymentservice:dao_postgres:CreateSubscription", "error", err)
		return "", err
	}

	slog.Info("paymentservice:dao_postgres:CreateSubscription", "subscriptionID", id, "userID", userID, "productID", productID)
	return id, nil
}

func (d *PostgresDAO) UpdateSubscription(subscriptionID, providerSubscriptionID, providerCustomerID, providerSubscriptionStatus, status string, currentPeriodStart, currentPeriodEnd int64, cancelAtPeriodEnd bool) error {
	now := time.Now().Format(time.RFC3339)

	query := `UPDATE subscriptions SET provider_subscription_id = $1, provider_customer_id = $2, provider_subscription_status = $3, status = $4, current_period_start = $5, current_period_end = $6, cancel_at_period_end = $7, updated_at = $8 WHERE id = $9`
	_, err := d.db.Exec(query, providerSubscriptionID, providerCustomerID, providerSubscriptionStatus, status, currentPeriodStart, currentPeriodEnd, cancelAtPeriodEnd, now, subscriptionID)
	if err != nil {
		slog.Error("paymentservice:dao_postgres:UpdateSubscription", "error", err)
		return err
	}

	slog.Info("paymentservice:dao_postgres:UpdateSubscription", "subscriptionID", subscriptionID, "status", status)
	return nil
}

func (d *PostgresDAO) GetSubscriptionByProviderCustomerID(providerCustomerID string) (*Subscription, error) {
	slog.Info("paymentservice:dao_postgres:GetSubscriptionByProviderCustomerID", "providerCustomerID", providerCustomerID)

	query := `SELECT * FROM subscriptions WHERE provider_customer_id = $1`
	subscription := &Subscription{}
	err := d.db.Get(subscription, query, providerCustomerID)
	if err != nil {
		slog.Error("paymentservice:dao_postgres:GetSubscriptionByProviderCustomerID", "error", err)
		return nil, err
	}

	return subscription, nil
}

func (d *PostgresDAO) GetSubscriptionByUserIDAndProductID(userID, productID string) (*Subscription, error) {
	slog.Info("paymentservice:dao_postgres:GetSubscriptionByUserIDAndProductID", "userID", userID, "productID", productID)

	query := `SELECT * FROM subscriptions WHERE user_id = $1 AND product_id = $2`
	subscription := &Subscription{}
	err := d.db.Get(subscription, query, userID, productID)
	if err != nil {
		slog.Error("paymentservice:dao_postgres:GetSubscriptionByUserIDAndProductID", "error", err)
		return nil, err
	}
	return subscription, nil
}

// User payment methods
func (d *PostgresDAO) CreateUserPayment(userID, productID, subscriptionID, paymentID, transactionMetadata string, isSuccess bool) (string, error) {
	id := uuid.New().String()
	now := time.Now().Format(time.RFC3339)

	query := `INSERT INTO user_payments (id, user_id, product_id, subscription_id, transaction_metadata, payment_id, is_success, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`
	_, err := d.db.Exec(query, id, userID, productID, subscriptionID, transactionMetadata, paymentID, isSuccess, now, now)
	if err != nil {
		slog.Error("paymentservice:dao_postgres:CreateUserPayment", "error", err)
		return "", err
	}

	slog.Info("paymentservice:dao_postgres:CreateUserPayment", "paymentID", id, "userID", userID, "productID", productID, "isSuccess", isSuccess)
	return id, nil
}

func (d *PostgresDAO) CheckUserProductAccess(userID, productID string) (bool, error) {
	slog.Info("paymentservice:dao_postgres:CheckUserProductAccess", "userID", userID, "productID", productID)

	query := `
        SELECT EXISTS(
            SELECT 1 FROM subscriptions 
            WHERE user_id = $1 AND product_id = $2 AND (
                (is_recurring = true AND status = 'active' AND current_period_end > EXTRACT(epoch FROM NOW())) OR
                (is_recurring = false)
            )
        )`

	var hasAccess bool
	err := d.db.Get(&hasAccess, query, userID, productID)
	if err != nil {
		slog.Error("paymentservice:dao_postgres:CheckUserProductAccess", "error", err)
		return false, err
	}

	slog.Info("paymentservice:dao_postgres:CheckUserProductAccess", "hasAccess", hasAccess)
	return hasAccess, nil
}
