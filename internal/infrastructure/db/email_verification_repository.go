package db

import (
	"context"
	"errors"
	"fmt"

	domain "github.com/amantyagi23/authbackend/internal/domain/user"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// EmailVerificationRepository is the pgxdb-backed implementation.
type EmailVerificationRepository struct {
	db  *pgxpool.Pool
	log *zap.Logger
}

// NewUserRepository constructs a pgxDB repository.
func NewEmailVerificationRepository(db *pgxpool.Pool, log *zap.Logger) *EmailVerificationRepository {
	return &EmailVerificationRepository{
		db:  db,
		log: log,
	}
}

func (r *EmailVerificationRepository) Create(
	ctx context.Context,
	verification *domain.EmailVerification,
) error {
	const query = `
		INSERT INTO email_verifications (
			id,
			user_id,
			token_hash,
			expires_at,
			used_at,
			created_at
		)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := r.db.Exec(
		ctx,
		query,
		verification.ID,
		verification.UserID,
		verification.TokenHash,
		verification.ExpiresAt,
		verification.UsedAt,
		verification.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf(
			"EmailVerificationRepository.Create: %w",
			err,
		)
	}

	return nil
}

func (r *EmailVerificationRepository) GetByID(
	ctx context.Context,
	id uuid.UUID,
) (*domain.EmailVerification, error) {
	const query = `
		SELECT
			id,
			user_id,
			token_hash,
			expires_at,
			used_at,
			created_at
		FROM email_verifications
		WHERE id = $1
		LIMIT 1
	`

	var verification domain.EmailVerification

	err := r.db.QueryRow(
		ctx,
		query,
		id,
	).Scan(
		&verification.ID,
		&verification.UserID,
		&verification.TokenHash,
		&verification.ExpiresAt,
		&verification.UsedAt,
		&verification.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrVerificationNotFound
		}

		return nil, fmt.Errorf(
			"EmailVerificationRepository.GetByID: %w",
			err,
		)
	}

	return &verification, nil
}

func (r *EmailVerificationRepository) GetByUserID(
	ctx context.Context,
	userID uuid.UUID,
) (*domain.EmailVerification, error) {
	const query = `
		SELECT
			id,
			user_id,
			token_hash,
			expires_at,
			used_at,
			created_at
		FROM email_verifications
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT 1
	`

	var verification domain.EmailVerification

	err := r.db.QueryRow(
		ctx,
		query,
		userID,
	).Scan(
		&verification.ID,
		&verification.UserID,
		&verification.TokenHash,
		&verification.ExpiresAt,
		&verification.UsedAt,
		&verification.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrVerificationNotFound
		}

		return nil, fmt.Errorf(
			"EmailVerificationRepository.GetByUserID: %w",
			err,
		)
	}

	return &verification, nil
}

func (r *EmailVerificationRepository) GetByTokenHash(
	ctx context.Context,
	tokenHash string,
) (*domain.EmailVerification, error) {
	const query = `
		SELECT
			id,
			user_id,
			token_hash,
			expires_at,
			used_at,
			created_at
		FROM email_verifications
		WHERE token_hash = $1
		LIMIT 1
	`

	var verification domain.EmailVerification

	err := r.db.QueryRow(
		ctx,
		query,
		tokenHash,
	).Scan(
		&verification.ID,
		&verification.UserID,
		&verification.TokenHash,
		&verification.ExpiresAt,
		&verification.UsedAt,
		&verification.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrVerificationNotFound
		}

		return nil, fmt.Errorf(
			"EmailVerificationRepository.GetByTokenHash: %w",
			err,
		)
	}

	return &verification, nil
}

func (r *EmailVerificationRepository) MarkUsed(
	ctx context.Context,
	id uuid.UUID,
) error {
	const query = `
		UPDATE email_verifications
		SET used_at = NOW()
		WHERE id = $1
			AND used_at IS NULL
	`

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf(
			"EmailVerificationRepository.MarkUsed: %w",
			err,
		)
	}

	if result.RowsAffected() == 0 {
		return domain.ErrVerificationAlreadyUsed
	}

	return nil
}

func (r *EmailVerificationRepository) DeleteByUserID(
	ctx context.Context,
	userID uuid.UUID,
) error {
	const query = `
		DELETE FROM email_verifications
		WHERE user_id = $1
	`

	_, err := r.db.Exec(ctx, query, userID)

	if err != nil {
		return fmt.Errorf(
			"EmailVerificationRepository.DeleteByUserID: %w",
			err,
		)
	}

	return nil
}
