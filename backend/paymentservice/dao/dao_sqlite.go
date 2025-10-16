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

func (d *SQLiteDAO) CreateProduct(stripeProductID string, razorpayProductID string, userID string, name string, description string, amountInSmallestUnit int64, currency string) (string, error) {
	id := uuid.New().String()
	slog.Info("paymentservice:dao_sqlite:CreateProduct", "userID", userID, "name", name, "description", description, "cost", amountInSmallestUnit, "currency", currency)

	query := `INSERT INTO products (id, stripe_product_id, razorpay_product_id, user_id, name, description, price, currency, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	now := time.Now().Format(time.RFC3339)
	_, err := d.db.Exec(query, id, stripeProductID, razorpayProductID, userID, name, description, amountInSmallestUnit, currency, now, now)
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

	// Simple INSERT OR IGNORE followed by UPDATE - much cleaner with unique index
	// First, try to insert a new record
	insertQuery := `INSERT OR IGNORE INTO user_purchases (id, session_id, user_id, product_id, transaction_metadata, is_success, provider, created_at, updated_at) 
					VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := d.db.Exec(insertQuery, id, sessionID, userID, productID, transaction_metadata, is_success, provider, now, now)
	if err != nil {
		slog.Error("paymentservice:dao_sqlite:CreateUserPurchase", "error", "failed to insert", "details", err)
		return "", err
	}

	// Then update the record (will update existing or the just-inserted record)
	updateQuery := `UPDATE user_purchases 
					SET user_id = ?, product_id = ?, transaction_metadata = ?, is_success = ?, provider = ?, updated_at = ?
					WHERE session_id = ?`
	_, err = d.db.Exec(updateQuery, userID, productID, transaction_metadata, is_success, provider, now, sessionID)
	if err != nil {
		slog.Error("paymentservice:dao_sqlite:CreateUserPurchase", "error", "failed to update", "details", err)
		return "", err
	}

	// Get the actual ID
	var actualID string
	err = d.db.Get(&actualID, "SELECT id FROM user_purchases WHERE session_id = ?", sessionID)
	if err != nil {
		slog.Error("paymentservice:dao_sqlite:CreateUserPurchase", "error", "failed to get actual ID", "details", err)
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
