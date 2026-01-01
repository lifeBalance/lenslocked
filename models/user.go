package models

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID           uint
	Email        string
	PasswordHash string
}

type UserService struct {
	DB *sql.DB
}

func (us *UserService) Create(email, password string) (*User, error) {
	email = strings.ToLower(email)
	// hash the password
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}
	hashString := string(hashedBytes)

	user := User{
		Email:        email,
		PasswordHash: hashString,
	}
	row := us.DB.QueryRow(`
	INSERT INTO users (email, password_hash)
	VALUES ($1, $2)
	RETURNING id;
	`, email, hashString)
	err = row.Scan(&user.ID)
	if err != nil {
		fmt.Println(err)          // ERROR: duplicate key value violates unique constraint "users_email_key" (SQLSTATE 23505) 23505
		var pgErr *pgconn.PgError // Variable needed to use errors.As
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			fmt.Println(pgErr) // ERROR: duplicate key value violates unique constraint "users_email_key" (SQLSTATE 23505) 23505
			return nil, ErrEmailTaken
		}
		return nil, fmt.Errorf("models: create: %w", err) // too much info for the baddies
	}
	return &user, nil
}

func (us *UserService) Authenticate(email, password string) (*User, error) {
	email = strings.ToLower(email)
	user := User{
		Email: email,
	}
	// fetch user from DB
	row := us.DB.QueryRow(`
	SELECT id, password_hash
	FROM users
	WHERE email=$1
	`, email)
	err := row.Scan(&user.ID, &user.PasswordHash)
	if err != nil {
		return nil, fmt.Errorf("authenticate: %w", err)
	}

	// compare the password hashes
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return nil, fmt.Errorf("authenticate: %w", err)
	}

	return &user, nil
}

func (us *UserService) UpdatePassword(userId uint, password string) error {
	// Hash the pwd
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("update password: %w", err)
	}
	pwdHash := string(hashedBytes)
	// Store the hashed pwd
	_, err = us.DB.Exec(`
		UPDATE users
		SET password_hash = $2
		WHERE id = $1;
	`, userId, pwdHash)
	if err != nil {
		return fmt.Errorf("update password: %w", err)
	}
	return nil
}
