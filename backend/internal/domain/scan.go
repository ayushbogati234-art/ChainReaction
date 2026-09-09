package domain

import (
	"errors"
	"time"
)

var ErrNotFound = errors.New("resource not found")

type Scan struct {
	ID           string     `json:"id"`
	ProjectID    string     `json:"project_id"`
	CommitSHA    string     `json:"commit_sha"`
	Status       string     `json:"status"`
	Stage        string     `json:"stage"`
	Progress     int        `json:"progress"`
	Incomplete   bool       `json:"incomplete"`
	Limitations  []string   `json:"limitations"`
	ErrorCode    *string    `json:"error_code"`
	ErrorMessage *string    `json:"error_message"`
	CreatedAt    time.Time  `json:"created_at"`
	StartedAt    *time.Time `json:"started_at"`
	CompletedAt  *time.Time `json:"completed_at"`
}
