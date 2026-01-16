package repository

import (
	"github.com/ArtemChadaev/RedGo/internal/domain"
	"github.com/jmoiron/sqlx"
)

type Repository struct {
	Incidents domain.IncidentRepository
}

func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{
		Incidents: NewIncidentRepository(db),
	}
}
