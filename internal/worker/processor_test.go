package worker

import (
	"context"
	"testing"
	"time"

	"github.com/sarabjeet/golang-backend-task/internal/models"
	"github.com/sarabjeet/golang-backend-task/internal/storage"
)

type MockStorage struct {
	jobs map[string]*models.Job
}

func NewMockStorage() *MockStorage {
	return &MockStorage{
		jobs: make(map[string]*models.Job),
	}
}

func (m *MockStorage) GetJob(ctx context.Context, jobID string) (*models.Job, error) {
	job, ok := m.jobs[jobID]
	if !ok {
		return nil, storage.ErrJobNotFound
	}
	return job, nil
}

func (m *MockStorage) SaveJob(ctx context.Context, job *models.Job) error {
	m.jobs[job.JobID] = job
	return nil
}

func (m *MockStorage) UpdateJob(ctx context.Context, job *models.Job) error {
	if _, ok := m.jobs[job.JobID]; !ok {
		return storage.ErrJobNotFound
	}
	m.jobs[job.JobID] = job
	return nil
}

func (m *MockStorage) UpdateJobStatus(ctx context.Context, jobID string, status models.JobStatus) error {
	job, ok := m.jobs[jobID]
	if !ok {
		return storage.ErrJobNotFound
	}
	job.Status = status
	job.UpdatedAt = time.Now()
	return nil
}

func (m *MockStorage) UpdateJobWithResult(ctx context.Context, jobID string, status models.JobStatus, result *models.Result, errorMsg string) error {
	job, ok := m.jobs[jobID]
	if !ok {
		return storage.ErrJobNotFound
	}
	job.Status = status
	job.Result = result
	job.Error = errorMsg
	job.UpdatedAt = time.Now()
	return nil
}

func (m *MockStorage) IncrementRetryCount(ctx context.Context, jobID string) error {
	job, ok := m.jobs[jobID]
	if !ok {
		return storage.ErrJobNotFound
	}
	job.RetryCount++
	job.UpdatedAt = time.Now()
	return nil
}

func (m *MockStorage) CreateIndexes(ctx context.Context) error {
	return nil
}

func (m *MockStorage) Close(ctx context.Context) error {
	return nil
}

func TestNewProcessor(t *testing.T) {
	storage := NewMockStorage()
	processor := NewProcessor(storage, 3)

	if processor == nil {
		t.Fatal("Expected processor, got nil")
	}

	if processor.maxRetries != 3 {
		t.Errorf("Expected maxRetries 3, got %d", processor.maxRetries)
	}
}

func TestProcessJob_Success(t *testing.T) {
	storage := NewMockStorage()
	processor := NewProcessor(storage, 3)

	jobID := "test-job-1"
	fileContent := "CLAIM*CLM001*MEM123*2500\nCLAIM*CLM002*MEM456*3000"

	// Create a job in storage
	job := models.NewJob(jobID, "test.edi")
	ctx := context.Background()
	storage.SaveJob(ctx, job)

	// Process the job
	err := processor.ProcessJob(ctx, jobID, fileContent)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Verify job was updated
	updatedJob, _ := storage.GetJob(ctx, jobID)
	if updatedJob.Status != models.StatusCompleted {
		t.Errorf("Expected status completed, got %s", updatedJob.Status)
	}

	if updatedJob.Result == nil {
		t.Fatal("Expected result, got nil")
	}

	if updatedJob.Result.Summary.TotalClaims != 2 {
		t.Errorf("Expected 2 claims, got %d", updatedJob.Result.Summary.TotalClaims)
	}

	if updatedJob.Result.Summary.TotalAmount != 5500 {
		t.Errorf("Expected total 5500, got %.2f", updatedJob.Result.Summary.TotalAmount)
	}
}

func TestProcessJob_ParseError(t *testing.T) {
	storage := NewMockStorage()
	processor := NewProcessor(storage, 3)

	jobID := "test-job-2"
	fileContent := "INVALID*DATA*FORMAT"

	// Create a job in storage
	job := models.NewJob(jobID, "test.edi")
	ctx := context.Background()
	storage.SaveJob(ctx, job)

	// Process the job
	err := processor.ProcessJob(ctx, jobID, fileContent)
	if err == nil {
		t.Error("Expected error for invalid content, got nil")
	}

	// Verify job retry count was incremented
	updatedJob, _ := storage.GetJob(ctx, jobID)
	if updatedJob.RetryCount != 1 {
		t.Errorf("Expected retry count 1, got %d", updatedJob.RetryCount)
	}

	if updatedJob.Status != models.StatusPending {
		t.Errorf("Expected status pending for retry, got %s", updatedJob.Status)
	}
}

func TestProcessJob_MaxRetriesReached(t *testing.T) {
	storage := NewMockStorage()
	processor := NewProcessor(storage, 3)

	jobID := "test-job-3"
	fileContent := "INVALID*DATA*FORMAT"

	// Create a job with max retries already reached
	job := models.NewJob(jobID, "test.edi")
	job.RetryCount = 3
	ctx := context.Background()
	storage.SaveJob(ctx, job)

	// Process the job
	err := processor.ProcessJob(ctx, jobID, fileContent)
	if err == nil {
		t.Error("Expected error for max retries, got nil")
	}

	// Verify job was marked as failed
	updatedJob, _ := storage.GetJob(ctx, jobID)
	if updatedJob.Status != models.StatusFailed {
		t.Errorf("Expected status failed, got %s", updatedJob.Status)
	}

	if updatedJob.Error == "" {
		t.Error("Expected error message, got empty string")
	}
}

func TestProcessJob_AlreadyCompleted(t *testing.T) {
	storage := NewMockStorage()
	processor := NewProcessor(storage, 3)

	jobID := "test-job-4"
	fileContent := "CLAIM*CLM001*MEM123*2500"

	// Create a job that's already completed
	job := models.NewJob(jobID, "test.edi")
	job.Status = models.StatusCompleted
	ctx := context.Background()
	storage.SaveJob(ctx, job)

	// Process the job
	err := processor.ProcessJob(ctx, jobID, fileContent)
	if err != nil {
		t.Errorf("Expected no error for completed job, got: %v", err)
	}

	// Verify job status didn't change
	updatedJob, _ := storage.GetJob(ctx, jobID)
	if updatedJob.Status != models.StatusCompleted {
		t.Errorf("Expected status to remain completed, got %s", updatedJob.Status)
	}
}

func TestProcessJob_JobNotFound(t *testing.T) {
	storage := NewMockStorage()
	processor := NewProcessor(storage, 3)

	jobID := "non-existent-job"
	fileContent := "CLAIM*CLM001*MEM123*2500"
	ctx := context.Background()

	// Process non-existent job
	err := processor.ProcessJob(ctx, jobID, fileContent)
	if err == nil {
		t.Error("Expected error for non-existent job, got nil")
	}
}

func TestProcessJob_EmptyFileContent(t *testing.T) {
	storage := NewMockStorage()
	processor := NewProcessor(storage, 3)

	jobID := "test-job-5"
	fileContent := ""

	// Create a job in storage
	job := models.NewJob(jobID, "test.edi")
	ctx := context.Background()
	storage.SaveJob(ctx, job)

	// Process the job with empty content
	err := processor.ProcessJob(ctx, jobID, fileContent)
	if err == nil {
		t.Error("Expected error for empty content, got nil")
	}

	// Verify job retry count was incremented
	updatedJob, _ := storage.GetJob(ctx, jobID)
	if updatedJob.RetryCount != 1 {
		t.Errorf("Expected retry count 1, got %d", updatedJob.RetryCount)
	}
}

func TestProcessJob_RetryLogic(t *testing.T) {
	storage := NewMockStorage()
	processor := NewProcessor(storage, 3)

	jobID := "test-job-6"
	fileContent := "INVALID*CONTENT"

	// Create a job in storage
	job := models.NewJob(jobID, "test.edi")
	ctx := context.Background()
	storage.SaveJob(ctx, job)

	// Process multiple times to trigger retries
	for i := 0; i < 3; i++ {
		err := processor.ProcessJob(ctx, jobID, fileContent)
		if err == nil {
			t.Errorf("Expected error on attempt %d, got nil", i+1)
		}

		updatedJob, _ := storage.GetJob(ctx, jobID)
		expectedRetryCount := i + 1
		if updatedJob.RetryCount != expectedRetryCount {
			t.Errorf("Expected retry count %d on attempt %d, got %d",
				expectedRetryCount, i+1, updatedJob.RetryCount)
		}

		if i < 2 && updatedJob.Status != models.StatusPending {
			t.Errorf("Expected status pending on attempt %d, got %s", i+1, updatedJob.Status)
		}
	}

	// After 3 retries, job should be failed
	finalJob, _ := storage.GetJob(ctx, jobID)
	if finalJob.Status != models.StatusFailed {
		t.Errorf("Expected status failed after max retries, got %s", finalJob.Status)
	}
}

func TestProcessJob_MixedContent(t *testing.T) {
	storage := NewMockStorage()
	processor := NewProcessor(storage, 3)

	jobID := "test-job-7"
	// Mixed valid and invalid lines
	fileContent := `CLAIM*CLM001*MEM123*2500
INVALID*LINE
CLAIM*CLM002*MEM456*3000`

	// Create a job in storage
	job := models.NewJob(jobID, "test.edi")
	ctx := context.Background()
	storage.SaveJob(ctx, job)

	// Process the job
	err := processor.ProcessJob(ctx, jobID, fileContent)
	if err != nil {
		t.Fatalf("Expected no error for mixed content, got: %v", err)
	}

	// Verify job was completed with valid claims only
	updatedJob, _ := storage.GetJob(ctx, jobID)
	if updatedJob.Status != models.StatusCompleted {
		t.Errorf("Expected status completed, got %s", updatedJob.Status)
	}

	if updatedJob.Result == nil {
		t.Fatal("Expected result, got nil")
	}

	if updatedJob.Result.Summary.TotalClaims != 2 {
		t.Errorf("Expected 2 valid claims, got %d", updatedJob.Result.Summary.TotalClaims)
	}
}
