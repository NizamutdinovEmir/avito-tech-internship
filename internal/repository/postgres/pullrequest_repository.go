package postgres

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"avito-tech-internship/internal/domain"
	"avito-tech-internship/internal/repository"
)

type pullRequestRepository struct {
	db *sql.DB
}

// NewPullRequestRepository creates a new PostgreSQL pull request repository
func NewPullRequestRepository(db *sql.DB) *pullRequestRepository {
	return &pullRequestRepository{db: db}
}

func (r *pullRequestRepository) CreatePR(pr *domain.PullRequest) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback() // Ignore error - transaction may already be committed
	}()

	now := time.Now()
	// Create PR
	_, err = tx.Exec(
		`INSERT INTO pull_requests (pull_request_id, pull_request_name, author_id, status, created_at) 
		 VALUES ($1, $2, $3, $4, $5)`,
		pr.PullRequestID, pr.PullRequestName, pr.AuthorID, pr.Status, now,
	)
	if err != nil {
		return fmt.Errorf("failed to create PR: %w", err)
	}

	// Assign reviewers
	for _, reviewerID := range pr.AssignedReviewers {
		_, err = tx.Exec(
			"INSERT INTO pr_reviewers (pull_request_id, user_id) VALUES ($1, $2)",
			pr.PullRequestID, reviewerID,
		)
		if err != nil {
			return fmt.Errorf("failed to assign reviewer %s: %w", reviewerID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit PR creation: %w", err)
	}

	pr.CreatedAt = &now
	return nil
}

func (r *pullRequestRepository) GetPR(prID string) (*domain.PullRequest, error) {
	var pr domain.PullRequest
	var createdAt, mergedAt sql.NullTime

	err := r.db.QueryRow(
		`SELECT pull_request_id, pull_request_name, author_id, status, created_at, merged_at 
		 FROM pull_requests WHERE pull_request_id = $1`,
		prID,
	).Scan(&pr.PullRequestID, &pr.PullRequestName, &pr.AuthorID, &pr.Status, &createdAt, &mergedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, repository.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get PR: %w", err)
	}

	if createdAt.Valid {
		pr.CreatedAt = &createdAt.Time
	}
	if mergedAt.Valid {
		pr.MergedAt = &mergedAt.Time
	}

	// Get assigned reviewers
	rows, err := r.db.Query(
		"SELECT user_id FROM pr_reviewers WHERE pull_request_id = $1 ORDER BY user_id",
		prID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query reviewers: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var reviewerID string
		if err := rows.Scan(&reviewerID); err != nil {
			return nil, fmt.Errorf("failed to scan reviewer: %w", err)
		}
		pr.AssignedReviewers = append(pr.AssignedReviewers, reviewerID)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating reviewers: %w", err)
	}

	return &pr, nil
}

func (r *pullRequestRepository) UpdatePR(pr *domain.PullRequest) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback() // Ignore error - transaction may already be committed
	}()

	// Update PR
	_, err = tx.Exec(
		`UPDATE pull_requests 
		 SET pull_request_name = $1, status = $2, merged_at = $3 
		 WHERE pull_request_id = $4`,
		pr.PullRequestName, pr.Status, pr.MergedAt, pr.PullRequestID,
	)
	if err != nil {
		return fmt.Errorf("failed to update PR: %w", err)
	}

	// Delete old reviewers
	_, err = tx.Exec("DELETE FROM pr_reviewers WHERE pull_request_id = $1", pr.PullRequestID)
	if err != nil {
		return fmt.Errorf("failed to delete old reviewers: %w", err)
	}

	// Insert new reviewers
	for _, reviewerID := range pr.AssignedReviewers {
		_, err = tx.Exec(
			"INSERT INTO pr_reviewers (pull_request_id, user_id) VALUES ($1, $2)",
			pr.PullRequestID, reviewerID,
		)
		if err != nil {
			return fmt.Errorf("failed to assign reviewer %s: %w", reviewerID, err)
		}
	}

	return tx.Commit()
}

func (r *pullRequestRepository) MergePR(prID string) (*domain.PullRequest, error) {
	pr, err := r.GetPR(prID)
	if err != nil {
		return nil, err
	}

	// If already merged, return current state (idempotent)
	if pr.Status == domain.PRStatusMerged {
		return pr, nil
	}

	now := time.Now()
	pr.Status = domain.PRStatusMerged
	pr.MergedAt = &now

	if err := r.UpdatePR(pr); err != nil {
		return nil, err
	}

	return pr, nil
}

func (r *pullRequestRepository) PRExists(prID string) (bool, error) {
	var exists bool
	err := r.db.QueryRow(
		"SELECT EXISTS(SELECT 1 FROM pull_requests WHERE pull_request_id = $1)",
		prID,
	).Scan(&exists)
	return exists, err
}

func (r *pullRequestRepository) GetPRsByReviewer(userID string) ([]*domain.PullRequestShort, error) {
	rows, err := r.db.Query(
		`SELECT pr.pull_request_id, pr.pull_request_name, pr.author_id, pr.status
		 FROM pull_requests pr
		 INNER JOIN pr_reviewers prr ON pr.pull_request_id = prr.pull_request_id
		 WHERE prr.user_id = $1
		 ORDER BY pr.created_at DESC`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query PRs: %w", err)
	}
	defer rows.Close()

	var prs []*domain.PullRequestShort
	for rows.Next() {
		pr := &domain.PullRequestShort{}
		if err := rows.Scan(&pr.PullRequestID, &pr.PullRequestName, &pr.AuthorID, &pr.Status); err != nil {
			return nil, fmt.Errorf("failed to scan PR: %w", err)
		}
		prs = append(prs, pr)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating PRs: %w", err)
	}

	return prs, nil
}

func (r *pullRequestRepository) ReassignReviewer(prID string, oldUserID string, newUserID string) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback() // Ignore error - transaction may already be committed
	}()

	var count int
	err = tx.QueryRow(
		"SELECT COUNT(*) FROM pr_reviewers WHERE pull_request_id = $1 AND user_id = $2",
		prID, oldUserID,
	).Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check reviewer assignment: %w", err)
	}
	if count == 0 {
		return fmt.Errorf("reviewer not assigned to this PR")
	}

	_, err = tx.Exec(
		"UPDATE pr_reviewers SET user_id = $1 WHERE pull_request_id = $2 AND user_id = $3",
		newUserID, prID, oldUserID,
	)
	if err != nil {
		return fmt.Errorf("failed to reassign reviewer: %w", err)
	}

	return tx.Commit()
}

func (r *pullRequestRepository) GetStats() (*domain.Stats, error) {
	stats := &domain.Stats{}

	err := r.db.QueryRow("SELECT COUNT(*) FROM pull_requests").Scan(&stats.TotalPRs)
	if err != nil {
		return nil, fmt.Errorf("failed to get total PRs: %w", err)
	}

	err = r.db.QueryRow("SELECT COUNT(*) FROM users").Scan(&stats.TotalUsers)
	if err != nil {
		return nil, fmt.Errorf("failed to get total users: %w", err)
	}

	var avgReviewers sql.NullFloat64
	err = r.db.QueryRow(`
		SELECT COALESCE(AVG(reviewer_count), 0) 
		FROM (
			SELECT pull_request_id, COUNT(*) as reviewer_count 
			FROM pr_reviewers 
			GROUP BY pull_request_id
		) subq
	`).Scan(&avgReviewers)
	if err != nil {
		return nil, fmt.Errorf("failed to get average reviewers: %w", err)
	}
	if avgReviewers.Valid {
		stats.AverageReviewersPerPR = avgReviewers.Float64
	}

	rows, err := r.db.Query(`
		SELECT u.user_id, u.username, COUNT(prr.user_id) as assignment_count
		FROM users u
		LEFT JOIN pr_reviewers prr ON u.user_id = prr.user_id
		GROUP BY u.user_id, u.username
		ORDER BY assignment_count DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to get assignments by user: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var userStat domain.UserAssignmentStats
		if scanErr := rows.Scan(&userStat.UserID, &userStat.Username, &userStat.AssignmentCount); scanErr != nil {
			return nil, fmt.Errorf("failed to scan user stats: %w", scanErr)
		}
		stats.AssignmentsByUser = append(stats.AssignmentsByUser, userStat)
	}
	if scanErr := rows.Err(); scanErr != nil {
		return nil, fmt.Errorf("error iterating user stats: %w", scanErr)
	}

	prRows, err := r.db.Query(`
		SELECT pr.pull_request_id, pr.pull_request_name, COUNT(prr.user_id) as reviewer_count
		FROM pull_requests pr
		LEFT JOIN pr_reviewers prr ON pr.pull_request_id = prr.pull_request_id
		GROUP BY pr.pull_request_id, pr.pull_request_name
		ORDER BY reviewer_count DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to get reviewers per PR: %w", err)
	}
	defer prRows.Close()

	for prRows.Next() {
		var prStat domain.PRReviewerStats
		if err := prRows.Scan(&prStat.PRID, &prStat.PRName, &prStat.ReviewerCount); err != nil {
			return nil, fmt.Errorf("failed to scan PR stats: %w", err)
		}
		stats.ReviewersPerPR = append(stats.ReviewersPerPR, prStat)
	}
	if err := prRows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating PR stats: %w", err)
	}

	return stats, nil
}

func (r *pullRequestRepository) GetOpenPRsByReviewers(userIDs []string) ([]*domain.PullRequest, error) {
	if len(userIDs) == 0 {
		return []*domain.PullRequest{}, nil
	}

	placeholders := make([]string, len(userIDs))
	args := make([]interface{}, len(userIDs))
	for i, userID := range userIDs {
		args[i] = userID
		placeholders[i] = fmt.Sprintf("$%d", i+1)
	}

	query := fmt.Sprintf(`
		SELECT DISTINCT pr.pull_request_id, pr.pull_request_name, pr.author_id, pr.status, pr.created_at, pr.merged_at
		FROM pull_requests pr
		INNER JOIN pr_reviewers prr ON pr.pull_request_id = prr.pull_request_id
		WHERE pr.status = 'OPEN' AND prr.user_id IN (%s)
		ORDER BY pr.created_at DESC
	`, strings.Join(placeholders, ", "))

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query open PRs: %w", err)
	}
	defer rows.Close()

	var prs []*domain.PullRequest
	prMap := make(map[string]*domain.PullRequest)

	for rows.Next() {
		var pr domain.PullRequest
		var createdAt, mergedAt sql.NullTime

		if scanErr := rows.Scan(
			&pr.PullRequestID,
			&pr.PullRequestName,
			&pr.AuthorID,
			&pr.Status,
			&createdAt,
			&mergedAt,
		); scanErr != nil {
			return nil, fmt.Errorf("failed to scan PR: %w", scanErr)
		}

		if createdAt.Valid {
			pr.CreatedAt = &createdAt.Time
		}
		if mergedAt.Valid {
			pr.MergedAt = &mergedAt.Time
		}

		if _, exists := prMap[pr.PullRequestID]; !exists {
			prMap[pr.PullRequestID] = &pr
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating PRs: %w", err)
	}

	for _, pr := range prMap {
		prs = append(prs, pr)
	}

	for _, pr := range prs {
		reviewerRows, err := r.db.Query(
			"SELECT user_id FROM pr_reviewers WHERE pull_request_id = $1 ORDER BY user_id",
			pr.PullRequestID,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to query reviewers for PR %s: %w", pr.PullRequestID, err)
		}

		for reviewerRows.Next() {
			var reviewerID string
			if err := reviewerRows.Scan(&reviewerID); err != nil {
				reviewerRows.Close()
				return nil, fmt.Errorf("failed to scan reviewer: %w", err)
			}
			pr.AssignedReviewers = append(pr.AssignedReviewers, reviewerID)
		}
		reviewerRows.Close()
	}

	return prs, nil
}
