package repository

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/ArtemChadaev/RedGo/internal/domain"
	"github.com/redis/go-redis/v9"
)

type incidentCasheRepository struct {
	redis *redis.Client
}

func NewIncidentCasheRepository(redis *redis.Client) domain.IncidentCacheRepository {
	return &incidentCasheRepository{redis: redis}
}

const activeIncidentsKey = "incidents:active"

func (r *incidentCasheRepository) GetActive(ctx context.Context) ([]domain.Incident, error) {
	val, err := r.redis.Get(ctx, activeIncidentsKey).Result()

	if errors.Is(err, redis.Nil) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var incidents []domain.Incident
	if err := json.Unmarshal([]byte(val), &incidents); err != nil {
		return nil, err
	}

	return incidents, nil
}

func (r *incidentCasheRepository) SetActive(ctx context.Context, incidents []domain.Incident) error {
	data, err := json.Marshal(incidents)
	if err != nil {
		return err
	}

	return r.redis.Set(ctx, activeIncidentsKey, data, 10*time.Minute).Err()
}

func (r *incidentCasheRepository) DeleteActive(ctx context.Context) error {
	return r.redis.Del(ctx, activeIncidentsKey).Err()
}

func (r *incidentCasheRepository) PingRedis(ctx context.Context) error {
	// Для Redis отправляется команда PING, ответом на которую должно быть PONG
	return r.redis.Ping(ctx).Err()
}
