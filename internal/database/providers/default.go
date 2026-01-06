package providers

import (
	"database/sql"
	"log/slog"
	"slices"
	"strconv"

	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

type Database struct {
	Driver string
	DB     *sql.DB
	closed bool
}

func (d *Database) GetQuery(query string) string {
	switch d.Driver {
	case "sqlite3":
		return query
	case "postgres":
		oQuery := ""
		counter := 1
		for i := 0; i < len(query); i++ {
			char := query[i]
			if char == '?' {
				oQuery += "$" + strconv.Itoa(counter)
				counter++
				continue
			}
			oQuery += string(char)
		}
		return oQuery
	default:
		return query
	}
}

func InitDB(driver, parameters string) (*Database, error) {
	if !slices.Contains([]string{"sqlite3", "postgres"}, driver) {
		panic("Only supported databases are sqlite3 and postgres")
	}

	db, err := sql.Open(driver, parameters)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, err
	}

	slog.Info("Database successfuly connected")
	return &Database{Driver: driver, DB: db, closed: false}, nil
}

func (db *Database) Close() error {
	if !db.closed {
		db.closed = true
		return db.DB.Close()
	}
	return nil
}
