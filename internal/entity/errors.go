package entity

import "errors"

// Domain errors
var (
	ErrTeamExists   = errors.New("team_name already exists")
	ErrPRExists     = errors.New("PR id already exists")
	ErrPRMerged     = errors.New("cannot reassign on merged PR")
	ErrNotAssigned  = errors.New("reviewer is not assigned to this PR")
	ErrNoCandidate  = errors.New("no active replacement candidate in team")
	ErrNotFound     = errors.New("resource not found")
)

// ErrorCode represents error codes for API responses
type ErrorCode string

const (
	ErrorCodeTeamExists  ErrorCode = "TEAM_EXISTS"
	ErrorCodePRExists    ErrorCode = "PR_EXISTS"
	ErrorCodePRMerged   ErrorCode = "PR_MERGED"
	ErrorCodeNotAssigned ErrorCode = "NOT_ASSIGNED"
	ErrorCodeNoCandidate ErrorCode = "NO_CANDIDATE"
	ErrorCodeNotFound    ErrorCode = "NOT_FOUND"
)

// GetErrorCode returns the error code for a given error
func GetErrorCode(err error) ErrorCode {
	switch err {
	case ErrTeamExists:
		return ErrorCodeTeamExists
	case ErrPRExists:
		return ErrorCodePRExists
	case ErrPRMerged:
		return ErrorCodePRMerged
	case ErrNotAssigned:
		return ErrorCodeNotAssigned
	case ErrNoCandidate:
		return ErrorCodeNoCandidate
	case ErrNotFound:
		return ErrorCodeNotFound
	default:
		return ErrorCodeNotFound
	}
}

