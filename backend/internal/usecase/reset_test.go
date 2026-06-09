package usecase_test

import (
	"context"
	"testing"
	"time"

	"github.com/obsidian-engine/youtube-comment-user-list/backend/internal/adapter/memory"
	"github.com/obsidian-engine/youtube-comment-user-list/backend/internal/domain"
	"github.com/obsidian-engine/youtube-comment-user-list/backend/internal/usecase"
	"github.com/obsidian-engine/youtube-comment-user-list/backend/internal/usecase/snapshot"
)

func TestReset_ClearsUsersAndSetsWaiting(t *testing.T) {
	ctx := context.Background()

	users := memory.NewUserRepo()
	_ = users.UpsertWithJoinTime("ch1", "Alice", time.Now())
	_ = users.UpsertWithJoinTime("ch2", "Bob", time.Now())

	state := memory.NewStateRepo()
	_ = state.Set(ctx, domain.LiveState{Status: domain.StatusActive, VideoID: "v123"})

	uc := &usecase.Reset{Users: users, State: state, Snap: &snapshot.NopCoordinator{}}

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

// TestReset_ClearsComments: Reset 実行でコメントもクリアされる
func TestReset_ClearsComments(t *testing.T) {
	ctx := context.Background()

	users := memory.NewUserRepo()
	comments := memory.NewCommentRepo()
	_ = comments.Add(domain.Comment{ID: "c1", ChannelID: "ch1", DisplayName: "Alice", Message: "hello", PublishedAt: time.Now()})
	_ = comments.Add(domain.Comment{ID: "c2", ChannelID: "ch2", DisplayName: "Bob", Message: "world", PublishedAt: time.Now()})

	state := memory.NewStateRepo()
	_ = state.Set(ctx, domain.LiveState{Status: domain.StatusActive, VideoID: "v123"})

	uc := &usecase.Reset{Users: users, Comments: comments, State: state, Snap: &snapshot.NopCoordinator{}}

	_, err := uc.Execute(ctx)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if got := comments.Count(); got != 0 {
		t.Errorf("Comments.Count() = %d, want 0 (should be cleared on Reset)", got)
	}
}
