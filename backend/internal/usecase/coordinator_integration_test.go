package usecase_test

import (
	"context"
	"testing"
	"time"

	"github.com/obsidian-engine/youtube-comment-user-list/backend/internal/adapter/memory"
	"github.com/obsidian-engine/youtube-comment-user-list/backend/internal/domain"
	"github.com/obsidian-engine/youtube-comment-user-list/backend/internal/port"
	"github.com/obsidian-engine/youtube-comment-user-list/backend/internal/usecase"
	"github.com/obsidian-engine/youtube-comment-user-list/backend/internal/usecase/snapshot"
)

// inMemorySink は port.SnapshotSink の in-memory 実装 (integration test 用)。
type inMemorySink struct {
	snapshots map[string]*port.Snapshot
	current   *port.CurrentPointer
}

func newInMemorySink() *inMemorySink {
	return &inMemorySink{snapshots: make(map[string]*port.Snapshot)}
}

func (s *inMemorySink) Load(_ context.Context, videoID string) (*port.Snapshot, error) {
	snap, ok := s.snapshots[videoID]
	if !ok {
		return nil, nil
	}
	cp := *snap
	return &cp, nil
}
func (s *inMemorySink) Save(_ context.Context, snap *port.Snapshot) error {
	cp := *snap
	s.snapshots[snap.VideoID] = &cp
	return nil
}
func (s *inMemorySink) LoadCurrent(_ context.Context) (*port.CurrentPointer, error) {
	if s.current == nil {
		return nil, nil
	}
	cp := *s.current
	return &cp, nil
}
func (s *inMemorySink) SaveCurrent(_ context.Context, ptr *port.CurrentPointer) error {
	cp := *ptr
	s.current = &cp
	return nil
}
func (s *inMemorySink) List(_ context.Context) ([]port.SnapshotSummary, error) {
	summaries := make([]port.SnapshotSummary, 0, len(s.snapshots))
	for _, snap := range s.snapshots {
		summaries = append(summaries, port.SnapshotSummary{
			VideoID: snap.VideoID, SavedAt: snap.SavedAt,
			UserCount: len(snap.Users), CommentCount: len(snap.Comments),
		})
	}
	return summaries, nil
}

// TestReserve_thenRestore_prevUsersDoNotResurrect: e2e シナリオ。
// 旧配信の users/comments が in-memory に残った状態から別 videoId で Reserve を実行し、
// Flush → 新しい Coordinator で Restore しても旧 users が復活しないことを検証する。
// 本 test が緑なら、リロード / 再デプロイ後の「前配信データ復活」バグを再発させない。
func TestReserve_thenRestore_prevUsersDoNotResurrect(t *testing.T) {
	ctx := context.Background()
	sink := newInMemorySink()

	// --- Step 1: 旧配信 "prev-vid" が終了し、users/comments が in-memory に残っている状態を作る ---
	users := memory.NewUserRepo()
	comments := memory.NewCommentRepo()
	state := memory.NewStateRepo()
	_ = users.UpsertWithJoinTime("ch-prev", "prev user", time.Unix(1000, 0))
	_ = comments.Add(domain.Comment{ID: "c-prev", ChannelID: "ch-prev", DisplayName: "prev user", Message: "old"})
	_ = state.Set(ctx, domain.LiveState{Status: domain.StatusWaiting, VideoID: "prev-vid"})

	coord := snapshot.NewCoordinator(sink, users, comments, state, 30*time.Second)
	coord.SetVideo("prev-vid", "prev-chat", "", "")

	// --- Step 2: 別 videoId "new-vid" で Reserve を実行 ---
	yt := &fakeYTForReserve{details: port.VideoLiveDetails{
		IsLiveContent:      true,
		ScheduledStartTime: time.Unix(2000, 0),
	}}
	uc := &usecase.Reserve{
		YT: yt, Users: users, Comments: comments,
		State: state, Clock: fixedClock{t: time.Unix(1500, 0)}, Snap: coord,
	}
	if _, err := uc.Execute(ctx, usecase.ReserveInput{VideoID: "new-vid"}); err != nil {
		t.Fatalf("Reserve failed: %v", err)
	}

	// Reserve 内部で Flush 済み。sink に new-vid の snapshot が users=0 で残っているはず
	saved, _ := sink.Load(ctx, "new-vid")
	if saved == nil {
		t.Fatal("expected snapshot at new-vid, got nil")
	}
	if len(saved.Users) != 0 {
		t.Errorf("snapshot.Users = %d, want 0 (prev users must not leak into new-vid snapshot)", len(saved.Users))
	}
	if len(saved.Comments) != 0 {
		t.Errorf("snapshot.Comments = %d, want 0 (prev comments must not leak)", len(saved.Comments))
	}

	// --- Step 3: 別プロセス再起動を模擬。新規 in-memory repo + 新 Coordinator で Restore ---
	users2 := memory.NewUserRepo()
	comments2 := memory.NewCommentRepo()
	state2 := memory.NewStateRepo()
	coord2 := snapshot.NewCoordinator(sink, users2, comments2, state2, 30*time.Second)
	if err := coord2.Restore(ctx); err != nil {
		t.Fatalf("Restore failed: %v", err)
	}

	if users2.Count() != 0 {
		t.Errorf("after Restore users.Count = %d, want 0 (prev users resurrected via GCS)", users2.Count())
	}
	if comments2.Count() != 0 {
		t.Errorf("after Restore comments.Count = %d, want 0", comments2.Count())
	}
	restoredState, _ := state2.Get(ctx)
	if restoredState.VideoID != "new-vid" {
		t.Errorf("restored state.VideoID = %q, want new-vid", restoredState.VideoID)
	}
	if restoredState.Status != domain.StatusReserved {
		t.Errorf("restored state.Status = %q, want RESERVED", restoredState.Status)
	}
}

// mockCoord は Coordinator の呼出し回数・引数を記録するテスト用 mock。
type mockCoord struct {
	markDirty int
	flush     int
	setVideo  [][2]string
	calls     []string // 呼出し順を記録
}

func (m *mockCoord) Restore(_ context.Context) error                      { return nil }
func (m *mockCoord) RestoreFor(_ context.Context, _ string) (bool, error) { return false, nil }
func (m *mockCoord) SetVideo(videoID, liveChatID, _, _ string) {
	m.setVideo = append(m.setVideo, [2]string{videoID, liveChatID})
	m.calls = append(m.calls, "SetVideo")
}
func (m *mockCoord) MarkDirty() {
	m.markDirty++
	m.calls = append(m.calls, "MarkDirty")
}
func (m *mockCoord) Flush(_ context.Context) error {
	m.flush++
	m.calls = append(m.calls, "Flush")
	return nil
}
func (m *mockCoord) Start(_ context.Context) {}
func (m *mockCoord) Stop()                   {}
func (m *mockCoord) LastSavedAt() time.Time  { return time.Time{} }

// TestSwitchVideo_CoordinatorCallOrder: 旧 video Flush → SetVideo → MarkDirty → Flush の順序を検証
func TestSwitchVideo_CoordinatorCallOrder(t *testing.T) {
	t.Helper()
	ctx := context.Background()

	users := memory.NewUserRepo()
	state := memory.NewStateRepo()
	yt := &fakeYT{}
	clock := fixedClock{t: time.Unix(2000, 0)}
	coord := &mockCoord{}

	uc := &usecase.SwitchVideo{
		YT:    yt,
		Users: users,
		State: state,
		Clock: clock,
		Snap:  coord,
	}

	_, err := uc.Execute(ctx, usecase.SwitchVideoInput{VideoID: "new-video"})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// 期待呼出し順: Flush(旧) → SetVideo → MarkDirty → Flush(新)
	want := []string{"Flush", "SetVideo", "MarkDirty", "Flush"}
	if len(coord.calls) != len(want) {
		t.Fatalf("calls = %v, want %v", coord.calls, want)
	}
	for i, call := range coord.calls {
		if call != want[i] {
			t.Errorf("calls[%d] = %q, want %q", i, call, want[i])
		}
	}

	// SetVideo の引数確認
	if len(coord.setVideo) != 1 {
		t.Fatalf("SetVideo called %d times, want 1", len(coord.setVideo))
	}
	if coord.setVideo[0][0] != "new-video" {
		t.Errorf("SetVideo videoID = %q, want %q", coord.setVideo[0][0], "new-video")
	}
	if coord.setVideo[0][1] != "live:abc" {
		t.Errorf("SetVideo liveChatID = %q, want %q", coord.setVideo[0][1], "live:abc")
	}

	// Flush は 2 回呼ばれる
	if coord.flush != 2 {
		t.Errorf("flush count = %d, want 2", coord.flush)
	}
}

// TestCoordinator_MultiVideoSnapshotIsolation: V1 save → SetVideo(V2) → V2 save → Load(V1) で V1 snapshot が破壊されない
func TestCoordinator_MultiVideoSnapshotIsolation(t *testing.T) {
	t.Helper()
	ctx := context.Background()

	users := memory.NewUserRepo()
	comments := memory.NewCommentRepo()
	state := memory.NewStateRepo()

	sink := newFakeSinkForUsecase()
	coord := snapshot.NewCoordinator(sink, users, comments, state, 0)

	// V1 配信: ユーザー追加 → save
	_ = users.UpsertWithJoinTime("ch1", "Alice", time.Now())
	_ = state.Set(ctx, domain.LiveState{Status: domain.StatusActive, VideoID: "v1", LiveChatID: "chat-v1"})
	coord.SetVideo("v1", "chat-v1", "", "")
	coord.MarkDirty()
	if err := coord.Flush(ctx); err != nil {
		t.Fatalf("V1 Flush failed: %v", err)
	}

	// V2 に切替: users クリア → 別ユーザー追加 → save
	users.Clear()
	comments.Clear()
	_ = users.UpsertWithJoinTime("ch2", "Bob", time.Now())
	_ = users.UpsertWithJoinTime("ch3", "Carol", time.Now())
	_ = state.Set(ctx, domain.LiveState{Status: domain.StatusActive, VideoID: "v2", LiveChatID: "chat-v2"})
	coord.SetVideo("v2", "chat-v2", "", "")
	coord.MarkDirty()
	if err := coord.Flush(ctx); err != nil {
		t.Fatalf("V2 Flush failed: %v", err)
	}

	// V1 snapshot が破壊されていないか確認
	v1Snap, err := sink.Load(ctx, "v1")
	if err != nil {
		t.Fatalf("sink.Load(v1) error: %v", err)
	}
	if v1Snap == nil {
		t.Fatal("V1 snapshot was destroyed after V2 save")
	}
	if len(v1Snap.Users) != 1 {
		t.Errorf("V1 snapshot users = %d, want 1", len(v1Snap.Users))
	}
	if v1Snap.Users[0].ChannelID != "ch1" {
		t.Errorf("V1 snapshot user = %q, want ch1", v1Snap.Users[0].ChannelID)
	}

	// V2 snapshot も正しく保存されているか
	v2Snap, err := sink.Load(ctx, "v2")
	if err != nil {
		t.Fatalf("sink.Load(v2) error: %v", err)
	}
	if v2Snap == nil {
		t.Fatal("V2 snapshot not found")
	}
	if len(v2Snap.Users) != 2 {
		t.Errorf("V2 snapshot users = %d, want 2", len(v2Snap.Users))
	}
}

// TestPull_MarkDirty_CalledOnlyWhenDiff: 差分あり時のみ MarkDirty が呼ばれることを検証
func TestPull_MarkDirty_CalledOnlyWhenDiff(t *testing.T) {
	t.Helper()

	tests := []struct {
		name              string
		items             []port.ChatMessage
		wantMarkDirtyCall int
	}{
		{
			name: "差分あり: MarkDirty 呼ばれる",
			items: []port.ChatMessage{
				{ID: "msg1", ChannelID: "ch1", DisplayName: "Alice", PublishedAt: time.Date(2023, 1, 1, 11, 0, 0, 0, time.UTC)},
			},
			wantMarkDirtyCall: 1,
		},
		{
			name:              "差分なし (items 空): MarkDirty 呼ばれない",
			items:             []port.ChatMessage{},
			wantMarkDirtyCall: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Helper()
			ctx := context.Background()
			users := memory.NewUserRepo()
			comments := memory.NewCommentRepo()
			state := memory.NewStateRepo()
			_ = state.Set(ctx, domain.LiveState{
				Status:     domain.StatusActive,
				VideoID:    "v",
				LiveChatID: "live:chat",
			})
			clock := &fakeClock{now: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)}
			coord := &mockCoord{}

			yt := &fakeYTForPull{items: tt.items, ended: false}
			uc := &usecase.Pull{
				YT:       yt,
				Users:    users,
				Comments: comments,
				State:    state,
				Clock:    clock,
				Snap:     coord,
			}

			_, err := uc.Execute(ctx)
			if err != nil {
				t.Fatalf("Execute failed: %v", err)
			}

			if coord.markDirty != tt.wantMarkDirtyCall {
				t.Errorf("markDirty calls = %d, want %d", coord.markDirty, tt.wantMarkDirtyCall)
			}
		})
	}
}
