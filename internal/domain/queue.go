package domain

import "context"

const (
	WebhookQueueKey = "webhooks:queue"
	WebhookDLQKey   = "webhooks:dlq" // Очередь для задач, которые не удалось выполнить
	MaxRetries      = 5              // Максимальное количество попыток
)

// WebhookTask представляет данные, которые полетят в очередь Redis
type WebhookTask struct {
	IncidentID int     `json:"incident_id"`
	UserID     int     `json:"user_id"`
	X          float64 `json:"x"`
	Y          float64 `json:"y"`
	Retries    int     `json:"retries"`
}

// QueueRepository — интерфейс для работы с очередью задач
type QueueRepository interface {
	PushWebhookTask(ctx context.Context, task WebhookTask) error
}
