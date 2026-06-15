package server

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/google/uuid"

	"ingest/internal/queue"
)

const (
	maxUploadBytes = 5 << 30
)

var allowedTypes = map[string]bool{
	"video/mp4":       true,
	"video/quicktime": true,
	"video/webm":      true,
}

type BlobStore interface {
	Put(ctx context.Context, key string, src io.Reader, size int64, contentType string) (int64, error)
}

type UploadHandler struct {
	storage  BlobStore
	producer *queue.JobProducer
	maxBytes int64
}

func NewUploadHandler(s BlobStore, p *queue.JobProducer) *UploadHandler {
	return &UploadHandler{storage: s, producer: p, maxBytes: maxUploadBytes}
}

func (h *UploadHandler) Upload(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := r.Header.Get("X-User-Id")

	r.Body = http.MaxBytesReader(w, r.Body, h.maxBytes)

	if err := r.ParseMultipartForm(32 << 20); err != nil {
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			writeError(w, http.StatusRequestEntityTooLarge, "file too large", maxErr.Limit)
			return
		}
		writeError(w, http.StatusBadRequest, "bad multipart", err)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, "missing 'file' field", err)
		return
	}
	defer file.Close()

	contentType := header.Header.Get("Content-Type")
	if !allowedTypes[contentType] {
		writeError(w, http.StatusUnsupportedMediaType, "unsupported content type", nil)
		return
	}

	jobID := uuid.NewString()
	key := fmt.Sprintf("uploads/%s/%s", userID, jobID)

	size, err := h.storage.Put(ctx, key, file, header.Size, contentType)
	if err != nil {
		slog.ErrorContext(ctx, "blob store put failed", "err", err, "key", key)
		writeError(w, http.StatusBadGateway, "storage unavailable", nil)
		return
	}

	job := queue.Job{
		JobID:       jobID,
		OwnerID:     userID,
		InputKey:    key,
		Size:        size,
		ContentType: contentType,
		Kind:        queue.JobKindTranscode,
		EnqueuedAt:  queue.NowMillis(),
		Attempts:    0,
	}

	if err := h.producer.Enqueue(ctx, job); err != nil {
		slog.ErrorContext(ctx, "enqueue failed; orphan blob", "err", err, "key", key, "job_id", jobID)
		writeError(w, http.StatusBadGateway, "queue unavailable", nil)
		return
	}

	w.Header().Set("Location", "/v1/jobs/"+jobID)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	fmt.Fprintf(w, `{"job_id":%q,"status":"queued"}`, jobID)
}
