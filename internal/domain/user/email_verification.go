package user

import (
	"time"

	"github.com/google/uuid"
)

type EmailVerification struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	TokenHash string
	ExpiresAt time.Time
	UsedAt    *time.Time
	CreatedAt time.Time
}

func NewEmailVerification(
	userID uuid.UUID,
	tokenHash string,
	expiresAt time.Time,
) (*EmailVerification, error) {
	if userID == uuid.Nil {
		return nil, ErrUserNotFound
	}

	if tokenHash == "" {
		return nil, ErrInvalidVerificationToken
	}

	return &EmailVerification{
		ID:        uuid.New(),
		UserID:    userID,
		TokenHash: tokenHash,
		ExpiresAt: expiresAt,
		CreatedAt: time.Now().UTC(),
	}, nil
}

func (e *EmailVerification) IsExpired() bool {
	return time.Now().UTC().After(e.ExpiresAt)
}

func (e *EmailVerification) IsUsed() bool {
	return e.UsedAt != nil
}

func (e *EmailVerification) MarkUsed() error {
	if e.IsUsed() {
		return ErrVerificationAlreadyUsed
	}

	if e.IsExpired() {
		return ErrVerificationExpired
	}

	now := time.Now().UTC()
	e.UsedAt = &now

	return nil
}
