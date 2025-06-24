package dao

import (
	"database/sql"
	"embed"
	"fmt"
	"log"
	"log/slog"

	// sqlite_vec "github.com/asg017/sqlite-vec-go-bindings/cgo"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite"

	_ "github.com/mattn/go-sqlite3"

	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

const MIGRATION_TABLE = "chatservice_migrations"
const SEED_MIGRATION_TABLE = "chatservice_seed"

//go:embed db/sqlite/scripts/migrations
var sqliteMigrationFiles embed.FS

//go:embed db/sqlite/scripts/seed
var sqliteSeedFiles embed.FS

func MigrateDB_UsingConnection(sqlDB *sql.DB, files embed.FS, directoryInFS string, migrationsTable string) error {
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

func MigrateSQLite(dbURL string) error {
	slog.Info("Migrating database", "dbURL", dbURL)
	// sqlite_vec.Auto()

	sqlDB, err := sql.Open("sqlite", dbURL)
	if err != nil {
		slog.Error("error", "err", err)
	}

	return MigrateDB_UsingConnection(sqlDB, sqliteMigrationFiles, "db/sqlite/scripts/migrations", MIGRATION_TABLE)
}

func SeedSqlite(dbURL string) error {
	slog.Info("Seeding database", "dbURL", dbURL)
	sqlDB, err := sql.Open("sqlite", dbURL)
	if err != nil {
		slog.Error("error", "err", err)
	}

	return MigrateDB_UsingConnection(sqlDB, sqliteSeedFiles, "db/sqlite/scripts/seed", SEED_MIGRATION_TABLE)
}
