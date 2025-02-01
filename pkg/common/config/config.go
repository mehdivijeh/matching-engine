package config

import (
	"fmt"
	"os"
	"strings"
)

type Config struct {
	Kafka *KafkaConfig
	App   *AppConfig
}

type AppConfig struct {
	LogLevel   string
	ServerPort string
}

type KafkaConfig struct {
	Brokers      []string
	OrdersTopic  string
	MatchesTopic string
	GroupID      string
}

func Load() (*Config, error) {
	kafkaBrokers := getEnv("KAFKA_BROKERS", "kafka:29092")
	if kafkaBrokers == "" {
		return nil, fmt.Errorf("KAFKA_BROKERS environment variable is required")
	}

	config := &Config{
		App: &AppConfig{
			LogLevel:   getEnv("LOG_LEVEL", "info"),
			ServerPort: getEnv("SERVER_PORT", "8080"),
		},
		Kafka: &KafkaConfig{
			Brokers:      strings.Split(kafkaBrokers, ","),
			OrdersTopic:  getEnv("KAFKA_ORDERS_TOPIC", "orders"),
			MatchesTopic: getEnv("KAFKA_MATCHES_TOPIC", "matches"),
			GroupID:      getEnv("KAFKA_GROUP_ID", "matching-engine-group"),
		},
	}

	if err := config.validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return config, nil
}

func (c *Config) validate() error {
	if len(c.Kafka.Brokers) == 0 {
		return fmt.Errorf("at least one Kafka broker is required")
	}
	if c.Kafka.OrdersTopic == "" {
		return fmt.Errorf("orders topic is required")
	}
	if c.Kafka.MatchesTopic == "" {
		return fmt.Errorf("matches topic is required")
	}
	return nil
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists && value != "" {
		return value
	}
	return defaultValue
}
