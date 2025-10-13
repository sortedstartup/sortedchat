package dao

import (

	// sqlite_vec "github.com/asg017/sqlite-vec-go-bindings/cgo"

	"log/slog"
	"time"

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

func (d *SQLiteDAO) ListProducts(userID string) ([]*Product, error) {
	slog.Info("paymentservice:dao_sqlite:ListProducts", "userID", userID)

	query := `SELECT * FROM products WHERE user_id = ?`
	productList, err := d.db.Queryx(query, userID)
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
