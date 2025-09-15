package usecase_test

import (
	"context"
	"testing"
	"time"

	"github.com/obsidian-engine/youtube-comment-user-list/backend/internal/adapter/memory"
	"github.com/obsidian-engine/youtube-comment-user-list/backend/internal/domain"
	"github.com/obsidian-engine/youtube-comment-user-list/backend/internal/port"
	"github.com/obsidian-engine/youtube-comment-user-list/backend/internal/usecase"
)

type fakeYTForPull struct {
	items []port.ChatMessage
	ended bool
}

func (f *fakeYTForPull) GetActiveLiveChatID(ctx context.Context, videoID string) (string, error) {
	return "live:abc", nil
}
func (f *fakeYTForPull) ListLiveChatMessages(ctx context.Context, liveChatID string) ([]port.ChatMessage, bool, error) {
	return f.items, f.ended, nil
}

type fakeClock struct {
	now time.Time
}

func (f *fakeClock) Now() time.Time {
	return f.now
}

func TestPull_AddsUsers_NormalFlow(t *testing.T) {
	ctx := context.Background()
	users := memory.NewUserRepo()
	state := memory.NewStateRepo()
	_ = state.Set(ctx, domain.LiveState{Status: domain.StatusActive, VideoID: "v", LiveChatID: "live:abc"})
	yt := &fakeYTForPull{items: []port.ChatMessage{{ChannelID: "ch1", DisplayName: "Alice"}}, ended: false}
	clock := &fakeClock{now: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)}

	uc := &usecase.Pull{YT: yt, Users: users, State: state, Clock: clock}
	out, err := uc.Execute(ctx)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// ユーザーが追加されたか確認
	if out.AddedCount != 1 {
		t.Errorf("AddedCount = %d, want 1", out.AddedCount)
	}
	if out.AutoReset {
		t.Errorf("AutoReset = true, want false")
	}
	if users.Count() != 1 {
		t.Errorf("Users.Count() = %d, want 1", users.Count())
	}
}

func TestPull_Ended_AutoReset(t *testing.T) {
	ctx := context.Background()
	users := memory.NewUserRepo()
	_ = users.UpsertWithJoinTime("ch1", "Alice", time.Now()) // 事前にユーザーを追加
	state := memory.NewStateRepo()
	_ = state.Set(ctx, domain.LiveState{Status: domain.StatusActive, VideoID: "v", LiveChatID: "live:abc"})
	yt := &fakeYTForPull{items: nil, ended: true}
	clock := &fakeClock{now: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)}

	uc := &usecase.Pull{YT: yt, Users: users, State: state, Clock: clock}
	out, err := uc.Execute(ctx)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// AutoResetがtrueか確認
	if !out.AutoReset {
		t.Errorf("AutoReset = false, want true")
	}

	// ユーザーがクリアされたか確認
	if users.Count() != 0 {
		t.Errorf("Users.Count() = %d, want 0", users.Count())
	}

	// StateがWAITINGに戻ったか確認
	currentState, _ := state.Get(ctx)
	if currentState.Status != domain.StatusWaiting {
		t.Errorf("State.Status = %v, want %v", currentState.Status, domain.StatusWaiting)
	}
}
