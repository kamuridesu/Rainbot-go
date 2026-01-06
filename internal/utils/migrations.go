package utils

import (
	"fmt"
	"log/slog"
	"os"
	"path"
	"slices"
	"strconv"
	"strings"

	"github.com/kamuridesu/rainbot-go/internal/database/providers"
)

func LoadScripts(path string) []os.DirEntry {
	entries, err := os.ReadDir(path)
	if err != nil {
		panic(err)
	}

	slices.SortStableFunc(entries, func(a, b os.DirEntry) int {
		nameA, errA := strconv.Atoi(strings.TrimSuffix(a.Name(), ".sql"))
		nameB, errB := strconv.Atoi(strings.TrimSuffix(b.Name(), ".sql"))
		if errA != nil || errB != nil {
			return 0
		}
		return nameA - nameB
	})

	return entries
}

func migrate(db *providers.Database, migrationsDir string) {
	files := LoadScripts(migrationsDir)
	if len(files) == 0 {
		slog.Error("No migrations found!")
		os.Exit(1)
	}
	tx, err := db.DB.Begin()

	if err != nil {
		panic(err)
	}

	defer func() {
		if p := recover(); p != nil {
			slog.Error("Migration panicked, rolling back trasaction")
			tx.Rollback()
			panic(p)
		} else if err != nil {
			slog.Error("Error while migrating: " + err.Error())
			if rbErr := tx.Rollback(); rbErr != nil {
				slog.Error("error while rolling back trasaction")
			}
		} else {
			slog.Info("Finished migrations, now commiting")
			err = tx.Commit()
		}
	}()

	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".sql") {
			slog.Warn("Skipping non-sql file: " + file.Name())
			continue
		}

		filePath := path.Join(migrationsDir, file.Name())
		content, readErr := os.ReadFile(filePath)
		if readErr != nil {
			panic("fail to read migration file " + filePath)
		}

		body := string(content)
		slog.Info(fmt.Sprintf("Running transaction for file: %s", filePath))

		if _, execErr := tx.Exec(body); execErr != nil {
			panic("failed to run migration on file " + filePath + " err is " + execErr.Error())
		}
	}

}

func MigrateSqlite(db *providers.Database) {
	migrationsDir := "migrations/sqlite"
	migrate(db, migrationsDir)
}

func MigratePostgres(db *providers.Database) {
	migrate(db, "migrations/postgres")
}

func Migrate() {

	defer func() {
		if p := recover(); p != nil {
			slog.Error(fmt.Sprintf("An error happened while processing migrations: %v", p))
			os.Exit(1)
		}
	}()

	dbDriver := os.Getenv("DB_DRIVER")
	dbParams := os.Getenv("DB_PARAMS")

	if dbDriver == "" || dbParams == "" {
		panic("DB_DRIVER and/or DB_PARAMS cannot be empty")
	}

	db, err := providers.InitDB(dbDriver, dbParams)
	if err != nil {
		panic(err)
	}

	switch dbDriver {
	case "sqlite3":
		MigrateSqlite(db)
	case "postgres":
		MigratePostgres(db)
	}

}
