package api

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sarabjeet/golang-backend-task/internal/models"
	"github.com/sarabjeet/golang-backend-task/internal/queue"
	"github.com/sarabjeet/golang-backend-task/internal/storage"
)

func TestHealthCheck(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := &Handler{}
	router := gin.New()
	router.GET("/health", handler.HealthCheck)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if response["status"] != "healthy" {
		t.Errorf("Expected status 'healthy', got %v", response["status"])
	}
}

func TestCreateJob_InvalidFile(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := &Handler{}
	router := gin.New()
	router.POST("/jobs", handler.CreateJob)

	// Request without file
	req := httptest.NewRequest(http.MethodPost, "/jobs", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if response["error"] == nil {
		t.Error("Expected error in response")
	}
}

func TestCreateJob_InvalidFileExtension(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := &Handler{}
	router := gin.New()
	router.POST("/jobs", handler.CreateJob)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", "test.txt")
	part.Write([]byte("test content"))
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/jobs", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if response["error"] == nil {
		t.Error("Expected error for invalid file extension")
	}
}

func TestCreateJob_EmptyFile(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := &Handler{}
	router := gin.New()
	router.POST("/jobs", handler.CreateJob)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", "test.edi")
	part.Write([]byte(""))
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/jobs", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if response["error"] == nil {
		t.Error("Expected error for empty file")
	}
}

func TestGetJobStatus_InvalidJobID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := &Handler{}
	router := gin.New()
	router.GET("/jobs/:job_id", handler.GetJobStatus)

	req := httptest.NewRequest(http.MethodGet, "/jobs/", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}
}

func TestGetJobResult_InvalidJobID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := &Handler{}
	router := gin.New()
	router.GET("/jobs/:job_id/result", handler.GetJobResult)

	req := httptest.NewRequest(http.MethodGet, "/jobs/", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound && w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 404 or 400, got %d", w.Code)
	}
}

func createTestFile(filename, content string) *bytes.Buffer {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", filename)
	io.WriteString(part, content)
	writer.Close()
	return body
}

func TestLoggerMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	middleware := LoggerMiddleware()
	if middleware == nil {
		t.Fatal("Expected middleware function, got nil")
	}

	router := gin.New()
	router.Use(middleware)
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestCORSMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	middleware := CORSMiddleware()
	if middleware == nil {
		t.Fatal("Expected middleware function, got nil")
	}

	router := gin.New()
	router.Use(middleware)
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

	req := httptest.NewRequest(http.MethodOptions, "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("Expected status 204 for OPTIONS, got %d", w.Code)
	}
	req = httptest.NewRequest(http.MethodGet, "/test", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	corsHeader := w.Header().Get("Access-Control-Allow-Origin")
	if corsHeader != "*" {
		t.Errorf("Expected CORS header '*', got '%s'", corsHeader)
	}
}

func TestSetupRouter(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	if router == nil {
		t.Fatal("Failed to create router")
	}
}

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

type MockQueue struct {
	messages []string
}

func NewMockQueue() *MockQueue {
	return &MockQueue{
		messages: make([]string, 0),
	}
}

func (m *MockQueue) Enqueue(ctx context.Context, jobMsg *queue.JobMessage) error {
	data, _ := json.Marshal(jobMsg)
	m.messages = append(m.messages, string(data))
	return nil
}

func (m *MockQueue) Dequeue(ctx context.Context) (string, error) {
	if len(m.messages) == 0 {
		return "", nil
	}
	msg := m.messages[0]
	m.messages = m.messages[1:]
	return msg, nil
}

func (m *MockQueue) QueueLength(ctx context.Context) (int64, error) {
	return int64(len(m.messages)), nil
}

func (m *MockQueue) Close() error {
	return nil
}
