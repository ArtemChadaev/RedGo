package service

import (
	"context"

	"github.com/ArtemChadaev/RedGo/internal/domain"
)

type IncidentConfig struct {
	StatsWindow     int     // за сколько минут считать юзеров
	DetectionRadius float64 // радиус обнаружения в игровых единицах
}
type incidentService struct {
	repo  domain.IncidentRepository
	cash  domain.IncidentCacheRepository
	queue domain.QueueRepository
	cfg   IncidentConfig
}

func NewIncidentService(repo domain.IncidentRepository, cash domain.IncidentCacheRepository, queue domain.QueueRepository, cfg IncidentConfig) domain.IncidentService {
	return &incidentService{
		repo:  repo,
		cash:  cash,
		queue: queue,
		cfg:   cfg,
	}
}

func (s *incidentService) CreateIncident(ctx context.Context, inc *domain.Incident) error {
	return s.repo.Create(ctx, inc)
}

func (s *incidentService) GetIncidents(ctx context.Context, page, pageSize int) ([]domain.Incident, error) {
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

func (s *incidentService) GetIncidentByID(ctx context.Context, id int) (*domain.Incident, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *incidentService) UpdateIncident(ctx context.Context, id int, inc *domain.Incident) error {
	inc.ID = id
	return s.repo.Update(ctx, inc)
}

func (s *incidentService) DeleteIncident(ctx context.Context, id int) error {
	return s.repo.Delete(ctx, id)
}

func (s *incidentService) CheckLocation(ctx context.Context, userID int, x, y float64) ([]domain.Incident, error) {
	if err := s.repo.SaveCheck(ctx, userID, x, y); err != nil {
		return nil, err
	}

	return s.repo.GetCircle(ctx, x, y, s.cfg.DetectionRadius)
}

func (s *incidentService) GetStats(ctx context.Context) (int, error) {
	return s.repo.GetStats(ctx, s.cfg.StatsWindow)
}
