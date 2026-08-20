// Package usecase contains the application logic for the user-service.
//
// The usecase layer orchestrates user objects and repository calls.
// It knows about business workflows but has zero HTTP/transport knowledge.
//
// Dependency flow:
//
//	delivery/http  →  usecase.UserUsecase (interface)
//	                        ↓
//	                  user + repository.UserRepository
package usecase

import (
	"context"
	"fmt"

	"github.com/amantyagi23/authbackend/internal/domain/user"
	"github.com/amantyagi23/authbackend/internal/repository"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// CreateUserInput is the DTO for the CreateUser use case.
// Using a dedicated input struct decouples the use case from HTTP request shapes.
type CreateUserInput struct {
	Name     string `validate:"required,min=2,max=100"`
	Email    string `validate:"required,email"`
	Password string `validate:"required,min=8"`
}

// UserUsecase is the application-layer contract.
// Handlers depend on this interface, not on the concrete struct.
//
//go:generate mockery --name=UserUsecase --output=../../mocks --outpkg=mocks
type UserUsecase interface {
	CreateUser(ctx context.Context, input CreateUserInput) (*user.UserResponse, error)
	GetUser(ctx context.Context, id uuid.UUID) (*user.User, error)
	GetUserByEmail(ctx context.Context, email string) (*user.User, error)
	UpdateUser(ctx context.Context, user *user.User) error
}

// userUsecase is the concrete implementation of UserUsecase.
type userUsecase struct {
	repo repository.UserRepository
	log  *zap.Logger
}

// NewUserUsecase constructs a userUsecase with its dependencies injected.
func NewUserUsecase(repo repository.UserRepository, log *zap.Logger) UserUsecase {
	return &userUsecase{repo: repo, log: log}
}

// CreateUser validates input, delegates entity creation to the user,
// checks for duplicates, and persists the new user.
func (uc *userUsecase) CreateUser(ctx context.Context, input CreateUserInput) (*user.UserResponse, error) {
	uc.log.Info("CreateUser: called", zap.String("email", input.Email))

	// 1. Check for duplicate email — a business rule, not a DB constraint concern.
	exists, err := uc.repo.ExistsByEmail(ctx, input.Email)
	if err != nil {
		return nil, fmt.Errorf("CreateUser: check email: %w", err)
	}
	if exists {
		return nil, user.ErrEmailTaken
	}

	// 2. Delegate entity construction (including password hashing) to the user.
	user, err := user.NewUser(input.Name, input.Email, input.Password)
	if err != nil {
		return nil, fmt.Errorf("CreateUser: user: %w", err)
	}

	// 3. Persist via the repository interface.
	if err := uc.repo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("CreateUser: persist: %w", err)
	}

	uc.log.Info("CreateUser: success", zap.String("user_id", (user.UserID).String()))

	resp := user.Sanitized()

	return &resp, nil
}

// GetUser fetches a single user by ID, returning a sanitized response.
func (uc *userUsecase) GetUser(ctx context.Context, id uuid.UUID) (*user.User, error) {
	uc.log.Info("GetUser: called", zap.String("user_id", id.String()))

	user, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("GetUser: %w", err)
	}

	return user, nil
}

func (uc *userUsecase) GetUserByEmail(ctx context.Context, email string) (*user.User, error) {
	uc.log.Info("GetUserByEmail: called", zap.String("email", email))

	user, err := uc.repo.GetByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("GetUserByEmail: %w", err)
	}

	return user, nil
}
func (uc *userUsecase) UpdateUser(ctx context.Context, user *user.User) error {
	uc.log.Info("UpdateUser: called", zap.String("email", user.Email))

	err := uc.repo.Update(ctx, user)
	if err != nil {
		return fmt.Errorf("UpdateUser: %w", err)
	}

	return nil
}
