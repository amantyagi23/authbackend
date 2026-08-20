// Domain errors are expressed as typed sentinels.
// The usecase layer wraps them; the delivery layer maps them to HTTP codes.
// This keeps HTTP concerns out of the domain entirely.
package user

import "errors"

var (
	ErrNameRequired             = errors.New("user: name is required")
	ErrEmailRequired            = errors.New("user: email is required")
	ErrPasswordTooShort         = errors.New("user: password must be at least 8 characters")
	ErrUserNotFound             = errors.New("user: not found")
	ErrEmailTaken               = errors.New("user: email already in use")
	ErrNoTokenFound             = errors.New("user session: no token found")
	ErrInvalidVerificationToken = errors.New("Verification Token: no token found")
	ErrVerificationAlreadyUsed  = errors.New("Verification Token: token already used")
	ErrVerificationExpired      = errors.New("Verification Token: token expired")
	ErrVerificationNotFound     = errors.New("Verification Token: no token found")
)
