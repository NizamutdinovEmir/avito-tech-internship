package repository

import "avito-tech-internship/internal/domain"

// UserRepository defines the interface for user operations
type UserRepository interface {
	// GetUser retrieves a user by ID
	GetUser(userID string) (*domain.User, error)

	// SetIsActive updates the is_active flag for a user
	SetIsActive(userID string, isActive bool) (*domain.User, error)

	// GetActiveUsersByTeam returns all active users in a team (excluding specified user IDs)
	GetActiveUsersByTeam(teamName string, excludeUserIDs []string) ([]*domain.User, error)

	// CreateOrUpdateUser creates a new user or updates existing one
	CreateOrUpdateUser(user *domain.User) error

	// BulkSetIsActive updates is_active flag for multiple users
	BulkSetIsActive(userIDs []string, isActive bool) error
}
