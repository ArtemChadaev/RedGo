package repository

import (
	"github.com/ArtemChadaev/RedGo/internal/domain"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
)

type Repository struct {
	Incidents     domain.IncidentRepository
	IncidentCashe domain.IncidentCacheRepository
	Queues        domain.QueueRepository
}

func NewRepository(db *sqlx.DB, redis *redis.Client) *Repository {
	return &Repository{
		Incidents:     NewIncidentRepository(db),
		IncidentCashe: NewIncidentCasheRepository(redis),
		Queues:        NewIncidentQueueRepository(redis),
	}
}
