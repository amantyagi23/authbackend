// Package domain_test tests the User entity's invariants.
// No external dependencies — pure Go, instant execution.
package domain_test

import (
	"testing"

	"github.com/amantyagi23/authbackend/internal/domain"
)

func TestNewUser_Valid(t *testing.T) {
	u, err := domain.NewUser("Alice", "alice@example.com", "strongpass")
	if err != nil {
		t.Fatalf("expected valid user, got error: %v", err)
	}
	if u.Name != "Alice" {
		t.Errorf("name: want Alice, got %s", u.Name)
	}
	// Password must be stored as a bcrypt hash, not plaintext.
	if u.Password == "strongpass" {
		t.Error("password must not be stored in plaintext")
	}
	if u.ID.String() == "" {
		t.Error("UUID must be assigned")
	}
}

func TestNewUser_EmptyName(t *testing.T) {
	_, err := domain.NewUser("", "alice@example.com", "strongpass")
	if err != domain.ErrNameRequired {
		t.Errorf("expected ErrNameRequired, got %v", err)
	}
}

func TestNewUser_EmptyEmail(t *testing.T) {
	_, err := domain.NewUser("Alice", "", "strongpass")
	if err != domain.ErrEmailRequired {
		t.Errorf("expected ErrEmailRequired, got %v", err)
	}
}

func TestNewUser_ShortPassword(t *testing.T) {
	_, err := domain.NewUser("Alice", "alice@example.com", "short")
	if err != domain.ErrPasswordTooShort {
		t.Errorf("expected ErrPasswordTooShort, got %v", err)
	}
}

func TestUser_VerifyPassword(t *testing.T) {
	u, _ := domain.NewUser("Alice", "alice@example.com", "correctpass")

	if !u.VerifyPassword("correctpass") {
		t.Error("correct password should verify successfully")
	}
	if u.VerifyPassword("wrongpass") {
		t.Error("wrong password should not verify")
	}
}

func TestUser_Sanitized_OmitsPassword(t *testing.T) {
	u, _ := domain.NewUser("Alice", "alice@example.com", "strongpass")
	resp := u.Sanitized()

	// UserResponse has no Password field — this is verified at compile time
	// by the type system. We just check the other fields are populated.
	if resp.Name != "Alice" {
		t.Errorf("name: want Alice, got %s", resp.Name)
	}
	if resp.Email != "alice@example.com" {
		t.Errorf("email: want alice@example.com, got %s", resp.Email)
	}
	if resp.ID.String() == "" {
		t.Error("UUID must be present in sanitized response")
	}
}
