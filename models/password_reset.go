package models

import (
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/lifebalance/lenslocked/rand"
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
	lowercasedEmail := strings.ToLower(email)
	var userId int
	// fetch user from DB
	row := svc.DB.QueryRow(`
	SELECT id
	FROM users
	WHERE email=$1;
	`, lowercasedEmail)
	err := row.Scan(&userId)
	if err != nil {
		// TODO: return informative error about user not existing
		return nil, fmt.Errorf("create: %w", err)
	}
	// Generate token
	bytesPerToken := svc.BytesPerToken
	bytesPerToken = max(MinBytesPerToken, bytesPerToken)

	token, err := rand.RandomBase64String(bytesPerToken)
	if err != nil {
		return nil, fmt.Errorf("create: %w", err)
	}
	// Hash the token
	tokenHash := svc.hashToken(token)
	duration := max(svc.Duration, DefaultResetDuration)

	// Create an instance of PasswordReset
	pwdReset := PasswordReset{
		UserID:    userId,
		Token:     token,
		TokenHash: tokenHash,
		ExpiresAt: time.Now().Add(duration),
	}

	// Store PasswordReset in DB
	row = svc.DB.QueryRow(`
		INSERT INTO password_resets (user_id, token_hash, expires_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id) DO
		UPDATE
		SET token_hash = $2, expires_at = $3
		RETURNING id;
	`, pwdReset.UserID, pwdReset.TokenHash, pwdReset.ExpiresAt)
	err = row.Scan(&pwdReset.ID)
	if err != nil {
		return nil, fmt.Errorf("create %w", err)
	}
	// Return a ref. to the PasswordReset instance
	return &pwdReset, nil
}

func (svc *PasswordResetService) Consume(resetToken string) (*User, error) {
	var user User
	var pwdReset PasswordReset
	// Query DB for the reset token, and the user
	tokenHash := svc.hashToken(resetToken)
	row := svc.DB.QueryRow(`
	SELECT password_resets.id,
		password_resets.expires_at,
		users.id,
		users.email,
		users.password_hash
	FROM password_resets
		JOIN users ON users.id = password_resets.user_id
	WHERE password_resets.token_hash=$1;
	`, tokenHash)
	err := row.Scan(
		&pwdReset.ID,
		&pwdReset.ExpiresAt,
		&user.ID,
		&user.Email,
		&user.PasswordHash)
	if err != nil {
		return nil, fmt.Errorf("consume %w", err)
	}

	// Check expiry date of token
	if time.Now().After(pwdReset.ExpiresAt) {
		return nil, fmt.Errorf("token expired: %v", resetToken)
	}
	// Delete token from DB
	err = svc.deleteResetToken(pwdReset.ID)
	if err != nil {
		return nil, fmt.Errorf("consume: %w", err)
	}

	return &user, nil
}

func (svc *PasswordResetService) hashToken(token string) string {
	tokenHash := sha256.Sum256([]byte(token))
	return base64.URLEncoding.EncodeToString(tokenHash[:])
}

func (svc *PasswordResetService) deleteResetToken(tokenId int) error {
	_, err := svc.DB.Exec(`
		DELETE FROM password_resets
		WHERE id = $1;
	`, tokenId)
	if err != nil {
		return fmt.Errorf("delete: %w", err)
	}
	return nil
}
