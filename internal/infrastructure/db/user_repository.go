package db

import (
	"context"
	"errors"
	"fmt"
	"time"

	domain "github.com/amantyagi23/authbackend/internal/domain/user"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// UserRepository is the pgxdb-backed implementation.
type UserRepository struct {
	db  *pgxpool.Pool
	log *zap.Logger
}

// NewUserRepository constructs a pgxDB repository.
func NewUserRepository(db *pgxpool.Pool, log *zap.Logger) *UserRepository {
	return &UserRepository{
		db:  db,
		log: log,
	}
}

// Create inserts a new user.
func (r *UserRepository) Create(ctx context.Context, user *domain.User) error {
	if user.UserID == uuid.Nil {
		user.UserID = uuid.New()
	}

	if user.CreatedAt.IsZero() {
		user.CreatedAt = time.Now()
	}

	err := r.db.QueryRow(
		ctx,
		`
		INSERT INTO users (
			user_id,
			name,
			email,
			password
		)
		VALUES ($1, $2, $3, $4)
		RETURNING
			user_id,
			created_at,
			updated_at
		`,
		user.UserID,
		user.Name,
		user.Email,
		user.Password,
	).Scan(
		&user.UserID,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		fmt.Print(err.Error())
		if isPGXDuplicateKey(err) {
			return domain.ErrEmailTaken
		}
		return fmt.Errorf("UserRepository.Create: %w", err)
	}

	return nil
}

func (r *UserRepository) Update(
	ctx context.Context,
	user *domain.User,
) error {
	const query = `
		UPDATE users
		SET
			name = $2,
			email = $3,
			password = $4,
			profile_pic = $5,
			is_email_verified = $6,
			is_active = $7,
			updated_at = NOW()
		WHERE user_id = $1
		RETURNING
			user_id,
			name,
			email,
			password,
			profile_pic,
			is_email_verified,
			is_active,
			created_at,
			updated_at
	`

	err := r.db.QueryRow(
		ctx,
		query,
		user.UserID,
		user.Name,
		user.Email,
		user.Password,
		user.ProfilePic,
		user.IsEmailVerified,
		user.IsActive,
	).Scan(
		&user.UserID,
		&user.Name,
		&user.Email,
		&user.Password,
		&user.ProfilePic,
		&user.IsEmailVerified,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrUserNotFound
		}

		if isPGXDuplicateKey(err) {
			return domain.ErrEmailTaken
		}

		return fmt.Errorf("UserRepository.Update: %w", err)
	}

	return nil
}

// GetByID fetches user by UUID.
func (repo *UserRepository) GetByID(ctx context.Context, userId uuid.UUID) (*domain.User, error) {

	const query = `
		SELECT
			*
		FROM users
		WHERE user_id = $1
		LIMIT 1
	`

	var user domain.User

	err := repo.db.QueryRow(ctx, query, userId).Scan(
		&user.UserID,
		&user.Name,
		&user.Email,
		&user.Password,
		&user.ProfilePic,
		&user.IsEmailVerified,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}

		return nil, err
	}

	return &user, nil
}

// GetByEmail fetches user by email.
func (repo *UserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {

	const query = `
		SELECT
			*
		FROM users
		WHERE email = $1
		LIMIT 1
	`

	var user domain.User

	err := repo.db.QueryRow(ctx, query, email).Scan(
		&user.UserID,
		&user.Name,
		&user.Email,
		&user.Password,
		&user.ProfilePic,
		&user.IsEmailVerified,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}

		return nil, err
	}

	return &user, nil

}

// ExistsByEmail checks if email exists.
func (repo *UserRepository) ExistsByEmail(
	ctx context.Context,
	email string,
) (bool, error) {
	const query = `
		SELECT EXISTS (
			SELECT 1
			FROM users
			WHERE email = $1
		)
	`

	var exists bool

	err := repo.db.QueryRow(ctx, query, email).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("UserRepository.ExistsByEmail: %w", err)
	}

	return exists, nil
}

func isPGXDuplicateKey(err error) bool {
	if err == nil {
		return false
	}

	var pgErr *pgconn.PgError

	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}

	return false
}
