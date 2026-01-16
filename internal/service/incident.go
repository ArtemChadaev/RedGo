package service

import (
	"context"

	"github.com/ArtemChadaev/RedGo/internal/domain"
)

type IncidentService struct {
	repo domain.IncidentRepository
}

func NewIncidentService(repo domain.IncidentRepository) *IncidentService {
	return &IncidentService{repo: repo}
}

func (i IncidentService) CreateIncident(ctx context.Context, inc *domain.Incident) error {
	//TODO implement me
	panic("implement me")
}

func (i IncidentService) GetIncidents(ctx context.Context, page, pageSize int) ([]domain.Incident, error) {
	//TODO implement me
	panic("implement me")
}

func (i IncidentService) GetIncidentByID(ctx context.Context, id int64) (*domain.Incident, error) {
	//TODO implement me
	panic("implement me")
}

func (i IncidentService) UpdateIncident(ctx context.Context, id int64, inc *domain.Incident) error {
	//TODO implement me
	panic("implement me")
}

func (i IncidentService) DeleteIncident(ctx context.Context, id int64) error {
	//TODO implement me
	panic("implement me")
}

func (i IncidentService) CheckLocation(ctx context.Context, playerX, playerY float64) ([]domain.Incident, error) {
	//TODO implement me
	panic("implement me")
}

func (i IncidentService) GetStats(ctx context.Context) (int, error) {
	//TODO implement me
	panic("implement me")
}
