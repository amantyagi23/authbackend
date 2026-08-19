package domain

import (
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// User is the core aggregate root of the user-service bounded context.
type User struct {
	UserID          uuid.UUID
	Name            string
	Email           string
	Password        string // bcrypt hash — never the raw password
	ProfilePic      *string
	IsEmailVerified bool
	IsActive        bool
	CreatedAt       time.Time
	UpdatedAt       time.Time
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
		UserID:          uuid.New(),
		Name:            name,
		Email:           email,
		ProfilePic:      nil,
		IsEmailVerified: false,
		IsActive:        false,
		Password:        string(hash),
		CreatedAt:       time.Now().UTC(),
		UpdatedAt:       time.Now().UTC(),
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
	var profilePic string
	if u.ProfilePic != nil {
		profilePic = *u.ProfilePic
	}

	return UserResponse{
		UserID:          u.UserID,
		Name:            u.Name,
		Email:           u.Email,
		ProfilePic:      profilePic,
		IsEmailVerified: u.IsEmailVerified,
		IsActive:        u.IsActive,
		CreatedAt:       u.CreatedAt,
		UpdatedAt:       u.UpdatedAt,
	}
}

// UserResponse is the public-safe representation of a User.
// The password hash is deliberately omitted.
type UserResponse struct {
	UserID          uuid.UUID `json:"user_id"`
	Name            string    `json:"name"`
	Email           string    `json:"email"`
	ProfilePic      string
	IsEmailVerified bool
	IsActive        bool
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}
