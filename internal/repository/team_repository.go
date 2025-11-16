package repository

import "avito-tech-internship/internal/domain"

// TeamRepository defines the interface for team operations
type TeamRepository interface {
	// CreateTeam creates a new team with members (creates/updates users)
	CreateTeam(team *domain.Team) error

	// GetTeam retrieves a team by name with all its members
	GetTeam(teamName string) (*domain.Team, error)

	// TeamExists checks if a team with given name exists
	TeamExists(teamName string) (bool, error)
}
