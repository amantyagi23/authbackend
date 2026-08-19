package database

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

// Config holds MongoDB settings.
type Config struct {
	URI            string
	MaxPoolSize    uint64
	MinPoolSize    uint64
	ConnectTimeout time.Duration
	SocketTimeout  time.Duration
}

// DefaultConfig returns sane defaults.
func DefaultConfig(uri string) Config {
	return Config{
		URI:            uri,
		MaxPoolSize:    25,
		MinPoolSize:    5,
		ConnectTimeout: 10 * time.Second,
		SocketTimeout:  10 * time.Second,
	}
}

// Connect initializes MongoDB client.
func Connect(ctx context.Context, cfg Config, log *zap.Logger) (*mongo.Client, error) {
	// Apply Mongo options
	clientOptions := options.Client().
		ApplyURI(cfg.URI).
		SetMaxPoolSize(cfg.MaxPoolSize).
		SetMinPoolSize(cfg.MinPoolSize).
		SetConnectTimeout(cfg.ConnectTimeout).
		SetSocketTimeout(cfg.SocketTimeout)

	// Create client
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("database: connect error: %w", err)
	}

	// Ping to verify connection
	pingCtx, cancel := context.WithTimeout(ctx, cfg.ConnectTimeout)
	defer cancel()

	if err := client.Ping(pingCtx, nil); err != nil {
		_ = client.Disconnect(ctx)
		return nil, fmt.Errorf("database: ping failed: %w", err)
	}

	log.Info("MongoDB connected",
		zap.String("uri", cfg.URI),
		zap.Uint64("maxPoolSize", cfg.MaxPoolSize),
	)

	return client, nil
}
