package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	fiberlogger "github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/joho/godotenv"
	"go.uber.org/zap"

	"github.com/amantyagi23/authbackend/internal/config"
	"github.com/amantyagi23/authbackend/internal/handler/user"
	"github.com/amantyagi23/authbackend/internal/infrastructure/db"
	"github.com/amantyagi23/authbackend/internal/usecase"
	"github.com/amantyagi23/authbackend/pkg/response"

	"github.com/amantyagi23/authbackend/internal/platform/database"
	"github.com/amantyagi23/authbackend/internal/platform/logger"
	"github.com/amantyagi23/authbackend/internal/platform/redis"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
		return
	}
	cfg := config.Load()

	log, syncLog, err := logger.New(cfg.AppName, cfg.LogLevel, cfg.Production)
	if err != nil {
		panic(err)
	}

	defer syncLog()

	log.Info("Auth Backend Starting",
		zap.String("port", cfg.Port),
	)
	redisClient, err := redis.NewRedisClient(context.Background(), redis.Config{Host: cfg.RedisHost, Port: cfg.RedisPost, Password: "", DB: cfg.RedisDB}, log)
	if err != nil {
		log.Fatal("Redis Connection failed", zap.Error(err))
	}

	defer redisClient.Close()

	psqlDBCfg := database.DefaultConfig("postgresql://postgres:12345678@localhost:5432/authbackend")

	psqlClient, err := database.Connect(context.Background(), psqlDBCfg, log)
	if err != nil {
		log.Fatal("failed to connect database", zap.Error(err))
	}

	defer psqlClient.Close()
	userRepo := db.NewUserRepository(psqlClient, log)
	userSessionRepo := db.NewUserSessionRepository(psqlClient, log)
	uc := usecase.NewUserUsecase(userRepo, log)
	usuc := usecase.NewUserSessionUsecase(userSessionRepo, log)
	userHandler := user.NewUserHandler(uc, usuc, log)

	app := fiber.New(fiber.Config{
		AppName:               cfg.AppName,
		ReadTimeout:           10 * time.Second,
		WriteTimeout:          10 * time.Second,
		IdleTimeout:           30 * time.Second,
		DisableStartupMessage: false,
		// Custom error handler: logs the error and returns a clean JSON body.
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			log.Error("unhandled fiber error", zap.Error(err))
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return c.Status(code).JSON(fiber.Map{
				"success": false,
				"error":   fiber.Map{"code": "FIBER_ERROR", "message": err.Error()},
			})
		},
	})

	app.Use(logger.Middleware(log))
	app.Use(cors.New(cors.Config{
		AllowOrigins:     "http://localhost:5173, http://localhost:3000",
		AllowMethods:     "GET,POST,PUT,DELETE",
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		AllowCredentials: true,
	}))
	// Global middleware
	app.Use(recover.New(recover.Config{
		EnableStackTrace: true,
	})) // recover from panics
	app.Use(requestid.New()) // inject X-Request-ID
	app.Use(fiberlogger.New(fiberlogger.Config{
		Format: "[${time}] ${status} - ${method} ${path} | ${latency} | rid=${locals:requestid}\n",
	}))
	// Routes

	api := app.Group(cfg.APIPrefix)

	api.Get("/", func(c *fiber.Ctx) error {
		return response.OK(c, "Welcome to Auth Apis", nil)
	})
	userHandler.RegisterRoutes(api)

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := app.Listen(":" + cfg.Port); err != nil {
			log.Fatal("server listen error", zap.Error(err))
		}

	}()

	log.Info("user-service ready", zap.String("port", cfg.Port))

	<-quit
	log.Info("shutting down gracefully…")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := app.ShutdownWithContext(shutdownCtx); err != nil {
		log.Error("graceful shutdown failed", zap.Error(err))
	}

	log.Info("user-service stopped")
}
