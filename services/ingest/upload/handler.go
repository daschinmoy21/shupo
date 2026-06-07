package upload

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os/user"

	"github.com/google/uuid"
)

const (
	maxUploadSize = 5 << 30 //5gb
	allowedTypes  = "video/mp4,video/mkv,video/quicktime,video/webm"
)

type VideoHandler struct {
	storage  BlobStorage
	producer *JobProducer
	maxBytes int64
}

func NewVideoHandler(s BlobStorage, p *JobProducer) *VideoHandler {
	return &VideoHandler{storage: s, producer: p, maxBytes: maxUploadSize}
}

func (h *VideoHandler) Upload(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := r.Header.Get("X-User-Id")

	r.Body = http.MaxBytesReader(w, r.Body, h.maxBytes)
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			writeError(w, http.StatusRequestEntityTooLarge, "File too large", maxErr.Limit)
			return
		}
		writeError(w, http.StatusBadRequest, "bad multipart", err)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, "Missing file field", err)
		return
	}
	defer file.Close()

	if !validContentType(header.Header.Get("Content-Type")) {
		writeError(w, http.StatusBadRequest, "Unsupported content type", err)
		return
	}
	jobId := uuid.NewString()
	key := fmt.Sprintf("upload/%s/%s", userId, jobId)

	size, err := h.storage.Put(ctx, key, file, header.Size, header.Header.Get("Content-Type"))
	if err != nil {
		writeError(w, http.StatusBadGateway, "Storage unavailable", err)
		return
	}

	job := Job{
		ID:         jobID,
		OwnerID:    userID,
		Key:        key,
		Size:       size,
		ContentTyp: header.Header.Get("Content-Type"),
		Kind:       "transcode",
		EnqueuedAt: time.Now().UnixMilli(),
	}

	if err := h.producer.Enqueue(ctx, job); err != nil {
		slog.ErrorContext(ctx, "Enqueue failed;Orphan blob ", "err", err, "key", key, "jobId", jobId)
		writeError(w, http.StatusBadGateway, "queue unavailable", err)
		return
	}
	w.Header().Set("Location", "/v1/jobs"+jobId)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	fmt.Fprintf(w, `{"jobId":%q,"status":"queued"}`, jobID)
}
