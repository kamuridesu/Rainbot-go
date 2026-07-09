package utils

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/kamuridesu/rainbot-go/core/database/providers"
)

func LoadScripts(dir string) ([]os.DirEntry, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read migrations directory: %w", err)
	}

	var sqlFiles []os.DirEntry
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".sql") {
			sqlFiles = append(sqlFiles, e)
		}
	}

	slices.SortStableFunc(sqlFiles, func(a, b os.DirEntry) int {
		prefixA := strings.SplitN(a.Name(), "_", 2)[0]
		prefixA = strings.TrimSuffix(prefixA, ".sql")

		prefixB := strings.SplitN(b.Name(), "_", 2)[0]
		prefixB = strings.TrimSuffix(prefixB, ".sql")

		nameA, _ := strconv.Atoi(prefixA)
		nameB, _ := strconv.Atoi(prefixB)
		return nameA - nameB
	})

	return sqlFiles, nil
}

func migrate(db *providers.Database, migrationsDir string) error {
	files, err := LoadScripts(migrationsDir)
	if err != nil {
		return err
	}
	if len(files) == 0 {
		slog.Warn("No migrations found in " + migrationsDir)
		return nil
	}

	_, err = db.DB.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (version TEXT PRIMARY KEY)`)
	if err != nil {
		return fmt.Errorf("failed to create schema_migrations table: %w", err)
	}

	rows, err := db.DB.Query(`SELECT version FROM schema_migrations`)
	if err != nil {
		return fmt.Errorf("failed to fetch applied migrations: %w", err)
	}
	defer rows.Close()

	applied := make(map[string]bool)
	for rows.Next() {
		var version string
		if err := rows.Scan(&version); err == nil {
			applied[version] = true
		}
	}

	tx, err := db.DB.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	migrationsRun := 0
	for _, file := range files {
		filename := file.Name()
		if applied[filename] {
			continue
		}

		filePath := filepath.Join(migrationsDir, filename)
		content, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", filePath, err)
		}

		slog.Info(fmt.Sprintf("Applying migration: %s", filename))

		if _, err = tx.Exec(string(content)); err != nil {
			return fmt.Errorf("migration failed on %s: %w", filename, err)
		}

		insertQuery := fmt.Sprintf(`INSERT INTO schema_migrations (version) VALUES ('%s')`, filename)
		if _, err = tx.Exec(insertQuery); err != nil {
			return fmt.Errorf("failed to record migration %s: %w", filename, err)
		}

		migrationsRun++
	}

	if migrationsRun > 0 {
		if err = tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit migrations: %w", err)
		}
		slog.Info(fmt.Sprintf("Successfully applied %d migrations.", migrationsRun))
	} else {
		slog.Info("Database is up to date.")
	}

	return nil
}

func MigrateSqlite(db *providers.Database) error {
	return migrate(db, "migrations/sqlite")
}

func MigratePostgres(db *providers.Database) error {
	return migrate(db, "migrations/postgres")
}

func Migrate() {
	dbDriver := os.Getenv("DB_DRIVER")
	dbParams := os.Getenv("DB_PARAMS")

	if dbDriver == "" || dbParams == "" {
		slog.Error("DB_DRIVER and DB_PARAMS environment variables cannot be empty")
		os.Exit(1)
	}

	db, err := providers.InitDB(dbDriver, dbParams)
	if err != nil {
		slog.Error("Database initialization failed", "error", err)
		os.Exit(1)
	}

	switch dbDriver {
	case "sqlite3":
		err = MigrateSqlite(db)
	case "postgres":
		err = MigratePostgres(db)
	default:
		slog.Error("Unsupported database driver: " + dbDriver)
		os.Exit(1)
	}

	if err != nil {
		slog.Error("Migration failed", "error", err)
		os.Exit(1)
	}
}
