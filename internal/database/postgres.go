package database

import (
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // pq adapter for sql

	"github.com/teamxiv/growbot-api/internal/config"
)

// NewPostgres connects to the database and returns a query generator
func NewPostgres(cfg config.DatabaseConfig) (*sqlx.DB, error) {
	db, err := sqlx.Connect("postgres", cfg.ConnectionString)
	if err != nil {
		return nil, err
	}

	return db, nil
}
