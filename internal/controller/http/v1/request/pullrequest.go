package request

// CreatePRRequest -.
type CreatePRRequest struct {
	PullRequestID   string `json:"pull_request_id" validate:"required"`
	PullRequestName string `json:"pull_request_name" validate:"required"`
	AuthorID        string `json:"author_id" validate:"required"`
}

// MergePRRequest -.
type MergePRRequest struct {
	PullRequestID string `json:"pull_request_id" validate:"required"`
}

// ReassignReviewerRequest -.
type ReassignReviewerRequest struct {
	PullRequestID string `json:"pull_request_id" validate:"required"`
	OldUserID     string `json:"old_user_id" validate:"required"`
}

