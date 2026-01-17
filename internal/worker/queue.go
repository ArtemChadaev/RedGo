package worker

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
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

	wg sync.WaitGroup
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
	w.wg.Add(1)
	defer w.wg.Done()

	log.Printf("Autoscaler started. Min: %d, Max: %d", min, max)

	// Запускаем минимальное кол-во воркеров сразу
	for i := 0; i < min; i++ {
		w.addWorker(ctx)
	}

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			qLen, err := w.redis.LLen(ctx, domain.WebhookQueueKey).Result()
			if err != nil {
				continue
			}

			current := atomic.LoadInt32(&w.activeWorkers)

			// 2. РАССЧИТЫВАЕМ ЦЕЛЕВОЕ КОЛИЧЕСТВО (target)
			target := int32(qLen / 30)

			// Ограничиваем target рамками min/max
			if target < int32(min) {
				target = int32(min)
			}
			if target > int32(max) {
				target = int32(max)
			}

			// Если очередь пуста, сбрасываем до минимума
			if qLen == 0 {
				target = int32(min)
			}

			// 3. МАСШТАБИРУЕМ ПАЧКОЙ (БОЛЬШЕ ЗА РАЗ)
			if target > current {
				diff := target - current
				log.Printf("Scaling UP: +%d workers (Queue: %d, Total: %d)", diff, qLen, target)
				for i := 0; i < int(diff); i++ {
					w.addWorker(ctx)
				}
			} else if target < current {
				diff := current - target
				// Плавное снижение: убираем не более 5 за раз, чтобы не дергать систему
				if diff > 5 {
					diff = 5
				}
				log.Printf("Scaling DOWN: -%d workers (Target: %d)", diff, target)
				for i := 0; i < int(diff); i++ {
					w.removeWorker()
				}
			}
		}
	}
}

func (w *WebhookWorker) addWorker(ctx context.Context) {
	w.mu.Lock()
	defer w.mu.Unlock()

	workerCtx, cancel := context.WithCancel(ctx)
	w.workerCancel = append(w.workerCancel, cancel)

	atomic.AddInt32(&w.activeWorkers, 1)

	// Регистрируем новый воркер
	w.wg.Add(1)
	go func() {
		defer w.wg.Done()
		w.runWorkerLoop(workerCtx)
	}()
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
				w.handleFailure(task)
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

// handleFailure обрабатывает ошибки: планирует пере повтор (ZSet) или отправляет в DLQ
func (w *WebhookWorker) handleFailure(task domain.WebhookTask) {
	task.Retries++

	// ВАЖНО: Создаем новый контекст на 2 секунды для финальной записи в Redis.
	// Мы НЕ используем входящий ctx, так как он может быть уже отменен (canceled).
	cleanupCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if task.Retries >= domain.MaxRetries {
		log.Printf("Task %d FAILED after %d attempts. Moving to DLQ.", task.IncidentID, domain.MaxRetries)
		data, _ := json.Marshal(task)
		// Используем cleanupCtx вместо ctx
		if err := w.redis.RPush(cleanupCtx, domain.WebhookDLQKey, data).Err(); err != nil {
			log.Printf("Critical: failed to push to DLQ: %v", err)
		}
		return
	}

	delay := time.Duration(math.Pow(2, float64(task.Retries))) * time.Second
	executeAt := time.Now().Add(delay).Unix()
	data, _ := json.Marshal(task)

	// Используем cleanupCtx вместо ctx
	err := w.redis.ZAdd(cleanupCtx, "webhooks:delayed", redis.Z{
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
	w.wg.Add(1)
	defer w.wg.Done()

	log.Println("Scheduler started (ZSet -> Queue)")
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("Stopping scheduler...")
			return
		case <-ticker.C:
			// Запускаем Lua-скрипт.
			// Переносим все задачи, чей Score (время) <= текущему времени
			now := time.Now().Unix()

			// KEYS[1] = webhooks:delayed, KEYS[2] = webhooks:queue, ARGV[1] = now
			count, err := transferScript.Run(ctx, w.redis,
				[]string{"webhooks:delayed", domain.WebhookQueueKey},
				now,
			).Int()

			if err != nil && !errors.Is(err, redis.Nil) {
				log.Printf("Scheduler Lua error: %v", err)
			} else if count > 0 {
				log.Printf("Scheduler: moved %d tasks to main queue", count)
			}
		}
	}
}

func (w *WebhookWorker) Wait() {
	w.wg.Wait()
}

// Stats содержит информацию о текущей нагрузке
type Stats struct {
	PendingTasks  int64 `json:"pending_tasks"`  // В основной очереди
	DelayedTasks  int64 `json:"delayed_tasks"`  // На повторе (ZSet)
	ActiveWorkers int32 `json:"active_workers"` // Живые горутины
}

func (w *WebhookWorker) GetStats(ctx context.Context) (Stats, error) {
	// 1. Сколько задач ждут прямо сейчас
	qLen, err := w.redis.LLen(ctx, domain.WebhookQueueKey).Result()
	if err != nil {
		return Stats{}, fmt.Errorf("failed to get queue len: %w", err)
	}

	// 2. Сколько задач "спят" и ждут времени переповтора.
	// Используем константу или строку "webhooks:delayed"
	dLen, err := w.redis.ZCard(ctx, "webhooks:delayed").Result()
	if err != nil {
		return Stats{}, fmt.Errorf("failed to get delayed len: %w", err)
	}

	return Stats{
		PendingTasks:  qLen,
		DelayedTasks:  dLen,
		ActiveWorkers: atomic.LoadInt32(&w.activeWorkers),
	}, nil
}
