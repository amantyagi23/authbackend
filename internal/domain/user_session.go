// Package domain contains the User entity and its business rules.
// This layer has ZERO external dependencies — no frameworks, no DB drivers.
// Everything else depends on this layer; this layer depends on nothing.
//
// DDD principle: the domain is the source of truth for what a User IS
// and what invariants must always hold.
package domain

import (
	"time"

	"github.com/google/uuid"
)

// User is the core aggregate root of the user-service bounded context.
type UserSession struct {
	SessionID uuid.UUID
	UserID    uuid.UUID
	Token     string
	IsDeleted bool
	ExpiredAt time.Time
	CreatedAt time.Time
}

// NewUser is the factory function for creating a valid User.
// It enforces all creation invariants (hashing the password, assigning an ID).
func NewSession(userId, token string, expiredAt time.Time) (*UserSession, error) {
	if token == "" {
		return nil, ErrNoTokenFound
	}

	return &UserSession{
		SessionID: uuid.New(),
		UserID:    uuid.New(),
		Token:     token,
		IsDeleted: false,
		ExpiredAt: time.Now().UTC(),
		CreatedAt: time.Now().UTC(),
	}, nil
}
