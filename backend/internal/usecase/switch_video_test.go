package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/obsidian-engine/youtube-comment-user-list/backend/internal/adapter/memory"
	"github.com/obsidian-engine/youtube-comment-user-list/backend/internal/domain"
	"github.com/obsidian-engine/youtube-comment-user-list/backend/internal/port"
	"github.com/obsidian-engine/youtube-comment-user-list/backend/internal/usecase"
	"github.com/obsidian-engine/youtube-comment-user-list/backend/internal/usecase/snapshot"
)

type fakeYT struct{}

func (f *fakeYT) GetActiveLiveChatID(ctx context.Context, videoID string) (string, error) {
	return "live:abc", nil
}

type failingYT struct{}

func (f *failingYT) GetActiveLiveChatID(ctx context.Context, videoID string) (string, error) {
	return "", errors.New("api error: live broadcast not found")
}

func (f *failingYT) ListLiveChatMessages(ctx context.Context, liveChatID string, pageToken string) ([]port.ChatMessage, string, int64, int, bool, error) {
	return nil, "", 0, 0, false, nil
}

func (f *failingYT) GetChannelDisplayNames(ctx context.Context, channelIDs []string) (map[string]string, error) {
	return nil, nil
}
func (f *fakeYT) ListLiveChatMessages(ctx context.Context, liveChatID string, pageToken string) ([]port.ChatMessage, string, int64, int, bool, error) {
	return nil, "", 0, 0, false, nil
}
func (f *fakeYT) GetChannelDisplayNames(ctx context.Context, channelIDs []string) (map[string]string, error) {
	return nil, nil
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

	uc := &usecase.SwitchVideo{YT: yt, Users: users, State: state, Clock: clock, Snap: &snapshot.NopCoordinator{}}

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

func TestSwitchVideo_SameVideoID_PreservesUsersAndStartedAt(t *testing.T) {
	ctx := context.Background()

	users := memory.NewUserRepo()
	_ = users.UpsertWithJoinTime("ch1", "Alice", time.Now())
	_ = users.UpsertWithJoinTime("ch2", "Bob", time.Now())

	originalStartedAt := time.Date(2023, 1, 1, 10, 0, 0, 0, time.UTC)
	state := memory.NewStateRepo()
	_ = state.Set(ctx, domain.LiveState{
		Status:     domain.StatusWaiting,
		VideoID:    "video123",
		LiveChatID: "live:abc",
		StartedAt:  originalStartedAt,
		EndedAt:    time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
	})

	yt := &fakeYT{}
	clock := fixedClock{t: time.Date(2023, 1, 1, 13, 0, 0, 0, time.UTC)}

	uc := &usecase.SwitchVideo{YT: yt, Users: users, State: state, Clock: clock, Snap: &snapshot.NopCoordinator{}}

	if _, err := uc.Execute(ctx, usecase.SwitchVideoInput{VideoID: "video123"}); err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// 同一 videoId への切替なので users は保持される
	if got := users.Count(); got != 2 {
		t.Errorf("Users.Count() = %d, want 2 (preserved on same-video re-switch)", got)
	}

	// StartedAt も元の値を維持する（監視時間カウンタが巻き戻らないため）
	currentState, _ := state.Get(ctx)
	if !currentState.StartedAt.Equal(originalStartedAt) {
		t.Errorf("State.StartedAt = %v, want %v (preserved)", currentState.StartedAt, originalStartedAt)
	}
	if currentState.Status != domain.StatusActive {
		t.Errorf("State.Status = %v, want %v", currentState.Status, domain.StatusActive)
	}
}

func TestSwitchVideo_APIError_SameVideoID_RestoresFromMemory(t *testing.T) {
	ctx := context.Background()

	originalStartedAt := time.Date(2023, 1, 1, 10, 0, 0, 0, time.UTC)
	users := memory.NewUserRepo()
	_ = users.UpsertWithJoinTime("ch1", "Alice", time.Now())
	_ = users.UpsertWithJoinTime("ch2", "Bob", time.Now())

	state := memory.NewStateRepo()
	_ = state.Set(ctx, domain.LiveState{
		Status:     domain.StatusWaiting,
		VideoID:    "video123",
		LiveChatID: "live:abc",
		StartedAt:  originalStartedAt,
	})

	yt := &failingYT{}
	now := time.Date(2023, 1, 1, 14, 0, 0, 0, time.UTC)
	clock := fixedClock{t: now}

	uc := &usecase.SwitchVideo{YT: yt, Users: users, State: state, Clock: clock, Snap: &snapshot.NopCoordinator{}}

	out, err := uc.Execute(ctx, usecase.SwitchVideoInput{VideoID: "video123"})
	if err != nil {
		t.Fatalf("Execute should not return error on fallback, got: %v", err)
	}

	if out.State.Status != domain.StatusWaiting {
		t.Errorf("Output.State.Status = %v, want %v", out.State.Status, domain.StatusWaiting)
	}
	if out.State.VideoID != "video123" {
		t.Errorf("Output.State.VideoID = %v, want video123", out.State.VideoID)
	}
	if !out.State.StartedAt.Equal(originalStartedAt) {
		t.Errorf("Output.State.StartedAt = %v, want %v (preserved)", out.State.StartedAt, originalStartedAt)
	}
	if !out.State.EndedAt.Equal(now) {
		t.Errorf("Output.State.EndedAt = %v, want %v", out.State.EndedAt, now)
	}
	// users は維持されている
	if got := users.Count(); got != 2 {
		t.Errorf("Users.Count() = %d, want 2 (preserved on fallback)", got)
	}
}

func TestSwitchVideo_APIError_NewVideoID_ReturnsError(t *testing.T) {
	ctx := context.Background()

	users := memory.NewUserRepo()
	_ = users.UpsertWithJoinTime("ch1", "Alice", time.Now())

	state := memory.NewStateRepo()
	_ = state.Set(ctx, domain.LiveState{
		Status:  domain.StatusWaiting,
		VideoID: "video-old",
	})

	yt := &failingYT{}
	clock := fixedClock{t: time.Now()}

	uc := &usecase.SwitchVideo{YT: yt, Users: users, State: state, Clock: clock, Snap: &snapshot.NopCoordinator{}}

	_, err := uc.Execute(ctx, usecase.SwitchVideoInput{VideoID: "video-new"})
	if err == nil {
		t.Fatal("Execute should return error when videoId differs and API fails")
	}
}

func TestSwitchVideo_APIError_SameVideoID_NoUsers_ReturnsError(t *testing.T) {
	ctx := context.Background()

	users := memory.NewUserRepo() // 空

	state := memory.NewStateRepo()
	_ = state.Set(ctx, domain.LiveState{
		Status:  domain.StatusWaiting,
		VideoID: "video123",
	})

	yt := &failingYT{}
	clock := fixedClock{t: time.Now()}

	uc := &usecase.SwitchVideo{YT: yt, Users: users, State: state, Clock: clock, Snap: &snapshot.NopCoordinator{}}

	_, err := uc.Execute(ctx, usecase.SwitchVideoInput{VideoID: "video123"})
	if err == nil {
		t.Fatal("Execute should return error when users are empty even on same videoId")
	}
}
