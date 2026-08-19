package db

import (
	"context"
	"fmt"
	"time"

	"github.com/amantyagi23/authbackend/internal/domain"
	"github.com/amantyagi23/authbackend/internal/repository"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

// UserSessionRepository is the mongodb-backed implementation.
type UserSessionRepository struct {
	collection *mongo.Collection
	log        *zap.Logger
}

// NewUserSessionRepository constructs a MongoDB repository.
func NewUserSessionRepository(client *mongo.Client, dbName string, log *zap.Logger) repository.UserSessionRepository {
	collection := client.Database(dbName).Collection("user_sessions")

	return &UserSessionRepository{
		collection: collection,
		log:        log,
	}
}

// Create inserts a new user.
func (r *UserSessionRepository) Create(ctx context.Context, userSession *domain.UserSession) error {
	if userSession.SessionID == uuid.Nil {
		userSession.SessionID = uuid.New()
	}

	if userSession.CreatedAt.IsZero() {
		userSession.CreatedAt = time.Now()
	}

	_, err := r.collection.InsertOne(ctx, userSession)
	if err != nil {
		return fmt.Errorf("UserSessionRepository.Create: %w", err)
	}

	return nil
}

// // GetByID fetches user by UUID.
// func (r *UserSessionRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
// 	var user domain.User

// 	err := r.collection.FindOne(ctx, bson.M{
// 		"id": id,
// 	}).Decode(&user)

// 	if err != nil {
// 		if errors.Is(err, mongo.ErrNoDocuments) {
// 			return nil, domain.ErrUserNotFound
// 		}
// 		return nil, fmt.Errorf("UserSessionRepository.GetByID: %w", err)
// 	}

// 	return &user, nil
// }

// // GetByEmail fetches user by email.
// func (r *UserSessionRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
// 	var user domain.User

// 	err := r.collection.FindOne(ctx, bson.M{
// 		"email": email,
// 	}).Decode(&user)

// 	if err != nil {
// 		if errors.Is(err, mongo.ErrNoDocuments) {
// 			return nil, domain.ErrUserNotFound
// 		}
// 		return nil, fmt.Errorf("UserSessionRepository.GetByEmail: %w", err)
// 	}

// 	return &user, nil
// }

// // ExistsByEmail checks if email exists.
// func (r *UserSessionRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
// 	count, err := r.collection.CountDocuments(ctx, bson.M{
// 		"email": email,
// 	})
// 	if err != nil {
// 		return false, fmt.Errorf("UserSessionRepository.ExistsByEmail: %w", err)
// 	}

// 	return count > 0, nil
// }

// func isMongoDuplicateKey(err error) bool {
// 	if err == nil {
// 		return false
// 	}

// 	// Mongo duplicate key error code = 11000
// 	var mongoErr mongo.WriteException
// 	if errors.As(err, &mongoErr) {
// 		for _, e := range mongoErr.WriteErrors {
// 			if e.Code == 11000 {
// 				return true
// 			}
// 		}
// 	}

// 	return false
// }
