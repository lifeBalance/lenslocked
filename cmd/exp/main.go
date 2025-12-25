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
	// var returnedId int
	// name := "Bob"
	// email := "bob@test.com"
	// row := conn.QueryRow(context.Background(), `
	// INSERT INTO users (name, email)
	// VALUES ($1, $2)
	// RETURNING id;
	// `, name, email)
	// err = row.Scan(&returnedId)
	// // row.Err()
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Println("User Created! id:", returnedId, row)
	// fmt.Println("row -->", row)

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

	// Get user by Id
	// id := 4 // use id of an existing row
	// var name string
	// var email string
	// row := conn.QueryRow(context.Background(), `
	// SELECT name, email
	// FROM users
	// WHERE id=$1
	// `, id)
	// err = row.Scan(&name, &email)
	// if errors.Is(err, sql.ErrNoRows) {
	// 	fmt.Println("-> No rows <-") // if the user id doesn't exist panic
	// }
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Printf("User: %s, %s\n", name, email)

	// insert multiple records
	// userId := 4
	// for i := range 5 {
	// 	amount := i * 100
	// 	desc := fmt.Sprintf("Fake order: #%d", i)
	// 	_, err := conn.Exec(context.Background(), `
	// 	INSERT INTO orders (user_id, amount, description)
	// 	VALUES($1, $2, $3)
	// 	`, userId, amount, desc)
	// 	if err != nil {
	// 		panic(err)
	// 	}
	// }
	// fmt.Println("Created fake orders")

	// query multiple records
	type Order struct {
		ID          int
		UserID      int
		Amount      int
		Description string
	}
	var orders []Order
	userId := 4
	rows, err := conn.Query(context.Background(), `
	SELECT id, amount, description
	FROM orders
	WHERE user_id=$1
	`, userId)
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	for rows.Next() {
		var order Order
		order.UserID = userId
		err := rows.Scan(&order.ID, &order.Amount, &order.Description)
		if err != nil {
			panic(err)
		}
		orders = append(orders, order)
	}
	// check for error
	err = rows.Err()
	if err != nil {
		panic(err)
	}
	fmt.Println("orders:", orders)
}
