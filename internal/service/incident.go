package service

import (
	"context"

	"github.com/ArtemChadaev/RedGo/internal/domain"
)

type IncidentConfig struct {
	StatsWindow     int     // за сколько минут считать юзеров
	DetectionRadius float64 // радиус обнаружения в игровых единицах
}
type IncidentService struct {
	repo domain.IncidentRepository
	cfg  IncidentConfig
}

func NewIncidentService(repo domain.IncidentRepository, cfg IncidentConfig) *IncidentService {
	return &IncidentService{
		repo: repo,
		cfg:  cfg,
	}
}

func (s *IncidentService) CreateIncident(ctx context.Context, inc *domain.Incident) error {
	return s.repo.Create(ctx, inc)
}

func (s *IncidentService) GetIncidents(ctx context.Context, page, pageSize int) ([]domain.Incident, error) {
	limit := pageSize
	if limit <= 0 {
		limit = 10000
	}

	offset := (page - 1) * pageSize
	if offset < 0 {
		offset = 0
	}
	return s.repo.GetAll(ctx, limit, offset)
}

func (s *IncidentService) GetIncidentByID(ctx context.Context, id int) (*domain.Incident, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *IncidentService) UpdateIncident(ctx context.Context, id int, inc *domain.Incident) error {
	inc.ID = id
	return s.repo.Update(ctx, inc)
}

func (s *IncidentService) DeleteIncident(ctx context.Context, id int) error {
	return s.repo.Delete(ctx, id)
}

func (s *IncidentService) CheckLocation(ctx context.Context, playerX, playerY float64) ([]domain.Incident, error) {
	return s.repo.GetCircle(ctx, playerX, playerY, s.cfg.DetectionRadius)
}

func (s *IncidentService) GetStats(ctx context.Context) (int, error) {
	return s.repo.GetStats(ctx, s.cfg.StatsWindow)
}
