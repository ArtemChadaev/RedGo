package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ArtemChadaev/RedGo/internal/config"
	"github.com/ArtemChadaev/RedGo/internal/domain"
	"github.com/ArtemChadaev/RedGo/internal/handler"
	"github.com/ArtemChadaev/RedGo/internal/repository"
	"github.com/ArtemChadaev/RedGo/internal/service"
)

func main() {
	// 0. Загрузка конфигурации один раз при старте
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Ошибка загрузки конфига: %s", err)
	}

	// 1. Подключение к БД, используя данные из конфига
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

	// 2. Подключение к Redis, используя данные из конфига
	redisClient, err := repository.NewRedisClient(repository.RedisConfig{
		Addr:     cfg.RedisHost + ":" + cfg.RedisPort,
		Password: cfg.RedisPassword,
		DB:       0,
	})
	if err != nil {
		log.Fatalf("failed to initialize redis: %s", err.Error())
	}

	// 3. Инициализация слоев (Dependency Injection)
	repos := repository.NewRepository(db)

	// Настройки для бизнес-логики сервиса инцидентов
	incCfg := service.IncidentConfig{
		StatsWindow:     cfg.StatsWindow,
		DetectionRadius: cfg.DetectionRadius,
	}

	services := service.NewService(repos, redisClient, incCfg)

	// Инициализируем роуты, передавая API-ключ для Middleware
	handlers := handler.NewHandler(services, redisClient)

	srv := new(domain.Server)

	// Запуск сервера на порту из конфига
	go func() {
		if err := srv.Run(cfg.Port, handlers.Routes(cfg.ApiKey)); err != nil {
			log.Fatalf("error occurred while running http server: %s", err.Error())
		}
	}()

	log.Printf("Server started on port %s", cfg.Port)

	// Ожидание сигнала завершения
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
	<-ch

	fmt.Println("Stopping server...")

	// Создаем контекст с таймаутом (например, 5 секунд)
	// Это время, которое мы даем серверу на "сборы"
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Вызываем Shutdown у твоей структуры из domain/server.go
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server Shutdown Failed:%+v", err)
	}

	// Здесь также можно закрыть базу и редис
	// db.Close()
	// redisClient.Close()

	fmt.Println("Server exited properly")
	// TODO: srv.Shutdown для плавного выхода
}
