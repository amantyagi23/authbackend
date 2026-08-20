package user

import (
	"errors"
	"fmt"
	"path/filepath"
	"time"

	"github.com/amantyagi23/authbackend/internal/domain/user"
	"github.com/amantyagi23/authbackend/internal/middleware"
	"github.com/amantyagi23/authbackend/internal/usecase"
	"github.com/amantyagi23/authbackend/pkg/jwt"
	"github.com/amantyagi23/authbackend/pkg/response"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// UserHandler holds the handler methods for user routes.
type UserHandler struct {
	userUsecase              usecase.UserUsecase
	userSessionUsecase       usecase.UserSessionUsecase
	emailVerificationUsecase usecase.EmailVerificationUsecase
	validate                 *validator.Validate
	log                      *zap.Logger
}

// NewUserHandler constructs a UserHandler with injected dependencies.
func NewUserHandler(userUsecase usecase.UserUsecase, userSessionUsecase usecase.UserSessionUsecase, emailVerificationUsecase usecase.EmailVerificationUsecase, log *zap.Logger) *UserHandler {
	return &UserHandler{
		userUsecase:              userUsecase,
		userSessionUsecase:       userSessionUsecase,
		emailVerificationUsecase: emailVerificationUsecase,
		validate:                 validator.New(),
		log:                      log,
	}
}

// RegisterRoutes attaches user routes to the given Fiber router group.
func (h *UserHandler) RegisterRoutes(router fiber.Router) {
	user := router.Group("/users")
	user.Post("/", h.CreateUser)
	user.Get("/getme", middleware.EnsureAuthentication(h.log), h.GetMe)
	user.Post("/login", h.Login)
	user.Get("logout", middleware.EnsureAuthentication(h.log), h.Logout)
	user.Get("/refresh", h.RefreshSession)
	user.Post("/change-password", middleware.EnsureAuthentication(h.log), h.ChangePassword)
	user.Post("/upload-image", middleware.EnsureAuthentication(h.log), h.UploadProfileImage)
	user.Get("/verify-email", h.VerifyEmail)
}

// createUserRequest is the HTTP request body shape for user creation.
type createUserRequest struct {
	Name     string `json:"fullName"     validate:"required,min=2,max=100"`
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

type loginUserRequest struct {
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

type ChangePasswordRequest struct {
	OldPassword     string `json:"oldPassword" validate:"required,min=8"`
	NewPassword     string `json:"newpassword" validate:"required,min=8"`
	ConfirmPassword string `json:"confirmPassword" validate:"required,min=8"`
}

// CreateUser godoc
// POST /api/v1/users
// Creates a new user account. Returns 201 on success.
func (h *UserHandler) CreateUser(c *fiber.Ctx) error {
	var req createUserRequest
	fmt.Println(req.Email)
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "PARSE_ERROR", "invalid request body")
	}

	if err := h.validate.Struct(&req); err != nil {
		return response.BadRequest(c, "VALIDATION_ERROR", err.Error())
	}

	user, err := h.userUsecase.CreateUser(c.Context(), usecase.CreateUserInput{
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
	})

	if err != nil {
		return h.mapError(c, err)
	}

	_, token, err := h.emailVerificationUsecase.Create(c.Context(), usecase.CreateEmailVerificationInput{
		UserID: user.UserID,
	})

	if err != nil {
		return h.mapError(c, err)
	}

	fmt.Println(token)

	return response.Created(c, "User Created", nil)
}

func (h *UserHandler) VerifyEmail(c *fiber.Ctx) error {
	token := c.Queries()["token"]
	if token == "" {
		return response.NotFound(c, "No Token Found")
	}

	emailVerification, err := h.emailVerificationUsecase.Verify(c.Context(), token)
	if err != nil {
		fmt.Println(err)
		return h.mapError(c, err)
	}
	user, err := h.userUsecase.GetUser(c.Context(), emailVerification.UserID)
	if err != nil {
		return h.mapError(c, err)
	}

	user.SetEmailVerification(true)

	err = h.userUsecase.UpdateUser(c.Context(), user)
	if err != nil {
		return h.mapError(c, err)
	}
	return response.OK(c, "Email Verified", nil)
}

func (h *UserHandler) GetMe(c *fiber.Ctx) error {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		return response.InternalError(c)
	}
	user, err := h.userUsecase.GetUser(c.Context(), userID)
	if err != nil {
		return h.mapError(c, err)
	}

	return response.OK(c, "", user.Sanitized())
}

func (h *UserHandler) Login(c *fiber.Ctx) error {
	var req loginUserRequest

	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "PARSE_ERROR", "invalid request body")
	}

	if err := h.validate.Struct(&req); err != nil {
		return response.BadRequest(c, "VALIDATION_ERROR", err.Error())
	}

	user, err := h.userUsecase.GetUserByEmail(c.Context(), req.Email)
	if err != nil {
		return h.mapError(c, err)
	}

	if user.IsEmailVerified == false {
		return response.Conflict(c, "Email Not Verified")
	}

	match := user.VerifyPassword(req.Password)
	if !match {
		return response.Conflict(c, "Invalid Email Or Password")
	}

	sessionId := uuid.New()
	accessToken, err := jwt.GenerateAccessToken(user.UserID, sessionId)

	refreshToken := uuid.New().String()

	if err != nil {
		return h.mapError(c, err)
	}

	err = h.userSessionUsecase.CreateUserSession(c.Context(), usecase.CreateUserSession{
		SessionId:             sessionId,
		UserId:                user.UserID,
		AccessToken:           accessToken,
		RefreshToken:          refreshToken,
		AccessTokenExpiredAt:  time.Now().Add(3600 * time.Second),
		RefreshTokenExpiredAt: time.Now().Add(3600 * time.Second),
	})

	if err != nil {
		return h.mapError(c, err)
	}

	c.Cookie(&fiber.Cookie{
		Name:     "auth_token",
		Value:    accessToken,
		HTTPOnly: true,
		Secure:   false, // true in production (HTTPS)
		SameSite: "Lax",
		MaxAge:   3600, // 1 hour
	})

	c.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		HTTPOnly: true,
		Secure:   false, // true in production (HTTPS)
		SameSite: "Lax",
		MaxAge:   3600, // 1 hour
	})

	return response.OK(c, "Login Successfully", nil)

}

func (h *UserHandler) RefreshSession(c *fiber.Ctx) error {
	token := c.Cookies("refresh_token")

	if token == "" {
		return response.Unauthorized(c, "Token Not Found")
	}

	userSession, err := h.userSessionUsecase.GetUserSessionByRefreshToken(c.Context(), token)
	if err != nil {
		return h.mapError(c, err)

	}
	accessToken, err := jwt.GenerateAccessToken(userSession.UserID, userSession.SessionID)

	err = h.userSessionUsecase.UpdateAccessToken(c.Context(), userSession.SessionID, accessToken)
	if err != nil {
		return h.mapError(c, err)

	}
	c.Cookie(&fiber.Cookie{
		Name:     "auth_token",
		Value:    accessToken,
		HTTPOnly: true,
		Secure:   false, // true in production (HTTPS)
		SameSite: "Lax",
		MaxAge:   3600, // 1 hour
	})

	return response.OK(c, "Session Refreshed", nil)
}

func (h *UserHandler) Logout(c *fiber.Ctx) error {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		return response.InternalError(c)
	}

	user, err := h.userUsecase.GetUser(c.Context(), userID)

	if err != nil {
		return h.mapError(c, err)

	}

	h.userSessionUsecase.DeactivateUserSession(c.Context(), user.UserID)

	c.Cookie(&fiber.Cookie{
		Name:     "auth_token",
		Value:    "",
		HTTPOnly: true,
		Secure:   false, // true in production (HTTPS)
		SameSite: "Lax",
		MaxAge:   -1,
		Expires:  time.Now().Add(-1 * time.Hour),
	})

	c.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    "refreshToken",
		Expires:  time.Now().Add(-1 * time.Hour),
		HTTPOnly: true,
		Secure:   false, // true in production (HTTPS)
		SameSite: "Lax",
		MaxAge:   -1,
	})

	return response.OK(c, "Logged Out Successfully", nil)
}

func (h *UserHandler) ChangePassword(c *fiber.Ctx) error {
	var req ChangePasswordRequest

	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		return response.InternalError(c)
	}

	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(
			c,
			"PARSE_ERROR",
			"invalid request body",
		)
	}

	if err := h.validate.Struct(&req); err != nil {
		return response.BadRequest(
			c,
			"VALIDATION_ERROR",
			err.Error(),
		)
	}

	if req.NewPassword != req.ConfirmPassword {
		return response.BadRequest(
			c,
			"PASSWORD_MISMATCH",
			"new password and confirm password do not match",
		)
	}

	if req.OldPassword == req.NewPassword {
		return response.BadRequest(
			c,
			"SAME_PASSWORD",
			"new password must be different from old password",
		)
	}

	user, err := h.userUsecase.GetUser(c.Context(), userID)
	if err != nil {
		return h.mapError(c, err)
	}

	if !user.VerifyPassword(req.OldPassword) {
		return response.BadRequest(
			c,
			"INVALID_PASSWORD",
			"old password is incorrect",
		)
	}

	err = user.SetPassword(req.NewPassword)
	if err != nil {
		return h.mapError(c, err)
	}

	err = h.userUsecase.UpdateUser(c.Context(), user)
	if err != nil {
		return h.mapError(c, err)
	}

	return response.OK(
		c,
		"Password changed successfully",
		nil,
	)
}

func (h *UserHandler) UploadProfileImage(c *fiber.Ctx) error {
	userId, ok := middleware.UserIDFromContext(c)

	if !ok {
		return response.InternalError(c)
	}

	user, err := h.userUsecase.GetUser(c.Context(), userId)

	if err != nil {
		return h.mapError(c, err)
	}

	file, err := c.FormFile("image")
	if err != nil {
		h.log.Error("failed to get uploaded image",
			zap.Error(err),
		)
		return response.BadRequest(c, "400", "Invalid Image")
	}
	filename := uuid.NewString() + filepath.Ext(file.Filename)

	path := filepath.Join("uploads", filename)

	err = user.SetProfilePic(path)
	if err != nil {
		return h.mapError(c, err)
	}

	// Save locally
	err = c.SaveFile(file, path)
	if err != nil {
		h.log.Error("failed to save uploaded image",
			zap.Error(err),
		)
		return response.InternalError(c)
	}

	err = h.userUsecase.UpdateUser(c.Context(), user)
	if err != nil {
		return h.mapError(c, err)
	}

	return response.OK(
		c,
		"Image uploaded successfully",
		nil,
	)
}

// mapError translates user / usecase errors to HTTP responses.
// This is the ONLY place where user errors meet HTTP status codes.
func (h *UserHandler) mapError(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, user.ErrUserNotFound):
		return response.NotFound(c, "user not found")
	case errors.Is(err, user.ErrEmailTaken):
		return response.Conflict(c, "email is already in use")
	case errors.Is(err, user.ErrNameRequired),
		errors.Is(err, user.ErrEmailRequired),
		errors.Is(err, user.ErrPasswordTooShort):
		return response.BadRequest(c, "DOMAIN_ERROR", err.Error())
	default:
		h.log.Error("unhandled error", zap.Error(err))
		return response.Conflict(c, err.Error())
	}
}
