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

func (d *PostgresDAO) CreateProduct(stripeProductID string, razorpayProductID string, userID string, name string, description string, amountInCents int64, currency string) (string, error) {
	id := uuid.New().String()
	slog.Info("paymentservice:dao_postgres:CreateProduct", "userID", userID, "name", name, "description", description, "cost", amountInCents, "currency", currency)

	query := `INSERT INTO products (id,stripe_product_id, razorpay_product_id, user_id, name, description, price, currency, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`
	_, err := d.db.Exec(query, id, stripeProductID, razorpayProductID, userID, name, description, amountInCents, currency, time.Now().Format(time.RFC3339), time.Now().Format(time.RFC3339))
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
			  ON CONFLICT (session_id) 
			  DO UPDATE SET 
				  user_id = EXCLUDED.user_id,
				  product_id = EXCLUDED.product_id,
				  transaction_metadata = EXCLUDED.transaction_metadata,
				  is_success = EXCLUDED.is_success,
				  provider = EXCLUDED.provider,
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
