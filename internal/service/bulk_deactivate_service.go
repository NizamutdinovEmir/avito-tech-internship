package service

import (
	"fmt"

	"avito-tech-internship/internal/repository"
)

// BulkDeactivateService handles bulk deactivation of users with safe PR reassignment
type BulkDeactivateService struct {
	userRepo repository.UserRepository
	prRepo   repository.PullRequestRepository
	teamRepo repository.TeamRepository
}

func NewBulkDeactivateService(
	userRepo repository.UserRepository,
	prRepo repository.PullRequestRepository,
	teamRepo repository.TeamRepository,
) *BulkDeactivateService {
	return &BulkDeactivateService{
		userRepo: userRepo,
		prRepo:   prRepo,
		teamRepo: teamRepo,
	}
}

// BulkDeactivate deactivates multiple users in a team and safely reassigns reviewers in open PRs
func (s *BulkDeactivateService) BulkDeactivate(teamName string, userIDs []string) error {
	if len(userIDs) == 0 {
		return fmt.Errorf("no users provided")
	}

	_, err := s.teamRepo.GetTeam(teamName)
	if err != nil {
		return fmt.Errorf("team not found: %w", err)
	}

	for _, userID := range userIDs {
		user, getUserErr := s.userRepo.GetUser(userID)
		if getUserErr != nil {
			return fmt.Errorf("user %s not found: %w", userID, getUserErr)
		}
		if user.TeamName != teamName {
			return fmt.Errorf("user %s does not belong to team %s", userID, teamName)
		}
	}

	openPRs, err := s.prRepo.GetOpenPRsByReviewers(userIDs)
	if err != nil {
		return fmt.Errorf("failed to get open PRs: %w", err)
	}

	if err := s.userRepo.BulkSetIsActive(userIDs, false); err != nil {
		return fmt.Errorf("failed to deactivate users: %w", err)
	}

	for _, pr := range openPRs {
		deactivatedReviewers := make([]string, 0)
		for _, reviewerID := range pr.AssignedReviewers {
			for _, deactivatedID := range userIDs {
				if reviewerID == deactivatedID {
					deactivatedReviewers = append(deactivatedReviewers, reviewerID)
					break
				}
			}
		}

		for _, oldReviewerID := range deactivatedReviewers {
			oldReviewer, err := s.userRepo.GetUser(oldReviewerID)
			if err != nil {
				continue // Skip if user not found
			}

			excludeIDs := append(userIDs, pr.AssignedReviewers...)
			candidates, err := s.userRepo.GetActiveUsersByTeam(oldReviewer.TeamName, excludeIDs)
			if err != nil || len(candidates) == 0 {
				continue
			}

			newReviewerID := candidates[0].UserID

			if err := s.prRepo.ReassignReviewer(pr.PullRequestID, oldReviewerID, newReviewerID); err != nil {
				// Log error but continue with other PRs
				// In production, you might want to rollback or handle this differently
				continue
			}
		}
	}

	return nil
}
