package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ArtemChadaev/RedGo/internal/config"
	"github.com/ArtemChadaev/RedGo/internal/domain"
	"github.com/ArtemChadaev/RedGo/internal/handler"
	"github.com/ArtemChadaev/RedGo/internal/repository"
	"github.com/ArtemChadaev/RedGo/internal/service"
	"github.com/ArtemChadaev/RedGo/internal/worker"
)

func main() {
	// 0. Конфиг и базовый контекст для сигналов ОС
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to config: %s", err)
	}

	// Создаем контекст, который отменится при Ctrl+C или docker stop
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// 1. Инициализация ресурсов (БД и Redis)
	db, err := repository.NewPostgresDB(repository.PostgresConfig{
		Host:     cfg.DBHost,
		Port:     cfg.DBPort,
		Username: cfg.DBUser,
		Database: cfg.DBName,
		Password: cfg.DBPassword,
		SSLMode:  "disable",
	})
	if err != nil {
		log.Fatalf("failed to initialize db: %s", err.Error())
	}

	redisClient, err := repository.NewRedisClient(repository.RedisConfig{
		Addr:     cfg.RedisHost + ":" + cfg.RedisPort,
		Password: cfg.RedisPassword,
		DB:       0,
	})
	if err != nil {
		log.Fatalf("failed to initialize redis: %s", err.Error())
	}

	// 2. Инициализация Воркера
	// Мы передаем управление WaitGroup внутрь структуры WebhookWorker
	webhookWorker := worker.NewWebhookWorker(redisClient, cfg.WebhookURL)

	// Запускаем фоновые процессы воркера
	go webhookWorker.RunScheduler(ctx)
	go webhookWorker.StartAutoscaler(ctx, 2, 200)

	// 3. Инициализация слоев (Repository -> Service -> Handler)
	repos := repository.NewRepository(db, redisClient)
	incCfg := service.IncidentConfig{
		StatsWindow:     cfg.StatsWindow,
		DetectionRadius: cfg.DetectionRadius,
	}
	services := service.NewService(repos, incCfg)

	handlers := handler.NewHandler(services, webhookWorker)

	// 4. Запуск HTTP сервера в отдельной горутине
	srv := new(domain.Server)
	go func() {
		if err := srv.Run(cfg.Port, handlers.Routes(cfg.ApiKey)); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("HTTP server error: %v", err)
		}
	}()

	log.Printf("Server started on port %s", cfg.Port)

	// --- ОЖИДАНИЕ ЗАВЕРШЕНИЯ ---
	<-ctx.Done() // Блокируемся здесь, пока не придет сигнал (SIGINT/SIGTERM)
	log.Println("Shutting down gracefully...")

	// 1. Останавливаем HTTP-сервер (перестаем принимать новые входящие инциденты)
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server Shutdown Failed: %+v", err)
	}

	// 2. Ждем, пока воркеры доделают задачи, отправят ретраи в Redis и выйдут
	log.Println("Waiting for workers to finish current tasks...")

	// Используем канал для таймаута ожидания воркеров
	waitCh := make(chan struct{})
	go func() {
		webhookWorker.Wait() // Этот метод внутри вызывает wg.Wait()
		close(waitCh)
	}()

	select {
	case <-waitCh:
		log.Println("All workers exited cleanly")
	case <-time.After(15 * time.Second): // Даем воркерам чуть больше времени, чем серверу
		log.Println("Workers shutdown timed out, force closing resources...")
	}

	// 3. Только когда воркеры закончили работу с БД/Redis, закрываем соединения
	if err := redisClient.Close(); err != nil {
		log.Printf("Failed to close Redis: %v", err)
	}
	if err := db.Close(); err != nil {
		log.Printf("Failed to close DB: %v", err)
	}

	fmt.Println("Server exited properly")
}
