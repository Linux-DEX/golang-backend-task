package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type JobStatus string

const (
	StatusPending    JobStatus = "pending"
	StatusProcessing JobStatus = "processing"
	StatusCompleted  JobStatus = "completed"
	StatusFailed     JobStatus = "failed"
)

type Job struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	JobID      string             `bson:"job_id" json:"job_id"`
	FileName   string             `bson:"file_name,omitempty" json:"file_name,omitempty"`
	Status     JobStatus          `bson:"status" json:"status"`
	Result     *Result            `bson:"result,omitempty" json:"result,omitempty"`
	Error      string             `bson:"error,omitempty" json:"error,omitempty"`
	RetryCount int                `bson:"retry_count" json:"retry_count"`
	CreatedAt  time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt  time.Time          `bson:"updated_at" json:"updated_at"`
}

type Claim struct {
	ClaimID  string  `bson:"claim_id" json:"claim_id"`
	MemberID string  `bson:"member_id" json:"member_id"`
	Amount   float64 `bson:"amount" json:"amount"`
}

type Summary struct {
	TotalClaims int     `bson:"total_claims" json:"total_claims"`
	TotalAmount float64 `bson:"total_amount" json:"total_amount"`
}

type Result struct {
	Claims  []Claim `bson:"claims" json:"claims"`
	Summary Summary `bson:"summary" json:"summary"`
}

func NewJob(jobID, fileName string) *Job {
	return &Job{
		ID:         primitive.ObjectID{},
		JobID:      jobID,
		FileName:   fileName,
		Status:     StatusPending,
		RetryCount: 0,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
}

func (j *Job) UpdateStatus(status JobStatus, errorMsg string) {
	j.Status = status
	j.Error = errorMsg
	j.UpdatedAt = time.Now()
}

type JobResponse struct {
	JobID      string    `json:"job_id"`
	Status     JobStatus `json:"status"`
	Result     *Result   `json:"result,omitempty"`
	Error      string    `json:"error,omitempty"`
	RetryCount int       `json:"retry_count"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

func (j *Job) ToResponse() *JobResponse {
	return &JobResponse{
		JobID:      j.ID.Hex(),
		Status:     j.Status,
		Result:     j.Result,
		Error:      j.Error,
		RetryCount: j.RetryCount,
		CreatedAt:  j.CreatedAt,
		UpdatedAt:  j.UpdatedAt,
	}
}

type ResultResponse struct {
	JobID   string  `json:"job_id"`
	Status  string  `json:"status"`
	Claims  []Claim `json:"claims"`
	Summary Summary `json:"summary"`
}

func (r *Result) ToResponse() *ResultResponse {
	return &ResultResponse{
		Status:  "completed",
		Claims:  r.Claims,
		Summary: r.Summary,
	}
}
