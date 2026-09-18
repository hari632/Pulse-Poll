package database

import (
	"context"
	"crypto/tls"
	"os"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisDatabase struct {
	Client *redis.Client
}

// ConnectRedis connects to Redis via URL or host:port with optional TLS support.
func ConnectRedis(addrOrURL string, password string) (*RedisDatabase, error) {
	var opts *redis.Options

	trimmed := strings.TrimSpace(addrOrURL)

	// Check if full URL (e.g. redis:// or rediss:// for TLS cloud instances)
	if strings.HasPrefix(trimmed, "redis://") || strings.HasPrefix(trimmed, "rediss://") {
		parsed, err := redis.ParseURL(trimmed)
		if err != nil {
			return nil, err
		}
		opts = parsed
	} else {
		user := os.Getenv("REDIS_USER")
		useTLS := strings.EqualFold(os.Getenv("REDIS_TLS"), "true")

		opts = &redis.Options{
			Addr:     trimmed,
			Username: user,
			Password: password,
			DB:       0,
		}

		if useTLS {
			opts.TLSConfig = &tls.Config{
				MinVersion: tls.VersionTLS12,
			}
		}
	}

	client := redis.NewClient(opts)

	ctx, cancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return &RedisDatabase{
		Client: client,
	}, nil
}