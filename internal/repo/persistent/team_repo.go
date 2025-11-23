package persistent

import (
	"context"
	"fmt"

	"github.com/finstape/pr-reviews/internal/entity"
	"github.com/finstape/pr-reviews/pkg/postgres"
)

// TeamRepo handles team data persistence.
type TeamRepo struct {
	*postgres.Postgres
}

// NewTeamRepo creates a new TeamRepo instance.
func NewTeamRepo(pg *postgres.Postgres) *TeamRepo {
	return &TeamRepo{pg}
}

// CreateTeam creates a team and its members
func (r *TeamRepo) CreateTeam(ctx context.Context, team entity.Team) error {
	tx, err := r.Pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("TeamRepo - CreateTeam - Begin: %w", err)
	}
	defer tx.Rollback(ctx)

	// Insert team
	sql, args, err := r.Builder.
		Insert("teams").
		Columns("team_name").
		Values(team.TeamName).
		ToSql()
	if err != nil {
		return fmt.Errorf("TeamRepo - CreateTeam - BuildInsert: %w", err)
	}

	_, err = tx.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("TeamRepo - CreateTeam - Exec team: %w", err)
	}

	// Insert or update users
	for _, member := range team.Members {
		sql, args, err := r.Builder.
			Insert("users").
			Columns("user_id", "username", "team_name", "is_active").
			Values(member.UserID, member.Username, team.TeamName, member.IsActive).
			Suffix("ON CONFLICT (user_id) DO UPDATE SET username = EXCLUDED.username, team_name = EXCLUDED.team_name, is_active = EXCLUDED.is_active, updated_at = CURRENT_TIMESTAMP").
			ToSql()
		if err != nil {
			return fmt.Errorf("TeamRepo - CreateTeam - BuildInsert user: %w", err)
		}

		_, err = tx.Exec(ctx, sql, args...)
		if err != nil {
			return fmt.Errorf("TeamRepo - CreateTeam - Exec user: %w", err)
		}
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("TeamRepo - CreateTeam - Commit: %w", err)
	}

	return nil
}

// GetTeam retrieves a team with its members
func (r *TeamRepo) GetTeam(ctx context.Context, teamName string) (entity.Team, error) {
	// Get team members
	sql, args, err := r.Builder.
		Select("user_id", "username", "is_active").
		From("users").
		Where("team_name = ?", teamName).
		OrderBy("user_id").
		ToSql()
	if err != nil {
		return entity.Team{}, fmt.Errorf("TeamRepo - GetTeam - BuildSelect: %w", err)
	}

	rows, err := r.Pool.Query(ctx, sql, args...)
	if err != nil {
		return entity.Team{}, fmt.Errorf("TeamRepo - GetTeam - Query: %w", err)
	}
	defer rows.Close()

	var members []entity.TeamMember
	for rows.Next() {
		var member entity.TeamMember
		if err := rows.Scan(&member.UserID, &member.Username, &member.IsActive); err != nil {
			return entity.Team{}, fmt.Errorf("TeamRepo - GetTeam - Scan: %w", err)
		}
		members = append(members, member)
	}

	if err = rows.Err(); err != nil {
		return entity.Team{}, fmt.Errorf("TeamRepo - GetTeam - RowsErr: %w", err)
	}

	return entity.Team{
		TeamName: teamName,
		Members:  members,
	}, nil
}

// TeamExists checks if a team exists
func (r *TeamRepo) TeamExists(ctx context.Context, teamName string) (bool, error) {
	sql, args, err := r.Builder.
		Select("1").
		From("teams").
		Where("team_name = ?", teamName).
		ToSql()
	if err != nil {
		return false, fmt.Errorf("TeamRepo - TeamExists - BuildSelect: %w", err)
	}

	var exists int
	err = r.Pool.QueryRow(ctx, sql, args...).Scan(&exists)
	if err != nil {
		// If no rows, team doesn't exist
		return false, nil
	}

	return exists == 1, nil
}

