package service

import (
	"errors"
	"fmt"

	"avito-tech-internship/internal/domain"
	"avito-tech-internship/internal/repository"
)

var (
	ErrTeamExists   = errors.New("team already exists")
	ErrTeamNotFound = errors.New("team not found")
)

type TeamService struct {
	teamRepo repository.TeamRepository
}

func NewTeamService(teamRepo repository.TeamRepository) *TeamService {
	return &TeamService{teamRepo: teamRepo}
}

// CreateTeam creates a new team with members (creates/updates users)
func (s *TeamService) CreateTeam(team *domain.Team) error {
	exists, err := s.teamRepo.TeamExists(team.TeamName)
	if err != nil {
		return fmt.Errorf("failed to check team existence: %w", err)
	}
	if exists {
		return ErrTeamExists
	}

	if err := s.teamRepo.CreateTeam(team); err != nil {
		return fmt.Errorf("failed to create team: %w", err)
	}

	return nil
}

// GetTeam retrieves a team by name with all its members
func (s *TeamService) GetTeam(teamName string) (*domain.Team, error) {
	team, err := s.teamRepo.GetTeam(teamName)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrTeamNotFound
		}
		return nil, fmt.Errorf("failed to get team: %w", err)
	}
	return team, nil
}
