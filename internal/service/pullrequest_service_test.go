package service

import (
	"testing"

	"avito-tech-internship/internal/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockPullRequestRepository is a mock implementation of PullRequestRepository
type MockPullRequestRepository struct {
	mock.Mock
}

func (m *MockPullRequestRepository) CreatePR(pr *domain.PullRequest) error {
	args := m.Called(pr)
	return args.Error(0)
}

func (m *MockPullRequestRepository) GetPR(prID string) (*domain.PullRequest, error) {
	args := m.Called(prID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.PullRequest), args.Error(1)
}

func (m *MockPullRequestRepository) UpdatePR(pr *domain.PullRequest) error {
	args := m.Called(pr)
	return args.Error(0)
}

func (m *MockPullRequestRepository) MergePR(prID string) (*domain.PullRequest, error) {
	args := m.Called(prID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.PullRequest), args.Error(1)
}

func (m *MockPullRequestRepository) PRExists(prID string) (bool, error) {
	args := m.Called(prID)
	return args.Bool(0), args.Error(1)
}

func (m *MockPullRequestRepository) GetPRsByReviewer(userID string) ([]*domain.PullRequestShort, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.PullRequestShort), args.Error(1)
}

func (m *MockPullRequestRepository) ReassignReviewer(prID string, oldUserID string, newUserID string) error {
	args := m.Called(prID, oldUserID, newUserID)
	return args.Error(0)
}

func (m *MockPullRequestRepository) GetStats() (*domain.Stats, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Stats), args.Error(1)
}

func (m *MockPullRequestRepository) GetOpenPRsByReviewers(userIDs []string) ([]*domain.PullRequest, error) {
	args := m.Called(userIDs)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.PullRequest), args.Error(1)
}

// MockUserRepository is a mock implementation of UserRepository
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) GetUser(userID string) (*domain.User, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserRepository) SetIsActive(userID string, isActive bool) (*domain.User, error) {
	args := m.Called(userID, isActive)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserRepository) GetActiveUsersByTeam(teamName string, excludeUserIDs []string) ([]*domain.User, error) {
	args := m.Called(teamName, excludeUserIDs)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.User), args.Error(1)
}

func (m *MockUserRepository) CreateOrUpdateUser(user *domain.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepository) BulkSetIsActive(userIDs []string, isActive bool) error {
	args := m.Called(userIDs, isActive)
	return args.Error(0)
}

// MockTeamRepository is a mock implementation of TeamRepository
type MockTeamRepository struct {
	mock.Mock
}

func (m *MockTeamRepository) CreateTeam(team *domain.Team) error {
	args := m.Called(team)
	return args.Error(0)
}

func (m *MockTeamRepository) GetTeam(teamName string) (*domain.Team, error) {
	args := m.Called(teamName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Team), args.Error(1)
}

func (m *MockTeamRepository) TeamExists(teamName string) (bool, error) {
	args := m.Called(teamName)
	return args.Bool(0), args.Error(1)
}

func TestPullRequestService_CreatePR(t *testing.T) {
	mockPRRepo := new(MockPullRequestRepository)
	mockUserRepo := new(MockUserRepository)
	mockTeamRepo := new(MockTeamRepository)

	service := NewPullRequestService(mockPRRepo, mockUserRepo, mockTeamRepo)

	// Setup mocks
	author := &domain.User{
		UserID:   "u1",
		Username: "Alice",
		TeamName: "backend",
		IsActive: true,
	}

	candidates := []*domain.User{
		{UserID: "u2", Username: "Bob", TeamName: "backend", IsActive: true},
		{UserID: "u3", Username: "Charlie", TeamName: "backend", IsActive: true},
	}

	mockPRRepo.On("PRExists", "pr-1").Return(false, nil)
	mockUserRepo.On("GetUser", "u1").Return(author, nil)
	mockUserRepo.On("GetActiveUsersByTeam", "backend", []string{"u1"}).Return(candidates, nil)
	mockPRRepo.On("CreatePR", mock.AnythingOfType("*domain.PullRequest")).Return(nil)

	pr := &domain.PullRequest{
		PullRequestID:   "pr-1",
		PullRequestName: "Test PR",
		AuthorID:        "u1",
	}

	err := service.CreatePR(pr)
	assert.NoError(t, err)
	assert.Equal(t, domain.PRStatusOpen, pr.Status)
	assert.LessOrEqual(t, len(pr.AssignedReviewers), 2)

	mockPRRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
}

func TestPullRequestService_MergePR_Idempotent(t *testing.T) {
	mockPRRepo := new(MockPullRequestRepository)
	mockUserRepo := new(MockUserRepository)
	mockTeamRepo := new(MockTeamRepository)

	service := NewPullRequestService(mockPRRepo, mockUserRepo, mockTeamRepo)

	// First merge
	mergedPR := &domain.PullRequest{
		PullRequestID: "pr-1",
		Status:        domain.PRStatusMerged,
	}

	mockPRRepo.On("MergePR", "pr-1").Return(mergedPR, nil)

	result1, err := service.MergePR("pr-1")
	assert.NoError(t, err)
	assert.Equal(t, domain.PRStatusMerged, result1.Status)

	// Second merge (should be idempotent)
	result2, err := service.MergePR("pr-1")
	assert.NoError(t, err)
	assert.Equal(t, domain.PRStatusMerged, result2.Status)

	mockPRRepo.AssertExpectations(t)
}
