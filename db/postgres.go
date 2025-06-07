// db/postgres.go
package db

import (
	"database/sql"
	"fmt"
	"log"
	"sync"

	_ "github.com/lib/pq"
)

var (
	db   *sql.DB
	once sync.Once
)

func Connect(host string, port int, user, password, dbname string, sslmode string) (*sql.DB, error) {
	var err error
	once.Do(func() {
		connStr := fmt.Sprintf(
			"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
			host, port, user, password, dbname, sslmode,
		)

		db, err = sql.Open("postgres", connStr)
		if err != nil {
			err = fmt.Errorf("failed to open DB connection: %w", err)
			return
		}

		db.SetMaxOpenConns(25)
		db.SetMaxIdleConns(25)
		db.SetConnMaxIdleTime(0)

		err = db.Ping()
		if err != nil {
			err = fmt.Errorf("failed to ping DB: %w", err)
			return
		}

		log.Println("PostgreSQL connected successfully")
	})
	return db, err
}

func Close() error {
	if db != nil {
		return db.Close()
	}
	return nil
}
