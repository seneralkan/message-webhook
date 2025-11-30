package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type IRedisInstance interface {
	Client() *redis.Client
	Close() error
	Ping(ctx context.Context) error
}

type redisInstance struct {
	client *redis.Client
}

func NewRedisInstance(host, port, password string, db int) (IRedisInstance, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", host, port),
		Password: password,
		DB:       db,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	return &redisInstance{client: client}, nil
}

func (r *redisInstance) Client() *redis.Client {
	return r.client
}

func (r *redisInstance) Close() error {
	return r.client.Close()
}

func (r *redisInstance) Ping(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}
