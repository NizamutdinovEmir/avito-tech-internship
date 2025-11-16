package service

import (
	"errors"
	"fmt"
	"math/rand"

	"avito-tech-internship/internal/domain"
	"avito-tech-internship/internal/repository"
)

var (
	ErrPRNotFound     = errors.New("PR not found")
	ErrPRExists       = errors.New("PR already exists")
	ErrPRMerged       = errors.New("PR is merged")
	ErrNotAssigned    = errors.New("reviewer is not assigned")
	ErrNoCandidate    = errors.New("no active replacement candidate")
	ErrAuthorNotFound = errors.New("author not found")
)

type PullRequestService struct {
	prRepo   repository.PullRequestRepository
	userRepo repository.UserRepository
	teamRepo repository.TeamRepository
}

func NewPullRequestService(
	prRepo repository.PullRequestRepository,
	userRepo repository.UserRepository,
	teamRepo repository.TeamRepository,
) *PullRequestService {
	return &PullRequestService{
		prRepo:   prRepo,
		userRepo: userRepo,
		teamRepo: teamRepo,
	}
}

// CreatePR creates a new PR and automatically assigns up to 2 active reviewers from author's team
func (s *PullRequestService) CreatePR(pr *domain.PullRequest) error {
	exists, err := s.prRepo.PRExists(pr.PullRequestID)
	if err != nil {
		return fmt.Errorf("failed to check PR existence: %w", err)
	}
	if exists {
		return ErrPRExists
	}

	author, err := s.userRepo.GetUser(pr.AuthorID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrAuthorNotFound
		}
		return fmt.Errorf("failed to get author: %w", err)
	}

	candidates, err := s.userRepo.GetActiveUsersByTeam(author.TeamName, []string{pr.AuthorID})
	if err != nil {
		return fmt.Errorf("failed to get active users: %w", err)
	}

	pr.AssignedReviewers = s.selectReviewers(candidates, 2)

	pr.Status = domain.PRStatusOpen

	if err := s.prRepo.CreatePR(pr); err != nil {
		return fmt.Errorf("failed to create PR: %w", err)
	}

	return nil
}

// MergePR marks a PR as merged (idempotent operation)
func (s *PullRequestService) MergePR(prID string) (*domain.PullRequest, error) {
	pr, err := s.prRepo.MergePR(prID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrPRNotFound
		}
		return nil, fmt.Errorf("failed to merge PR: %w", err)
	}
	return pr, nil
}

// ReassignReviewer replaces one reviewer with another random active user from the replaced reviewer's team
func (s *PullRequestService) ReassignReviewer(prID string, oldUserID string) (*domain.PullRequest, string, error) {
	// Get PR
	pr, err := s.prRepo.GetPR(prID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, "", ErrPRNotFound
		}
		return nil, "", fmt.Errorf("failed to get PR: %w", err)
	}

	if pr.Status == domain.PRStatusMerged {
		return nil, "", ErrPRMerged
	}

	found := false
	for _, reviewerID := range pr.AssignedReviewers {
		if reviewerID == oldUserID {
			found = true
			break
		}
	}
	if !found {
		return nil, "", ErrNotAssigned
	}

	oldReviewer, err := s.userRepo.GetUser(oldUserID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, "", ErrUserNotFound
		}
		return nil, "", fmt.Errorf("failed to get old reviewer: %w", err)
	}

	excludeIDs := []string{oldUserID}
	for _, reviewerID := range pr.AssignedReviewers {
		if reviewerID != oldUserID {
			excludeIDs = append(excludeIDs, reviewerID)
		}
	}

	candidates, err := s.userRepo.GetActiveUsersByTeam(oldReviewer.TeamName, excludeIDs)
	if err != nil {
		return nil, "", fmt.Errorf("failed to get active users: %w", err)
	}

	if len(candidates) == 0 {
		return nil, "", ErrNoCandidate
	}

	newReviewer := s.selectReviewers(candidates, 1)
	if len(newReviewer) == 0 {
		return nil, "", ErrNoCandidate
	}
	newUserID := newReviewer[0]

	if reassignErr := s.prRepo.ReassignReviewer(prID, oldUserID, newUserID); reassignErr != nil {
		return nil, "", fmt.Errorf("failed to reassign reviewer: %w", reassignErr)
	}

	updatedPR, err := s.prRepo.GetPR(prID)
	if err != nil {
		return nil, "", fmt.Errorf("failed to get updated PR: %w", err)
	}

	return updatedPR, newUserID, nil
}

// GetPRsByReviewer returns all PRs where the user is assigned as reviewer
func (s *PullRequestService) GetPRsByReviewer(userID string) ([]*domain.PullRequestShort, error) {
	_, err := s.userRepo.GetUser(userID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	prs, err := s.prRepo.GetPRsByReviewer(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get PRs by reviewer: %w", err)
	}

	return prs, nil
}

// selectReviewers randomly selects up to maxCount reviewers from candidates
func (s *PullRequestService) selectReviewers(candidates []*domain.User, maxCount int) []string {
	if len(candidates) == 0 {
		return []string{}
	}

	shuffled := make([]*domain.User, len(candidates))
	copy(shuffled, candidates)
	rand.Shuffle(len(shuffled), func(i, j int) {
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	})

	count := maxCount
	if len(shuffled) < maxCount {
		count = len(shuffled)
	}

	reviewers := make([]string, 0, count)
	for i := 0; i < count; i++ {
		reviewers = append(reviewers, shuffled[i].UserID)
	}

	return reviewers
}

// GetPR retrieves a PR by ID
func (s *PullRequestService) GetPR(prID string) (*domain.PullRequest, error) {
	pr, err := s.prRepo.GetPR(prID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrPRNotFound
		}
		return nil, fmt.Errorf("failed to get PR: %w", err)
	}
	return pr, nil
}

// GetStats retrieves statistics about PR assignments
func (s *PullRequestService) GetStats() (*domain.Stats, error) {
	stats, err := s.prRepo.GetStats()
	if err != nil {
		return nil, fmt.Errorf("failed to get stats: %w", err)
	}
	return stats, nil
}
