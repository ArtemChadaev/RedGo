package worker

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ArtemChadaev/RedGo/internal/domain"
	"github.com/redis/go-redis/v9"
)

var transferScript = redis.NewScript(`
    local tasks = redis.call('ZRANGEBYSCORE', KEYS[1], '-inf', ARGV[1])
    if #tasks > 0 then
        for _, task in ipairs(tasks) do
            redis.call('ZREM', KEYS[1], task)
            redis.call('RPUSH', KEYS[2], task)
        end
    end
    return #tasks
`)

type WebhookWorker struct {
	redis      *redis.Client
	webhookURL string
	client     *http.Client

	// Поля для автоскейлинга
	activeWorkers int32                // Атомарный счетчик живых воркеров
	workerCancel  []context.CancelFunc // Функции для остановки лишних воркеров
	mu            sync.Mutex
}

func NewWebhookWorker(redis *redis.Client, url string) *WebhookWorker {
	return &WebhookWorker{
		redis:      redis,
		webhookURL: url,
		client: &http.Client{
			Timeout: 10 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        100,
				IdleConnTimeout:     90 * time.Second,
				MaxIdleConnsPerHost: 20,
			},
		},
	}
}

// StartAutoscaler — ЕДИНСТВЕННАЯ точка входа. Сама управляет мощностью.
func (w *WebhookWorker) StartAutoscaler(ctx context.Context, min, max int) {
	log.Printf("Autoscaler started. Min: %d, Max: %d", min, max)

	// Запускаем минимальное кол-во воркеров сразу
	for i := 0; i < min; i++ {
		w.addWorker(ctx)
	}

	ticker := time.NewTicker(5 * time.Second) // Проверяем очередь каждые 5 сек
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// 1. Узнаем длину очереди
			qLen, err := w.redis.LLen(ctx, domain.WebhookQueueKey).Result()
			if err != nil {
				continue
			}

			current := atomic.LoadInt32(&w.activeWorkers)

			// 2. Логика масштабирования
			// Если задач много (> 50 на каждый воркер) — добавляем
			if qLen > int64(current*50) && current < int32(max) {
				log.Printf("Queue is heavy (%d tasks). Scaling UP...", qLen)
				w.addWorker(ctx)
			}

			// Если очередь пуста и воркеров больше минимума — убираем один
			if qLen == 0 && current > int32(min) {
				log.Printf("Queue is empty. Scaling DOWN...")
				w.removeWorker()
			}
		}
	}
}

func (w *WebhookWorker) addWorker(ctx context.Context) {
	w.mu.Lock()
	defer w.mu.Unlock()

	// Создаем отдельный контекст для воркера, чтобы его можно было убить отдельно
	workerCtx, cancel := context.WithCancel(ctx)
	w.workerCancel = append(w.workerCancel, cancel)

	atomic.AddInt32(&w.activeWorkers, 1)
	go w.runWorkerLoop(workerCtx)
}

func (w *WebhookWorker) removeWorker() {
	w.mu.Lock()
	defer w.mu.Unlock()

	if len(w.workerCancel) > 0 {
		// Берем последнюю функцию отмены и вызываем её
		cancel := w.workerCancel[len(w.workerCancel)-1]
		cancel()
		w.workerCancel = w.workerCancel[:len(w.workerCancel)-1]
		atomic.AddInt32(&w.activeWorkers, -1)
	}
}

func (w *WebhookWorker) runWorkerLoop(ctx context.Context) {
	id := atomic.LoadInt32(&w.activeWorkers)
	log.Printf("Worker #%d started", id)

	for {
		select {
		case <-ctx.Done():
			log.Printf("Worker #%d stopped", id)
			return
		default:
			result, err := w.redis.BLPop(ctx, 5*time.Second, domain.WebhookQueueKey).Result()
			if err != nil {
				continue // Здесь BLPop прервется сам, если вызвать cancel() контекста
			}

			var task domain.WebhookTask
			if err := json.Unmarshal([]byte(result[1]), &task); err != nil {
				log.Printf("CRITICAL: poison pill in queue! Failed to unmarshal: %v. Data: %s", err, result[1])
				continue // Пропускаем битую задачу и идем за следующей
			}

			if err := w.processTask(ctx, task); err != nil {
				w.handleFailure(ctx, task, err)
			}
		}
	}
}

// processTask выполняет непосредственную отправку HTTP POST запроса
func (w *WebhookWorker) processTask(ctx context.Context, task domain.WebhookTask) error {
	body, _ := json.Marshal(task)

	req, err := http.NewRequestWithContext(ctx, "POST", w.webhookURL, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("request build error: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := w.client.Do(req)
	if err != nil {
		return fmt.Errorf("network error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("server error: status %d", resp.StatusCode)
	}

	log.Printf("Successfully sent webhook for incident %d (User %d)", task.IncidentID, task.UserID)
	return nil
}

// handleFailure обрабатывает ошибки: планирует переповтор (ZSet) или отправляет в DLQ
func (w *WebhookWorker) handleFailure(ctx context.Context, task domain.WebhookTask, taskErr error) {
	task.Retries++

	// Если попытки исчерпаны — в очередь "мертвых" сообщений
	if task.Retries >= domain.MaxRetries {
		log.Printf("Task %d FAILED after %d attempts. Moving to DLQ. Last error: %v",
			task.IncidentID, domain.MaxRetries, taskErr)

		data, _ := json.Marshal(task)
		if err := w.redis.RPush(ctx, domain.WebhookDLQKey, data).Err(); err != nil {
			log.Printf("Critical: failed to push to DLQ: %v", err)
		}
		return
	}

	// Рассчитываем время следующей попытки (Exponential Backoff: 2, 4, 8, 16... секунд)
	delay := time.Duration(math.Pow(2, float64(task.Retries))) * time.Second
	executeAt := time.Now().Add(delay).Unix()

	data, _ := json.Marshal(task)

	// Сохраняем в ZSet (отложенная очередь)
	// Score — это время в формате Unix, когда задача должна "проснуться"
	err := w.redis.ZAdd(ctx, "webhooks:delayed", redis.Z{
		Score:  float64(executeAt),
		Member: data,
	}).Err()

	if err != nil {
		log.Printf("Failed to schedule retry for task %d: %v", task.IncidentID, err)
	} else {
		log.Printf("Task %d scheduled for retry #%d in %v", task.IncidentID, task.Retries, delay)
	}
}

// RunScheduler мониторит ZSet и атомарно переносит готовые задачи в основную очередь
func (w *WebhookWorker) RunScheduler(ctx context.Context) {
	log.Println("Scheduler started (ZSet -> Queue)")
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("Stopping scheduler...")
			return
		case <-ticker.C:
			// Запускаем Lua-скрипт
			// Переносим все задачи, чей Score (время) <= текущему времени
			now := time.Now().Unix()

			// KEYS[1] = webhooks:delayed, KEYS[2] = webhooks:queue, ARGV[1] = now
			count, err := transferScript.Run(ctx, w.redis,
				[]string{"webhooks:delayed", domain.WebhookQueueKey},
				now,
			).Int()

			if err != nil && err != redis.Nil {
				log.Printf("Scheduler Lua error: %v", err)
			} else if count > 0 {
				log.Printf("Scheduler: moved %d tasks to main queue", count)
			}
		}
	}
}
