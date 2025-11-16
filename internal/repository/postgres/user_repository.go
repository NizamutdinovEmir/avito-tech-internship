package postgres

import (
	"database/sql"
	"fmt"
	"strings"

	"avito-tech-internship/internal/domain"
	"avito-tech-internship/internal/repository"
)

type userRepository struct {
	db *sql.DB
}

// NewUserRepository creates a new PostgreSQL user repository
func NewUserRepository(db *sql.DB) *userRepository {
	return &userRepository{db: db}
}

func (r *userRepository) GetUser(userID string) (*domain.User, error) {
	var user domain.User
	err := r.db.QueryRow(
		"SELECT user_id, username, team_name, is_active FROM users WHERE user_id = $1",
		userID,
	).Scan(&user.UserID, &user.Username, &user.TeamName, &user.IsActive)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, repository.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return &user, nil
}

func (r *userRepository) SetIsActive(userID string, isActive bool) (*domain.User, error) {
	_, err := r.db.Exec(
		"UPDATE users SET is_active = $1, updated_at = CURRENT_TIMESTAMP WHERE user_id = $2",
		isActive, userID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update user activity: %w", err)
	}

	return r.GetUser(userID)
}

func (r *userRepository) GetActiveUsersByTeam(teamName string, excludeUserIDs []string) ([]*domain.User, error) {
	query := "SELECT user_id, username, team_name, is_active FROM users WHERE team_name = $1 AND is_active = true"
	args := []interface{}{teamName}

	if len(excludeUserIDs) > 0 {
		placeholders := make([]string, len(excludeUserIDs))
		for i, id := range excludeUserIDs {
			args = append(args, id)
			placeholders[i] = fmt.Sprintf("$%d", i+2) // +2 because $1 is teamName
		}
		query += fmt.Sprintf(" AND user_id NOT IN (%s)", strings.Join(placeholders, ", "))
	}

	query += " ORDER BY user_id"

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query active users: %w", err)
	}
	defer rows.Close()

	var users []*domain.User
	for rows.Next() {
		var user domain.User
		if err := rows.Scan(&user.UserID, &user.Username, &user.TeamName, &user.IsActive); err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, &user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating users: %w", err)
	}

	return users, nil
}

func (r *userRepository) CreateOrUpdateUser(user *domain.User) error {
	_, err := r.db.Exec(
		`INSERT INTO users (user_id, username, team_name, is_active) 
		 VALUES ($1, $2, $3, $4)
		 ON CONFLICT (user_id) 
		 DO UPDATE SET username = $2, team_name = $3, is_active = $4, updated_at = CURRENT_TIMESTAMP`,
		user.UserID, user.Username, user.TeamName, user.IsActive,
	)
	return err
}

func (r *userRepository) BulkSetIsActive(userIDs []string, isActive bool) error {
	if len(userIDs) == 0 {
		return nil
	}

	placeholders := make([]string, len(userIDs))
	args := make([]interface{}, len(userIDs)+1)
	args[0] = isActive
	for i, userID := range userIDs {
		args[i+1] = userID
		placeholders[i] = fmt.Sprintf("$%d", i+2)
	}

	query := fmt.Sprintf(`
		UPDATE users 
		SET is_active = $1, updated_at = CURRENT_TIMESTAMP 
		WHERE user_id IN (%s)
	`, strings.Join(placeholders, ", "))

	_, err := r.db.Exec(query, args...)
	return err
}
