package usecase_test

import (
	"context"
	"testing"
	"time"

	"github.com/obsidian-engine/youtube-comment-user-list/backend/internal/adapter/memory"
	"github.com/obsidian-engine/youtube-comment-user-list/backend/internal/domain"
	"github.com/obsidian-engine/youtube-comment-user-list/backend/internal/usecase"
)

func TestReset_ClearsUsersAndSetsWaiting(t *testing.T) {
	ctx := context.Background()

	users := memory.NewUserRepo()
	_ = users.UpsertWithJoinTime("ch1", "Alice", time.Now())
	_ = users.UpsertWithJoinTime("ch2", "Bob", time.Now())

	state := memory.NewStateRepo()
	_ = state.Set(ctx, domain.LiveState{Status: domain.StatusActive, VideoID: "v123"})

	uc := &usecase.Reset{Users: users, State: state}

	// Execute実行
	out, err := uc.Execute(ctx)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// ユーザーがクリアされたか確認
	if got := users.Count(); got != 0 {
		t.Errorf("Users.Count() = %d, want 0 (should be cleared)", got)
	}

	// StateがWAITINGに設定されたか確認
	currentState, _ := state.Get(ctx)
	if currentState.Status != domain.StatusWaiting {
		t.Errorf("State.Status = %v, want %v", currentState.Status, domain.StatusWaiting)
	}

	// 返り値の確認
	if out.State.Status != domain.StatusWaiting {
		t.Errorf("Output.State.Status = %v, want %v", out.State.Status, domain.StatusWaiting)
	}
}
