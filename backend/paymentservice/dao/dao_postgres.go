package dao

import (
	"fmt"
	"log/slog"
	"time"

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

func (d *PostgresDAO) CreateProduct(id string, userID string, name string, description string, cost string, currency string) (string, error) {
	slog.Info("paymentservice:dao_postgres:CreateProduct", "userID", userID, "name", name, "description", description, "cost", cost, "currency", currency)

	query := `INSERT INTO products (id, user_id, name, description, price, currency, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	_, err := d.db.Exec(query, id, userID, name, description, cost, currency, time.Now().Format(time.RFC3339), time.Now().Format(time.RFC3339))
	if err != nil {
		slog.Error("paymentservice:dao_postgres:CreateProduct", "error", err)
		return "", err
	}

	return id, nil
}

func (d *PostgresDAO) ListProducts(userID string) ([]*Product, error) {
	slog.Info("paymentservice:dao_postgres:ListProducts", "userID", userID)

	query := `SELECT * FROM products WHERE user_id = $1`
	productList, err := d.db.Queryx(query, userID)
	if err != nil {
		slog.Error("paymentservice:dao_postgres:ListProducts", "error", err)
		return nil, err
	}

	products := []*Product{}
	for productList.Next() {
		product := &Product{}
		err := productList.Scan(product)
		if err != nil {
			slog.Error("paymentservice:dao_postgres:ListProducts", "error", err)
			return nil, err
		}
		products = append(products, product)
	}
	return products, nil
}
