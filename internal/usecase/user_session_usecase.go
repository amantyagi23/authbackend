package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/amantyagi23/authbackend/internal/domain"
	"github.com/amantyagi23/authbackend/internal/repository"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type CreateUserSession struct {
	UserId                uuid.UUID
	AccessToken           string
	RefreshToken          string
	AccessTokenExpiredAt  time.Time
	RefreshTokenExpiredAt time.Time
}

// UserUsecase is the application-layer contract.
// Handlers depend on this interface, not on the concrete strusuct.
//
//go:generate mockery --name=UserUsecase --output=../../mocks --outpkg=mocks
type UserSessionUsecase interface {
	CreateUserSession(ctx context.Context, input CreateUserSession) error
	GetUserSessionByUserId(ctx context.Context, id uuid.UUID) (*domain.UserSession, error)
	DeactivateUserSession(ctx context.Context, id uuid.UUID) error
}

type userSessionUsecase struct {
	repo repository.UserSessionRepository
	log  *zap.Logger
}

func NewUserSessionUsecase(repo repository.UserSessionRepository, log *zap.Logger) UserSessionUsecase {
	return &userSessionUsecase{repo: repo, log: log}
}

// CreateUser validates input, delegates entity creation to the domain,
// checks for duplicates, and persists the new user.
func (usuc *userSessionUsecase) CreateUserSession(ctx context.Context, input CreateUserSession) error {
	usuc.log.Info("CreateUserSession: called", zap.String("UserId", input.UserId.String()))

	// 2. Delegate entity construsuction (including password hashing) to the domain.
	userSession, err := domain.NewSession(input.UserId, input.AccessToken, input.RefreshToken, input.AccessTokenExpiredAt, input.RefreshTokenExpiredAt, "", "", "", "", "")
	if err != nil {
		return fmt.Errorf("CreateUserSession: domain: %w", err)
	}

	// 3. Persist via the repository interface.
	if err := usuc.repo.Create(ctx, userSession); err != nil {
		return fmt.Errorf("CreateUserSession: persist: %w", err)
	}

	usuc.log.Info("CreateUserSession: susuccess", zap.String("user_id", userSession.SessionID.String()))

	return nil
}

// GetUser fetches a single user by ID, returning a sanitized response.
func (usuc *userSessionUsecase) GetUserSessionByUserId(ctx context.Context, id uuid.UUID) (*domain.UserSession, error) {
	usuc.log.Info("GetUser: called", zap.String("user_id", id.String()))

	// user, err := usuc.repo.GetByID(ctx, id)
	// if err != nil {
	// 	return nil, fmt.Errorf("GetUser: %w", err)
	// }

	return nil, nil
}

func (usuc *userSessionUsecase) DeactivateUserSession(ctx context.Context, id uuid.UUID) error {
	usuc.log.Info("DeactivateUserSession: called", zap.String("email", id.String()))

	// user, err := usuc.repo.GetByEmail(ctx, email)
	// if err != nil {
	// 	return nil, fmt.Errorf("GetUserByEmail: %w", err)
	// }

	return nil
}
