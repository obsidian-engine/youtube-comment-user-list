package snapshot

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/obsidian-engine/youtube-comment-user-list/backend/internal/domain"
	"github.com/obsidian-engine/youtube-comment-user-list/backend/internal/port"
)

// Coordinator はスナップショット永続化を調整します。
// MarkDirty で dirty フラグを立て、throttle 経過後に自動 save します。
type Coordinator interface {
	Restore(ctx context.Context) error
	SetVideo(videoID, liveChatID string)
	MarkDirty()
	Flush(ctx context.Context) error
	Start(ctx context.Context)
	Stop()
}

// coordinator は GCS sink を持つ Coordinator 実装です。
type coordinator struct {
	sink        port.SnapshotSink
	userRepo    port.UserSnapshotSource
	commentRepo port.CommentSnapshotSource
	stateRepo   port.StateRepo
	throttle    time.Duration

	mu         sync.Mutex
	saveMu     sync.Mutex // save を直列化する
	videoID    string
	liveChatID string
	dirty      bool
	lastSaved  time.Time

	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// NewCoordinator は coordinator を生成します。
func NewCoordinator(
	sink port.SnapshotSink,
	ur port.UserSnapshotSource,
	cr port.CommentSnapshotSource,
	sr port.StateRepo,
	throttle time.Duration,
) Coordinator {
	return &coordinator{
		sink:        sink,
		userRepo:    ur,
		commentRepo: cr,
		stateRepo:   sr,
		throttle:    throttle,
	}
}

// Restore は起動時に current pointer を読み、snapshot を in-memory repo に復元します。
// GCS auth / unmarshal 失敗など致命的エラーは err を返します。
// snapshot 不在 / 空 videoID は正常ケースとして nil を返します。
func (c *coordinator) Restore(ctx context.Context) error {
	ptr, err := c.sink.LoadCurrent(ctx)
	if err != nil {
		return fmt.Errorf("snapshot: load current: %w", err)
	}
	if ptr == nil || ptr.VideoID == "" {
		return nil
	}

	snap, err := c.sink.Load(ctx, ptr.VideoID)
	if err != nil {
		return fmt.Errorf("snapshot: load snapshot %s: %w", ptr.VideoID, err)
	}
	if snap == nil {
		log.Printf("[WARN] snapshot: current.json points to %s but snapshot not found", ptr.VideoID)
		return nil
	}

	c.userRepo.LoadFrom(port.UserSnapshot{Users: snap.Users, ProcessedMsgs: snap.ProcessedMsgs})
	c.commentRepo.LoadFrom(snap.Comments)

	if snap.State != nil && c.stateRepo != nil {
		if err := c.stateRepo.Set(ctx, *snap.State); err != nil {
			return fmt.Errorf("snapshot: restore state: %w", err)
		}
	}

	c.mu.Lock()
	c.videoID = snap.VideoID
	c.liveChatID = snap.LiveChatID
	c.mu.Unlock()

	log.Printf("[INFO] snapshot: restored videoId=%s users=%d comments=%d",
		snap.VideoID, len(snap.Users), len(snap.Comments))
	return nil
}

// SetVideo は video 切替時に呼びます。
func (c *coordinator) SetVideo(videoID, liveChatID string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.videoID = videoID
	c.liveChatID = liveChatID
	c.dirty = false
}

// MarkDirty は pull 差分発生時に呼びます。
func (c *coordinator) MarkDirty() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.dirty = true
}

// Flush は throttle を無視して即時 save します（SIGTERM / video 切替時）。
// videoID が空の場合は Reset 経路として current.json を空 videoID で上書きします。
func (c *coordinator) Flush(ctx context.Context) error {
	c.mu.Lock()
	videoID := c.videoID
	liveChatID := c.liveChatID
	c.mu.Unlock()

	if videoID == "" {
		// Reset 経路: current.json を空 videoID で上書きして旧 videoID を消す
		// saveMu で save() 経路と直列化し、background save との並走で SaveCurrent 順序が逆転するのを防ぐ
		c.saveMu.Lock()
		err := c.sink.SaveCurrent(ctx, &port.CurrentPointer{VideoID: "", SavedAt: time.Now()})
		c.saveMu.Unlock()
		if err != nil {
			return fmt.Errorf("snapshot: flush save current (empty): %w", err)
		}
		c.mu.Lock()
		c.dirty = false
		c.lastSaved = time.Now()
		c.mu.Unlock()
		return nil
	}

	if err := c.save(ctx, videoID, liveChatID); err != nil {
		return fmt.Errorf("snapshot: flush: %w", err)
	}
	// save 成功後にのみ dirty をクリア
	c.mu.Lock()
	c.dirty = false
	c.lastSaved = time.Now()
	c.mu.Unlock()
	return nil
}

// Start は throttle 監視 goroutine を起動します。
func (c *coordinator) Start(ctx context.Context) {
	innerCtx, cancel := context.WithCancel(ctx)
	c.cancel = cancel

	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-innerCtx.Done():
				return
			case <-ticker.C:
				c.mu.Lock()
				shouldSave := c.dirty && time.Since(c.lastSaved) > c.throttle
				videoID := c.videoID
				liveChatID := c.liveChatID
				c.mu.Unlock()

				if shouldSave && videoID != "" {
					if err := c.save(innerCtx, videoID, liveChatID); err != nil {
						log.Printf("[WARN] snapshot: background save failed: %v", err)
						// save 失敗時は dirty を維持して次 tick で再試行
					} else {
						// save 成功後にのみ dirty をクリア (Flush と同じ pattern)
						c.mu.Lock()
						c.dirty = false
						c.lastSaved = time.Now()
						c.mu.Unlock()
					}
				}
			}
		}
	}()
}

// Stop は background goroutine を停止します。
func (c *coordinator) Stop() {
	if c.cancel != nil {
		c.cancel()
	}
	c.wg.Wait()
}

// save は snapshot を組み立てて sink に書き込みます。
// saveMu で直列化し、並列 save による上書き race を防ぎます。
func (c *coordinator) save(ctx context.Context, videoID, liveChatID string) error {
	c.saveMu.Lock()
	defer c.saveMu.Unlock()

	userSnap := c.userRepo.Dump()
	comments := c.commentRepo.Dump()

	var liveState *domain.LiveState
	if c.stateRepo != nil {
		st, err := c.stateRepo.Get(ctx)
		if err != nil {
			log.Printf("[WARN] snapshot: state.Get failed, saving without state: %v", err)
		} else {
			liveState = &st
		}
	}

	snap := &port.Snapshot{
		SchemaVersion: 1,
		VideoID:       videoID,
		LiveChatID:    liveChatID,
		SavedAt:       time.Now(),
		Users:         userSnap.Users,
		Comments:      comments,
		ProcessedMsgs: userSnap.ProcessedMsgs,
		State:         liveState,
	}

	if err := c.sink.Save(ctx, snap); err != nil {
		return fmt.Errorf("save snapshot: %w", err)
	}

	ptr := &port.CurrentPointer{VideoID: videoID, SavedAt: time.Now()}
	if err := c.sink.SaveCurrent(ctx, ptr); err != nil {
		return fmt.Errorf("save current pointer: %w", err)
	}

	return nil
}

// NopCoordinator は GCS_BUCKET が空の場合に使う no-op 実装です。
type NopCoordinator struct{}

func (n *NopCoordinator) Restore(_ context.Context) error { return nil }
func (n *NopCoordinator) SetVideo(_, _ string)            {}
func (n *NopCoordinator) MarkDirty()                      {}
func (n *NopCoordinator) Flush(_ context.Context) error   { return nil }
func (n *NopCoordinator) Start(_ context.Context)         {}
func (n *NopCoordinator) Stop()                           {}
