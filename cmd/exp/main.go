package main

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

type PostgresConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Database string
	SSLMode  string
}

func (cfg PostgresConfig) Stringify() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s", cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Database, cfg.SSLMode)
}

func main() {
	cfg := PostgresConfig{
		Host:     "localhost",
		Port:     "5433",
		User:     "bob",
		Database: "lenslocked",
		Password: "1234",
		SSLMode:  "disable",
	}
	conn, err := pgx.Connect(context.Background(), cfg.Stringify())
	if err != nil {
		panic(err)
	}
	defer conn.Close(context.Background())

	if err := conn.Ping(context.Background()); err != nil {
		panic(err)
	}

	fmt.Println("Connected!")
}
