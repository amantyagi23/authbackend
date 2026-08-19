package config

import (
	"os"
	"time"
)

type Config struct {
	AppName   string
	Port      string
	APIPrefix string

	MongodbURI      string
	DatabaseName    string
	DatabaseTimeOut time.Duration
	FrontendURL     string
	JWTSecret       string
	TokenHeader     string
	LogLevel        string
	RedisHost       string
	RedisPost       string
	RedisDB         int
}

func Load() *Config {
	return &Config{
		AppName:         "Stealth Chat",
		LogLevel:        os.Getenv("LOG_LEVEL"),
		Port:            "4500",
		APIPrefix:       os.Getenv("API_PREFIX"),
		TokenHeader:     "auth_token",
		MongodbURI:      "mongodb://localhost:27017",
		DatabaseName:    "stealthchat",
		DatabaseTimeOut: 10 * time.Second,
		FrontendURL:     os.Getenv("FRONTEND_URL"),
		JWTSecret:       os.Getenv("JWT_SECRET"),
		RedisHost:       os.Getenv("REDIS_HOST"),
		RedisDB:         0,
		RedisPost:       os.Getenv("REDIS_PORT"),
	}
}
