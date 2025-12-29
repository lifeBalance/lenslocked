package context

import (
	"context"

	"github.com/lifebalance/lenslocked/models"
)

// Avoid collisions with other packages’ context values.
// Best practice: never use plain strings as context keys; use an unexported type.
type key string

const (
	userKey key = "user"
)

// Returns a new context carrying the authenticated *models.User
// Uses context.WithValue so downstream code (handlers, services) can
// retrieve the current user from the request context.
// To Use it, we set on the request (e.g., in auth middleware):
//   - ctx := WithUser(r.Context(), user)
//   - r = r.WithContext(ctx); next.ServeHTTP(w, r)
func WithUser(ctx context.Context, user *models.User) context.Context {
	return context.WithValue(ctx, userKey, user)
}

// Fetch the current user from the request context.
//   - ctx.Value(userKey) returns an interface{} for that key.
//   - val.(*models.User) is a type assertion to *models.User.
//   - If the key is missing or the value isn’t a *models.User, ok is false and it returns nil.
//
// - Requirements:
//   - Must use the exact same key (userKey) and type used in WithUser when storing.
//
// - Usage:
//   - In middleware: r = r.WithContext(WithUser(r.Context(), user))
//   - In handlers: u := User(r.Context()); if u == nil { /* no user */ }
func User(ctx context.Context) *models.User {
	val := ctx.Value(userKey)
	user, ok := val.(*models.User)
	if !ok {
		return nil
	}
	return user
}
