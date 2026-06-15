package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/redis/go-redis/v9"

	"ingest/internal/config"
	"ingest/internal/queue"
	"ingest/internal/server"
	"ingest/internal/storage"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	cfg := config.LoadConfig()

	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       0,
	})
	defer rdb.Close()

	store, err := storage.NewMinioStore(cfg.MinIOEndpoint, cfg.MinIOAccessKey, cfg.MinIOSecretKey, cfg.MinIOBucket)
	if err != nil {
		slog.Error("blob store init failed", "err", err)
		os.Exit(1)
	}

	producer := queue.NewJobProducer(rdb)
	limiter := server.NewRateLimiter(rdb, cfg.RateLimit, cfg.RateWindow)
	handler := server.NewUploadHandler(store, producer)

	srv := server.NewServer(handler, limiter.Middleware())
	if err := server.Run(context.Background(), ":"+cfg.Port, srv.Handler()); err != nil {
		slog.Error("server exited with error", "err", err)
		os.Exit(1)
	}
}
