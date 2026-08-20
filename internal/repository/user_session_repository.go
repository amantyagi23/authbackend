package repository

import (
	"context"

	"github.com/amantyagi23/authbackend/internal/domain/user"
	"github.com/google/uuid"
)

type UserSessionRepository interface {
	Create(ctx context.Context, userSession *user.UserSession) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetByRefreshToken(ctx context.Context, refreshTokenHash string) (*user.UserSession, error)
	UpdateAccessToken(ctx context.Context, sessionId uuid.UUID, accessToken string) error
}
