package app

import (
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func NewDB(config *Config) (*sqlx.DB, error) {
	db, err := sqlx.Open("postgres", config.Database)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}
