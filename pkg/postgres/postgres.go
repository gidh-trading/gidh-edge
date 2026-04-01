// pkg/postgres/postgres.go
package postgres

import (
	"database/sql"
	"time"

	_ "github.com/lib/pq"
)

func New(connString string) (*sql.DB, error) {
	db, err := sql.Open("postgres", connString)
	if err != nil {
		return nil, err
	}

	// High-speed settings for trading data
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}
