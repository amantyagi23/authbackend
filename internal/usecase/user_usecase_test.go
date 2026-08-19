// Package usecase_test contains unit tests for the user usecase.
// The PostgreSQL repository is replaced with an in-memory fake so tests
// run fast and don't need a live database.
package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/amantyagi23/authbackend/internal/domain"
	"github.com/amantyagi23/authbackend/internal/usecase"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// ── In-memory fake repository ─────────────────────────────────────────────────

type fakeUserRepo struct {
	users map[uuid.UUID]*domain.User
}

func newFakeRepo() *fakeUserRepo {
	return &fakeUserRepo{users: make(map[uuid.UUID]*domain.User)}
}

func (r *fakeUserRepo) Create(_ context.Context, u *domain.User) error {
	for _, existing := range r.users {
		if existing.Email == u.Email {
			return domain.ErrEmailTaken
		}
	}
	r.users[u.UserID] = u
	return nil
}

func (r *fakeUserRepo) GetByID(_ context.Context, id uuid.UUID) (*domain.User, error) {
	u, ok := r.users[id]
	if !ok {
		return nil, domain.ErrUserNotFound
	}
	return u, nil
}

func (r *fakeUserRepo) GetByEmail(_ context.Context, email string) (*domain.User, error) {
	for _, u := range r.users {
		if u.Email == email {
			return u, nil
		}
	}
	return nil, domain.ErrUserNotFound
}

func (r *fakeUserRepo) ExistsByEmail(_ context.Context, email string) (bool, error) {
	for _, u := range r.users {
		if u.Email == email {
			return true, nil
		}
	}
	return false, nil
}

// ── Helper ────────────────────────────────────────────────────────────────────

func newUC() (usecase.UserUsecase, *fakeUserRepo) {
	repo := newFakeRepo()
	log := zap.NewNop()
	return usecase.NewUserUsecase(repo, log), repo
}

// ── Tests ─────────────────────────────────────────────────────────────────────

func TestCreateUser_Success(t *testing.T) {
	uc, _ := newUC()

	resp, err := uc.CreateUser(context.Background(), usecase.CreateUserInput{
		Name:     "Alice",
		Email:    "alice@example.com",
		Password: "supersecret",
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.Email != "alice@example.com" {
		t.Errorf("expected email alice@example.com, got %s", resp.Email)
	}
	if resp.UserID == uuid.Nil {
		t.Error("expected a non-nil UUID")
	}
}

func TestCreateUser_DuplicateEmail(t *testing.T) {
	uc, _ := newUC()
	input := usecase.CreateUserInput{
		Name:     "Alice",
		Email:    "alice@example.com",
		Password: "supersecret",
	}

	if _, err := uc.CreateUser(context.Background(), input); err != nil {
		t.Fatalf("first create should succeed, got %v", err)
	}

	_, err := uc.CreateUser(context.Background(), input)
	if !errors.Is(err, domain.ErrEmailTaken) {
		t.Errorf("expected ErrEmailTaken, got %v", err)
	}
}

func TestGetUser_Success(t *testing.T) {
	uc, _ := newUC()

	created, err := uc.CreateUser(context.Background(), usecase.CreateUserInput{
		Name:     "Bob",
		Email:    "bob@example.com",
		Password: "password123",
	})
	if err != nil {
		t.Fatal(err)
	}

	fetched, err := uc.GetUser(context.Background(), created.UserID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if fetched.UserID != created.UserID {
		t.Errorf("UserID mismatch: want %s got %s", created.UserID, fetched.UserID)
	}
}

func TestGetUser_NotFound(t *testing.T) {
	uc, _ := newUC()

	_, err := uc.GetUser(context.Background(), uuid.New())
	if !errors.Is(err, domain.ErrUserNotFound) {
		t.Errorf("expected ErrUserNotFound, got %v", err)
	}
}

func TestCreateUser_PasswordTooShort(t *testing.T) {
	uc, _ := newUC()

	_, err := uc.CreateUser(context.Background(), usecase.CreateUserInput{
		Name:     "Carol",
		Email:    "carol@example.com",
		Password: "short",
	})

	if !errors.Is(err, domain.ErrPasswordTooShort) {
		t.Errorf("expected ErrPasswordTooShort, got %v", err)
	}
}
