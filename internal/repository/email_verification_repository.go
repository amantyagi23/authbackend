package repository

import (
	"context"

	domain "github.com/amantyagi23/authbackend/internal/domain/user"
	"github.com/google/uuid"
)

type EmailVerificationRepository interface {
	Create(ctx context.Context, verification *domain.EmailVerification) error

	GetByID(ctx context.Context, id uuid.UUID) (*domain.EmailVerification, error)

	GetByUserID(ctx context.Context, userID uuid.UUID) (*domain.EmailVerification, error)

	GetByTokenHash(ctx context.Context, tokenHash string) (*domain.EmailVerification, error)

	MarkUsed(ctx context.Context, id uuid.UUID) error

	DeleteByUserID(ctx context.Context, userID uuid.UUID) error
}
