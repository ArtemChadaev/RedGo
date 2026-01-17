package domain

import (
	"context"
)

type IncidentStatus string

const (
	StatusActive   IncidentStatus = "active"
	StatusInactive IncidentStatus = "inactive"
)

type Incident struct {
	ID          int            `json:"id" db:"id"`
	Description string         `json:"description" db:"description"`
	X           *float64       `json:"x" binding:"required" db:"x"`
	Y           *float64       `json:"y" binding:"required" db:"y"`
	Status      IncidentStatus `json:"status" binding:"omitempty,oneof=active inactive" db:"status"`
}

type UpdateIncidentInput struct {
	X           *float64        `json:"x"`
	Y           *float64        `json:"y"`
	Description *string         `json:"description"`
	Status      *IncidentStatus `json:"status" binding:"omitempty,oneof=active inactive"`
}

type IncidentRepository interface {
	Create(ctx context.Context, inc *Incident) error                     // Для POST /
	GetAll(ctx context.Context, limit, offset int) ([]Incident, error)   // Для GET /
	GetByID(ctx context.Context, id int) (*Incident, error)              // Для GET /:id
	Update(ctx context.Context, id int, input UpdateIncidentInput) error // Для PUT /:id
	Delete(ctx context.Context, id int) error                            // Для DELETE /:id (смена статуса)

	// GetStats Метод для получения количества уникальных пользователей из истории проверок
	GetStats(ctx context.Context, windowMinutes int) (int, error) // Для GET /stats

	// SaveCheck Сохранение проверок
	SaveCheck(ctx context.Context, userID int, x, y float64) error

	// GetAllActive Нужен для получения всех активных записей для кэша
	GetAllActive(ctx context.Context) ([]Incident, error)

	PingDB(ctx context.Context) error
}

type IncidentCacheRepository interface {
	GetActive(ctx context.Context) ([]Incident, error)
	SetActive(ctx context.Context, incidents []Incident) error
	DeleteActive(ctx context.Context) error

	PingRedis(ctx context.Context) error
}

type IncidentService interface {
	CreateIncident(ctx context.Context, inc *Incident) error
	GetIncidents(ctx context.Context, page, pageSize int) ([]Incident, error)
	GetIncidentByID(ctx context.Context, id int) (*Incident, error)
	Update(ctx context.Context, id int, input UpdateIncidentInput) error
	DeleteIncident(ctx context.Context, id int) error

	// CheckLocation Логика проверки координат игрока: попал ли он в радиус опасности
	CheckLocation(ctx context.Context, userID int, x, y float64) ([]Incident, error)

	// GetStats Получение статистики (уникальные пользователи)
	GetStats(ctx context.Context) (int, error)

	HealthCheckDB(ctx context.Context) error
	HealthCheckRedis(ctx context.Context) error
}
