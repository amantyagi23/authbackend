package user

import (
	"errors"
	"fmt"
	"time"

	"github.com/amantyagi23/authbackend/internal/domain"
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
	uc       usecase.UserUsecase
	usuc     usecase.UserSessionUsecase
	validate *validator.Validate
	log      *zap.Logger
}

// NewUserHandler constructs a UserHandler with injected dependencies.
func NewUserHandler(uc usecase.UserUsecase, usuc usecase.UserSessionUsecase, log *zap.Logger) *UserHandler {
	return &UserHandler{
		uc:       uc,
		usuc:     usuc,
		validate: validator.New(),
		log:      log,
	}
}

// RegisterRoutes attaches user routes to the given Fiber router group.
func (h *UserHandler) RegisterRoutes(router fiber.Router) {
	user := router.Group("/users")
	user.Post("/", h.CreateUser)
	user.Get("/getme", middleware.EnsureAuthentication(h.log), h.GetMe)
	user.Post("/login", h.Login)
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

	_, err := h.uc.CreateUser(c.Context(), usecase.CreateUserInput{
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		return h.mapError(c, err)
	}

	return response.Created(c, "User Created", nil)
}

func (h *UserHandler) GetMe(c *fiber.Ctx) error {
	email := c.Params("id")

	user, err := h.uc.GetUserByEmail(c.Context(), email)
	if err != nil {
		return h.mapError(c, err)
	}

	return response.OK(c, "", user)
}

func (h *UserHandler) Login(c *fiber.Ctx) error {
	var req loginUserRequest

	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "PARSE_ERROR", "invalid request body")
	}

	if err := h.validate.Struct(&req); err != nil {
		return response.BadRequest(c, "VALIDATION_ERROR", err.Error())
	}

	user, err := h.uc.GetUserByEmail(c.Context(), req.Email)
	if err != nil {
		return h.mapError(c, err)
	}

	match := user.VerifyPassword(req.Password)
	if !match {
		return response.Conflict(c, "Invalid Email Or Password")
	}

	accessToken, err := jwt.GenerateToken((user.UserID).String())
	refreshToken := uuid.New().String()

	if err != nil {
		return h.mapError(c, err)
	}

	err = h.usuc.CreateUserSession(c.Context(), usecase.CreateUserSession{
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

// mapError translates domain / usecase errors to HTTP responses.
// This is the ONLY place where domain errors meet HTTP status codes.
func (h *UserHandler) mapError(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, domain.ErrUserNotFound):
		return response.NotFound(c, "user not found")
	case errors.Is(err, domain.ErrEmailTaken):
		return response.Conflict(c, "email is already in use")
	case errors.Is(err, domain.ErrNameRequired),
		errors.Is(err, domain.ErrEmailRequired),
		errors.Is(err, domain.ErrPasswordTooShort):
		return response.BadRequest(c, "DOMAIN_ERROR", err.Error())
	default:
		h.log.Error("unhandled error", zap.Error(err))
		return response.InternalError(c)
	}
}
