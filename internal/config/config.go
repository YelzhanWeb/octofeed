package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type EnvConfig struct{}

func NewEnvConfig() *EnvConfig {
	return &EnvConfig{}
}

func (c *EnvConfig) GetPostgresHost() string {
	return getEnv("POSTGRES_HOST", "localhost")
}

func (c *EnvConfig) GetPostgresPort() string {
	return getEnv("POSTGRES_PORT", "5432")
}

func (c *EnvConfig) GetPostgresUser() string {
	return getEnv("POSTGRES_USER", "postgres")
}

func (c *EnvConfig) GetPostgresPassword() string {
	return getEnv("POSTGRES_PASSWORD", "postgres")
}

func (c *EnvConfig) GetPostgresDBName() string {
	return getEnv("POSTGRES_DBNAME", "rsshub")
}

func (c *EnvConfig) GetPostgresSSLMode() string {
	return getEnv("POSTGRES_SSLMODE", "disable")
}

func (c *EnvConfig) GetDefaultInterval() time.Duration {
	intervalStr := getEnv("CLI_APP_TIMER_INTERVAL", "3m")
	duration, err := time.ParseDuration(intervalStr)
	if err != nil {
		return 3 * time.Minute
	}
	return duration
}

func (c *EnvConfig) GetDefaultWorkersCount() int {
	workersStr := getEnv("CLI_APP_WORKERS_COUNT", "3")
	workers, err := strconv.Atoi(workersStr)
	if err != nil {
		return 3
	}
	return workers
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func (c *EnvConfig) GetDSN() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.GetPostgresHost(),
		c.GetPostgresPort(),
		c.GetPostgresUser(),
		c.GetPostgresPassword(),
		c.GetPostgresDBName(),
		c.GetPostgresSSLMode(),
	)
}
