package dao

import (
	"database/sql"
	"embed"
	"fmt"
	"log"
	"log/slog"

	sqlite_vec "github.com/asg017/sqlite-vec-go-bindings/cgo"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/database/sqlite"

	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"

	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

const MIGRATION_TABLE = "userservice_migrations"
const SEED_MIGRATION_TABLE = "userservice_seed"

//go:embed db/sqlite/scripts/migrations
var sqliteMigrationFiles embed.FS

//go:embed db/sqlite/scripts/seed
var sqliteSeedFiles embed.FS

//go:embed db/postgres/scripts/migrations
var postgresMigrationFiles embed.FS

//go:embed db/postgres/scripts/seed
var postgresSeedFiles embed.FS

func MigrateDB_UsingConnection_SQLite(sqlDB *sql.DB, files embed.FS, directoryInFS string, migrationsTable string) error {
	_files, err := iofs.New(files, directoryInFS)
	if err != nil {
		log.Fatal(err)
	}

	dbInstance, err := sqlite.WithInstance(sqlDB, &sqlite.Config{MigrationsTable: migrationsTable})
	if err != nil {
		slog.Error("error", "err", err)
		return err
	}

	//TODO: externalize in config
	m, err := migrate.NewWithInstance("iofs", _files, "DUMMY", dbInstance)

	if err != nil {
		slog.Error("error", "err", err)
		return fmt.Errorf("failed creating new migration: %w", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		slog.Error("error", "err", err)
		return fmt.Errorf("failed while migrating: %w", err)
	}

	return nil
}

func MigrateDB_UsingConnection_Postgres(sqlDB *sql.DB, files embed.FS, directoryInFS string, migrationsTable string) error {
	_files, err := iofs.New(files, directoryInFS)
	if err != nil {
		log.Fatal(err)
	}

	dbInstance, err := postgres.WithInstance(sqlDB, &postgres.Config{MigrationsTable: migrationsTable})
	if err != nil {
		slog.Error("error", "err", err)
		return err
	}

	m, err := migrate.NewWithInstance("iofs", _files, "DUMMY", dbInstance)

	if err != nil {
		slog.Error("error", "err", err)
		return fmt.Errorf("failed creating new migration: %w", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		slog.Error("error", "err", err)
		return fmt.Errorf("failed while migrating: %w", err)
	}

	return nil
}

func MigrateSQLite(dbURL string) error {
	slog.Info("AuthService: Migrating SQLite database", "dbURL", dbURL)
	sqlite_vec.Auto()
	sqlDB, err := sql.Open("sqlite3", dbURL)
	if err != nil {
		slog.Error("error", "err", err)
		return err
	}
	defer sqlDB.Close()

	return MigrateDB_UsingConnection_SQLite(sqlDB, sqliteMigrationFiles, "db/sqlite/scripts/migrations", MIGRATION_TABLE)
}

func SeedSqlite(dbURL string) error {
	slog.Info("AuthService: Seeding SQLite database", "dbURL", dbURL)
	sqlite_vec.Auto()
	sqlDB, err := sql.Open("sqlite3", dbURL)
	if err != nil {
		slog.Error("error", "err", err)
		return err
	}
	defer sqlDB.Close()

	return MigrateDB_UsingConnection_SQLite(sqlDB, sqliteSeedFiles, "db/sqlite/scripts/seed", SEED_MIGRATION_TABLE)
}

func MigratePostgres(dbURL string) error {
	slog.Info("AuthService: Connecting to PostgreSQL database")
	sqlDB, err := sql.Open("postgres", dbURL)
	if err != nil {
		slog.Error("error", "err", err)
		return err
	}
	defer sqlDB.Close()

	return MigrateDB_UsingConnection_Postgres(sqlDB, postgresMigrationFiles, "db/postgres/scripts/migrations", MIGRATION_TABLE)
}

func SeedPostgres(dbURL string) error {
	slog.Info("AuthService: Seeding PostgreSQL database", "dbURL", dbURL)
	sqlDB, err := sql.Open("postgres", dbURL)
	if err != nil {
		slog.Error("error", "err", err)
		return err
	}
	defer sqlDB.Close()

	return MigrateDB_UsingConnection_Postgres(sqlDB, postgresSeedFiles, "db/postgres/scripts/seed", SEED_MIGRATION_TABLE)
}
