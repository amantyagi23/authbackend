// Package repository defines the UserRepository interface.
//
// Clean Architecture rule: the usecase layer OWNS this interface.
// The infrastructure layer provides the concrete implementation.
// This inverts the dependency — infra depends on domain/usecase, never the reverse.
//
// Dependency flow:
//
//	usecase  →  repository.UserRepository (interface, defined here)
//	                 ↑
//	            infrastructure/db (implements)
package repository

import (
	"context"

	"github.com/amantyagi23/authbackend/internal/domain"
	"github.com/google/uuid"
)

//go:generate mockery --name=UserRepository --output=../../mocks --outpkg=mocks

// UserRepository is the persistence contract for the User aggregate.
// Any storage backend (Postgres, Redis, in-memory) must satisfy this interface.
type UserRepository interface {
	// Create persists a new user. Returns ErrEmailTaken if the email is duplicate.
	Create(ctx context.Context, user *domain.User) error

	// GetByID fetches a user by its UUID. Returns ErrUserNotFound if absent.
	GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error)

	// GetByEmail fetches a user by email. Returns ErrUserNotFound if absent.
	GetByEmail(ctx context.Context, email string) (*domain.User, error)

	// ExistsByEmail returns true if the email is already registered.
	ExistsByEmail(ctx context.Context, email string) (bool, error)
}
