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

	// Connecting to db
	conn, err := pgx.Connect(context.Background(), cfg.Stringify())
	if err != nil {
		panic(err)
	}
	defer conn.Close(context.Background())

	if err := conn.Ping(context.Background()); err != nil {
		panic(err)
	}
	fmt.Println("Connected!")

	// Create users table
	_, err = conn.Exec(context.Background(), `
	CREATE TABLE IF NOT EXISTS users(
		id SERIAL PRIMARY KEY,
		name TEXT,
		email TEXT UNIQUE NOT NULL
	);

	CREATE TABLE IF NOT EXISTS orders(
		id SERIAL PRIMARY KEY,
		user_id INT NOT NULL,
		amount INT,
		description TEXT
	);
	`)
	if err != nil {
		panic(err)
	}
	fmt.Println("Tables Created!")

	// Insert data
	_, err = conn.Exec(context.Background(), `
	INSERT INTO users (name, email)
	VALUES ('Bob', 'bob@test.com');
	`)
	if err != nil {
		panic(err)
	}
	fmt.Println("User Created!")

	// Another user
	name := "Liza"
	email := "liza@test.com"
	_, err = conn.Exec(context.Background(), `
	INSERT INTO users (name, email)
	VALUES ($1, $2);
	`, name, email)
	if err != nil {
		panic(err)
	}
	fmt.Println("Another user Created!")

	// SQL injection demo ☠️
	// name := "'',''); DROP TABLE users; --"
	// email := "haha@test.com"
	// query := fmt.Sprintf(`
	// INSERT INTO users (name, email)
	// VALUES (%s, %s);`, name, email)
	// _, err = conn.Exec(context.Background(), query)
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Println("You've been Pwned!")
}
