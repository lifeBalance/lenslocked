package models

import "errors"

var (
	ErrEmailTaken = errors.New("email address already taken")
	ErrNotFound   = errors.New("not found")
)
