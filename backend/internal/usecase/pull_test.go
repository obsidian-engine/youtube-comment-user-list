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
		if user.ChannelID == "ch1" {
			if user.CommentCount != 2 {
				t.Errorf("ch1 CommentCount = %d, want 2", user.CommentCount)
			}
			if user.FirstCommentedAt.IsZero() {
				t.Errorf("ch1 FirstCommentedAt is zero, want non-zero")
			}
		} else if user.ChannelID == "ch2" {
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

func TestPull_MultipleUsers_AddedCorrectly(t *testing.T) {
	ctx := context.Background()
	users := memory.NewUserRepo()
	state := memory.NewStateRepo()
	_ = state.Set(ctx, domain.LiveState{Status: domain.StatusActive, VideoID: "v", LiveChatID: "live:abc"})
	yt := &fakeYTForPull{
		items: []port.ChatMessage{
			{ID: "msg1", ChannelID: "ch1", DisplayName: "Alice", PublishedAt: time.Date(2023, 1, 1, 11, 30, 0, 0, time.UTC)},
			{ID: "msg2", ChannelID: "ch2", DisplayName: "Bob", PublishedAt: time.Date(2023, 1, 1, 11, 35, 0, 0, time.UTC)},
			{ID: "msg3", ChannelID: "ch3", DisplayName: "Charlie", PublishedAt: time.Date(2023, 1, 1, 11, 40, 0, 0, time.UTC)},
		},
		ended: false,
	}
	clock := &fakeClock{now: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)}

	uc := &usecase.Pull{YT: yt, Users: users, State: state, Clock: clock}
	out, err := uc.Execute(ctx)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// 複数ユーザーが追加されたか確認
	if out.AddedCount != 3 {
		t.Errorf("AddedCount = %d, want 3", out.AddedCount)
	}
	if out.AutoReset {
		t.Errorf("AutoReset = true, want false")
	}
	if users.Count() != 3 {
		t.Errorf("Users.Count() = %d, want 3", users.Count())
	}
}

// 重複インクリメントの防止をテストする新しいテスト
func TestPull_DuplicateMessages_NoDuplicateIncrement(t *testing.T) {
	ctx := context.Background()
	users := memory.NewUserRepo()
	state := memory.NewStateRepo()
	_ = state.Set(ctx, domain.LiveState{Status: domain.StatusActive, VideoID: "v", LiveChatID: "live:abc"})
	
	// 同じメッセージIDを複数回受信するシナリオ（例：複数回のpull実行）
	yt := &fakeYTForPull{items: []port.ChatMessage{
		{ID: "msg1", ChannelID: "ch1", DisplayName: "Alice", PublishedAt: time.Date(2023, 1, 1, 11, 30, 0, 0, time.UTC)},
		{ID: "msg1", ChannelID: "ch1", DisplayName: "Alice", PublishedAt: time.Date(2023, 1, 1, 11, 30, 0, 0, time.UTC)}, // 重複メッセージ
		{ID: "msg2", ChannelID: "ch1", DisplayName: "Alice", PublishedAt: time.Date(2023, 1, 1, 11, 35, 0, 0, time.UTC)},
	}, ended: false}
	clock := &fakeClock{now: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)}

	uc := &usecase.Pull{YT: yt, Users: users, State: state, Clock: clock}
	out, err := uc.Execute(ctx)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// 3つのメッセージが処理されるが、重複により実際には2つのメッセージのみがカウントされる
	if out.AddedCount != 3 {
		t.Errorf("AddedCount = %d, want 3 (processing count, not unique count)", out.AddedCount)
	}
	
	// ユーザー数は1（同じユーザー）
	if users.Count() != 1 {
		t.Errorf("Users.Count() = %d, want 1", users.Count())
	}
	
	userList := users.ListUsersSortedByJoinTime()
	if len(userList) != 1 {
		t.Fatalf("UserList length = %d, want 1", len(userList))
	}
	
	// 重複メッセージが排除されるため、コメント数は2
	if userList[0].CommentCount != 2 {
		t.Errorf("CommentCount = %d, want 2 (msg1 once + msg2)", userList[0].CommentCount)
	}
}
