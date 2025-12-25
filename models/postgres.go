package models

import (
	"database/sql"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type PostgresConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Database string
	SSLMode  string
}

func DefaultPostgresConfig() PostgresConfig {
	return PostgresConfig{
		Host:     "localhost",
		Port:     "5433",
		User:     "bob",
		Database: "lenslocked",
		Password: "1234",
		SSLMode:  "disable",
	}
}

func (cfg PostgresConfig) Stringify() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s", cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Database, cfg.SSLMode)
}

func Open(cfg PostgresConfig) (*sql.DB, error) {
	conn, err := sql.Open("pgx", cfg.Stringify())
	if err != nil {
		return nil, fmt.Errorf("open: %w", err)
	}

	return conn, nil
}
