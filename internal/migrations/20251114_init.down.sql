-- Drop indexes
DROP INDEX IF EXISTS idx_pr_reviewers_user_id;
DROP INDEX IF EXISTS idx_pr_reviewers_pr_id;
DROP INDEX IF EXISTS idx_pull_requests_status;
DROP INDEX IF EXISTS idx_pull_requests_author;
DROP INDEX IF EXISTS idx_users_team_active;
DROP INDEX IF EXISTS idx_users_is_active;
DROP INDEX IF EXISTS idx_users_team_name;

-- Drop tables in reverse order
DROP TABLE IF EXISTS pr_reviewers;
DROP TABLE IF EXISTS pull_requests;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS teams;

