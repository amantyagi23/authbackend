package usecase

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"time"

	"github.com/amantyagi23/authbackend/internal/domain/user"
	"github.com/amantyagi23/authbackend/internal/repository"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func generateToken() (string, error) {
	b := make([]byte, 32)

	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(b), nil
}

func hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))

	return hex.EncodeToString(hash[:])
}

type CreateEmailVerificationInput struct {
	UserID uuid.UUID
}

type EmailVerificationUsecase interface {
	Create(ctx context.Context, input CreateEmailVerificationInput) (*user.EmailVerification, string, error)
	Verify(ctx context.Context, token string) (*user.EmailVerification, error)
}

type emailVerificationUsecase struct {
	repo repository.EmailVerificationRepository
	log  *zap.Logger
}

func NewEmailVerificationUsecase(
	repo repository.EmailVerificationRepository,
	log *zap.Logger,
) EmailVerificationUsecase {
	return &emailVerificationUsecase{
		repo: repo,
		log:  log,
	}
}

func (u *emailVerificationUsecase) Create(ctx context.Context, input CreateEmailVerificationInput) (*user.EmailVerification, string, error) {

	if input.UserID == uuid.Nil {
		return nil, "", user.ErrUserNotFound
	}

	// Generate raw token.
	rawToken, err := generateToken()
	if err != nil {
		return nil, "", err
	}

	// Hash token before storing.
	tokenHash := hashToken(rawToken)

	// Token valid for 30 minutes.
	expiresAt := time.Now().UTC().Add(30 * time.Minute)

	verification, err := user.NewEmailVerification(
		input.UserID,
		tokenHash,
		expiresAt,
	)
	if err != nil {
		return nil, "", err
	}

	// Remove previous active verification.
	if err := u.repo.DeleteByUserID(ctx, input.UserID); err != nil {
		return nil, "", err
	}

	if err := u.repo.Create(ctx, verification); err != nil {
		return nil, "", err
	}

	// Raw token is returned only so it can be sent by email.
	return verification, rawToken, nil
}

func (u *emailVerificationUsecase) Verify(ctx context.Context, token string) (*user.EmailVerification, error) {
	if token == "" {
		return nil, user.ErrInvalidVerificationToken
	}

	tokenHash := hashToken(token)

	verification, err := u.repo.GetByTokenHash(
		ctx,
		tokenHash,
	)
	if err != nil {
		return nil, err
	}

	if err := verification.MarkUsed(); err != nil {
		return nil, err
	}

	if err := u.repo.MarkUsed(
		ctx,
		verification.ID,
	); err != nil {
		return nil, err
	}

	return verification, nil
}
