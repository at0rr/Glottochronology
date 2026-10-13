package redis

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
)

const servicePrefix = "glotto."

type Client struct {
	client *redis.Client
}

func New(ctx context.Context) (*Client, error) {
	port, _ := strconv.Atoi(os.Getenv("REDIS_PORT"))

	rc := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_HOST") + ":" + strconv.Itoa(port),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0,
	})

	if _, err := rc.Ping(ctx).Result(); err != nil {
		return nil, fmt.Errorf("cant ping redis: %w", err)
	}

	return &Client{client: rc}, nil
}

func (c *Client) Close() error {
	return c.client.Close()
}

// --- blacklist ---

func jwtKey(token string) string {
	return servicePrefix + "jwt.blacklist." + token
}

func (c *Client) WriteJWTToBlacklist(ctx context.Context, jwtStr string, ttl time.Duration) error {
	return c.client.Set(ctx, jwtKey(jwtStr), true, ttl).Err()
}

func (c *Client) CheckJWTInBlacklist(ctx context.Context, jwtStr string) (bool, error) {
	n, err := c.client.Exists(ctx, jwtKey(jwtStr)).Result()
	if err != nil {
		return false, err
	}
	return n > 0, nil
}
