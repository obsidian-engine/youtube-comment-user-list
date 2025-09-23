package memory

import (
	"testing"
	"time"
)

func TestUpsertWithMessageUpdated(t *testing.T) {
	repo := NewUserRepo()
	joinedAt := time.Now()

	t.Run("first message should return updated=true", func(t *testing.T) {
		updated, err := repo.UpsertWithMessageUpdated("user1", "User One", joinedAt, "msg1")
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if !updated {
			t.Errorf("Expected updated=true for first message, got %v", updated)
		}

		// Verify user was created
		users := repo.ListUsersSortedByJoinTime()
		if len(users) != 1 {
			t.Fatalf("Expected 1 user, got %d", len(users))
		}
		if users[0].ChannelID != "user1" {
			t.Errorf("Expected ChannelID=user1, got %s", users[0].ChannelID)
		}
		if users[0].CommentCount != 1 {
			t.Errorf("Expected CommentCount=1, got %d", users[0].CommentCount)
		}
	})

	t.Run("duplicate message should return updated=false", func(t *testing.T) {
		updated, err := repo.UpsertWithMessageUpdated("user1", "User One Updated", joinedAt, "msg1")
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if updated {
			t.Errorf("Expected updated=false for duplicate message, got %v", updated)
		}

		// Verify user count didn't change
		users := repo.ListUsersSortedByJoinTime()
		if len(users) != 1 {
			t.Fatalf("Expected 1 user, got %d", len(users))
		}
		// Verify comment count didn't increase
		if users[0].CommentCount != 1 {
			t.Errorf("Expected CommentCount=1 (no change), got %d", users[0].CommentCount)
		}
	})

	t.Run("new message from existing user should return updated=true", func(t *testing.T) {
		updated, err := repo.UpsertWithMessageUpdated("user1", "User One Updated", joinedAt, "msg2")
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if !updated {
			t.Errorf("Expected updated=true for new message from existing user, got %v", updated)
		}

		// Verify comment count increased
		users := repo.ListUsersSortedByJoinTime()
		if len(users) != 1 {
			t.Fatalf("Expected 1 user, got %d", len(users))
		}
		if users[0].CommentCount != 2 {
			t.Errorf("Expected CommentCount=2, got %d", users[0].CommentCount)
		}
		if users[0].DisplayName != "User One Updated" {
			t.Errorf("Expected DisplayName updated, got %s", users[0].DisplayName)
		}
	})

	t.Run("new user with new message should return updated=true", func(t *testing.T) {
		updated, err := repo.UpsertWithMessageUpdated("user2", "User Two", joinedAt, "msg3")
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if !updated {
			t.Errorf("Expected updated=true for new user, got %v", updated)
		}

		// Verify new user was created
		users := repo.ListUsersSortedByJoinTime()
		if len(users) != 2 {
			t.Fatalf("Expected 2 users, got %d", len(users))
		}
	})
}

// Test backward compatibility: old UpsertWithMessage should still exist and work
func TestUpsertWithMessageBackwardCompatibility(t *testing.T) {
	repo := NewUserRepo()
	joinedAt := time.Now()

	// Test that old method still works
	err := repo.UpsertWithMessage("user1", "User One", joinedAt, "msg1")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	users := repo.ListUsersSortedByJoinTime()
	if len(users) != 1 {
		t.Fatalf("Expected 1 user, got %d", len(users))
	}

	// Test duplicate with old method
	err = repo.UpsertWithMessage("user1", "User One", joinedAt, "msg1")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	users = repo.ListUsersSortedByJoinTime()
	if len(users) != 1 {
		t.Fatalf("Expected still 1 user, got %d", len(users))
	}
	if users[0].CommentCount != 1 {
		t.Errorf("Expected CommentCount=1 (no increase for duplicate), got %d", users[0].CommentCount)
	}
}