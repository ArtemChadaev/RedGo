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
	cashe domain.IncidentCacheRepository
	queue domain.QueueRepository
	cfg   IncidentConfig
}

func NewIncidentService(repo domain.IncidentRepository, cashe domain.IncidentCacheRepository, queue domain.QueueRepository, cfg IncidentConfig) domain.IncidentService {
	return &incidentService{
		repo:  repo,
		cashe: cashe,
		queue: queue,
		cfg:   cfg,
	}
}

func (s *incidentService) CreateIncident(ctx context.Context, inc *domain.Incident) error {
	if err := s.repo.Create(ctx, inc); err != nil {
		return err
	}
	if err := s.cashe.DeleteActive(ctx); err != nil {
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

func (s *incidentService) Update(ctx context.Context, id int, input domain.UpdateIncidentInput) error {
	// 1. Вызываем метод репозитория с id и структурой для обновления
	// Мы больше не присваиваем id внутрь структуры, а передаем его вторым аргументом
	if err := s.repo.Update(ctx, id, input); err != nil {
		return err
	}

	// 2. Инвалидация кэша
	// После обновления данных в БД старый кэш "GetAllActive" становится неактуальным
	if err := s.cashe.DeleteActive(ctx); err != nil {
		// Логируем ошибку, но не прерываем выполнение, так как БД уже обновлена
		log.Printf("WARNING: failed to delete active cache after update: %v", err)
	}

	return nil
}

func (s *incidentService) DeleteIncident(ctx context.Context, id int) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}

	if err := s.cashe.DeleteActive(ctx); err != nil {
		log.Printf("failed to delete active cache: %v", err)
	}
	return nil
}

func (s *incidentService) CheckLocation(ctx context.Context, userID int, x, y float64) ([]domain.Incident, error) {
	if err := s.repo.SaveCheck(ctx, userID, x, y); err != nil {
		return nil, err
	}

	// Здесь кэш важен, так как запросов много
	incidents, err := s.cashe.GetActive(ctx)
	if err != nil || incidents == nil {
		incidents, err = s.repo.GetAllActive(ctx)
		if err == nil {
			_ = s.cashe.SetActive(ctx, incidents)
		}
	}

	var nearby []domain.Incident
	radiusSq := s.cfg.DetectionRadius * s.cfg.DetectionRadius

	for _, inc := range incidents {
		if inc.X == nil || inc.Y == nil {
			log.Printf("WARNING: incident %d has nil coordinates", inc.ID)
			continue
		}

		dx := x - *inc.X
		dy := y - *inc.Y

		if dx*dx+dy*dy <= radiusSq {
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

func (s *incidentService) HealthCheckDB(ctx context.Context) error {
	return s.repo.PingDB(ctx)
}

func (s *incidentService) HealthCheckRedis(ctx context.Context) error {
	return s.cashe.PingRedis(ctx)
}
