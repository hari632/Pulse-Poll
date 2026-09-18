package repositories

import (
	"context"
	"encoding/json"

	"github.com/redis/go-redis/v9"

	"pulse-backend/models"
)

const VoteEventsChannel = "pulsepoll:vote-events"

type RedisEventPublisher struct {
	client *redis.Client
}

func NewRedisEventPublisher(client *redis.Client) *RedisEventPublisher {
	return &RedisEventPublisher{
		client: client,
	}
}

func (p *RedisEventPublisher) PublishVoteEvent(
	ctx context.Context,
	event models.VoteEvent,
) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return err
	}

	return p.client.Publish(
		ctx,
		VoteEventsChannel,
		payload,
	).Err()
}