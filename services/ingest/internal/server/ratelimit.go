package server

import (
	"context"
	_ "embed"
	"net/http"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

//go:embed ratelimit.lua
var ratelimitScript string

type RateLimiter struct {
	rdb    *redis.Client
	script *redis.Script
	limit  int
	window time.Duration
}

func NewRateLimiter(rdb *redis.Client, limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		rdb:    rdb,
		script: redis.NewScript(ratelimitScript),
		limit:  limit,
		window: window,
	}
}

func (rl *RateLimiter) Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user := r.Header.Get("X-User-Id")
			if user == "" {
				next.ServeHTTP(w, r)
				return
			}

			bucket := time.Now().UTC().Truncate(rl.window).Unix()
			key := "rl:" + user + ":" + strconv.FormatInt(bucket, 10)

			ctx, cancel := context.WithTimeout(r.Context(), 50*time.Millisecond)
			defer cancel()

			n, err := rl.script.Run(ctx, rl.rdb, []string{key}, rl.limit, int(rl.window.Seconds())).Int()
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}

			w.Header().Set("X-RateLimit-Limit", strconv.Itoa(rl.limit))
			w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(max(0, rl.limit-int(n))))
			w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(bucket+int64(rl.window.Seconds()), 10))

			if n > rl.limit {
				w.Header().Set("Retry-After", strconv.Itoa(int(rl.window.Seconds())))
				http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
