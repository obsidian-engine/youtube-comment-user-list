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

// buildCoordWithSink は fakeSinkForUsecase を使って snapshot.Coordinator を生成するヘルパー。
func buildCoordWithSink(sink *fakeSinkForUsecase, users *memory.UserRepo, comments *memory.CommentRepo, state port.StateRepo) snapshot.Coordinator {
	return snapshot.NewCoordinator(sink, users, comments, state, 0)
}

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

// TestSwitchVideo_DifferentVideo_ClearsComments: 別 video 切替で CommentRepo がクリアされる
func TestSwitchVideo_DifferentVideo_ClearsComments(t *testing.T) {
	ctx := context.Background()

	users := memory.NewUserRepo()
	comments := memory.NewCommentRepo()
	_ = comments.Add(domain.Comment{ID: "c1", ChannelID: "ch1", DisplayName: "Alice", Message: "hello", PublishedAt: time.Now()})
	_ = comments.Add(domain.Comment{ID: "c2", ChannelID: "ch2", DisplayName: "Bob", Message: "world", PublishedAt: time.Now()})
	state := memory.NewStateRepo()
	_ = state.Set(ctx, domain.LiveState{Status: domain.StatusActive, VideoID: "video-old"})

	yt := &fakeYT{}
	clock := fixedClock{t: time.Now()}
	sink := newFakeSinkForUsecase()
	coord := buildCoordWithSink(sink, users, comments, state)

	uc := &usecase.SwitchVideo{YT: yt, Users: users, Comments: comments, State: state, Clock: clock, Snap: coord}

	_, err := uc.Execute(ctx, usecase.SwitchVideoInput{VideoID: "video-new"})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if got := comments.Count(); got != 0 {
		t.Errorf("Comments.Count() = %d, want 0 (should be cleared on different video)", got)
	}
}

// TestSwitchVideo_V1_to_V2_to_V1_RestoresFromGCS: V1→V2→V1 切替でユーザーが GCS から復元される
func TestSwitchVideo_V1_to_V2_to_V1_RestoresFromGCS(t *testing.T) {
	ctx := context.Background()

	users := memory.NewUserRepo()
	comments := memory.NewCommentRepo()
	state := memory.NewStateRepo()
	_ = state.Set(ctx, domain.LiveState{Status: domain.StatusActive, VideoID: "v1", LiveChatID: "chat-v1"})

	yt := &fakeYT{}
	clock := fixedClock{t: time.Now()}
	sink := newFakeSinkForUsecase()
	coord := buildCoordWithSink(sink, users, comments, state)

	// V1 配信: ユーザー追加
	_ = users.UpsertWithJoinTime("ch1", "Alice", time.Now())
	_ = users.UpsertWithJoinTime("ch2", "Bob", time.Now())
	_ = comments.Add(domain.Comment{ID: "c1", ChannelID: "ch1", DisplayName: "Alice", Message: "hello", PublishedAt: time.Now()})

	// V1 状態を coordinator に設定して Flush → V1.json 保存
	coord.SetVideo("v1", "chat-v1")
	coord.MarkDirty()
	if err := coord.Flush(ctx); err != nil {
		t.Fatalf("V1 Flush failed: %v", err)
	}

	// V2 に切替
	uc := &usecase.SwitchVideo{YT: yt, Users: users, Comments: comments, State: state, Clock: clock, Snap: coord}
	_, err := uc.Execute(ctx, usecase.SwitchVideoInput{VideoID: "v2"})
	if err != nil {
		t.Fatalf("Switch to V2 failed: %v", err)
	}

	// V2 切替後は users/comments がクリアされている
	if got := users.Count(); got != 0 {
		t.Errorf("after V2 switch: Users.Count() = %d, want 0", got)
	}
	if got := comments.Count(); got != 0 {
		t.Errorf("after V2 switch: Comments.Count() = %d, want 0", got)
	}

	// V1 に戻す → GCS から復元
	_, err = uc.Execute(ctx, usecase.SwitchVideoInput{VideoID: "v1"})
	if err != nil {
		t.Fatalf("Switch back to V1 failed: %v", err)
	}

	// V1 の users が復元されている
	if got := users.Count(); got != 2 {
		t.Errorf("after V1 restore: Users.Count() = %d, want 2 (restored from GCS)", got)
	}
	if got := comments.Count(); got != 1 {
		t.Errorf("after V1 restore: Comments.Count() = %d, want 1 (restored from GCS)", got)
	}
}

// TestSwitchVideo_APIError_RestoreFromGCS_succeeds: V1 終了済 + Users 空 + GCS に V1 snapshot あり → 復元成功
func TestSwitchVideo_APIError_RestoreFromGCS_succeeds(t *testing.T) {
	ctx := context.Background()

	users := memory.NewUserRepo()
	comments := memory.NewCommentRepo()
	state := memory.NewStateRepo()
	// prevState は別 video (cold start 後などを想定)
	_ = state.Set(ctx, domain.LiveState{
		Status:  domain.StatusWaiting,
		VideoID: "v0",
	})

	yt := &failingYT{}
	now := time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC)
	clock := fixedClock{t: now}

	// GCS 側に v1 snapshot を事前保存
	sink := newFakeSinkForUsecase()
	v1StartedAt := time.Date(2024, 6, 1, 10, 0, 0, 0, time.UTC)
	savedSnap := &port.Snapshot{
		VideoID:    "v1",
		LiveChatID: "chat-v1",
		Users: []domain.User{
			{ChannelID: "ch1", DisplayName: "Alice"},
			{ChannelID: "ch2", DisplayName: "Bob"},
		},
		Comments: []domain.Comment{
			{ID: "c1", ChannelID: "ch1", DisplayName: "Alice", Message: "hello", PublishedAt: now},
		},
		State: &domain.LiveState{
			Status:     domain.StatusActive,
			VideoID:    "v1",
			LiveChatID: "chat-v1",
			StartedAt:  v1StartedAt,
		},
	}
	_ = sink.Save(ctx, savedSnap)

	coord := buildCoordWithSink(sink, users, comments, state)
	uc := &usecase.SwitchVideo{YT: yt, Users: users, Comments: comments, State: state, Clock: clock, Snap: coord}

	out, err := uc.Execute(ctx, usecase.SwitchVideoInput{VideoID: "v1"})
	if err != nil {
		t.Fatalf("Execute should succeed (GCS restore), got: %v", err)
	}

	if out.State.Status != domain.StatusWaiting {
		t.Errorf("State.Status = %v, want %v", out.State.Status, domain.StatusWaiting)
	}
	if out.State.VideoID != "v1" {
		t.Errorf("State.VideoID = %v, want v1", out.State.VideoID)
	}
	if got := users.Count(); got != 2 {
		t.Errorf("Users.Count() = %d, want 2 (restored from GCS)", got)
	}
	if got := comments.Count(); got != 1 {
		t.Errorf("Comments.Count() = %d, want 1 (restored from GCS)", got)
	}
	// StartedAt は復元値を使う
	if !out.State.StartedAt.Equal(v1StartedAt) {
		t.Errorf("State.StartedAt = %v, want %v (from GCS snapshot)", out.State.StartedAt, v1StartedAt)
	}
	// EndedAt は now
	if !out.State.EndedAt.Equal(now) {
		t.Errorf("State.EndedAt = %v, want %v", out.State.EndedAt, now)
	}
}

// TestSwitchVideo_APIError_RestoreFromGCS_notFound: GCS にも snapshot なし → error return
func TestSwitchVideo_APIError_RestoreFromGCS_notFound(t *testing.T) {
	ctx := context.Background()

	users := memory.NewUserRepo()
	comments := memory.NewCommentRepo()
	state := memory.NewStateRepo()
	_ = state.Set(ctx, domain.LiveState{
		Status:  domain.StatusWaiting,
		VideoID: "v0",
	})

	yt := &failingYT{}
	clock := fixedClock{t: time.Now()}

	// GCS に v1 snapshot なし (空の sink)
	sink := newFakeSinkForUsecase()
	coord := buildCoordWithSink(sink, users, comments, state)
	uc := &usecase.SwitchVideo{YT: yt, Users: users, Comments: comments, State: state, Clock: clock, Snap: coord}

	_, err := uc.Execute(ctx, usecase.SwitchVideoInput{VideoID: "v1"})
	if err == nil {
		t.Fatal("Execute should return error when GCS also has no snapshot")
	}
}

// TestSwitchVideo_APIError_inMemoryFallback_priorityOverGCS: 1st fallback (同 videoId + Users 残存) が GCS より優先される
func TestSwitchVideo_APIError_inMemoryFallback_priorityOverGCS(t *testing.T) {
	ctx := context.Background()

	users := memory.NewUserRepo()
	_ = users.UpsertWithJoinTime("ch1", "Alice", time.Now())
	_ = users.UpsertWithJoinTime("ch2", "Bob", time.Now())
	comments := memory.NewCommentRepo()

	originalStartedAt := time.Date(2024, 1, 1, 9, 0, 0, 0, time.UTC)
	state := memory.NewStateRepo()
	_ = state.Set(ctx, domain.LiveState{
		Status:     domain.StatusWaiting,
		VideoID:    "v1",
		LiveChatID: "chat-v1",
		StartedAt:  originalStartedAt,
	})

	yt := &failingYT{}
	now := time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC)
	clock := fixedClock{t: now}

	// GCS にも v1 snapshot あり（ただし 1st fallback が優先されるべき）
	sink := newFakeSinkForUsecase()
	gcsStartedAt := time.Date(2024, 1, 1, 8, 0, 0, 0, time.UTC)
	_ = sink.Save(ctx, &port.Snapshot{
		VideoID:    "v1",
		LiveChatID: "chat-v1",
		Users: []domain.User{
			{ChannelID: "ch3", DisplayName: "GCS-Only-User"},
		},
		State: &domain.LiveState{
			Status:    domain.StatusActive,
			VideoID:   "v1",
			StartedAt: gcsStartedAt,
		},
	})
	coord := buildCoordWithSink(sink, users, comments, state)
	uc := &usecase.SwitchVideo{YT: yt, Users: users, Comments: comments, State: state, Clock: clock, Snap: coord}

	out, err := uc.Execute(ctx, usecase.SwitchVideoInput{VideoID: "v1"})
	if err != nil {
		t.Fatalf("Execute should succeed (in-memory fallback), got: %v", err)
	}

	// 1st fallback が当たるので in-memory の users (2件) が残存
	if got := users.Count(); got != 2 {
		t.Errorf("Users.Count() = %d, want 2 (in-memory fallback, not GCS)", got)
	}
	// StartedAt は in-memory 由来 (originalStartedAt)
	if !out.State.StartedAt.Equal(originalStartedAt) {
		t.Errorf("State.StartedAt = %v, want %v (in-memory, not GCS)", out.State.StartedAt, originalStartedAt)
	}
	if out.State.Status != domain.StatusWaiting {
		t.Errorf("State.Status = %v, want %v", out.State.Status, domain.StatusWaiting)
	}
}

// TestSwitchVideo_SnapshotPreservedAcrossSwitch: V1→V2 切替後も V1.json が破壊されない（回帰テスト）
func TestSwitchVideo_SnapshotPreservedAcrossSwitch(t *testing.T) {
	ctx := context.Background()

	users := memory.NewUserRepo()
	comments := memory.NewCommentRepo()
	state := memory.NewStateRepo()
	_ = state.Set(ctx, domain.LiveState{Status: domain.StatusActive, VideoID: "v1", LiveChatID: "chat-v1"})

	yt := &fakeYT{}
	clock := fixedClock{t: time.Now()}
	sink := newFakeSinkForUsecase()
	coord := buildCoordWithSink(sink, users, comments, state)

	// V1 配信: ユーザー追加 → Flush
	_ = users.UpsertWithJoinTime("ch1", "Alice", time.Now())
	coord.SetVideo("v1", "chat-v1")
	coord.MarkDirty()
	if err := coord.Flush(ctx); err != nil {
		t.Fatalf("V1 Flush failed: %v", err)
	}

	// V2 に切替 → V1.json が上書きされないことを確認
	uc := &usecase.SwitchVideo{YT: yt, Users: users, Comments: comments, State: state, Clock: clock, Snap: coord}
	_, err := uc.Execute(ctx, usecase.SwitchVideoInput{VideoID: "v2"})
	if err != nil {
		t.Fatalf("Switch to V2 failed: %v", err)
	}

	// V1 snapshot は sink に残っているか
	v1Snap, err := sink.Load(ctx, "v1")
	if err != nil {
		t.Fatalf("sink.Load(v1) error: %v", err)
	}
	if v1Snap == nil {
		t.Fatal("V1 snapshot was lost after switching to V2 (regression)")
	}
	if len(v1Snap.Users) != 1 {
		t.Errorf("V1 snapshot users = %d, want 1", len(v1Snap.Users))
	}
}
