package config

import (
	"os"
)

type Config struct {
	Production   bool
	AppName      string
	Port         string
	APIPrefix    string
	JWTSecret    string
	TokenHeader  string
	LogLevel     string
	RedisHost    string
	RedisPost    string
	RedisDB      int
	PSQLUser     string
	PSQLPassword string
	PSQLDatabase string
	PSQLHost     string
	PSQLPort     string
	BackendUrl   string
}

func Load() *Config {
	return &Config{
		Production:   false,
		AppName:      "Auth Backend",
		LogLevel:     os.Getenv("LOG_LEVEL"),
		Port:         "4500",
		APIPrefix:    os.Getenv("API_PREFIX"),
		TokenHeader:  "auth_token",
		JWTSecret:    os.Getenv("JWT_SECRET"),
		RedisHost:    os.Getenv("REDIS_HOST"),
		RedisDB:      0,
		RedisPost:    os.Getenv("REDIS_PORT"),
		PSQLUser:     "",
		PSQLPassword: "",
		PSQLDatabase: "",
		PSQLHost:     "",
		PSQLPort:     "",
		BackendUrl:   "http://localhost:4500/api/v1/",
	}
}
