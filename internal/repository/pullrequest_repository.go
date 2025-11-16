package repository

import "avito-tech-internship/internal/domain"

// PullRequestRepository defines the interface for pull request operations
type PullRequestRepository interface {
	// CreatePR creates a new pull request
	CreatePR(pr *domain.PullRequest) error

	// GetPR retrieves a pull request by ID with assigned reviewers
	GetPR(prID string) (*domain.PullRequest, error)

	// UpdatePR updates an existing pull request
	UpdatePR(pr *domain.PullRequest) error

	// MergePR marks a PR as merged (idempotent)
	MergePR(prID string) (*domain.PullRequest, error)

	// PRExists checks if a PR with given ID exists
	PRExists(prID string) (bool, error)

	// GetPRsByReviewer returns all PRs where the user is assigned as reviewer
	GetPRsByReviewer(userID string) ([]*domain.PullRequestShort, error)

	// ReassignReviewer replaces one reviewer with another
	ReassignReviewer(prID string, oldUserID string, newUserID string) error

	// GetStats retrieves statistics about PR assignments
	GetStats() (*domain.Stats, error)

	// GetOpenPRsByReviewers returns all OPEN PRs where any of the given users are reviewers
	GetOpenPRsByReviewers(userIDs []string) ([]*domain.PullRequest, error)
}
