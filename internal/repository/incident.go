package repository

import (
	"context"

	"github.com/ArtemChadaev/RedGo/internal/domain"
	"github.com/jmoiron/sqlx"
)

type incidentRepository struct {
	db *sqlx.DB
}

func NewIncidentRepository(db *sqlx.DB) domain.IncidentRepository {
	return &incidentRepository{db: db}
}

func (r *incidentRepository) Create(ctx context.Context, inc *domain.Incident) error {
	query := `
		INSERT INTO incidents (description, x, y, status)
		VALUES (:description, :x, :y, :status)
		RETURNING id
	`
	rows, err := r.db.NamedQueryContext(ctx, query, inc)
	if err != nil {
		return err
	}
	defer rows.Close()

	if rows.Next() {
		if err := rows.Scan(&inc.ID); err != nil {
			return err
		}
	}

	return rows.Err()
}

func (r *incidentRepository) GetAll(ctx context.Context, limit, offset int) ([]domain.Incident, error) {
	//TODO implement me
	panic("implement me")
}

func (r *incidentRepository) GetByID(ctx context.Context, id int64) (*domain.Incident, error) {
	//TODO implement me
	panic("implement me")
}

func (r *incidentRepository) Update(ctx context.Context, inc *domain.Incident) error {
	//TODO implement me
	panic("implement me")
}

func (r *incidentRepository) Delete(ctx context.Context, id int64) error {
	//TODO implement me
	panic("implement me")
}

func (r *incidentRepository) GetStats(ctx context.Context, windowMinutes int) (int, error) {
	//TODO implement me
	panic("implement me")
}
