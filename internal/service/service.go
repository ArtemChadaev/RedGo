package service

import (
	"github.com/ArtemChadaev/RedGo/internal/domain"
	"github.com/ArtemChadaev/RedGo/internal/repository"
)

type Service struct {
	domain.IncidentService
}

func NewService(repos *repository.Repository, cfg IncidentConfig) *Service {
	incidentService := NewIncidentService(repos.Incidents, repos.IncidentCashe, repos.Queues, cfg)
	return &Service{
		IncidentService: incidentService,
	}
}
