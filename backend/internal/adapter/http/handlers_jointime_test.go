package http

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/obsidian-engine/youtube-comment-user-list/backend/internal/domain"
)

// TestHandlers_UsersEndpointWithJoinTime tests that /users.json returns users with join time
func TestHandlers_UsersEndpointWithJoinTime(t *testing.T) {
	// Create mock user repo that supports join time
	users := &MockUserRepoWithJoinTime{
		users: []domain.User{
			{
				ChannelID:   "UC1",
				DisplayName: "User1",
				JoinedAt:    time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
			},
			{
				ChannelID:   "UC2",
				DisplayName: "User2",
				JoinedAt:    time.Date(2023, 1, 1, 11, 55, 0, 0, time.UTC),
			},
			{
				ChannelID:   "UC3",
				DisplayName: "User3",
				JoinedAt:    time.Date(2023, 1, 1, 12, 5, 0, 0, time.UTC),
			},
		},
	}

	// Create handlers with mock repo
	h := &Handlers{
		Users: users,
	}

	// Create router
	router := NewRouter(h, "http://localhost:5173")

	// Test /users.json endpoint
	req := httptest.NewRequest("GET", "/users.json", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Check status code
	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", w.Code)
	}

	// Parse response
	var response []domain.User
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to parse JSON response: %v", err)
	}

	// Should return 3 users
	if len(response) != 3 {
		t.Fatalf("Expected 3 users, got %d", len(response))
	}

	// Should be sorted by join time (earliest first)
	expectedOrder := []string{"User2", "User1", "User3"}
	for i, expectedName := range expectedOrder {
		if response[i].DisplayName != expectedName {
			t.Errorf("Expected user %d to be '%s', got '%s'", i, expectedName, response[i].DisplayName)
		}
	}

	// Check that join time field is present and correct
	expectedJoinTimes := []time.Time{
		time.Date(2023, 1, 1, 11, 55, 0, 0, time.UTC), // User2
		time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),  // User1
		time.Date(2023, 1, 1, 12, 5, 0, 0, time.UTC),  // User3
	}

	for i, expectedTime := range expectedJoinTimes {
		if !response[i].JoinedAt.Equal(expectedTime) {
			t.Errorf("Expected user %d join time to be %v, got %v", i, expectedTime, response[i].JoinedAt)
		}
	}

	// Verify JSON field name is "joinedAt"
	var rawResponse []map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &rawResponse)
	if err != nil {
		t.Fatalf("Failed to parse JSON response as map: %v", err)
	}

	for i, user := range rawResponse {
		if _, exists := user["joinedAt"]; !exists {
			t.Errorf("User %d should have 'joinedAt' field", i)
		}
		if _, exists := user["channelId"]; !exists {
			t.Errorf("User %d should have 'channelId' field", i)
		}
		if _, exists := user["displayName"]; !exists {
			t.Errorf("User %d should have 'displayName' field", i)
		}
	}
}

// MockUserRepoWithJoinTime is a mock implementation for testing
type MockUserRepoWithJoinTime struct {
	users []domain.User
}

// Legacy methods removed - not needed

func (m *MockUserRepoWithJoinTime) ListUsersSortedByJoinTime() []domain.User {
	// Return users sorted by join time (earliest first)
	sorted := make([]domain.User, len(m.users))
	copy(sorted, m.users)

	// Sort by join time
	for i := 0; i < len(sorted)-1; i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[i].JoinedAt.After(sorted[j].JoinedAt) {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	return sorted
}

func (m *MockUserRepoWithJoinTime) Count() int {
	return len(m.users)
}

func (m *MockUserRepoWithJoinTime) UpsertWithJoinTime(channelID string, displayName string, joinedAt time.Time) error {
	// Not needed for this test but required by interface
	return nil
}

func (m *MockUserRepoWithJoinTime) Clear() {
	m.users = []domain.User{}
}