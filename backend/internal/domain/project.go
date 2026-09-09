package domain

import "time"

type Project struct {
	ID            string    `json:"id"`
	RepositoryURL string    `json:"repository_url"`
	CreatedAt     time.Time `json:"created_at"`
}
