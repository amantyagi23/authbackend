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
	"golang.org/x/crypto/bcrypt"
)

// User is the core aggregate root of the user-service bounded context.
type User struct {
	ID        uuid.UUID
	Name      string
	Email     string
	Password  string // bcrypt hash — never the raw password
	CreatedAt time.Time
}

// NewUser is the factory function for creating a valid User.
// It enforces all creation invariants (hashing the password, assigning an ID).
func NewUser(name, email, rawPassword string) (*User, error) {
	if name == "" {
		return nil, ErrNameRequired
	}
	if email == "" {
		return nil, ErrEmailRequired
	}
	if len(rawPassword) < 8 {
		return nil, ErrPasswordTooShort
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(rawPassword), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	return &User{
		ID:        uuid.New(),
		Name:      name,
		Email:     email,
		Password:  string(hash),
		CreatedAt: time.Now().UTC(),
	}, nil
}

// VerifyPassword checks a raw password against the stored bcrypt hash.
func (u *User) VerifyPassword(rawPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(rawPassword))
	return err == nil
}

// Sanitized returns a copy of the user without sensitive fields.
// Use this whenever returning a User to the outside world.
func (u *User) Sanitized() UserResponse {
	return UserResponse{
		ID:        u.ID,
		Name:      u.Name,
		Email:     u.Email,
		CreatedAt: u.CreatedAt,
	}
}

// UserResponse is the public-safe representation of a User.
// The password hash is deliberately omitted.
type UserResponse struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}
