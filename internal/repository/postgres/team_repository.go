package postgres

import (
	"database/sql"
	"fmt"

	"avito-tech-internship/internal/domain"
	"avito-tech-internship/internal/repository"
)

type teamRepository struct {
	db *sql.DB
}

// NewTeamRepository creates a new PostgreSQL team repository
func NewTeamRepository(db *sql.DB) *teamRepository {
	return &teamRepository{db: db}
}

func (r *teamRepository) CreateTeam(team *domain.Team) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback() // Ignore error - transaction may already be committed
	}()

	// Create team
	_, err = tx.Exec(
		"INSERT INTO teams (team_name) VALUES ($1) ON CONFLICT (team_name) DO NOTHING",
		team.TeamName,
	)
	if err != nil {
		return fmt.Errorf("failed to create team: %w", err)
	}

	// Create or update users
	for _, member := range team.Members {
		_, err = tx.Exec(
			`INSERT INTO users (user_id, username, team_name, is_active) 
			 VALUES ($1, $2, $3, $4)
			 ON CONFLICT (user_id) 
			 DO UPDATE SET username = $2, team_name = $3, is_active = $4, updated_at = CURRENT_TIMESTAMP`,
			member.UserID, member.Username, team.TeamName, member.IsActive,
		)
		if err != nil {
			return fmt.Errorf("failed to create/update user %s: %w", member.UserID, err)
		}
	}

	return tx.Commit()
}

func (r *teamRepository) GetTeam(teamName string) (*domain.Team, error) {
	// Check if team exists
	var exists bool
	err := r.db.QueryRow(
		"SELECT EXISTS(SELECT 1 FROM teams WHERE team_name = $1)",
		teamName,
	).Scan(&exists)
	if err != nil {
		return nil, fmt.Errorf("failed to check team existence: %w", err)
	}
	if !exists {
		return nil, repository.ErrNotFound
	}

	// Get team members
	rows, err := r.db.Query(
		"SELECT user_id, username, is_active FROM users WHERE team_name = $1 ORDER BY user_id",
		teamName,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query team members: %w", err)
	}
	defer rows.Close()

	var members []domain.TeamMember
	for rows.Next() {
		var member domain.TeamMember
		if err := rows.Scan(&member.UserID, &member.Username, &member.IsActive); err != nil {
			return nil, fmt.Errorf("failed to scan team member: %w", err)
		}
		members = append(members, member)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating team members: %w", err)
	}

	return &domain.Team{
		TeamName: teamName,
		Members:  members,
	}, nil
}

func (r *teamRepository) TeamExists(teamName string) (bool, error) {
	var exists bool
	err := r.db.QueryRow(
		"SELECT EXISTS(SELECT 1 FROM teams WHERE team_name = $1)",
		teamName,
	).Scan(&exists)
	return exists, err
}
