package memory

import (
	"testing"
	"time"

	"github.com/obsidian-engine/youtube-comment-user-list/backend/internal/domain"
)

// TestUserRepo_UpsertWithJoinTime tests that users are stored with join time
func TestUserRepo_UpsertWithJoinTime(t *testing.T) {
	repo := NewUserRepo()
	now := time.Now()

	// Test upserting a user should record join time
	err := repo.UpsertWithJoinTime("UC123", "TestUser1", now)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Test that the user is stored
	if repo.Count() != 1 {
		t.Fatalf("Expected 1 user, got %d", repo.Count())
	}
}

// TestUserRepo_ListUsersWithJoinTime tests that users are returned with join time info
func TestUserRepo_ListUsersWithJoinTime(t *testing.T) {
	repo := NewUserRepo()

	// Add users with different join times
	time1 := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
	time2 := time.Date(2023, 1, 1, 12, 5, 0, 0, time.UTC)
	time3 := time.Date(2023, 1, 1, 11, 55, 0, 0, time.UTC)

	repo.UpsertWithJoinTime("UC1", "User1", time1)
	repo.UpsertWithJoinTime("UC2", "User2", time2)
	repo.UpsertWithJoinTime("UC3", "User3", time3)

	// Get users sorted by join time (earliest first)
	users := repo.ListUsersSortedByJoinTime()

	// Verify we get domain.User structs
	var _ []domain.User = users

	// Should be 3 users
	if len(users) != 3 {
		t.Fatalf("Expected 3 users, got %d", len(users))
	}

	// Should be sorted by join time (earliest first)
	if users[0].DisplayName != "User3" {
		t.Errorf("Expected first user to be 'User3', got '%s'", users[0].DisplayName)
	}
	if users[1].DisplayName != "User1" {
		t.Errorf("Expected second user to be 'User1', got '%s'", users[1].DisplayName)
	}
	if users[2].DisplayName != "User2" {
		t.Errorf("Expected third user to be 'User2', got '%s'", users[2].DisplayName)
	}

	// Check join times
	if !users[0].JoinedAt.Equal(time3) {
		t.Errorf("Expected first user join time to be %v, got %v", time3, users[0].JoinedAt)
	}
	if !users[1].JoinedAt.Equal(time1) {
		t.Errorf("Expected second user join time to be %v, got %v", time1, users[1].JoinedAt)
	}
	if !users[2].JoinedAt.Equal(time2) {
		t.Errorf("Expected third user join time to be %v, got %v", time2, users[2].JoinedAt)
	}
}

// TestUserRepo_UpsertExistingUser tests that upserting an existing user doesn't change join time
func TestUserRepo_UpsertExistingUser(t *testing.T) {
	repo := NewUserRepo()

	originalJoinTime := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
	laterTime := time.Date(2023, 1, 1, 13, 0, 0, 0, time.UTC)

	// First upsert
	repo.UpsertWithJoinTime("UC1", "User1", originalJoinTime)

	// Second upsert with same channelID should not change join time
	repo.UpsertWithJoinTime("UC1", "User1Updated", laterTime)

	users := repo.ListUsersSortedByJoinTime()
	if len(users) != 1 {
		t.Fatalf("Expected 1 user, got %d", len(users))
	}

	// Join time should remain the original time
	if !users[0].JoinedAt.Equal(originalJoinTime) {
		t.Errorf("Expected join time to remain %v, got %v", originalJoinTime, users[0].JoinedAt)
	}

	// Display name should be updated
	if users[0].DisplayName != "User1Updated" {
		t.Errorf("Expected display name to be 'User1Updated', got '%s'", users[0].DisplayName)
	}
}