package snapshot_test

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/obsidian-engine/youtube-comment-user-list/backend/internal/adapter/memory"
	"github.com/obsidian-engine/youtube-comment-user-list/backend/internal/domain"
	"github.com/obsidian-engine/youtube-comment-user-list/backend/internal/port"
	"github.com/obsidian-engine/youtube-comment-user-list/backend/internal/usecase/snapshot"
)

func newTestRepos() (*memory.UserRepo, *memory.CommentRepo) {
	return memory.NewUserRepo(), memory.NewCommentRepo()
}

// TestFlush_ignoresThrottle: Flush は throttle 経過前でも即時 save する
func TestFlush_ignoresThrottle(t *testing.T) {
	t.Helper()
	sink := newFakeSink()
	ur, cr := newTestRepos()

	c := snapshot.NewCoordinator(sink, ur, cr, nil, 30*time.Second)
	c.SetVideo("vid1", "chat1")
	c.MarkDirty()

	if err := c.Flush(context.Background()); err != nil {
		t.Fatalf("Flush returned error: %v", err)
	}

	if sink.getSaveCount() != 1 {
		t.Errorf("expected 1 save after Flush, got %d", sink.getSaveCount())
	}
}

// TestFlush_noVideo: videoID が空の場合は save しない
func TestFlush_noVideo(t *testing.T) {
	t.Helper()
	sink := newFakeSink()
	ur, cr := newTestRepos()

	c := snapshot.NewCoordinator(sink, ur, cr, nil, 30*time.Second)
	// SetVideo を呼ばない

	if err := c.Flush(context.Background()); err != nil {
		t.Fatalf("Flush returned error: %v", err)
	}

	if sink.getSaveCount() != 0 {
		t.Errorf("expected 0 saves when videoID is empty, got %d", sink.getSaveCount())
	}
}

// TestMarkDirty_throttleCollapse: throttle 内の連続 MarkDirty は Start 後 1 回しか save しない
func TestMarkDirty_throttleCollapse(t *testing.T) {
	t.Helper()
	sink := newFakeSink()
	ur, cr := newTestRepos()

	// throttle 50ms で短く設定、ticker は 1s だが Start 後に直接 flush で確認
	c := snapshot.NewCoordinator(sink, ur, cr, nil, 50*time.Millisecond)
	c.SetVideo("vid1", "chat1")

	// 連続 MarkDirty → Flush で 1 save
	c.MarkDirty()
	c.MarkDirty()
	c.MarkDirty()

	if err := c.Flush(context.Background()); err != nil {
		t.Fatalf("Flush returned error: %v", err)
	}

	if sink.getSaveCount() != 1 {
		t.Errorf("expected exactly 1 save after multiple MarkDirty + Flush, got %d", sink.getSaveCount())
	}
}

// TestRestore_loadsRepoState: Restore が sink の snapshot を repo に反映する
func TestRestore_loadsRepoState(t *testing.T) {
	t.Helper()
	sink := newFakeSink()
	ur, cr := newTestRepos()

	// 事前に sink にスナップショットを仕込む
	now := time.Now()
	snap := &port.Snapshot{
		SchemaVersion: 1,
		VideoID:       "vid-restore",
		LiveChatID:    "chat-restore",
		Users: []domain.User{
			{
				ChannelID:         "ch1",
				DisplayName:       "Alice",
				JoinedAt:          now,
				CommentCount:      2,
				FirstCommentedAt:  now,
				LatestCommentedAt: now,
			},
		},
		Comments: []domain.Comment{
			{ID: "c1", ChannelID: "ch1", DisplayName: "Alice", Message: "hello", PublishedAt: now},
		},
		ProcessedMsgs: []string{"msg1", "msg2"},
	}
	_ = sink.Save(context.Background(), snap)
	_ = sink.SaveCurrent(context.Background(), &port.CurrentPointer{VideoID: "vid-restore"})

	// Restore
	c := snapshot.NewCoordinator(sink, ur, cr, nil, 30*time.Second)
	if err := c.Restore(context.Background()); err != nil {
		t.Fatalf("Restore returned error: %v", err)
	}

	// repo に反映されているか確認
	if ur.Count() != 1 {
		t.Errorf("expected 1 user after Restore, got %d", ur.Count())
	}
	if cr.Count() != 1 {
		t.Errorf("expected 1 comment after Restore, got %d", cr.Count())
	}

	users := ur.ListUsersSortedByJoinTime()
	if len(users) != 1 || users[0].ChannelID != "ch1" {
		t.Errorf("unexpected users: %+v", users)
	}
}

// TestRestore_restoresLiveState: snap.State が nil でない場合に StateRepo へ復元される
func TestRestore_restoresLiveState(t *testing.T) {
	t.Helper()
	sink := newFakeSink()
	ur, cr := newTestRepos()
	sr := memory.NewStateRepo()

	now := time.Now()
	state := domain.LiveState{
		Status:        domain.StatusActive,
		VideoID:       "vid-state",
		LiveChatID:    "chat-state",
		StartedAt:     now,
		NextPageToken: "token-xyz",
	}
	snap := &port.Snapshot{
		SchemaVersion: 1,
		VideoID:       "vid-state",
		LiveChatID:    "chat-state",
		State:         &state,
	}
	_ = sink.Save(context.Background(), snap)
	_ = sink.SaveCurrent(context.Background(), &port.CurrentPointer{VideoID: "vid-state"})

	c := snapshot.NewCoordinator(sink, ur, cr, sr, 30*time.Second)
	if err := c.Restore(context.Background()); err != nil {
		t.Fatalf("Restore returned error: %v", err)
	}

	got, err := sr.Get(context.Background())
	if err != nil {
		t.Fatalf("StateRepo.Get error: %v", err)
	}
	if got.VideoID != state.VideoID {
		t.Errorf("VideoID = %q, want %q", got.VideoID, state.VideoID)
	}
	if got.LiveChatID != state.LiveChatID {
		t.Errorf("LiveChatID = %q, want %q", got.LiveChatID, state.LiveChatID)
	}
	if got.Status != state.Status {
		t.Errorf("Status = %q, want %q", got.Status, state.Status)
	}
	if got.NextPageToken != state.NextPageToken {
		t.Errorf("NextPageToken = %q, want %q", got.NextPageToken, state.NextPageToken)
	}
}

// TestRestore_nilState_skipStateSet: snap.State が nil の場合 StateRepo に書かない
func TestRestore_nilState_skipStateSet(t *testing.T) {
	t.Helper()
	sink := newFakeSink()
	ur, cr := newTestRepos()
	sr := memory.NewStateRepo()

	snap := &port.Snapshot{
		SchemaVersion: 1,
		VideoID:       "vid-nostate",
		LiveChatID:    "chat-nostate",
		State:         nil, // 旧 snapshot 互換: State なし
	}
	_ = sink.Save(context.Background(), snap)
	_ = sink.SaveCurrent(context.Background(), &port.CurrentPointer{VideoID: "vid-nostate"})

	c := snapshot.NewCoordinator(sink, ur, cr, sr, 30*time.Second)
	if err := c.Restore(context.Background()); err != nil {
		t.Fatalf("Restore returned error: %v", err)
	}

	// StateRepo は初期値のまま (VideoID が空) であること
	got, err := sr.Get(context.Background())
	if err != nil {
		t.Fatalf("StateRepo.Get error: %v", err)
	}
	if got.VideoID != "" {
		t.Errorf("expected empty VideoID when snap.State is nil, got %q", got.VideoID)
	}
}

// TestSave_includesState: Flush 時に snap に State が含まれる
func TestSave_includesState(t *testing.T) {
	t.Helper()
	sink := newFakeSink()
	ur, cr := newTestRepos()
	sr := memory.NewStateRepo()

	liveState := domain.LiveState{
		Status:     domain.StatusActive,
		VideoID:    "vid-save",
		LiveChatID: "chat-save",
	}
	_ = sr.Set(context.Background(), liveState)

	c := snapshot.NewCoordinator(sink, ur, cr, sr, 30*time.Second)
	c.SetVideo("vid-save", "chat-save")
	c.MarkDirty()

	if err := c.Flush(context.Background()); err != nil {
		t.Fatalf("Flush returned error: %v", err)
	}

	saved, err := sink.Load(context.Background(), "vid-save")
	if err != nil {
		t.Fatalf("sink.Load error: %v", err)
	}
	if saved == nil {
		t.Fatal("expected snapshot to be saved, got nil")
	}
	if saved.State == nil {
		t.Fatal("expected snap.State to be set, got nil")
	}
	if saved.State.VideoID != liveState.VideoID {
		t.Errorf("snap.State.VideoID = %q, want %q", saved.State.VideoID, liveState.VideoID)
	}
	if saved.State.Status != liveState.Status {
		t.Errorf("snap.State.Status = %q, want %q", saved.State.Status, liveState.Status)
	}
}

// TestRestore_noCurrent: current.json なし → 空 state で続行（エラーなし）
func TestRestore_noCurrent(t *testing.T) {
	t.Helper()
	sink := newFakeSink()
	ur, cr := newTestRepos()

	c := snapshot.NewCoordinator(sink, ur, cr, nil, 30*time.Second)
	if err := c.Restore(context.Background()); err != nil {
		t.Fatalf("Restore should not return error when current.json is absent, got: %v", err)
	}

	if ur.Count() != 0 {
		t.Errorf("expected empty repo, got %d users", ur.Count())
	}
}

// TestBackground_savesAfterThrottle: background goroutine が throttle 経過後に save する
func TestBackground_savesAfterThrottle(t *testing.T) {
	t.Helper()
	sink := newFakeSink()
	ur, cr := newTestRepos()

	// throttle を短くして実時間テストを可能にする
	// ticker が 1s なので throttle < 1s でも ticker が trigger するまで待つ必要がある
	// ここでは throttle=0 (即トリガー) にして ticker 1 tick 待つ
	c := snapshot.NewCoordinator(sink, ur, cr, nil, 0)
	c.SetVideo("vid-bg", "chat-bg")

	ctx := context.Background()
	c.Start(ctx)
	defer c.Stop()

	c.MarkDirty()

	// ticker 1s + マージン
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if sink.getSaveCount() >= 1 {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	if sink.getSaveCount() < 1 {
		t.Errorf("expected at least 1 background save after throttle, got 0")
	}
}

// TestRestore_currentExistsButSnapshotAbsent: current.json はあるが対象 videoID の snapshot が存在しない → 空 state で続行
func TestRestore_currentExistsButSnapshotAbsent(t *testing.T) {
	t.Helper()
	sink := newFakeSink()
	ur, cr := newTestRepos()

	// current.json は存在するが snapshot は保存しない
	_ = sink.SaveCurrent(context.Background(), &port.CurrentPointer{VideoID: "missing-video"})

	c := snapshot.NewCoordinator(sink, ur, cr, nil, 30*time.Second)
	if err := c.Restore(context.Background()); err != nil {
		t.Fatalf("Restore should not return error when snapshot is absent, got: %v", err)
	}

	if ur.Count() != 0 {
		t.Errorf("expected empty user repo, got %d users", ur.Count())
	}
	if cr.Count() != 0 {
		t.Errorf("expected empty comment repo, got %d comments", cr.Count())
	}
}

// TestFlush_saveError_dirtyPreserved: Flush で save が失敗した場合、dirty が維持される (C1)
func TestFlush_saveError_dirtyPreserved(t *testing.T) {
	t.Helper()
	sink := newFakeSink()
	ur, cr := newTestRepos()

	c := snapshot.NewCoordinator(sink, ur, cr, nil, 30*time.Second)
	c.SetVideo("vid1", "chat1")
	c.MarkDirty()

	sink.setForceError(fmt.Errorf("gcs unavailable"))
	err := c.Flush(context.Background())
	if err == nil {
		t.Fatal("expected error from Flush when save fails, got nil")
	}

	// save 失敗後に dirty が維持されているか: 再度 Flush が save を呼ぶことで確認
	sink.setForceError(nil)
	if err2 := c.Flush(context.Background()); err2 != nil {
		t.Fatalf("second Flush (after error cleared) failed: %v", err2)
	}
	if sink.getSaveCount() != 1 {
		t.Errorf("expected 1 save on second Flush, got %d", sink.getSaveCount())
	}
}

// TestFlush_resetClearsCurrent: Reset 後 (videoID="") に Flush すると current.json が空 videoID で上書きされる (C2)
func TestFlush_resetClearsCurrent(t *testing.T) {
	t.Helper()
	sink := newFakeSink()
	ur, cr := newTestRepos()

	c := snapshot.NewCoordinator(sink, ur, cr, nil, 30*time.Second)
	c.SetVideo("vid1", "chat1")
	c.MarkDirty()
	_ = c.Flush(context.Background())

	// current.json に vid1 がセットされている
	ptr, _ := sink.LoadCurrent(context.Background())
	if ptr == nil || ptr.VideoID != "vid1" {
		t.Fatalf("expected current.json videoId=vid1, got %+v", ptr)
	}

	// Reset: videoID を空にして Flush
	c.SetVideo("", "")
	if err := c.Flush(context.Background()); err != nil {
		t.Fatalf("Flush with empty videoID returned error: %v", err)
	}

	ptr2, _ := sink.LoadCurrent(context.Background())
	if ptr2 == nil {
		t.Fatal("expected current.json to exist after reset flush, got nil")
	}
	if ptr2.VideoID != "" {
		t.Errorf("expected current.json videoId empty after reset, got %q", ptr2.VideoID)
	}
}

// TestFlush_parallelSave_noRace: 並列 Flush が race なく完了する (W3)
func TestFlush_parallelSave_noRace(t *testing.T) {
	t.Helper()
	sink := newFakeSink()
	ur, cr := newTestRepos()

	c := snapshot.NewCoordinator(sink, ur, cr, nil, 0)
	c.SetVideo("vid1", "chat1")

	const n = 10
	var wg sync.WaitGroup
	errs := make([]error, n)
	for i := range n {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			c.MarkDirty()
			errs[idx] = c.Flush(context.Background())
		}(i)
	}
	wg.Wait()

	for i, err := range errs {
		if err != nil {
			t.Errorf("goroutine %d: Flush returned error: %v", i, err)
		}
	}
}

// TestNopCoordinator: NopCoordinator は全 method が no-op でエラーなし
func TestNopCoordinator(t *testing.T) {
	t.Helper()
	nop := &snapshot.NopCoordinator{}

	if err := nop.Restore(context.Background()); err != nil {
		t.Errorf("NopCoordinator.Restore error: %v", err)
	}
	nop.SetVideo("v", "c")
	nop.MarkDirty()
	if err := nop.Flush(context.Background()); err != nil {
		t.Errorf("NopCoordinator.Flush error: %v", err)
	}
	nop.Start(context.Background())
	nop.Stop()
}
