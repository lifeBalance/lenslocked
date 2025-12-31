package models

import (
	"database/sql"
	"fmt"
	"time"
)

const (
	DefaultResetDuration = 1 * time.Hour
)

type PasswordReset struct {
	ID        int
	UserID    int
	Token     string // Only set when creating a pwd reset (not stored in db)
	TokenHash string
	ExpiresAt time.Time
}

/* BytesPerToken determines how many bytes our session tokens are gonna have. If this field is not set, MinBytesPerToken (session.go) will be used */
type PasswordResetService struct {
	DB            *sql.DB
	BytesPerToken int
	Duration      time.Duration // Defaults to DefaultResetDuration
}

func (svc *PasswordResetService) Create(email string) (*PasswordReset, error) {
	return nil, fmt.Errorf("TODO: implement PasswordResetService.Create")
}

func (svc *PasswordResetService) Consume(resetToken string) (*PasswordReset, error) {
	return nil, fmt.Errorf("TODO: implement PasswordResetService.Consume")
}
