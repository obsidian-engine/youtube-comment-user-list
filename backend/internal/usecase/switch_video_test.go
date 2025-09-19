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

type fakeYT struct{}

func (f *fakeYT) GetActiveLiveChatID(ctx context.Context, videoID string) (string, error) {
	return "live:abc", nil
}
func (f *fakeYT) ListLiveChatMessages(ctx context.Context, liveChatID string, pageToken string) ([]port.ChatMessage, string, int64, bool, error) {
	return nil, "", 0, false, nil
}

type fixedClock struct{ t time.Time }

func (c fixedClock) Now() time.Time { return c.t }

func TestSwitchVideo_UsersClearedAndStateActive(t *testing.T) {
	ctx := context.Background()

	users := memory.NewUserRepo()
	_ = users.UpsertWithJoinTime("ch1", "to-be-cleared", time.Now())
	state := memory.NewStateRepo()
	yt := &fakeYT{}
	clock := fixedClock{t: time.Unix(1000, 0)}

	uc := &usecase.SwitchVideo{YT: yt, Users: users, State: state, Clock: clock}

	// Execute実行
	out, err := uc.Execute(ctx, usecase.SwitchVideoInput{VideoID: "video123"})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// ユーザーがクリアされたか確認
	if got := users.Count(); got != 0 {
		t.Errorf("Users.Count() = %d, want 0 (should be cleared)", got)
	}

	// Stateが正しく設定されたか確認
	currentState, _ := state.Get(ctx)
	if currentState.Status != domain.StatusActive {
		t.Errorf("State.Status = %v, want %v", currentState.Status, domain.StatusActive)
	}
	if currentState.VideoID != "video123" {
		t.Errorf("State.VideoID = %v, want video123", currentState.VideoID)
	}
	if currentState.LiveChatID != "live:abc" {
		t.Errorf("State.LiveChatID = %v, want live:abc", currentState.LiveChatID)
	}
	if !currentState.StartedAt.Equal(clock.t) {
		t.Errorf("State.StartedAt = %v, want %v", currentState.StartedAt, clock.t)
	}

	// 返り値の確認
	if out.State.Status != domain.StatusActive {
		t.Errorf("Output.State.Status = %v, want %v", out.State.Status, domain.StatusActive)
	}
}
