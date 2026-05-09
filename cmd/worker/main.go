package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sarabjeet/golang-backend-task/internal/config"
	"github.com/sarabjeet/golang-backend-task/internal/queue"
	"github.com/sarabjeet/golang-backend-task/internal/storage"
	"github.com/sarabjeet/golang-backend-task/internal/worker"
)

func main() {
	log.Println("Starting EDI Worker Service")

	cfg := config.Load()
	log.Printf("Configuration loaded - MongoDB: %s, Redis: %s:%s\n",
		cfg.MongoDB.URI, cfg.Redis.Host, cfg.Redis.Port)

	// Initialize storage
	store, err := storage.New(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := store.Close(ctx); err != nil {
			log.Printf("Error closing storage: %v", err)
		}
	}()

	// Initialize queue
	q, err := queue.New(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize queue: %v", err)
	}
	defer func() {
		if err := q.Close(); err != nil {
			log.Printf("Error closing queue: %v", err)
		}
	}()

	// Initialize processor
	processor := worker.NewProcessor(store, cfg.Worker.MaxRetries)

	// Start metrics endpoint
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		log.Println("Metrics server started on :9091")
		if err := http.ListenAndServe(":9091", nil); err != nil {
			log.Printf("Metrics server error: %v", err)
		}
	}()

	// Setup graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start worker loop
	go workerLoop(ctx, q, processor, cfg)

	// Wait for shutdown signal
	<-sigChan
	log.Println("Shutdown signal received, stopping worker...")
	cancel()

	// Give some time for graceful shutdown
	time.Sleep(5 * time.Second)
	log.Println("Worker service stopped")
}

func workerLoop(ctx context.Context, q *queue.Queue, processor *worker.Processor, cfg *config.Config) {
	log.Println("Worker loop started")
	pollInterval := time.Duration(cfg.Worker.PollInterval) * time.Second

	for {
		select {
		case <-ctx.Done():
			log.Println("Worker loop stopped")
			return
		default:
			// Dequeue job
			jobData, err := q.Dequeue(ctx)
			if err != nil {
				log.Printf("Error dequeuing job: %v", err)
				time.Sleep(pollInterval)
				continue
			}

			if jobData == "" {
				time.Sleep(pollInterval)
				continue
			}

			// Parse job message
			var jobMsg queue.JobMessage
			if err := json.Unmarshal([]byte(jobData), &jobMsg); err != nil {
				// Fallback to legacy format (just job ID)
				processJob(ctx, processor, jobData, "")
			} else {
				processJob(ctx, processor, jobMsg.JobID, jobMsg.FileContent)
			}
		}
	}
}

func processJob(ctx context.Context, processor *worker.Processor, jobID, fileContent string) {
	startTime := time.Now()
	log.Printf("Processing job: %s", jobID)

	err := processor.ProcessJob(ctx, jobID, fileContent)
	duration := time.Since(startTime)

	if err != nil {
		log.Printf("Job %s failed after %v: %v", jobID, duration, err)
	} else {
		log.Printf("Job %s completed successfully in %v", jobID, duration)
	}
}
