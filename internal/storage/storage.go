package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/sarabjeet/golang-backend-task/internal/config"
	"github.com/sarabjeet/golang-backend-task/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	jobsCollection    = "jobs"
	resultsCollection = "results"
)

type Storage struct {
	client *mongo.Client
	db     *mongo.Database
}

type MongoDB = Storage

func New(cfg *config.Config) (*Storage, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(cfg.MongoDB.URI))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	if err := client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	db := client.Database(cfg.MongoDB.Database)

	return &Storage{
		client: client,
		db:     db,
	}, nil
}

func NewMongoDB(cfg *config.MongoDBConfig) (*MongoDB, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(cfg.URI))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	if err := client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	db := client.Database(cfg.Database)

	return &Storage{
		client: client,
		db:     db,
	}, nil
}

func (s *Storage) Close(ctx context.Context) error {
	return s.client.Disconnect(ctx)
}

func (s *Storage) CreateIndexes(ctx context.Context) error {
	jobIDIndex := mongo.IndexModel{
		Keys:    bson.D{{Key: "job_id", Value: 1}},
		Options: options.Index().SetUnique(true),
	}
	_, err := s.db.Collection(jobsCollection).Indexes().CreateOne(ctx, jobIDIndex)
	if err != nil {
		return fmt.Errorf("failed to create job_id index: %w", err)
	}

	statusIndex := mongo.IndexModel{
		Keys: bson.D{{Key: "status", Value: 1}},
	}
	_, err = s.db.Collection(jobsCollection).Indexes().CreateOne(ctx, statusIndex)
	if err != nil {
		return fmt.Errorf("failed to create status index: %w", err)
	}

	return nil
}

func (s *Storage) GetJob(ctx context.Context, jobID string) (*models.Job, error) {
	var job models.Job
	err := s.db.Collection(jobsCollection).FindOne(ctx, bson.M{"job_id": jobID}).Decode(&job)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("job not found")
		}
		return nil, fmt.Errorf("failed to get job: %w", err)
	}

	return &job, nil
}

func (s *Storage) SaveJob(ctx context.Context, job *models.Job) error {
	if job.ID.IsZero() {
		job.ID = primitive.NewObjectID()
	}
	if job.CreatedAt.IsZero() {
		job.CreatedAt = time.Now()
	}
	if job.UpdatedAt.IsZero() {
		job.UpdatedAt = time.Now()
	}

	_, err := s.db.Collection(jobsCollection).InsertOne(ctx, job)
	if err != nil {
		return fmt.Errorf("failed to save job: %w", err)
	}

	return nil
}

func (s *Storage) UpdateJob(ctx context.Context, job *models.Job) error {
	job.UpdatedAt = time.Now()

	filter := bson.M{"_id": job.ID}
	update := bson.M{"$set": job}

	result, err := s.db.Collection(jobsCollection).UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to update job: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("job not found")
	}

	return nil
}

func (s *Storage) UpdateJobStatus(ctx context.Context, jobID string, status models.JobStatus) error {
	update := bson.M{
		"$set": bson.M{
			"status":     status,
			"updated_at": time.Now(),
		},
	}

	result, err := s.db.Collection(jobsCollection).UpdateOne(
		ctx,
		bson.M{"job_id": jobID},
		update,
	)
	if err != nil {
		return fmt.Errorf("failed to update job status: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("job not found")
	}

	return nil
}

func (s *Storage) UpdateJobWithResult(ctx context.Context, jobID string, status models.JobStatus, result *models.Result, errorMsg string) error {
	update := bson.M{
		"$set": bson.M{
			"status":     status,
			"updated_at": time.Now(),
		},
	}

	if result != nil {
		update["$set"].(bson.M)["result"] = result
	}

	if errorMsg != "" {
		update["$set"].(bson.M)["error"] = errorMsg
	}

	filter := bson.M{"job_id": jobID}

	resultUpdate, err := s.db.Collection(jobsCollection).UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to update job with result: %w", err)
	}

	if resultUpdate.MatchedCount == 0 {
		return fmt.Errorf("job not found")
	}

	return nil
}

func (s *Storage) GetResult(ctx context.Context, jobID string) (*models.Result, error) {
	job, err := s.GetJob(ctx, jobID)
	if err != nil {
		return nil, err
	}

	if job.Result == nil {
		return nil, fmt.Errorf("result not found")
	}

	return job.Result, nil
}

func (s *Storage) IncrementRetryCount(ctx context.Context, jobID string) error {
	update := bson.M{
		"$inc": bson.M{
			"retry_count": 1,
		},
		"$set": bson.M{
			"updated_at": time.Now(),
		},
	}

	result, err := s.db.Collection(jobsCollection).UpdateOne(
		ctx,
		bson.M{"job_id": jobID},
		update,
	)
	if err != nil {
		return fmt.Errorf("failed to increment retry count: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("job not found")
	}

	return nil
}

func (s *Storage) CreateJob(ctx context.Context, job *models.Job) (string, error) {
	if job.ID.IsZero() {
		job.ID = primitive.NewObjectID()
	}
	if job.CreatedAt.IsZero() {
		job.CreatedAt = time.Now()
	}
	if job.UpdatedAt.IsZero() {
		job.UpdatedAt = time.Now()
	}

	result, err := s.db.Collection(jobsCollection).InsertOne(ctx, job)
	if err != nil {
		return "", fmt.Errorf("failed to create job: %w", err)
	}

	return result.InsertedID.(primitive.ObjectID).Hex(), nil
}
