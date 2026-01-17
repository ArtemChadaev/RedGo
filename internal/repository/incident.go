package repository

import (
	"context"
	"errors"

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
	incidents := make([]domain.Incident, 0, limit)

	query := `
		SELECT id, description, x, y, status 
		FROM incidents 
		ORDER BY id DESC 
		LIMIT $1 OFFSET $2
	`

	err := r.db.SelectContext(ctx, &incidents, query, limit, offset)
	if err != nil {
		return nil, err
	}

	return incidents, nil
}

func (r *incidentRepository) GetByID(ctx context.Context, id int) (*domain.Incident, error) {
	var incident domain.Incident
	query := `SELECT id, description, x, y, status FROM incidents WHERE id = $1`

	err := r.db.GetContext(ctx, &incident, query, id)
	if err != nil {
		return nil, err
	}

	return &incident, nil
}

func (r *incidentRepository) Update(ctx context.Context, id int, input domain.UpdateIncidentInput) error {
	// В PostgreSQL COALESCE идеально подходит для Partial Update
	query := `
        UPDATE incidents 
        SET 
            x = COALESCE($1, x), 
            y = COALESCE($2, y), 
            description = COALESCE($3, description),
            status = COALESCE($4, status)
        WHERE id = $5
    `

	// Передаем указатели напрямую.
	// Если в структуре поле nil, драйвер sql/pq отправит в базу NULL.
	result, err := r.db.ExecContext(ctx, query, input.X, input.Y, input.Description, input.Status, id)
	if err != nil {
		return err
	}

	// Проверяем, была ли обновлена хотя бы одна строка
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return errors.New("incident not found")
	}

	return nil
}

func (r *incidentRepository) Delete(ctx context.Context, id int) error {
	query := `UPDATE incidents SET status = $1 WHERE id = $2`

	_, err := r.db.ExecContext(ctx, query, domain.StatusInactive, id)
	return err
}

func (r *incidentRepository) GetStats(ctx context.Context, windowMinutes int) (int, error) {
	var count int
	query := `
       SELECT COUNT(DISTINCT user_id) 
       FROM location_checks 
       WHERE created_at >= NOW() - INTERVAL '1 minute' * $1
    `

	err := r.db.GetContext(ctx, &count, query, windowMinutes)
	return count, err
}

func (r *incidentRepository) SaveCheck(ctx context.Context, userID int, x, y float64) error {
	query := `INSERT INTO location_checks (user_id, x, y) VALUES ($1, $2, $3)`
	_, err := r.db.ExecContext(ctx, query, userID, x, y)
	return err
}

func (r *incidentRepository) GetAllActive(ctx context.Context) ([]domain.Incident, error) {
	query := `
        SELECT id, x, y, status 
        FROM incidents 
        WHERE status = $1
    `

	rows, err := r.db.QueryContext(ctx, query, domain.StatusActive)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var incidents []domain.Incident
	for rows.Next() {
		var inc domain.Incident
		if err := rows.Scan(&inc.ID, &inc.X, &inc.Y, &inc.Status); err != nil {
			return nil, err
		}
		incidents = append(incidents, inc)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return incidents, nil
}

func (r *incidentRepository) PingDB(ctx context.Context) error {
	// PingContext проверяет, что соединение с базой данных всё еще активно
	return r.db.PingContext(ctx)
}
