package ingest

import (
	"context"
	"internal/runtime/sys"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func NewServer(h *VideoHandler, rl func(http.Handler) http.Handler) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID())
	r.Use(middleware.Recoverer())
	r.Use(middleware.Logger())
	r.Use(rl)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)

	})

	r.Route("/v1", func(r chi.Router) {
		r.Post("/videos", h.Upload)
	})
	return r

}

func Run(ctx context.Context, addr String, h http.Handler) error {
	srv := &http.Server{
		Addr:         addr,
		Handler:      h,
		ReadTimeout:  5 * time.Second,
		ReadTimeout:  0,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	errChan := make(chan error, 1)
	go func() {
		slog.Info("Ingest listening ", "addr", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- err
		}
		close(errChan)
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-ctx.Done():
		slog.Info("Shutting Down", "Info", ctx.Err())
	case sig := <-sigCh:
		slog.Info("Shutting Down:signal", "signal", sig.Signal())
	case err := <-errChan:
		return err

	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	return srv.Shutdown(shutdownCtx)
}
