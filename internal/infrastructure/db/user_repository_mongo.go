package db

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/amantyagi23/authbackend/internal/domain"
	"github.com/amantyagi23/authbackend/internal/repository"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

// UserRepository is the mongodb-backed implementation.
type UserRepository struct {
	collection *mongo.Collection
	log        *zap.Logger
}

// NewUserRepository constructs a MongoDB repository.
func NewUserRepository(client *mongo.Client, dbName string, log *zap.Logger) repository.UserRepository {
	collection := client.Database(dbName).Collection("users")

	return &UserRepository{
		collection: collection,
		log:        log,
	}
}

// Create inserts a new user.
func (r *UserRepository) Create(ctx context.Context, user *domain.User) error {
	if user.ID == uuid.Nil {
		user.ID = uuid.New()
	}

	if user.CreatedAt.IsZero() {
		user.CreatedAt = time.Now()
	}

	_, err := r.collection.InsertOne(ctx, user)
	if err != nil {
		if isMongoDuplicateKey(err) {
			return domain.ErrEmailTaken
		}
		return fmt.Errorf("UserRepository.Create: %w", err)
	}

	return nil
}

// GetByID fetches user by UUID.
func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	var user domain.User

	err := r.collection.FindOne(ctx, bson.M{
		"id": id,
	}).Decode(&user)

	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("UserRepository.GetByID: %w", err)
	}

	return &user, nil
}

// GetByEmail fetches user by email.
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	var user domain.User

	err := r.collection.FindOne(ctx, bson.M{
		"email": email,
	}).Decode(&user)

	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("UserRepository.GetByEmail: %w", err)
	}

	return &user, nil
}

// ExistsByEmail checks if email exists.
func (r *UserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	count, err := r.collection.CountDocuments(ctx, bson.M{
		"email": email,
	})
	if err != nil {
		return false, fmt.Errorf("UserRepository.ExistsByEmail: %w", err)
	}

	return count > 0, nil
}

func isMongoDuplicateKey(err error) bool {
	if err == nil {
		return false
	}

	// Mongo duplicate key error code = 11000
	var mongoErr mongo.WriteException
	if errors.As(err, &mongoErr) {
		for _, e := range mongoErr.WriteErrors {
			if e.Code == 11000 {
				return true
			}
		}
	}

	return false
}
