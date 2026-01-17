package service

import (
	"context"
	"log"

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
	if err := s.repo.Create(ctx, inc); err != nil {
		return err
	}
	if err := s.cash.DeleteActive(ctx); err != nil {
		log.Printf("failed to delete active cache: %v", err)
	}
	return nil
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
	if err := s.repo.Update(ctx, inc); err != nil {
		return err
	}
	if err := s.cash.DeleteActive(ctx); err != nil {
		log.Printf("failed to delete active cache: %v", err)
	}
	return nil
}

func (s *incidentService) DeleteIncident(ctx context.Context, id int) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}

	if err := s.cash.DeleteActive(ctx); err != nil {
		log.Printf("failed to delete active cache: %v", err)
	}
	return nil
}

func (s *incidentService) CheckLocation(ctx context.Context, userID int, x, y float64) ([]domain.Incident, error) {
	if err := s.repo.SaveCheck(ctx, userID, x, y); err != nil {
		return nil, err
	}

	// Здесь кэш важен, так как запросов много
	incidents, err := s.cash.GetActive(ctx)
	if err != nil || incidents == nil {
		incidents, err = s.repo.GetAllActive(ctx)
		if err == nil {
			_ = s.cash.SetActive(ctx, incidents)
		}
	}

	var nearby []domain.Incident
	radiusSq := s.cfg.DetectionRadius * s.cfg.DetectionRadius

	for _, inc := range incidents {
		if (x-inc.X)*(x-inc.X)+(y-inc.Y)*(y-inc.Y) <= radiusSq {
			nearby = append(nearby, inc)

			if err := s.queue.PushWebhookTask(ctx, domain.WebhookTask{
				IncidentID: inc.ID,
				UserID:     userID,
				X:          x,
				Y:          y,
			}); err != nil {
				log.Printf("WARNING: failed to push webhook task for incident %d: %v", inc.ID, err)
			}
		}
	}

	return nearby, nil
}

func (s *incidentService) GetStats(ctx context.Context) (int, error) {
	return s.repo.GetStats(ctx, s.cfg.StatsWindow)
}
