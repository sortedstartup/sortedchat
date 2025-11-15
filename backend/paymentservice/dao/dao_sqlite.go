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

func (d *SQLiteDAO) CreateProduct(stripeProductID string, razorpayProductID string, userID string, name string, description string, amountInSmallestUnit int64, currency string, isRecurring bool, intervalCount int64, intervalPeriod string) (string, error) {
	id := uuid.New().String()
	slog.Info("paymentservice:dao_sqlite:CreateProduct", "userID", userID, "name", name, "description", description, "cost", amountInSmallestUnit, "currency", currency, "isRecurring", isRecurring, "intervalCount", intervalCount, "intervalPeriod", intervalPeriod)

	query := `INSERT INTO paymentservice_products (id, stripe_product_id, razorpay_product_id, user_id, name, description, price, currency, is_recurring, interval_count, interval_period, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
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

func (d *SQLiteDAO) ListProducts(userID string) ([]*Product, error) {
	slog.Info("paymentservice:dao_sqlite:ListProducts", "userID", userID)

	// Optimized query that joins products with subscription access in a single query
	query := `
		SELECT 
			p.*,
			EXISTS(
				SELECT 1
				FROM paymentservice_subscriptions s
				WHERE s.product_id = p.id
				  AND s.user_id = ?
				  AND (
					(s.is_recurring = 1 AND s.status = 'active' AND s.current_period_end > strftime('%s', 'now')) OR
					s.is_recurring = 0
				  )
			) AS has_access
		FROM paymentservice_products p`

	rows, err := d.db.Queryx(query, userID)
	if err != nil {
		slog.Error("paymentservice:dao_sqlite:ListProducts", "error", err)
		return nil, err
	}
	defer rows.Close()

	products := []*Product{}
	for rows.Next() {
		var row struct {
			Product
			HasAccess int64 `db:"has_access"`
		}
		if err := rows.StructScan(&row); err != nil {
			slog.Error("paymentservice:dao_sqlite:ListProducts", "error scanning row", err)
			return nil, err
		}
		product := row.Product
		product.HasAccess = row.HasAccess == 1
		products = append(products, &product)
	}
	return products, nil
}

func (d *SQLiteDAO) GetProductById(productID string) (*Product, error) {
	slog.Info("paymentservice:dao_sqlite:GetProductById", "productID", productID)

	query := `SELECT * FROM paymentservice_products WHERE id = ?`
	product := &Product{}
	err := d.db.Get(product, query, productID)
	if err != nil {
		slog.Error("paymentservice:dao_sqlite:GetProductById", "error", err)
		return nil, err
	}

	return product, nil
}

// Subscription methods
func (d *SQLiteDAO) CreateSubscription(eventID, userID, productID, provider, providerSubscriptionID, providerCustomerID, providerSubscriptionStatus, status string, currentPeriodStart, currentPeriodEnd int64, cancelAtPeriodEnd bool, isRecurring bool) (string, error) {
	slog.Info("paymentservice:dao_sqlite:CreateSubscription", "userID", userID, "productID", productID, "provider", provider, "providerSubscriptionID", providerSubscriptionID, "providerCustomerID", providerCustomerID, "providerSubscriptionStatus", providerSubscriptionStatus, "status", status, "currentPeriodStart", currentPeriodStart, "currentPeriodEnd", currentPeriodEnd, "cancelAtPeriodEnd", cancelAtPeriodEnd)

	id := uuid.New().String()
	now := time.Now().Format(time.RFC3339)

	query := `INSERT INTO paymentservice_subscriptions (id, event_id, user_id, product_id, provider, provider_subscription_id, provider_customer_id, provider_subscription_status, status, current_period_start, current_period_end, cancel_at_period_end, created_at, updated_at, is_recurring) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := d.db.Exec(query, id, eventID, userID, productID, provider, providerSubscriptionID, providerCustomerID, providerSubscriptionStatus, status, currentPeriodStart, currentPeriodEnd, cancelAtPeriodEnd, now, now, isRecurring)
	if err != nil {
		slog.Error("paymentservice:dao_sqlite:CreateSubscription", "error", err)
		return "", err
	}

	slog.Info("paymentservice:dao_sqlite:CreateSubscription", "subscriptionID", id, "userID", userID, "productID", productID)
	return id, nil
}

func (d *SQLiteDAO) UpdateSubscription(subscriptionID, providerSubscriptionID, providerCustomerID, providerSubscriptionStatus, status string, currentPeriodStart, currentPeriodEnd int64, cancelAtPeriodEnd bool) error {
	now := time.Now().Format(time.RFC3339)

	query := `UPDATE paymentservice_subscriptions SET provider_subscription_id = ?, provider_customer_id = ?, provider_subscription_status = ?, status = ?, current_period_start = ?, current_period_end = ?, cancel_at_period_end = ?, updated_at = ? WHERE id = ?`
	_, err := d.db.Exec(query, providerSubscriptionID, providerCustomerID, providerSubscriptionStatus, status, currentPeriodStart, currentPeriodEnd, cancelAtPeriodEnd, now, subscriptionID)
	if err != nil {
		slog.Error("paymentservice:dao_sqlite:UpdateSubscription", "error", err)
		return err
	}

	slog.Info("paymentservice:dao_sqlite:UpdateSubscription", "subscriptionID", subscriptionID, "status", status)
	return nil
}

func (d *SQLiteDAO) GetSubscriptionByProviderCustomerID(providerCustomerID string) (*Subscription, error) {
	slog.Info("paymentservice:dao_sqlite:GetSubscriptionByProviderCustomerID", "providerCustomerID", providerCustomerID)

	query := `SELECT * FROM paymentservice_subscriptions WHERE provider_customer_id = ?`
	subscription := &Subscription{}
	err := d.db.Get(subscription, query, providerCustomerID)
	if err != nil {
		slog.Error("paymentservice:dao_sqlite:GetSubscriptionByProviderCustomerID", "error", err)
		return nil, err
	}

	return subscription, nil
}

func (d *SQLiteDAO) GetSubscriptionByUserIDAndProductID(userID, productID string) (*Subscription, error) {
	slog.Info("paymentservice:dao_sqlite:GetSubscriptionByUserIDAndProductID", "userID", userID, "productID", productID)

	query := `SELECT * FROM paymentservice_subscriptions WHERE user_id = ? AND product_id = ?`
	subscription := &Subscription{}
	err := d.db.Get(subscription, query, userID, productID)
	if err != nil {
		slog.Error("paymentservice:dao_sqlite:GetSubscriptionByUserIDAndProductID", "error", err)
		return nil, err
	}

	return subscription, nil
}

// User payment methods
func (d *SQLiteDAO) CreateUserPayment(userID, productID, subscriptionID, paymentID, transactionMetadata string, isSuccess bool) (string, error) {
	id := uuid.New().String()
	now := time.Now().Format(time.RFC3339)

	query := `INSERT INTO paymentservice_user_payments (id, user_id, product_id, subscription_id, transaction_metadata, payment_id, is_success, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := d.db.Exec(query, id, userID, productID, subscriptionID, transactionMetadata, paymentID, isSuccess, now, now)
	if err != nil {
		slog.Error("paymentservice:dao_sqlite:CreateUserPayment", "error", err)
		return "", err
	}

	slog.Info("paymentservice:dao_sqlite:CreateUserPayment", "paymentID", id, "userID", userID, "productID", productID, "isSuccess", isSuccess)
	return id, nil
}

func (d *SQLiteDAO) CheckUserProductAccess(userID, productID string) (bool, error) {
	slog.Info("paymentservice:dao_sqlite:CheckUserProductAccess", "userID", userID, "productID", productID)

	query := `
        SELECT EXISTS(
            SELECT 1 FROM paymentservice_subscriptions 
            WHERE user_id = ? AND product_id = ? AND (
                (is_recurring = 1 AND status = 'active' AND current_period_end > strftime('%s', 'now')) OR
                (is_recurring = 0)
            )
        )`

	var hasAccess bool
	err := d.db.Get(&hasAccess, query, userID, productID)
	if err != nil {
		slog.Error("paymentservice:dao_sqlite:CheckUserProductAccess", "error", err)
		return false, err
	}

	slog.Info("paymentservice:dao_sqlite:CheckUserProductAccess", "hasAccess", hasAccess)
	return hasAccess, nil
}
