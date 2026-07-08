package repository

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisRepository implements PubSubRepository, falls back to in-memory brokers
type RedisRepository struct {
	isMemory    bool
	mu          sync.RWMutex
	subscribers map[string][]func(msg string)
	redisClient *redis.Client
}

// NewRedisRepository creates a PubSubRepository instance
func NewRedisRepository(useRedis bool, redisURL string) *RedisRepository {
	repo := &RedisRepository{
		isMemory:    !useRedis,
		subscribers: make(map[string][]func(msg string)),
	}

	if useRedis {
		opt, err := redis.ParseURL(redisURL)
		if err != nil {
			log.Printf("[REDIS] Failed to parse URL %s: %v. Falling back to in-memory mode.", redisURL, err)
			repo.isMemory = true
			return repo
		}

		rdb := redis.NewClient(opt)

		// Ping Redis to verify connectivity
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := rdb.Ping(ctx).Err(); err != nil {
			log.Printf("[REDIS] Failed to ping Redis at %s: %v. Falling back to in-memory mode.", opt.Addr, err)
			repo.isMemory = true
		} else {
			repo.redisClient = rdb
			log.Printf("[REDIS] Connected to Redis server at %s", opt.Addr)
		}
	}

	return repo
}

// Publish sends message to channel
func (r *RedisRepository) Publish(ctx context.Context, channel string, message interface{}) error {
	var msgStr string
	switch v := message.(type) {
	case string:
		msgStr = v
	default:
		bytes, err := json.Marshal(v)
		if err != nil {
			return err
		}
		msgStr = string(bytes)
	}

	if r.isMemory {
		r.mu.RLock()
		handlers, exists := r.subscribers[channel]
		r.mu.RUnlock()

		if !exists {
			return nil
		}

		for _, handler := range handlers {
			h := handler
			go h(msgStr)
		}
		return nil
	}

	return r.redisClient.Publish(ctx, channel, msgStr).Err()
}

// Subscribe listens to channel
func (r *RedisRepository) Subscribe(ctx context.Context, channel string, handler func(msg string)) error {
	if r.isMemory {
		r.mu.Lock()
		defer r.mu.Unlock()
		r.subscribers[channel] = append(r.subscribers[channel], handler)
		return nil
	}

	pubsub := r.redisClient.Subscribe(ctx, channel)

	go func() {
		defer pubsub.Close()
		ch := pubsub.Channel()
		for msg := range ch {
			handler(msg.Payload)
		}
	}()

	return nil
}

// IncrLimit increments a rate limit key in Redis (or returns 0, nil to trigger local memory fallback)
func (r *RedisRepository) IncrLimit(ctx context.Context, key string, expiration time.Duration) (int64, error) {
	if r.isMemory {
		return 0, nil
	}

	pipe := r.redisClient.Pipeline()
	incr := pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, expiration)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return 0, err
	}

	return incr.Val(), nil
}
