package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/amantyagi23/authbackend/internal/domain/user"
	"github.com/amantyagi23/authbackend/internal/repository"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type CreateUserSession struct {
	SessionId             uuid.UUID
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
	GetUserSessionByUserId(ctx context.Context, id uuid.UUID) (*user.UserSession, error)
	DeactivateUserSession(ctx context.Context, id uuid.UUID) error
	GetUserSessionByRefreshToken(ctx context.Context, refreshToken string) (*user.UserSession, error)
	UpdateAccessToken(ctx context.Context, sessionId uuid.UUID, accessToken string) error
}

type userSessionUsecase struct {
	repo repository.UserSessionRepository
	log  *zap.Logger
}

func NewUserSessionUsecase(repo repository.UserSessionRepository, log *zap.Logger) UserSessionUsecase {
	return &userSessionUsecase{repo: repo, log: log}
}

// CreateUser validates input, delegates entity creation to the user,
// checks for duplicates, and persists the new user.
func (usuc *userSessionUsecase) CreateUserSession(ctx context.Context, input CreateUserSession) error {
	usuc.log.Info("CreateUserSession: called", zap.String("UserId", input.UserId.String()))

	// 2. Delegate entity construsuction (including password hashing) to the user.
	userSession, err := user.NewSession(input.SessionId, input.UserId, input.AccessToken, input.RefreshToken, input.AccessTokenExpiredAt, input.RefreshTokenExpiredAt, "", "", "", "", "")
	if err != nil {
		return fmt.Errorf("CreateUserSession: user: %w", err)
	}

	// 3. Persist via the repository interface.
	if err := usuc.repo.Create(ctx, userSession); err != nil {
		return fmt.Errorf("CreateUserSession: persist: %w", err)
	}

	usuc.log.Info("CreateUserSession: susuccess", zap.String("user_id", userSession.SessionID.String()))

	return nil
}

// GetUser fetches a single user by ID, returning a sanitized response.
func (usuc *userSessionUsecase) GetUserSessionByUserId(ctx context.Context, id uuid.UUID) (*user.UserSession, error) {
	usuc.log.Info("GetUser: called", zap.String("user_id", id.String()))

	// user, err := usuc.repo.GetByID(ctx, id)
	// if err != nil {
	// 	return nil, fmt.Errorf("GetUser: %w", err)
	// }

	return nil, nil
}

func (usuc *userSessionUsecase) DeactivateUserSession(ctx context.Context, id uuid.UUID) error {
	usuc.log.Info("DeactivateUserSession: called", zap.String("email", id.String()))

	usuc.repo.Delete(ctx, id)

	return nil
}

func (usuc *userSessionUsecase) GetUserSessionByRefreshToken(ctx context.Context, refreshToken string) (*user.UserSession, error) {
	usuc.log.Info("GetUserSessionByRefreshToken: called", zap.String("RefreshToken", refreshToken))

	userSession, err := usuc.repo.GetByRefreshToken(ctx, refreshToken)

	if err != nil {
		return nil, fmt.Errorf("GetUserSessionByRefreshToken: %w", err)
	}

	return userSession, nil
}

func (usuc *userSessionUsecase) UpdateAccessToken(ctx context.Context, sessionId uuid.UUID, accessToken string) error {
	usuc.log.Info("UpdateAccessToken: called", zap.String("userSessionId", sessionId.String()))

	err := usuc.repo.UpdateAccessToken(ctx, sessionId, accessToken)
	if err != nil {
		return fmt.Errorf("UpdateAccessToken: %w", err)
	}
	return nil
}
