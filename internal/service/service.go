package service

import (
	"github.com/ArtemChadaev/RedGo/internal/repository"
	"github.com/redis/go-redis/v9"
)

type Service struct {
}

func NewService(repos *repository.Repository, redis *redis.Client) *Service {

	return &Service{}
}
