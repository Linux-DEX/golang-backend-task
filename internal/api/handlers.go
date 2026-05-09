package api

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sarabjeet/golang-backend-task/internal/metrics"
	"github.com/sarabjeet/golang-backend-task/internal/models"
	"github.com/sarabjeet/golang-backend-task/internal/queue"
	"github.com/sarabjeet/golang-backend-task/internal/storage"
)

type Handler struct {
	storage *storage.Storage
	queue   *queue.Queue
}

func NewHandler(storage *storage.Storage, queue *queue.Queue) *Handler {
	return &Handler{
		storage: storage,
		queue:   queue,
	}
}

func (h *Handler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "healthy",
		"time":   time.Now().Format(time.RFC3339),
	})
}

func (h *Handler) CreateJob(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		log.Printf("Error parsing file: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "File is required",
		})
		return
	}
	defer file.Close()
	fileName := header.Filename
	if len(fileName) < 4 || fileName[len(fileName)-4:] != ".edi" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid file format. Only .edi files are accepted",
		})
		return
	}
	fileContent, err := io.ReadAll(file)
	if err != nil {
		log.Printf("Error reading file: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to read file content",
		})
		return
	}

	if len(fileContent) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "File is empty",
		})
		return
	}
	jobID := uuid.New().String()
	job := models.NewJob(jobID, fileName)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = h.storage.SaveJob(ctx, job)
	if err != nil {
		log.Printf("Error saving job: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create job",
		})
		return
	}
	jobMsg := &queue.JobMessage{
		JobID:       jobID,
		FileName:    fileName,
		FileContent: string(fileContent),
		CreatedAt:   time.Now(),
	}

	err = h.queue.Enqueue(ctx, jobMsg)
	if err != nil {
		log.Printf("Error enqueueing job: %v", err)
		job.UpdateStatus(models.StatusFailed, fmt.Sprintf("Failed to enqueue job: %v", err))
		_ = h.storage.UpdateJob(ctx, job)

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to queue job for processing",
		})
		return
	}
	metrics.RecordJobCreated()

	log.Printf("Job created successfully: %s (file: %s)", jobID, fileName)

	c.JSON(http.StatusCreated, gin.H{
		"job_id":  jobID,
		"message": "Job created successfully and queued for processing",
	})
}

func (h *Handler) GetJobStatus(c *gin.Context) {
	jobID := c.Param("job_id")

	if jobID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Job ID is required",
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	job, err := h.storage.GetJob(ctx, jobID)
	if err != nil {
		log.Printf("Error getting job %s: %v", jobID, err)
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Job not found",
		})
		return
	}

	response := gin.H{
		"job_id":      job.JobID,
		"status":      job.Status,
		"retry_count": job.RetryCount,
		"created_at":  job.CreatedAt.Format(time.RFC3339),
		"updated_at":  job.UpdatedAt.Format(time.RFC3339),
	}

	if job.Error != "" {
		response["error"] = job.Error
	}

	c.JSON(http.StatusOK, response)
}

func (h *Handler) GetJobResult(c *gin.Context) {
	jobID := c.Param("job_id")

	if jobID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Job ID is required",
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	job, err := h.storage.GetJob(ctx, jobID)
	if err != nil {
		log.Printf("Error getting job %s: %v", jobID, err)
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Job not found",
		})
		return
	}
	switch job.Status {
	case models.StatusPending:
		c.JSON(http.StatusAccepted, gin.H{
			"status":  "pending",
			"message": "Job is waiting to be processed",
		})
		return
	case models.StatusProcessing:
		c.JSON(http.StatusAccepted, gin.H{
			"status":  "processing",
			"message": "Job is currently being processed",
		})
		return
	case models.StatusFailed:
		c.JSON(http.StatusOK, gin.H{
			"status": "failed",
			"error":  job.Error,
		})
		return
	case models.StatusCompleted:
		if job.Result == nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Job completed but result is missing",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status":  "completed",
			"claims":  job.Result.Claims,
			"summary": job.Result.Summary,
		})
		return
	default:
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Unknown job status",
		})
		return
	}
}

func (h *Handler) GetMetrics(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	queueSize, err := h.queue.QueueLength(ctx)
	if err != nil {
		log.Printf("Error getting queue size: %v", err)
		queueSize = -1
	}

	c.JSON(http.StatusOK, gin.H{
		"queue_size": queueSize,
		"time":       time.Now().Format(time.RFC3339),
	})
}
