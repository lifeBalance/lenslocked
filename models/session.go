package models

import (
	"database/sql"
)

/*
The Token field (unhashed) is only set when creating a NEW session. When looking up a session it will be left EMPTY, as we only store the TokenHash in the database.
*/
type Session struct {
	ID        uint
	UserID    uint
	Token     string // Set when creating a NEW sesion (not stored in DB)
	TokenHash string
}

type SessionService struct {
	DB *sql.DB
}

func (ss *SessionService) Create(userId int) (*Session, error) {
	// create the token
	// hash the token
	// save the hash to DB
	// return the user
	return nil, nil
}

func (ss *SessionService) User(token string) (*User, error) {
	// hash the token
	// find user in the DB, using the hashed token
	// fetch user from DB
	// return user
	return nil, nil
}
