package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// RedisClient wrapper for Redis operations
type RedisClient struct {
	client *redis.Client
	ctx    context.Context
}

// Config holds Redis configuration
type Config struct {
	Host     string
	Port     string
	Password string
	DB       int
}

// NewRedisClient creates and initializes a new Redis client
func NewRedisClient(ctx context.Context, config Config, log *zap.Logger) (*RedisClient, error) {

	client := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%s", config.Host, config.Port),
		Password:     config.Password,
		DB:           config.DB,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		PoolSize:     10,
		MinIdleConns: 5,
	})

	// Test the connection
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	log.Info("Redis connected",
		zap.String("Host", config.Host),
		zap.String("Post", config.Port),
	)
	return &RedisClient{
		client: client,
		ctx:    ctx,
	}, nil
}

// Close closes the Redis connection
func (r *RedisClient) Close() error {
	return r.client.Close()
}

// Set sets a key-value pair with expiration
func (r *RedisClient) Set(key string, value interface{}, expiration time.Duration) error {
	return r.client.Set(r.ctx, key, value, expiration).Err()
}

// Get gets a value by key
func (r *RedisClient) Get(key string) (string, error) {
	return r.client.Get(r.ctx, key).Result()
}

// SetJSON stores a JSON object
func (r *RedisClient) SetJSON(key string, data interface{}, expiration time.Duration) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	return r.Set(key, jsonData, expiration)
}

// GetJSON retrieves and unmarshals a JSON object
func (r *RedisClient) GetJSON(key string, dest interface{}) error {
	data, err := r.Get(key)
	if err != nil {
		return err
	}

	if err := json.Unmarshal([]byte(data), dest); err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %w", err)
	}
	return nil
}

// Delete removes one or more keys
func (r *RedisClient) Delete(keys ...string) error {
	return r.client.Del(r.ctx, keys...).Err()
}

// Exists checks if a key exists
func (r *RedisClient) Exists(key string) (bool, error) {
	result, err := r.client.Exists(r.ctx, key).Result()
	return result > 0, err
}

// Expire sets expiration time for a key
func (r *RedisClient) Expire(key string, expiration time.Duration) error {
	return r.client.Expire(r.ctx, key, expiration).Err()
}

// Increment increments a counter
func (r *RedisClient) Increment(key string) (int64, error) {
	return r.client.Incr(r.ctx, key).Result()
}

// LPush pushes to the left of a list
func (r *RedisClient) LPush(key string, values ...interface{}) error {
	return r.client.LPush(r.ctx, key, values...).Err()
}

// RPop pops from the right of a list
func (r *RedisClient) RPop(key string) (string, error) {
	return r.client.RPop(r.ctx, key).Result()
}

// GetAllKeys returns all keys matching a pattern
func (r *RedisClient) GetAllKeys(pattern string) ([]string, error) {
	return r.client.Keys(r.ctx, pattern).Result()
}

// FlushAll removes all keys
func (r *RedisClient) FlushAll() error {
	return r.client.FlushAll(r.ctx).Err()
}

// Example usage
// func main() {
// 	// Redis configuration
// 	config := Config{
// 		Host:     "localhost",
// 		Port:     "6379",
// 		Password: "", // Use password if Redis requires authentication
// 		DB:       0,  // Default database
// 	}

// 	// Initialize Redis client
// 	redisClient, err := NewRedisClient(config)
// 	if err != nil {
// 		log.Fatal("Failed to initialize Redis:", err)
// 	}
// 	defer redisClient.Close()

// 	// Example 1: Basic string operations
// 	fmt.Println("\n--- String Operations ---")

// 	// Set a value
// 	err = redisClient.Set("username", "john_doe", 10*time.Minute)
// 	if err != nil {
// 		log.Printf("Error setting value: %v", err)
// 	}

// 	// Get a value
// 	value, err := redisClient.Get("username")
// 	if err != nil {
// 		log.Printf("Error getting value: %v", err)
// 	} else {
// 		fmt.Printf("Got username: %s\n", value)
// 	}

// 	// Example 2: JSON operations
// 	fmt.Println("\n--- JSON Operations ---")

// 	type User struct {
// 		ID    int    `json:"id"`
// 		Name  string `json:"name"`
// 		Email string `json:"email"`
// 	}

// 	user := User{
// 		ID:    1,
// 		Name:  "Alice Smith",
// 		Email: "alice@example.com",
// 	}

// 	// Store JSON
// 	err = redisClient.SetJSON("user:1", user, 5*time.Minute)
// 	if err != nil {
// 		log.Printf("Error storing JSON: %v", err)
// 	}

// 	// Retrieve JSON
// 	var retrievedUser User
// 	err = redisClient.GetJSON("user:1", &retrievedUser)
// 	if err != nil {
// 		log.Printf("Error retrieving JSON: %v", err)
// 	} else {
// 		fmt.Printf("Retrieved user: %+v\n", retrievedUser)
// 	}

// 	// Example 3: Counter operations
// 	fmt.Println("\n--- Counter Operations ---")

// 	for i := 0; i < 5; i++ {
// 		count, err := redisClient.Increment("page_views")
// 		if err != nil {
// 			log.Printf("Error incrementing counter: %v", err)
// 		} else {
// 			fmt.Printf("Page views: %d\n", count)
// 		}
// 		time.Sleep(100 * time.Millisecond)
// 	}

// 	// Example 4: List operations
// 	fmt.Println("\n--- List Operations ---")

// 	// Add items to list
// 	err = redisClient.LPush("tasks", "task1", "task2", "task3")
// 	if err != nil {
// 		log.Printf("Error pushing to list: %v", err)
// 	}

// 	// Pop item from list
// 	task, err := redisClient.RPop("tasks")
// 	if err != nil {
// 		log.Printf("Error popping from list: %v", err)
// 	} else {
// 		fmt.Printf("Popped task: %s\n", task)
// 	}

// 	// Example 5: Check existence and expiration
// 	fmt.Println("\n--- Utility Operations ---")

// 	exists, err := redisClient.Exists("username")
// 	if err != nil {
// 		log.Printf("Error checking existence: %v", err)
// 	} else {
// 		fmt.Printf("Key 'username' exists: %t\n", exists)
// 	}

// 	// Get all keys
// 	keys, err := redisClient.GetAllKeys("*")
// 	if err != nil {
// 		log.Printf("Error getting keys: %v", err)
// 	} else {
// 		fmt.Printf("All keys: %v\n", keys)
// 	}

// 	// Example 6: Clean up
// 	fmt.Println("\n--- Clean Up ---")

// 	// Delete specific keys
// 	err = redisClient.Delete("username", "page_views")
// 	if err != nil {
// 		log.Printf("Error deleting keys: %v", err)
// 	}

// 	fmt.Println("Demo completed successfully!")
// }
