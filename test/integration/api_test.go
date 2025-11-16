package integration

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"avito-tech-internship/internal/domain"
	"avito-tech-internship/internal/migrations"
	"avito-tech-internship/internal/router"
	"avito-tech-internship/pkg/migrate"

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
		t.Skipf("Skipping E2E test: failed to connect to database: %v", err)
		return nil
	}

	if err := db.Ping(); err != nil {
		db.Close()
		t.Skipf("Skipping E2E test: database %s not available: %v", testDB, err)
		return nil
	}

	if err := migrate.RunMigrations(db, migrations.FS); err != nil {
		db.Close()
		t.Fatalf("Failed to run migrations: %v", err)
	}

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
			t.Logf("Note: Could not truncate %s: %v", table, err)
		}
	}
}

// TestE2EPlaceholder is a placeholder test that documents the E2E test structure.
// This test is skipped by default as it requires a full test database setup.
func TestE2EPlaceholder(t *testing.T) {
	t.Skip("E2E tests require full test database setup (testcontainers or similar)")
}

// TestCreateTeamE2E is a real E2E test that requires a running database
func TestCreateTeamE2E(t *testing.T) {
	db := setupTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()
	defer cleanupTestDB(t, db)

	router := router.SetupRouter(db)

	// Create team via API
	team := domain.Team{
		TeamName: "backend",
		Members: []domain.TeamMember{
			{UserID: "u1", Username: "Alice", IsActive: true},
			{UserID: "u2", Username: "Bob", IsActive: true},
		},
	}

	teamJSON, err := json.Marshal(team)
	require.NoError(t, err)

	req := httptest.NewRequest("POST", "/team/add", bytes.NewBuffer(teamJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code, "Expected status 201, got %d: %s", w.Code, w.Body.String())

	t.Logf("POST response body: %s", w.Body.String())

	var userCount int
	err = db.QueryRow("SELECT COUNT(*) FROM users WHERE team_name = $1", "backend").Scan(&userCount)
	require.NoError(t, err)
	t.Logf("Users in database for team 'backend': %d", userCount)
	assert.Equal(t, 2, userCount, "Expected 2 users in database")

	getReq := httptest.NewRequest("GET", "/team/get?team_name=backend", nil)
	getW := httptest.NewRecorder()

	router.ServeHTTP(getW, getReq)

	assert.Equal(t, http.StatusOK, getW.Code, "GET response: %s", getW.Body.String())

	t.Logf("GET response body: %s", getW.Body.String())

	var retrievedTeam domain.Team
	err = json.Unmarshal(getW.Body.Bytes(), &retrievedTeam)
	require.NoError(t, err, "Failed to unmarshal response: %s", getW.Body.String())
	t.Logf("Retrieved team: %+v", retrievedTeam)
	assert.Equal(t, "backend", retrievedTeam.TeamName)
	assert.Len(t, retrievedTeam.Members, 2, "Expected 2 members, got: %+v", retrievedTeam.Members)
}
