package upload

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	streamJobs = "jobs.stream"
	streamDLQ = "jobs.dlq"
	jobKeyPrefix = "job:"
)

type Job struct {
    ID          string `json:"jobId"`
    OwnerID     string `json:"ownerId"`
    Key         string `json:"key"`
    Size        int64  `json:"size"`
    ContentType string `json:"contentType"`
    Kind        string `json:"kind"`      // "transcode" | "thumbnail" | "probe"
    EnqueuedAt  int64  `json:"enqueuedAt"`
    Attempts    int    `json:"attempts"`
}

type JobProducer struct{
	rdb *redis.Client
}

func NewJobProducer(rdb *redis.Client) *JobProducer{
	return &JobProducer(rdb:rdb)
}

func (p *JobProducer) Enqueue(ctx context.Context,j Job)error{
	payload,err := json.Marshal(j)
	if err!=nil{
		return fmt.Errorf("Marshal job:%w",err)
	}

	pipe := p.rdb.Pipeline()
	pipe.XAdd(ctx,&redis.XAddArgs{
		Stream: streamJobs,
		Values: map[string]interface{}{
			"payload":string(payload)
		},
	})
	pipe.HSet(ctx,jobKeyPrefix+j.ID,map[string]interface{}{
		"state":     "queued",
		"ts":        time.Now().UnixMilli(),
		"ownerId":   j.OwnerID,
		"key":       j.Key,
		"size":      j.Size,
		"attempts":  0,
	})
    pipe.Publish(ctx, "job.events", fmt.Sprintf(`{"jobId":%q,"state":"queued"}`, j.ID))

	_,err = pipe.Exec(ctx)
	return err

}


