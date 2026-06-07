package ratelimit

import (
	"context"
	_ "embed"
	"net/http"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

//go:embed ratelimit.lua
var rateLimitScript string

type RateLimiter struct {
	rdb    *redis.Client
	script *redis.Script
	limit  int
	window time.Duration
}

func NewRateLimiter(rdb *redis.Client, limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		rdb:    rdb,
		script: redis.NewScript(rateLimitScript),
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

			res, err := rl.script.Run(ctx, rl.rdb,
				[]string{key},
				rl.limit,
				int(rl.window.Seconds()),
			).Slice()
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}

			count, _ := res[0].(int64)
			limit, _ := res[1].(int64)

			w.Header().Set("X-RateLimit-Limit", strconv.FormatInt(limit, 10))
			w.Header().Set("X-RateLimit-Remaining", strconv.FormatInt(max(0, limit-count), 10))
			w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(bucket+int64(rl.window.Seconds()), 10))

			if count > limit {
				w.Header().Set("Retry-After", strconv.Itoa(int(rl.window.Seconds())))
				http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
