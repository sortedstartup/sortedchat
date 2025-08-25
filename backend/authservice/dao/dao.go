package dao

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	"github.com/jmoiron/sqlx"
)

type UserDAO interface {
	DoesUserExist(userID string) (bool, error)
	CreateUserIfNotExists(userID, email, roles, oAuthProvider, oAuthUserID string, isFederated bool) (string, error)
	GetUserIDByEmail(email string) (string, error)
}

type UserDaoFactory struct {
}

// SQLiteDAO implements the DAO interface using SQLite and sqlx
type UserSqliteDAO struct {
	db *sqlx.DB
}

// NewSQLiteDAO creates a new SQLite DAO instance
func NewUserSqliteDAO(sqliteUrl string) (*UserSqliteDAO, error) {
	// sqlite_vec.Auto()

	db, err := sqlx.Open("sqlite3", sqliteUrl)
	if err != nil {
		return nil, err
	}

	return &UserSqliteDAO{db: db}, nil
}

func (u *UserSqliteDAO) DoesUserExist(userID string) (bool, error) {
	var count int
	query := `SELECT COUNT(*) FROM userservice_users WHERE user_id = ?`
	err := u.db.QueryRow(query, userID).Scan(&count)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	return count > 0, nil
}

func (u *UserSqliteDAO) CreateUserIfNotExists(userID, email, roles, oAuthProvider, oAuthUserID string, isFederated bool) (string, error) {
	// First check if user exists by email
	existingUserID, err := u.GetUserIDByEmail(email)
	if err == nil && existingUserID != "" {
		// User already exists, return existing user ID
		return existingUserID, nil
	}

	// Check if user already exists by userID
	slog.Debug("Checking if user already exists by userID", "userID", userID)
	exists, err := u.DoesUserExist(userID)
	if err != nil {
		slog.Error("Error checking if user already exists by userID", "error", err)
		return "", err
	}
	if exists {
		slog.Debug("User already exists by userID", "userID", userID)
		return userID, nil
	}

	// User doesn't exist, insert new user
	insertQuery := `
		INSERT INTO userservice_users 
		(user_id, email, roles, oauth_provider, oauth_user_id, is_federated)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	slog.Debug("Inserting new user", "userID", userID, "email", email, "roles", roles, "oAuthProvider", oAuthProvider, "oAuthUserID", oAuthUserID, "isFederated", isFederated)
	_, err = u.db.Exec(insertQuery, userID, email, roles, oAuthProvider, oAuthUserID, isFederated)
	if err != nil {
		slog.Error("Error inserting new user", "error", err)
		return "", err
	}
	return userID, nil
}

func (u *UserSqliteDAO) GetUserIDByEmail(email string) (string, error) {
	var userID string
	query := `SELECT user_id FROM userservice_users WHERE email = ?`
	fmt.Println("SQLITE db variable in GetUserIDByEmail", u.db)
	err := u.db.QueryRow(query, email).Scan(&userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", nil // User not found, but not an error
		}
		return "", err
	}
	return userID, nil
}

// PostgresDAO implements the DAO interface using PostgreSQL and sqlx
type UserPostgresDAO struct {
	db *sqlx.DB
}

// NewPostgresDAO creates a new PostgreSQL DAO instance
func NewUserPostgresDAO(config *PostgresConfig) (*UserPostgresDAO, error) {
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

	return &UserPostgresDAO{db: db}, nil
}

func (u *UserPostgresDAO) DoesUserExist(userID string) (bool, error) {
	var count int
	query := `SELECT COUNT(*) FROM userservice_users WHERE user_id = $1`
	err := u.db.QueryRow(query, userID).Scan(&count)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	return count > 0, nil
}

func (u *UserPostgresDAO) CreateUserIfNotExists(userID, email, roles, oAuthProvider, oAuthUserID string, isFederated bool) (string, error) {
	// First check if user exists by email
	existingUserID, err := u.GetUserIDByEmail(email)
	if err == nil && existingUserID != "" {
		// User already exists, return existing user ID
		return existingUserID, nil
	}

	// Check if user already exists by userID
	exists, err := u.DoesUserExist(userID)
	if err != nil {
		slog.Error("Error checking if user already exists by userID", "error", err)
		return "", err
	}
	if exists {
		slog.Debug("User already exists by userID", "userID", userID)
		return userID, nil
	}

	// User doesn't exist, insert new user
	insertQuery := `
		INSERT INTO userservice_users 
		(user_id, email, roles, oauth_provider, oauth_user_id, is_federated)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	slog.Debug("Inserting new user", "userID", userID, "email", email, "roles", roles, "oAuthProvider", oAuthProvider, "oAuthUserID", oAuthUserID, "isFederated", isFederated)
	_, err = u.db.Exec(insertQuery, userID, email, roles, oAuthProvider, oAuthUserID, isFederated)
	if err != nil {
		slog.Error("Error inserting new user", "error", err)
		return "", err
	}
	return userID, nil
}

func (u *UserPostgresDAO) GetUserIDByEmail(email string) (string, error) {
	var userID string
	query := `SELECT user_id FROM userservice_users WHERE email = $1`
	err := u.db.QueryRow(query, email).Scan(&userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", nil // User not found, but not an error
		}
		slog.Error("Error getting user ID by email", "error", err)
		return "", err
	}
	return userID, nil
}
