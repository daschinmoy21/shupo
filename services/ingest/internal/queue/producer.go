package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type JobKind string

const (
	JobKindThumbnail JobKind = "thumbnail"
	JobKindTranscode JobKind = "transcode"
	JobKindProbe     JobKind = "probe"
)

type Job struct {
	JobID       string  `json:"job_id"`
	OwnerID     string  `json:"owner_id"`
	InputKey    string  `json:"input_key"`
	Size        int64   `json:"size"`
	ContentType string  `json:"content_type"`
	Kind        JobKind `json:"kind"`
	EnqueuedAt  int64   `json:"enqueued_at"`
	Attempts    int     `json:"attempts"`
}

const (
	streamJobs   = "jobs.stream"
	streamDLQ    = "jobs.dlq"
	jobKeyPrefix = "job:"
)

type JobProducer struct {
	rdb *redis.Client
}

func NewJobProducer(rdb *redis.Client) *JobProducer {
	return &JobProducer{rdb: rdb}
}

func (p *JobProducer) Enqueue(ctx context.Context, j Job) error {
	payload, err := json.Marshal(j)
	if err != nil {
		return fmt.Errorf("marshal job: %w", err)
	}

	pipe := p.rdb.Pipeline()
	pipe.XAdd(ctx, &redis.XAddArgs{
		Stream: streamJobs,
		Values: map[string]interface{}{
			"payload": string(payload),
		},
	})
	pipe.HSet(ctx, jobKeyPrefix+j.JobID, map[string]interface{}{
		"state":      "queued",
		"ts":         time.Now().UnixMilli(),
		"owner_id":   j.OwnerID,
		"input_key":  j.InputKey,
		"size":       j.Size,
		"attempts":   0,
	})
	pipe.Publish(ctx, "job.events", fmt.Sprintf(
		`{"job_id":%q,"owner_id":%q,"state":"queued"}`, j.JobID, j.OwnerID,
	))

	_, err = pipe.Exec(ctx)
	return err
}

func NowMillis() int64 { return time.Now().UnixMilli() }
