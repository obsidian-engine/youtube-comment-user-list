package usecase_test

import (
	"context"
	"testing"
	"time"

	"github.com/obsidian-engine/youtube-comment-user-list/backend/internal/adapter/memory"
	"github.com/obsidian-engine/youtube-comment-user-list/backend/internal/domain"
	"github.com/obsidian-engine/youtube-comment-user-list/backend/internal/usecase"
)

func TestStatus_ReturnsCurrentStateAndUserCount(t *testing.T) {
	ctx := context.Background()

	users := memory.NewUserRepo()
	_ = users.Upsert("ch1", "Alice")
	_ = users.Upsert("ch2", "Bob")

	state := memory.NewStateRepo()
	startedAt := time.Unix(1000, 0)
	_ = state.Set(ctx, domain.LiveState{
		Status:    domain.StatusActive,
		VideoID:   "video123",
		StartedAt: startedAt,
	})

	uc := &usecase.Status{Users: users, State: state}

	// Execute実行
	out, err := uc.Execute(ctx)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// 結果確認
	if out.Status != domain.StatusActive {
		t.Errorf("Status = %v, want %v", out.Status, domain.StatusActive)
	}
	if out.Count != 2 {
		t.Errorf("Count = %d, want 2", out.Count)
	}
	if out.VideoID != "video123" {
		t.Errorf("VideoID = %v, want video123", out.VideoID)
	}
	if !out.StartedAt.Equal(startedAt) {
		t.Errorf("StartedAt = %v, want %v", out.StartedAt, startedAt)
	}
}