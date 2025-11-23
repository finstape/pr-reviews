package entity

import "time"

// PullRequestStatus represents the status of a pull request
type PullRequestStatus string

const (
	PullRequestStatusOpen   PullRequestStatus = "OPEN"
	PullRequestStatusMerged PullRequestStatus = "MERGED"
)

// PullRequest represents a pull request
type PullRequest struct {
	PullRequestID   string              `json:"pull_request_id"`
	PullRequestName string              `json:"pull_request_name"`
	AuthorID        string              `json:"author_id"`
	Status          PullRequestStatus   `json:"status"`
	AssignedReviewers []string          `json:"assigned_reviewers"`
	CreatedAt       *time.Time          `json:"createdAt,omitempty"`
	MergedAt        *time.Time          `json:"mergedAt,omitempty"`
}

// PullRequestShort represents a short version of pull request (without reviewers)
type PullRequestShort struct {
	PullRequestID   string            `json:"pull_request_id"`
	PullRequestName string            `json:"pull_request_name"`
	AuthorID        string            `json:"author_id"`
	Status          PullRequestStatus `json:"status"`
}

