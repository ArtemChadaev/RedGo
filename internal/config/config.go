package config

import (
	"log"

	"github.com/spf13/viper"
)

type Config struct {
	// Основные настройки приложения
	Port            string  `mapstructure:"PORT"`
	ApiKey          string  `mapstructure:"API_KEY"`
	StatsWindow     int     `mapstructure:"STATS_TIME_WINDOW_MINUTES"`
	DetectionRadius float64 `mapstructure:"DETECTION_RADIUS"`
	WebhookURL      string  `mapstructure:"WEBHOOK_URL"`

	// Настройки Postgres
	DBHost     string `mapstructure:"DB_HOST"`
	DBPort     string `mapstructure:"DB_PORT"`
	DBUser     string `mapstructure:"DB_USER"`
	DBName     string `mapstructure:"DB_NAME"`
	DBPassword string `mapstructure:"DB_PASSWORD"`

	// Настройки Redis
	RedisHost     string `mapstructure:"REDIS_HOST"`
	RedisPort     string `mapstructure:"REDIS_PORT"`
	RedisPassword string `mapstructure:"REDIS_PASSWORD"`
}

func Load() (*Config, error) {
	v := viper.New()
	log.Println("Loading config...")

	v.SetConfigName(".env")
	v.SetConfigType("env")
	v.AddConfigPath(".")
	v.AutomaticEnv()

	// Пытаемся прочитать файл
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Println("Note: .env file not found, using system environment variables")
		} else {
			return nil, err
		}
	}

	// ВАЖНО: Явно привязываем каждый ключ, чтобы Unmarshal сработал без файла
	keys := []string{
		"PORT", "API_KEY", "STATS_TIME_WINDOW_MINUTES", "DETECTION_RADIUS",
		"WEBHOOK_URL", "DB_HOST", "DB_PORT", "DB_USER", "DB_NAME",
		"DB_PASSWORD", "REDIS_HOST", "REDIS_PORT", "REDIS_PASSWORD",
	}
	for _, key := range keys {
		if err := v.BindEnv(key); err != nil {
			return nil, err
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
