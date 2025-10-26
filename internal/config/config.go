package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	App   AppConfig
	DB    DBConfig
	Redis RedisConfig
	Log   LogConfig
}

type AppConfig struct {
	Addr string
	Env  string
}

type DBConfig struct {
	User     string
	Password string
	Host     string
	Port     string
	Name     string
}

type RedisConfig struct {
	Addr     string
	DB       int
	Password string
}

type LogConfig struct {
	Level string
}

func Load() (*Config, error) {
	c := &Config{}

	// App
	c.App.Addr = getEnv("PORT", ":8080")
	c.App.Env = getEnv("APP_ENV", "development")

	// Log
	c.Log.Level = getEnv("LOG_LEVEL", "debug")

	// DB
	c.DB.Host = getEnv("DB_HOST", "db")
	c.DB.Port = getEnv("DB_PORT", "5432")
	c.DB.Name = getEnv("DB_NAME", "postgres")

	s, err := readSecretFile("/run/secrets/db_user")
	if err != nil {
		return nil, fmt.Errorf("read db user file: %w", err)
	}
	c.DB.User = s

	s, err = readSecretFile("/run/secrets/db_password")
	if err != nil {
		return nil, fmt.Errorf("read db password file: %w", err)
	}
	c.DB.Password = s

	// Redis
	c.Redis.Addr = getEnv("REDIS_ADDR", "redis:6379")
	c.Redis.DB, _ = strconv.Atoi(getEnv("REDIS_DB", "0"))

	s, err = readSecretFile("/run/secrets/redis_password")
	if err != nil {
		return nil, fmt.Errorf("read redis password file: %w", err)
	}
	c.Redis.Password = s

	return c, nil
}

func getEnv(key, def string) string {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	return v
}

func readSecretFile(path string) (string, error) {
	if path == "" {
		return "", nil
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(b)), nil
}
