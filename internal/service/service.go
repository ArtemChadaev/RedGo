package service

import (
	"github.com/ArtemChadaev/RedGo/internal/domain"
	"github.com/ArtemChadaev/RedGo/internal/repository"
	"github.com/redis/go-redis/v9"
)

type Service struct {
	domain.IncidentService
}

func NewService(repos *repository.Repository, redis *redis.Client) *Service {
	incidentService := NewIncidentService(repos.Incidents)
	return &Service{
		IncidentService: incidentService,
	}
}
