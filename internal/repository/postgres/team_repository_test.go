package postgres

import (
	"database/sql"
	"fmt"
	"os"
	"testing"

	"avito-tech-internship/internal/domain"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupTestDB creates a test database connection
func setupTestDB(t *testing.T) *sql.DB {
	// Use test database if specified, otherwise use main database
	testDB := os.Getenv("TEST_DB_NAME")
	if testDB == "" {
		testDB = "avito_db"
	}

	dsn := fmt.Sprintf("host=localhost port=5432 user=avito password=avito dbname=%s sslmode=disable", testDB)
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Skipf("Skipping test: failed to connect to test database: %v", err)
		return nil
	}

	// Test connection
	if err := db.Ping(); err != nil {
		db.Close()
		t.Skipf("Skipping test: database %s not available: %v", testDB, err)
		return nil
	}

	// Clean up tables before test
	cleanupTestDB(t, db)

	return db
}

func cleanupTestDB(t *testing.T, db *sql.DB) {
	if db == nil {
		return
	}
	tables := []string{"pr_reviewers", "pull_requests", "users", "teams", "schema_migrations"}
	for _, table := range tables {
		_, err := db.Exec("TRUNCATE TABLE " + table + " CASCADE")
		if err != nil {
			// Ignore errors for non-existent tables in test environment
			t.Logf("Note: Could not truncate %s (may not exist): %v", table, err)
		}
	}
}

func TestTeamRepository_CreateTeam(t *testing.T) {
	db := setupTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()
	defer cleanupTestDB(t, db)

	repo := NewTeamRepository(db)

	team := &domain.Team{
		TeamName: "test-team",
		Members: []domain.TeamMember{
			{UserID: "u1", Username: "Alice", IsActive: true},
			{UserID: "u2", Username: "Bob", IsActive: true},
		},
	}

	err := repo.CreateTeam(team)
	require.NoError(t, err)

	// Verify team was created
	exists, err := repo.TeamExists("test-team")
	require.NoError(t, err)
	assert.True(t, exists)
}

func TestTeamRepository_GetTeam(t *testing.T) {
	db := setupTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()
	defer cleanupTestDB(t, db)

	repo := NewTeamRepository(db)

	// Create team first
	team := &domain.Team{
		TeamName: "test-team",
		Members: []domain.TeamMember{
			{UserID: "u1", Username: "Alice", IsActive: true},
		},
	}
	err := repo.CreateTeam(team)
	require.NoError(t, err)

	// Get team
	retrievedTeam, err := repo.GetTeam("test-team")
	require.NoError(t, err)
	assert.Equal(t, "test-team", retrievedTeam.TeamName)
	assert.Len(t, retrievedTeam.Members, 1)
	assert.Equal(t, "u1", retrievedTeam.Members[0].UserID)
}

func TestTeamRepository_TeamExists(t *testing.T) {
	db := setupTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()
	defer cleanupTestDB(t, db)

	repo := NewTeamRepository(db)

	exists, err := repo.TeamExists("non-existent")
	require.NoError(t, err)
	assert.False(t, exists)
}
