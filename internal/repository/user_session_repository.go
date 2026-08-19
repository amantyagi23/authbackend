package repository

import (
	"context"

	"github.com/amantyagi23/authbackend/internal/domain"
)

type UserSessionRepository interface {
	Create(ctx context.Context, userSession *domain.UserSession) error
}
