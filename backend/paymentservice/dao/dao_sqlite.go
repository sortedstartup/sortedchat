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

func (d *SQLiteDAO) CreateProduct(id string, userID string, name string, description string, cost string, currency string) (string, error) {
	slog.Info("paymentservice:dao_sqlite:CreateProduct", "userID", userID, "name", name, "description", description, "cost", cost, "currency", currency)

	query := `INSERT INTO products (id, user_id, name, description, price, currency, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := d.db.Exec(query, id, userID, name, description, cost, currency, time.Now().Format(time.RFC3339), time.Now().Format(time.RFC3339))
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

func (d *SQLiteDAO) CreateUserPurchase(userID string, productID string, transaction_metadata string, is_success bool) (string, error) {
	id := uuid.New().String()
	slog.Info("paymentservice:dao_sqlite:CreateUserPurchase", "userID", userID, "productID", productID, "is_success", is_success)

	query := `INSERT INTO user_purchases (id, user_id, product_id, transaction_metadata, is_success, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)`
	_, err := d.db.Exec(query, id, userID, productID, transaction_metadata, is_success, time.Now().Format(time.RFC3339), time.Now().Format(time.RFC3339))
	if err != nil {
		slog.Error("paymentservice:dao_sqlite:CreateUserPurchase", "error", err)
		return "", err
	}

	return id, nil
}
