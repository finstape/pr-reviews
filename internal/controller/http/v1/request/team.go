package request

// CreateTeamRequest -.
type CreateTeamRequest struct {
	TeamName string                 `json:"team_name" validate:"required"`
	Members  []CreateTeamMemberRequest `json:"members" validate:"required,dive"`
}

// CreateTeamMemberRequest -.
type CreateTeamMemberRequest struct {
	UserID   string `json:"user_id" validate:"required"`
	Username string `json:"username" validate:"required"`
	IsActive bool   `json:"is_active"`
}

