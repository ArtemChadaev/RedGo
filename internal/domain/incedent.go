package domain

import (
	"context"
)

type Incident struct {
	ID          int     `json:"id" db:"id"`
	Description string  `json:"description" db:"description"`
	X           float64 `json:"x" db:"x"`
	Y           float64 `json:"y" db:"y"`
	Status      string  `json:"status" db:"status"`
}

type IncidentRepository interface {
	Create(ctx context.Context, inc *Incident) error                   // Для POST /
	GetAll(ctx context.Context, limit, offset int) ([]Incident, error) // Для GET /
	GetByID(ctx context.Context, id int64) (*Incident, error)          // Для GET /:id
	Update(ctx context.Context, inc *Incident) error                   // Для PUT /:id
	Delete(ctx context.Context, id int64) error                        // Для DELETE /:id (смена статуса)

	// Метод для получения количества уникальных пользователей из истории проверок
	GetStats(ctx context.Context, windowMinutes int) (int, error) // Для GET /stats
}

type IncidentService interface {
	CreateIncident(ctx context.Context, inc *Incident) error
	GetIncidents(ctx context.Context, page, pageSize int) ([]Incident, error)
	GetIncidentByID(ctx context.Context, id int64) (*Incident, error)
	UpdateIncident(ctx context.Context, id int64, inc *Incident) error
	DeleteIncident(ctx context.Context, id int64) error

	// Логика проверки координат игрока: попал ли он в радиус опасности
	CheckLocation(ctx context.Context, playerX, playerY float64) ([]Incident, error)

	// Получение статистики (уникальные пользователи)
	GetStats(ctx context.Context) (int, error)
}
