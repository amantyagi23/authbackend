package db

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/amantyagi23/authbackend/internal/domain"
	"github.com/amantyagi23/authbackend/internal/repository"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// UserSessionRepository is the postgres-backed implementation.
type UserSessionRepository struct {
	db  *pgxpool.Pool
	log *zap.Logger
}

// NewUserSessionRepository constructs a Postgres-backed user session repository.
func NewUserSessionRepository(db *pgxpool.Pool, log *zap.Logger) repository.UserSessionRepository {
	return &UserSessionRepository{
		db:  db,
		log: log,
	}
}

// Create inserts a new user session.
func (r *UserSessionRepository) Create(ctx context.Context, userSession *domain.UserSession) error {
	if userSession.SessionID == uuid.Nil {
		userSession.SessionID = uuid.New()
	}
	if userSession.CreatedAt.IsZero() {
		userSession.CreatedAt = time.Now().UTC()
	}
	if userSession.UpdatedAt.IsZero() {
		userSession.UpdatedAt = userSession.CreatedAt
	}

	_, err := r.db.Exec(
		ctx,
		`
		INSERT INTO user_sessions (
			id,
			user_id,
			access_token,
			refresh_token,
			is_revoked,
			user_agent,
			platform,
			device_id,
			device_name,
			ip_address,
			access_token_expired_at,
			refresh_token_expired_at,
			created_at,
			updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		`,
		userSession.SessionID,
		userSession.UserID,
		userSession.AccessToken,
		userSession.RefreshToken,
		userSession.IsRevoked,
		userSession.UserAgent,
		userSession.Platform,
		userSession.DeviceID,
		userSession.DeviceName,
		nullableIP(userSession.IPAddress),
		userSession.AccessTokenExpiredAt,
		userSession.RefreshTokenExpiredAt,
		userSession.CreatedAt,
		userSession.UpdatedAt,
	)
	if err != nil {
		r.log.Error("UserSessionRepository.Create failed", zap.Error(err))
		return fmt.Errorf("UserSessionRepository.Create: %w", err)
	}

	return nil
}

// GetByID fetches a session by its session ID.
func (r *UserSessionRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.UserSession, error) {
	row := r.db.QueryRow(ctx, selectSessionColumns+` WHERE id = $1 AND is_deleted = FALSE`, id)

	session, err := scanUserSession(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("UserSessionRepository.GetByID: %w", err)
	}

	return session, nil
}

// GetByUserID fetches the most recent active session for a user.
func (r *UserSessionRepository) GetByUserID(ctx context.Context, userID uuid.UUID) (*domain.UserSession, error) {
	row := r.db.QueryRow(
		ctx,
		selectSessionColumns+`
		WHERE user_id = $1 AND is_deleted = FALSE AND is_revoked = FALSE
		ORDER BY created_at DESC
		LIMIT 1`,
		userID,
	)

	session, err := scanUserSession(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("UserSessionRepository.GetByUserID: %w", err)
	}

	return session, nil
}

// GetByRefreshToken fetches a session by its (hashed) refresh token — useful for refresh-token rotation.
func (r *UserSessionRepository) GetByRefreshToken(ctx context.Context, refreshTokenHash string) (*domain.UserSession, error) {
	row := r.db.QueryRow(
		ctx,
		selectSessionColumns+`
		WHERE refresh_token = $1 AND is_deleted = FALSE AND is_revoked = FALSE`,
		refreshTokenHash,
	)

	session, err := scanUserSession(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("UserSessionRepository.GetByRefreshToken: %w", err)
	}

	return session, nil
}

// Revoke marks a session as revoked (soft logout — session row stays for audit).
func (r *UserSessionRepository) Revoke(ctx context.Context, id uuid.UUID) error {
	now := time.Now().UTC()

	cmd, err := r.db.Exec(
		ctx,
		`
		UPDATE user_sessions
		SET is_revoked = TRUE,
		    revoked_at = $2,
		    updated_at = $2
		WHERE id = $1 AND is_deleted = FALSE`,
		id,
		now,
	)
	if err != nil {
		return fmt.Errorf("UserSessionRepository.Revoke: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return domain.ErrUserNotFound
	}

	return nil
}

// Delete soft-deletes a session.
func (r *UserSessionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	now := time.Now().UTC()

	cmd, err := r.db.Exec(
		ctx,
		`
		UPDATE user_sessions
		SET is_deleted = TRUE,
		    deleted_at = $2,
		    updated_at = $2
		WHERE id = $1 AND is_deleted = FALSE`,
		id,
		now,
	)
	if err != nil {
		return fmt.Errorf("UserSessionRepository.Delete: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return domain.ErrUserNotFound
	}

	return nil
}

// TouchLastUsed updates last_used_at, e.g. on each authenticated request.
func (r *UserSessionRepository) TouchLastUsed(ctx context.Context, id uuid.UUID) error {
	now := time.Now().UTC()

	_, err := r.db.Exec(
		ctx,
		`UPDATE user_sessions SET last_used_at = $2, updated_at = $2 WHERE id = $1`,
		id,
		now,
	)
	if err != nil {
		return fmt.Errorf("UserSessionRepository.TouchLastUsed: %w", err)
	}

	return nil
}

const selectSessionColumns = `
SELECT
	id,
	user_id,
	access_token,
	refresh_token,
	access_token_expired_at,
	refresh_token_expired_at,
	is_revoked,
	revoked_at,
	is_deleted,
	deleted_at,
	ip_address,
	user_agent,
	device_id,
	device_name,
	platform,
	last_used_at,
	created_at,
	updated_at
FROM user_sessions
`

// scanUserSession scans a single row into a domain.UserSession.
func scanUserSession(row pgx.Row) (*domain.UserSession, error) {
	var s domain.UserSession
	var ip *string

	err := row.Scan(
		&s.SessionID,
		&s.UserID,
		&s.AccessToken,
		&s.RefreshToken,
		&s.AccessTokenExpiredAt,
		&s.RefreshTokenExpiredAt,
		&s.IsRevoked,
		&s.RevokedAt,
		&s.IsDeleted,
		&s.DeletedAt,
		&ip,
		&s.UserAgent,
		&s.DeviceID,
		&s.DeviceName,
		&s.Platform,
		&s.LastUsedAt,
		&s.CreatedAt,
		&s.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	if ip != nil {
		s.IPAddress = *ip
	}

	return &s, nil
}

// nullableIP returns nil for an empty string so the INET column gets NULL
// instead of failing to parse "" as an address.
func nullableIP(ip string) *string {
	if ip == "" {
		return nil
	}
	return &ip
}
