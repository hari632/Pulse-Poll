package repositories

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

type RedisVoteCounter struct {
	client *redis.Client
}

func NewRedisVoteCounter(client *redis.Client) *RedisVoteCounter {
	return &RedisVoteCounter{
		client: client,
	}
}

func (r *RedisVoteCounter) Increment(
	ctx context.Context,
	pollID string,
	optionID string,
) (int64, error) {
	key := fmt.Sprintf("poll:%s:results", pollID)

	return r.client.HIncrBy(
		ctx,
		key,
		optionID,
		1,
	).Result()
}

func (r *RedisVoteCounter) Get(
	ctx context.Context,
	pollID string,
	optionID string,
) (int64, error) {
	key := fmt.Sprintf("poll:%s:results", pollID)

	return r.client.HGet(
		ctx,
		key,
		optionID,
	).Int64()
}

func (r *RedisVoteCounter) Delete(
	ctx context.Context,
	pollID string,
) error {
	key := fmt.Sprintf("poll:%s:results", pollID)

	return r.client.Del(ctx, key).Err()
}