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

// ページトークンを返すフェイク
type fakeYTWithToken struct{ items []port.ChatMessage }

func (f *fakeYTWithToken) GetActiveLiveChatID(ctx context.Context, videoID string) (string, error) { return "", nil }
func (f *fakeYTWithToken) ListLiveChatMessages(ctx context.Context, liveChatID string, pageToken string) ([]port.ChatMessage, string, bool, error) {
	return f.items, "nxt", false, nil
}

// ページトークンを保存・読み出しする簡易フェイク（必要なら）
type tokenStateRepo struct {
	port.StateRepo
}

func (f *fakeYTForPull) GetActiveLiveChatID(ctx context.Context, videoID string) (string, error) {
	return "live:abc", nil
}
func (f *fakeYTForPull) ListLiveChatMessages(ctx context.Context, liveChatID string, pageToken string) ([]port.ChatMessage, string, bool, error) {
	return f.items, "", f.ended, nil
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
	yt := &fakeYTForPull{items: []port.ChatMessage{{ID: "msg1", ChannelID: "ch1", DisplayName: "Alice", PublishedAt: time.Date(2023, 1, 1, 11, 30, 0, 0, time.UTC)}}, ended: false}
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
	
	// 発言数が1であることを確認
	userList := users.ListUsersSortedByJoinTime()
	if len(userList) != 1 {
		t.Errorf("UserList length = %d, want 1", len(userList))
	}
	if userList[0].CommentCount != 1 {
		t.Errorf("CommentCount = %d, want 1", userList[0].CommentCount)
	}
	// FirstCommentedAtが設定されていることを確認
	if userList[0].FirstCommentedAt.IsZero() {
		t.Errorf("FirstCommentedAt is zero, want non-zero")
	}
}

func TestPull_MultipleComments_IncrementCount(t *testing.T) {
	ctx := context.Background()
	users := memory.NewUserRepo()
	state := memory.NewStateRepo()
	_ = state.Set(ctx, domain.LiveState{Status: domain.StatusActive, VideoID: "v", LiveChatID: "live:abc"})
	
	// ch1が2回、ch2が1回コメントするシナリオ
	yt := &fakeYTForPull{items: []port.ChatMessage{
		{ID: "msg1", ChannelID: "ch1", DisplayName: "Alice", PublishedAt: time.Date(2023, 1, 1, 11, 30, 0, 0, time.UTC)},
		{ID: "msg2", ChannelID: "ch2", DisplayName: "Bob", PublishedAt: time.Date(2023, 1, 1, 11, 35, 0, 0, time.UTC)},
		{ID: "msg3", ChannelID: "ch1", DisplayName: "Alice", PublishedAt: time.Date(2023, 1, 1, 11, 40, 0, 0, time.UTC)},
	}, ended: false}
	clock := &fakeClock{now: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)}

	uc := &usecase.Pull{YT: yt, Users: users, State: state, Clock: clock}
	_, err := uc.Execute(ctx)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// ユーザー数が2であることを確認
	if users.Count() != 2 {
		t.Errorf("Users.Count() = %d, want 2", users.Count())
	}
	
	userList := users.ListUsersSortedByJoinTime()
	// ch1の発言数が2であることを確認
	for _, user := range userList {
		switch user.ChannelID {
		case "ch1":
			if user.CommentCount != 2 {
				t.Errorf("ch1 CommentCount = %d, want 2", user.CommentCount)
			}
			if user.FirstCommentedAt.IsZero() {
				t.Errorf("ch1 FirstCommentedAt is zero, want non-zero")
			}
		case "ch2":
			if user.CommentCount != 1 {
				t.Errorf("ch2 CommentCount = %d, want 1", user.CommentCount)
			}
			if user.FirstCommentedAt.IsZero() {
				t.Errorf("ch2 FirstCommentedAt is zero, want non-zero")
			}
		}
	}
}

func TestPull_Ended_AutoReset(t *testing.T) {
	ctx := context.Background()
	users := memory.NewUserRepo()
	_ = users.UpsertWithJoinTime("ch1", "Alice", time.Now()) // 事前にユーザーを追加
	state := memory.NewStateRepo()
	startedAt := time.Date(2023, 1, 1, 10, 0, 0, 0, time.UTC)
	_ = state.Set(ctx, domain.LiveState{
		Status:     domain.StatusActive,
		VideoID:    "v",
		LiveChatID: "live:abc",
		StartedAt:  startedAt,
	})
	yt := &fakeYTForPull{items: nil, ended: true}
	endedAt := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
	clock := &fakeClock{now: endedAt}

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

	// EndedAt時刻が正確に設定されているか確認
	if !currentState.EndedAt.Equal(endedAt) {
		t.Errorf("State.EndedAt = %v, want %v", currentState.EndedAt, endedAt)
	}
}

func TestPull_WaitingState_NoOperation(t *testing.T) {
	ctx := context.Background()
	users := memory.NewUserRepo()
	state := memory.NewStateRepo()
	_ = state.Set(ctx, domain.LiveState{Status: domain.StatusWaiting, VideoID: "v", LiveChatID: ""})
	yt := &fakeYTForPull{items: []port.ChatMessage{{ID: "msg1", ChannelID: "ch1", DisplayName: "Alice", PublishedAt: time.Date(2023, 1, 1, 11, 30, 0, 0, time.UTC)}}, ended: false}
	clock := &fakeClock{now: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)}

	uc := &usecase.Pull{YT: yt, Users: users, State: state, Clock: clock}
	out, err := uc.Execute(ctx)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// WAITING状態では何も処理されない
	if out.AddedCount != 0 {
		t.Errorf("AddedCount = %d, want 0", out.AddedCount)
	}
	if out.AutoReset {
		t.Errorf("AutoReset = true, want false")
	}
	if users.Count() != 0 {
		t.Errorf("Users.Count() = %d, want 0", users.Count())
	}
}

func TestPull_SavesNextPageToken(t *testing.T) {
	users := memory.NewUserRepo()
	state := memory.NewStateRepo()
	clock := &fakeClock{now: time.Date(2023, 1, 1, 11, 30, 0, 0, time.UTC)}

	// フェイクYT: 1ページ返して nextToken は "nxt"
	yt := &fakeYTWithToken{items: []port.ChatMessage{{ID: "m1", ChannelID: "ch1", DisplayName: "Alice", PublishedAt: time.Date(2023, 1, 1, 11, 30, 0, 0, time.UTC)}}}

	// Active状態にして実行
	if err := state.Set(context.Background(), domain.LiveState{Status: domain.StatusActive, LiveChatID: "lc1"}); err != nil {
		t.Fatalf("state set error: %v", err)
	}

	uc := &usecase.Pull{YT: yt, Users: users, State: state, Clock: clock}
	if _, err := uc.Execute(context.Background()); err != nil {
		t.Fatalf("execute error: %v", err)
	}

	// 次ページトークンが保存されていること
	cur, _ := state.Get(context.Background())
	if cur.NextPageToken != "nxt" {
		t.Errorf("NextPageToken = %q, want %q", cur.NextPageToken, "nxt")
	}
}