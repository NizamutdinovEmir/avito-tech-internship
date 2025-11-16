package domain

// Stats represents statistics about the system
type Stats struct {
	TotalPRs              int                   `json:"total_prs"`
	TotalUsers            int                   `json:"total_users"`
	AverageReviewersPerPR float64               `json:"average_reviewers_per_pr"`
	AssignmentsByUser     []UserAssignmentStats `json:"assignments_by_user"`
	ReviewersPerPR        []PRReviewerStats     `json:"reviewers_per_pr"`
}

// UserAssignmentStats represents assignment statistics for a user
type UserAssignmentStats struct {
	UserID          string `json:"user_id"`
	Username        string `json:"username"`
	AssignmentCount int    `json:"assignment_count"`
}

// PRReviewerStats represents reviewer count for a PR
type PRReviewerStats struct {
	PRID          string `json:"pr_id"`
	PRName        string `json:"pr_name"`
	ReviewerCount int    `json:"reviewer_count"`
}
