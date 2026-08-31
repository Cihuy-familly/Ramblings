package storage

import (
	"context"
	"log"

	"github.com/redis/go-redis/v9"
)

// NewRedisClient creates a new Redis client and verifies the connection.
func NewRedisClient(redisURL string) *redis.Client {
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		log.Fatalf("Failed to parse Redis URL: %v", err)
	}

	client := redis.NewClient(opt)

	ctx := context.Background()
	_, err = client.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	log.Println("Redis connection established")
	return client
}