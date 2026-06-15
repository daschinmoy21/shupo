package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Port           string
	RedisAddr      string
	RedisPassword  string
	MinIOEndpoint  string
	MinIOAccessKey string
	MinIOSecretKey string
	MinIOBucket    string
	RateLimit      int
	RateWindow     time.Duration
}

func LoadConfig() Config {
	return Config{
		Port:           getenv("PORT", "8080"),
		RedisAddr:      getenv("REDIS_ADDR", "localhost:6379"),
		RedisPassword:  os.Getenv("REDIS_PASSWORD"),
		MinIOEndpoint:  getenv("MINIO_ENDPOINT", "localhost:9000"),
		MinIOAccessKey: getenv("MINIO_ACCESS_KEY", "minioadmin"),
		MinIOSecretKey: getenv("MINIO_SECRET_KEY", "minioadmin"),
		MinIOBucket:    getenv("MINIO_BUCKET", "uploads"),
		RateLimit:      getenvInt("RATE_LIMIT", 10),
		RateWindow:     time.Duration(getenvInt("RATE_WINDOW_SECONDS", 60)) * time.Second,
	}
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func getenvInt(k string, def int) int {
	if v := os.Getenv(k); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}
