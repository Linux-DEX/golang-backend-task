package worker

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/sarabjeet/golang-backend-task/internal/metrics"
	"github.com/sarabjeet/golang-backend-task/internal/models"
	"github.com/sarabjeet/golang-backend-task/internal/parser"
	"github.com/sarabjeet/golang-backend-task/internal/storage"
)

type Processor struct {
	storage    storage.StorageInterface
	maxRetries int
}

func NewProcessor(storage storage.StorageInterface, maxRetries int) *Processor {
	return &Processor{
		storage:    storage,
		maxRetries: maxRetries,
	}
}

func (p *Processor) ProcessJob(ctx context.Context, jobID, fileContent string) error {
	log.Printf("[Worker] Starting to process job: %s", jobID)
	job, err := p.storage.GetJob(ctx, jobID)
	if err != nil {
		return fmt.Errorf("failed to get job: %w", err)
	}
	if job.Status == models.StatusCompleted {
		log.Printf("[Worker] Job %s is already completed, skipping", jobID)
		return nil
	}

	if job.Status == models.StatusFailed && job.RetryCount >= p.maxRetries {
		log.Printf("[Worker] Job %s has exceeded max retries (%d), skipping", jobID, p.maxRetries)
		return nil
	}
	err = p.storage.UpdateJobStatus(ctx, jobID, models.StatusProcessing)
	if err != nil {
		return fmt.Errorf("failed to update job status to processing: %w", err)
	}
	result, parseErr := parser.ParseEDI(fileContent)

	if parseErr != nil {
		log.Printf("[Worker] Job %s parsing failed: %v", jobID, parseErr)
		err = p.storage.IncrementRetryCount(ctx, jobID)
		if err != nil {
			log.Printf("[Worker] Failed to increment retry count for job %s: %v", jobID, err)
		}
		job, err = p.storage.GetJob(ctx, jobID)
		if err != nil {
			return fmt.Errorf("failed to get job after retry: %w", err)
		}

		if job.RetryCount >= p.maxRetries {
			log.Printf("[Worker] Job %s reached max retries (%d), marking as failed", jobID, p.maxRetries)
			err = p.storage.UpdateJobWithResult(
				ctx,
				jobID,
				models.StatusFailed,
				nil,
				fmt.Sprintf("Failed after %d attempts: %v", job.RetryCount, parseErr),
			)
			if err != nil {
				return fmt.Errorf("failed to update job as failed: %w", err)
			}

			metrics.RecordJobFailed()
			return fmt.Errorf("job failed after %d attempts: %w", job.RetryCount, parseErr)
		}
		log.Printf("[Worker] Job %s will be retried (attempt %d/%d)", jobID, job.RetryCount, p.maxRetries)
		err = p.storage.UpdateJobWithResult(
			ctx,
			jobID,
			models.StatusPending,
			nil,
			fmt.Sprintf("Retry %d/%d: %v", job.RetryCount, p.maxRetries, parseErr),
		)
		if err != nil {
			return fmt.Errorf("failed to update job for retry: %w", err)
		}
		backoffDuration := time.Duration(1<<uint(job.RetryCount)) * time.Second
		log.Printf("[Worker] Waiting %v before retry for job %s", backoffDuration, jobID)
		time.Sleep(backoffDuration)

		return fmt.Errorf("job will be retried: %w", parseErr)
	}
	log.Printf("[Worker] Job %s processed successfully: %d claims, total amount: %.2f",
		jobID, result.Summary.TotalClaims, result.Summary.TotalAmount)

	err = p.storage.UpdateJobWithResult(ctx, jobID, models.StatusCompleted, result, "")
	if err != nil {
		return fmt.Errorf("failed to update job with result: %w", err)
	}

	metrics.RecordJobCompleted()

	log.Printf("[Worker] Job %s completed successfully", jobID)
	return nil
}

func (p *Processor) ProcessJobWithRetry(ctx context.Context, jobID, fileContent string) error {
	var lastErr error

	for attempt := 0; attempt <= p.maxRetries; attempt++ {
		if attempt > 0 {
			backoff := time.Duration(1<<uint(attempt-1)) * time.Second
			log.Printf("[Worker] Retry attempt %d/%d for job %s after %v", attempt, p.maxRetries, jobID, backoff)
			time.Sleep(backoff)
		}

		err := p.ProcessJob(ctx, jobID, fileContent)
		if err == nil {
			return nil
		}

		lastErr = err
		log.Printf("[Worker] Attempt %d/%d failed for job %s: %v", attempt+1, p.maxRetries+1, jobID, err)
	}
	log.Printf("[Worker] Job %s failed after all retry attempts", jobID)
	metrics.RecordJobFailed()
	return lastErr
}
