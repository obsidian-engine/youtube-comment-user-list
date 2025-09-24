package memory

import (
	"testing"
	"time"
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

	if err := repo.UpsertWithJoinTime("UC1", "User1", time1); err != nil {
		t.Fatalf("Failed to upsert UC1: %v", err)
	}
	if err := repo.UpsertWithJoinTime("UC2", "User2", time2); err != nil {
		t.Fatalf("Failed to upsert UC2: %v", err)
	}
	if err := repo.UpsertWithJoinTime("UC3", "User3", time3); err != nil {
		t.Fatalf("Failed to upsert UC3: %v", err)
	}

	// Get users sorted by join time (earliest first)
	users := repo.ListUsersSortedByJoinTime()

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
	if err := repo.UpsertWithJoinTime("UC1", "User1", originalJoinTime); err != nil {
		t.Fatalf("Failed to upsert UC1: %v", err)
	}

	// Second upsert with same channelID should not change join time
	if err := repo.UpsertWithJoinTime("UC1", "User1Updated", laterTime); err != nil {
		t.Fatalf("Failed to upsert UC1 second time: %v", err)
	}

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

func TestUserRepo_UpsertWithMessage_NewUser(t *testing.T) {
	repo := NewUserRepo()
	now := time.Now()

	// Test upserting a new user with message ID
	err := repo.UpsertWithMessage("UC123", "TestUser1", now, "msg1")
	if err != nil {
		t.Fatalf("UpsertWithMessage failed: %v", err)
	}

	users := repo.ListUsersSortedByJoinTime()
	if len(users) != 1 {
		t.Fatalf("Expected 1 user, got %d", len(users))
	}

	user := users[0]
	if user.ChannelID != "UC123" {
		t.Errorf("Expected ChannelID UC123, got %s", user.ChannelID)
	}
	if user.DisplayName != "TestUser1" {
		t.Errorf("Expected DisplayName TestUser1, got %s", user.DisplayName)
	}
	if user.CommentCount != 1 {
		t.Errorf("Expected CommentCount 1, got %d", user.CommentCount)
	}
}

func TestUserRepo_UpsertWithMessage_ExistingUser(t *testing.T) {
	repo := NewUserRepo()
	originalJoinTime := time.Now().Add(-1 * time.Hour)
	laterTime := time.Now()

	// First upsert
	if err := repo.UpsertWithMessage("UC1", "User1", originalJoinTime, "msg1"); err != nil {
		t.Fatalf("Failed to upsert UC1: %v", err)
	}

	// Second upsert with same channelID but different message ID should increment comment count
	if err := repo.UpsertWithMessage("UC1", "User1Updated", laterTime, "msg2"); err != nil {
		t.Fatalf("Failed to upsert UC1 second time: %v", err)
	}

	users := repo.ListUsersSortedByJoinTime()
	if len(users) != 1 {
		t.Fatalf("Expected 1 user, got %d", len(users))
	}

	user := users[0]
	if user.CommentCount != 2 {
		t.Errorf("Expected CommentCount 2, got %d", user.CommentCount)
	}
	if user.DisplayName != "User1Updated" {
		t.Errorf("Expected DisplayName User1Updated, got %s", user.DisplayName)
	}
	// Join time should remain original
	if !user.JoinedAt.Equal(originalJoinTime) {
		t.Errorf("Expected JoinedAt to remain %v, got %v", originalJoinTime, user.JoinedAt)
	}
}

func TestUserRepo_UpsertWithMessage_DuplicateMessage(t *testing.T) {
	repo := NewUserRepo()
	now := time.Now()

	// First upsert
	if err := repo.UpsertWithMessage("UC1", "User1", now, "msg1"); err != nil {
		t.Fatalf("Failed to upsert UC1: %v", err)
	}

	// Second upsert with same message ID should be ignored
	if err := repo.UpsertWithMessage("UC1", "User1", now, "msg1"); err != nil {
		t.Fatalf("Failed to upsert UC1 second time: %v", err)
	}

	users := repo.ListUsersSortedByJoinTime()
	if len(users) != 1 {
		t.Fatalf("Expected 1 user, got %d", len(users))
	}

	user := users[0]
	if user.CommentCount != 1 {
		t.Errorf("Expected CommentCount 1 (no increment for duplicate message), got %d", user.CommentCount)
	}
}

func TestUserRepo_UpsertWithMessage_MultipleUpdates_NoDuplicates(t *testing.T) {
	repo := NewUserRepo()
	now := time.Now()

	// Simulate multiple pull requests with same message (should only count once)
	for i := 0; i < 5; i++ {
		if err := repo.UpsertWithMessage("UC1", "User1", now, "msg1"); err != nil {
			t.Fatalf("Failed to upsert UC1 iteration %d: %v", i, err)
		}
	}

	users := repo.ListUsersSortedByJoinTime()
	if len(users) != 1 {
		t.Fatalf("Expected 1 user, got %d", len(users))
	}

	user := users[0]
	if user.CommentCount != 1 {
		t.Errorf("Expected CommentCount 1 (no duplicates), got %d", user.CommentCount)
	}

	// Add different message - should increment
	if err := repo.UpsertWithMessage("UC1", "User1", now, "msg2"); err != nil {
		t.Fatalf("Failed to upsert UC1 with msg2: %v", err)
	}

	users = repo.ListUsersSortedByJoinTime()
	user = users[0]
	if user.CommentCount != 2 {
		t.Errorf("Expected CommentCount 2 after different message, got %d", user.CommentCount)
	}
}

func TestUserRepo_Clear_RemovesProcessedMessages(t *testing.T) {
	repo := NewUserRepo()
	now := time.Now()

	// Add user with message
	if err := repo.UpsertWithMessage("UC1", "User1", now, "msg1"); err != nil {
		t.Fatalf("Failed to upsert UC1: %v", err)
	}

	// Clear should remove both users and processed messages
	repo.Clear()

	// Same message ID should now be processable again
	if err := repo.UpsertWithMessage("UC1", "User1", now, "msg1"); err != nil {
		t.Fatalf("Failed to upsert UC1 after clear: %v", err)
	}

	users := repo.ListUsersSortedByJoinTime()
	if len(users) != 1 {
		t.Fatalf("Expected 1 user after clear and re-add, got %d", len(users))
	}

	user := users[0]
	if user.CommentCount != 1 {
		t.Errorf("Expected CommentCount 1 after clear and re-add, got %d", user.CommentCount)
	}
}

func TestUserRepo_UpsertWithMessage_LatestCommentedAt(t *testing.T) {
	repo := NewUserRepo()
	firstTime := time.Now().Add(-2 * time.Hour)
	secondTime := time.Now().Add(-1 * time.Hour)
	thirdTime := time.Now()

	// First comment
	if err := repo.UpsertWithMessage("UC1", "User1", firstTime, "msg1"); err != nil {
		t.Fatalf("Failed to upsert UC1 first time: %v", err)
	}

	users := repo.ListUsersSortedByJoinTime()
	user := users[0]
	if user.LatestCommentedAt.IsZero() {
		t.Error("LatestCommentedAt should be set on first comment")
	}
	if !user.LatestCommentedAt.Equal(firstTime) {
		t.Errorf("Expected LatestCommentedAt %v, got %v", firstTime, user.LatestCommentedAt)
	}

	// Second comment (should update LatestCommentedAt)
	if err := repo.UpsertWithMessage("UC1", "User1", secondTime, "msg2"); err != nil {
		t.Fatalf("Failed to upsert UC1 second time: %v", err)
	}

	users = repo.ListUsersSortedByJoinTime()
	user = users[0]
	if !user.LatestCommentedAt.Equal(secondTime) {
		t.Errorf("Expected LatestCommentedAt to be updated to %v, got %v", secondTime, user.LatestCommentedAt)
	}

	// Third comment (should update LatestCommentedAt again)
	if err := repo.UpsertWithMessage("UC1", "User1", thirdTime, "msg3"); err != nil {
		t.Fatalf("Failed to upsert UC1 third time: %v", err)
	}

	users = repo.ListUsersSortedByJoinTime()
	user = users[0]
	if !user.LatestCommentedAt.Equal(thirdTime) {
		t.Errorf("Expected LatestCommentedAt to be updated to %v, got %v", thirdTime, user.LatestCommentedAt)
	}

	// Verify FirstCommentedAt remains unchanged
	if !user.FirstCommentedAt.Equal(firstTime) {
		t.Errorf("Expected FirstCommentedAt to remain %v, got %v", firstTime, user.FirstCommentedAt)
	}
}
