package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
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
	// 0. Конфиг и базовый контекст
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Ошибка загрузки конфига: %s", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// 1. Инициализация ресурсов
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

	// 2.
	repos := repository.NewRepository(db, redisClient)
	incCfg := service.IncidentConfig{
		StatsWindow:     cfg.StatsWindow,
		DetectionRadius: cfg.DetectionRadius,
	}
	services := service.NewService(repos, incCfg)
	handlers := handler.NewHandler(services)

	// 3. Инициализация Воркеров и WaitGroup
	var wg sync.WaitGroup // Счетчик для ожидания горутин
	w := worker.NewWebhookWorker(redisClient, cfg.WebhookURL)

	// Запускаем планировщик
	wg.Add(1)
	go func() {
		defer wg.Done()
		w.RunScheduler(ctx)
	}()

	// Запускаем автоскейлер
	wg.Add(1)
	go func() {
		defer wg.Done()
		w.StartAutoscaler(ctx, 2, 20)
	}()

	// 4. Запуск HTTP сервера
	srv := new(domain.Server)
	go func() {
		if err := srv.Run(cfg.Port, handlers.Routes(cfg.ApiKey)); err != nil && err != http.ErrServerClosed {
			log.Printf("HTTP server error: %s", err.Error())
		}
	}()

	log.Printf("Server started on port %s", cfg.Port)

	// --- ОЖИДАНИЕ ЗАВЕРШЕНИЯ ---
	<-ctx.Done()
	log.Println("Shutting down gracefully...")

	// 1. Сначала закрываем HTTP-сервер (чтобы не принимал новые запросы)
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server Shutdown Failed: %+v", err)
	}

	// 2. Ждем, пока воркеры доделают свои задачи и выйдут из циклов
	waitCh := make(chan struct{})
	go func() {
		wg.Wait()
		close(waitCh)
	}()

	// Либо ждем воркеров, либо убиваем по таймауту
	select {
	case <-waitCh:
		log.Println("Workers finished successfully")
	case <-time.After(10 * time.Second):
		log.Println("Workers shutdown timed out, force closing...")
	}

	// 3. И только в самом конце закрываем БД и Redis
	if err := db.Close(); err != nil {
		log.Printf("Failed to close DB: %v", err)
	}
	if err := redisClient.Close(); err != nil {
		log.Printf("Failed to close Redis: %v", err)
	}

	fmt.Println("Server exited properly")
}
