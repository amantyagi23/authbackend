package user

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type UserSession struct {
	SessionID uuid.UUID
	UserID    uuid.UUID

	// Store hashes rather than raw tokens in the database.
	AccessToken  string
	RefreshToken string

	AccessTokenExpiredAt  time.Time
	RefreshTokenExpiredAt time.Time

	// Session lifecycle
	IsRevoked bool
	RevokedAt *time.Time
	IsDeleted bool
	DeletedAt *time.Time

	// Security / client information
	IPAddress string
	UserAgent string

	// Optional device information
	DeviceID   *string
	DeviceName *string
	Platform   *string

	// Session activity
	LastUsedAt *time.Time

	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewSession(
	SessionID uuid.UUID,
	userID uuid.UUID,
	accessTokenHash string,
	refreshTokenHash string,
	accessTokenExpiredAt time.Time,
	refreshTokenExpiredAt time.Time,
	ipAddress string,
	userAgent string,
	deviceID string,
	deviceName string,
	platform string,
) (*UserSession, error) {

	if SessionID == uuid.Nil {
		return nil, errors.New("Session Id Not Found")
	}
	if userID == uuid.Nil {
		return nil, ErrUserNotFound
	}

	if accessTokenHash == "" {
		return nil, ErrNoTokenFound
	}

	if refreshTokenHash == "" {
		return nil, ErrNoTokenFound
	}

	now := time.Now().UTC()

	return &UserSession{
		SessionID: SessionID,
		UserID:    userID,

		AccessToken:  accessTokenHash,
		RefreshToken: refreshTokenHash,

		AccessTokenExpiredAt:  accessTokenExpiredAt,
		RefreshTokenExpiredAt: refreshTokenExpiredAt,

		IsRevoked: false,
		RevokedAt: nil,

		IsDeleted: false,
		DeletedAt: nil,

		IPAddress: ipAddress,
		UserAgent: userAgent,

		DeviceID:   &deviceID,
		DeviceName: &deviceName,
		Platform:   &platform,

		LastUsedAt: &now,

		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}
