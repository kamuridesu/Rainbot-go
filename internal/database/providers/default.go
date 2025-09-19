package providers

import (
	"database/sql"
	"log/slog"
	"strconv"

	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

type Database struct {
	Driver string
	DB     *sql.DB
}

func (d *Database) GetQuery(query string) string {
	switch d.Driver {
	case "sqlite":
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

func InitDB(driver string, parameters string) (*Database, error) {
	db, err := sql.Open(driver, parameters)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, err
	}

	slog.Info("Database successfuly connected")
	return &Database{Driver: driver, DB: db}, nil
}
