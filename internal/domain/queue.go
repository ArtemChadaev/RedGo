package domain

import "context"

// WebhookTask представляет данные, которые полетят в очередь Redis
type WebhookTask struct {
	IncidentID int     `json:"incident_id"`
	UserID     int     `json:"user_id"`
	X          float64 `json:"x"`
	Y          float64 `json:"y"`
}

// QueueRepository — интерфейс для работы с очередью задач
type QueueRepository interface {
	PushWebhookTask(ctx context.Context, task WebhookTask) error
}
