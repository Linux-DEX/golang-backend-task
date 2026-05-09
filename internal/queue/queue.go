package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/sarabjeet/golang-backend-task/internal/config"
)

const (
	jobQueueKey = "edi:jobs:queue"
)

type Queue struct {
	client *redis.Client
}

type RedisQueue = Queue

type JobMessage struct {
	JobID       string    `json:"job_id"`
	FileName    string    `json:"file_name"`
	FileContent string    `json:"file_content"`
	CreatedAt   time.Time `json:"created_at"`
}

func New(cfg *config.Config) (*Queue, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &Queue{
		client: client,
	}, nil
}

func NewRedisQueue(cfg *config.RedisConfig) (*RedisQueue, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &Queue{
		client: client,
	}, nil
}

func (q *Queue) Close() error {
	return q.client.Close()
}

func (q *Queue) Enqueue(ctx context.Context, jobMsg *JobMessage) error {
	data, err := json.Marshal(jobMsg)
	if err != nil {
		return fmt.Errorf("failed to marshal job message: %w", err)
	}

	err = q.client.RPush(ctx, jobQueueKey, data).Err()
	if err != nil {
		return fmt.Errorf("failed to enqueue job: %w", err)
	}
	return nil
}
func (q *Queue) EnqueueJobID(ctx context.Context, jobID string) error {
	err := q.client.RPush(ctx, jobQueueKey, jobID).Err()
	if err != nil {
		return fmt.Errorf("failed to enqueue job: %w", err)
	}
	return nil
}

func (q *Queue) Dequeue(ctx context.Context) (string, error) {
	result, err := q.client.LPop(ctx, jobQueueKey).Result()
	if err != nil {
		if err == redis.Nil {
			return "", nil
		}
		return "", fmt.Errorf("failed to dequeue job: %w", err)
	}
	return result, nil
}

func (q *Queue) DequeueBlocking(ctx context.Context, timeout time.Duration) (string, error) {
	result, err := q.client.BLPop(ctx, timeout, jobQueueKey).Result()
	if err != nil {
		if err == redis.Nil {
			return "", nil
		}
		return "", fmt.Errorf("failed to dequeue job: %w", err)
	}
	if len(result) < 2 {
		return "", nil
	}
	return result[1], nil
}
func (q *Queue) QueueLength(ctx context.Context) (int64, error) {
	length, err := q.client.LLen(ctx, jobQueueKey).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get queue length: %w", err)
	}
	return length, nil
}
func (q *Queue) Size(ctx context.Context) (int64, error) {
	return q.QueueLength(ctx)
}
