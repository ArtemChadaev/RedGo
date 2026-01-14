package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/ArtemChadaev/RedGo/internal/domain"
	"github.com/ArtemChadaev/RedGo/internal/handler"
	"github.com/ArtemChadaev/RedGo/internal/repository"
	"github.com/ArtemChadaev/RedGo/internal/service"
)

func main() {
	// 1. Подключение к БД (Postgres)
	db, err := repository.NewPostgresDB(repository.PostgresConfig{
		Host:     os.Getenv("DB_HOST"),
		Port:     os.Getenv("DB_PORT"),
		Username: os.Getenv("DB_USER"),
		Database: os.Getenv("DB_NAME"),
		Password: os.Getenv("DB_PASSWORD"),
		SSLMode:  "disable",
	})

	if err != nil {
		log.Fatalf("failed to initialize db: %s", err.Error())
	}

	// 2. Подключение к Redis
	redisClient, err := repository.NewRedisClient(repository.RedisConfig{
		Addr:     os.Getenv("REDIS_HOST") + ":" + os.Getenv("REDIS_PORT"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0,
	})
	if err != nil {
		log.Fatalf("failed to initialize redis: %s", err.Error())
	}

	// 3. Инициализация слоев
	repos := repository.NewRepository(db)
	services := service.NewService(repos, redisClient)
	handlers := handler.NewHandler(services, redisClient)

	srv := new(domain.Server)

	go func() {
		if err := srv.Run(os.Getenv("PORT"), handlers.Routes()); err != nil {
			log.Fatalf("error occurred while running http server: %s", err.Error())
		}
	}()

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)
	<-ch

	fmt.Println("stopping server")
}
