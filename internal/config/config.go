package config

import (
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
	viper.SetConfigFile(".env")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
