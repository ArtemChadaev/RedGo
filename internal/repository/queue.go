package repository

import (
	"context"
	"encoding/json"

	"github.com/ArtemChadaev/RedGo/internal/domain"
	"github.com/redis/go-redis/v9"
)

const webhookQueueKey = "webhooks:queue"

type incidentQueueRepository struct {
	redis *redis.Client
}

func NewIncidentQueueRepository(redis *redis.Client) domain.QueueRepository {
	return &incidentQueueRepository{redis: redis}
}

func (r *incidentQueueRepository) PushWebhookTask(ctx context.Context, task domain.WebhookTask) error {
	data, err := json.Marshal(task)
	if err != nil {
		return err
	}

	return r.redis.RPush(ctx, webhookQueueKey, data).Err()
}
