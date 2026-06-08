package snapshot

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/obsidian-engine/youtube-comment-user-list/backend/internal/adapter/memory"
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
	userRepo    *memory.UserRepo
	commentRepo *memory.CommentRepo
	throttle    time.Duration

	mu         sync.Mutex
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
	ur *memory.UserRepo,
	cr *memory.CommentRepo,
	throttle time.Duration,
) Coordinator {
	return &coordinator{
		sink:        sink,
		userRepo:    ur,
		commentRepo: cr,
		throttle:    throttle,
	}
}

// Restore は起動時に current pointer を読み、snapshot を in-memory repo に復元します。
// 読み込み失敗は warn ログのみで続行します（起動を止めない）。
func (c *coordinator) Restore(ctx context.Context) error {
	ptr, err := c.sink.LoadCurrent(ctx)
	if err != nil {
		log.Printf("[WARN] snapshot: LoadCurrent failed: %v", err)
		return nil
	}
	if ptr == nil {
		log.Printf("[INFO] snapshot: no current.json found, starting with empty state")
		return nil
	}

	snap, err := c.sink.Load(ctx, ptr.VideoID)
	if err != nil {
		log.Printf("[WARN] snapshot: Load(%s) failed: %v", ptr.VideoID, err)
		return nil
	}
	if snap == nil {
		log.Printf("[INFO] snapshot: no snapshot for videoId=%s, starting with empty state", ptr.VideoID)
		return nil
	}

	c.userRepo.LoadFrom(snap.Users, snap.ProcessedMsgs)
	c.commentRepo.LoadFrom(snap.Comments)

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
func (c *coordinator) Flush(ctx context.Context) error {
	c.mu.Lock()
	videoID := c.videoID
	liveChatID := c.liveChatID
	c.dirty = false
	c.lastSaved = time.Now()
	c.mu.Unlock()

	if videoID == "" {
		return nil
	}

	if err := c.save(ctx, videoID, liveChatID); err != nil {
		log.Printf("[WARN] snapshot: Flush save failed: %v", err)
		return fmt.Errorf("snapshot: flush: %w", err)
	}
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
				if shouldSave {
					c.dirty = false
					c.lastSaved = time.Now()
				}
				c.mu.Unlock()

				if shouldSave && videoID != "" {
					if err := c.save(innerCtx, videoID, liveChatID); err != nil {
						log.Printf("[WARN] snapshot: background save failed: %v", err)
						// 次 tick で再試行するため dirty を再セット
						c.mu.Lock()
						c.dirty = true
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
func (c *coordinator) save(ctx context.Context, videoID, liveChatID string) error {
	users, processedMsgs := c.userRepo.Dump()
	comments := c.commentRepo.Dump()

	snap := &port.Snapshot{
		SchemaVersion: 1,
		VideoID:       videoID,
		LiveChatID:    liveChatID,
		Users:         users,
		Comments:      comments,
		ProcessedMsgs: processedMsgs,
	}

	if err := c.sink.Save(ctx, snap); err != nil {
		return fmt.Errorf("save snapshot: %w", err)
	}

	ptr := &port.CurrentPointer{VideoID: videoID}
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
