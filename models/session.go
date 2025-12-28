package models

import (
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"

	"github.com/lifebalance/lenslocked/rand"
)

/*
# We use 32 bytes for session tokens.

A single byte stores eight bits (1 or 0), so we can store 256 possible values in a byte (0 to 255).

With 32 bytes, we have 32 * 8 = 256 bits of randomness, which means 2^256 possible values.
That's approximately 1.16 × 10^77 possible combinations—an astronomically large number that makes brute-force attacks computationally infeasible.

For context:
- 16 bytes (128 bits): ~3.4 × 10^38 combinations - considered secure for most applications
- 32 bytes (256 bits): ~1.16 × 10^77 combinations - provides an extra safety margin

Industry standards (NIST, OWASP) recommend at least 128 bits of entropy for session tokens. We use 256 bits (32 bytes) to provide additional security headroom and align with common cryptographic key sizes (like AES-256).

This makes guessing a valid session token virtually impossible, even with massive computational resources.
*/
const (
	MinBytesPerToken = 32
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

/* BytesPerToken determines how many bytes our session tokens are gonna have. If this field is not set, MinBytesPerToken will be used */
type SessionService struct {
	DB            *sql.DB
	BytesPerToken int
}

func (ss *SessionService) Upsert(userId int) (*Session, error) {
	bytesPerToken := ss.BytesPerToken
	// if bytesPerToken < MinBytesPerToken {
	// 	bytesPerToken = MinBytesPerToken
	// }
	bytesPerToken = max(MinBytesPerToken, bytesPerToken)

	// create the token
	token, err := rand.RandomBase64String(bytesPerToken)
	if err != nil {
		return nil, fmt.Errorf("create: %w", err)
	}
	// hash the token
	session := Session{
		UserID:    uint(userId),
		Token:     token,
		TokenHash: ss.hashToken(token),
	}
	// try to update the user session
	row := ss.DB.QueryRow(`
		UPDATE sessions
		SET token_hash = $2
		WHERE user_id = $1
		RETURNING id;
	`, session.UserID, session.TokenHash)
	err = row.Scan(&session.ID)

	// but if there's error, create it
	if errors.Is(err, sql.ErrNoRows) {
		row = ss.DB.QueryRow(`
			INSERT INTO sessions (user_id, token_hash)
			VALUES ($1, $2)
			RETURNING id;
		`, userId, session.TokenHash)
		err = row.Scan(&session.ID)
	}

	if err != nil {
		return nil, fmt.Errorf("create %w", err)
	}
	// return the session
	return &session, nil
}

func (ss *SessionService) User(token string) (*User, error) {
	// hash the token
	tokenHash := ss.hashToken(token)

	// find details of the logged-in user, using the hashed token.
	var user User
	row := ss.DB.QueryRow(`
		SELECT users.id, users.email, users.password_hash
		FROM sessions
		JOIN users ON users.id = sessions.user_id
		WHERE sessions.token_hash = $1;
	`, tokenHash)
	err := row.Scan(&user.ID, &user.Email, &user.PasswordHash)
	if err != nil {
		return nil, fmt.Errorf("user: %w", err)
	}

	return &user, nil
}

func (ss *SessionService) DeleteSession(token string) error {
	tokenHash := ss.hashToken(token)
	_, err := ss.DB.Exec(`
		DELETE FROM sessions
		WHERE token_hash = $1;
	`, tokenHash)
	if err != nil {
		return fmt.Errorf("delete: %w", err)
	}
	return nil
}

func (ss *SessionService) hashToken(token string) string {
	tokenHash := sha256.Sum256([]byte(token))
	return base64.URLEncoding.EncodeToString(tokenHash[:])
}
