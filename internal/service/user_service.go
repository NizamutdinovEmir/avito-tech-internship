package service

import (
	"errors"
	"fmt"

	"avito-tech-internship/internal/domain"
	"avito-tech-internship/internal/repository"
)

var (
	ErrUserNotFound = errors.New("user not found")
)

type UserService struct {
	userRepo repository.UserRepository
}

func NewUserService(userRepo repository.UserRepository) *UserService {
	return &UserService{userRepo: userRepo}
}

// SetIsActive updates the is_active flag for a user
func (s *UserService) SetIsActive(userID string, isActive bool) (*domain.User, error) {
	user, err := s.userRepo.SetIsActive(userID, isActive)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to set user activity: %w", err)
	}
	return user, nil
}

// GetUser retrieves a user by ID
func (s *UserService) GetUser(userID string) (*domain.User, error) {
	user, err := s.userRepo.GetUser(userID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return user, nil
}

// GetActiveUsersByTeam returns all active users in a team (excluding specified user IDs)
func (s *UserService) GetActiveUsersByTeam(teamName string, excludeUserIDs []string) ([]*domain.User, error) {
	users, err := s.userRepo.GetActiveUsersByTeam(teamName, excludeUserIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to get active users: %w", err)
	}
	return users, nil
}
