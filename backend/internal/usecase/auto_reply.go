package usecase

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/obsidian-engine/youtube-comment-user-list/backend/internal/adapter/logging"
	"github.com/obsidian-engine/youtube-comment-user-list/backend/internal/port"
	"github.com/obsidian-engine/youtube-comment-user-list/backend/internal/usecase/snapshot"
)

// ReplyRequest は自動返信を投稿する 1 件の依頼。
type ReplyRequest struct {
	VideoID     string
	ChannelID   string
	DisplayName string
}

// AutoReply はフラグを設定した視聴者の初めてのコメントを判定し、返信の記録に「投稿中」を保存する。
// どの失敗のときも投稿しない (二重投稿より、投稿しないまま終わる方を許すため)。
type AutoReply struct {
	Enabled bool
	Targets port.ReplyTargetStore
	Records port.ReplyRecordStore
	Snap    snapshot.Coordinator
	Clock   port.Clock
	// Enqueue は投稿の依頼を受け取る。nil のときは投稿せず、投稿する予定をログに出力するだけにする。
	Enqueue func(ctx context.Context, req ReplyRequest)

	mu sync.Mutex
	// recovered は、この backend の起動後に「投稿中」を「失敗」に変える処理を済ませた動画。
	recovered map[string]bool
}

// Consider は一覧に新しく登録した視聴者のコメントについて、投稿の対象かを判定する。
// 対象なら返信の記録に「投稿中」を保存してから Enqueue に渡す。
func (a *AutoReply) Consider(ctx context.Context, videoID string, msg port.ChatMessage) {
	if a == nil || !a.Enabled || videoID == "" {
		return
	}
	if msg.PublishedAtUnknown {
		return
	}
	if !a.Snap.RestoreHealthy() {
		logging.Log(ctx, "warn", "AUTO_REPLY", "skip: snapshot restore failed for video %s", videoID)
		return
	}

	doc, _, err := a.Targets.Load(ctx)
	if err != nil {
		logging.Log(ctx, "warn", "AUTO_REPLY", "skip: load reply targets: %v", err)
		return
	}
	target, ok := findTarget(doc.Targets, msg.ChannelID)
	// フラグを設定する前のコメントは「既にコメントした人」なので対象にしない
	if !ok || !msg.PublishedAt.After(target.SetAt) {
		return
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	marked, err := a.markPosting(ctx, videoID, msg.ChannelID)
	if errors.Is(err, port.ErrConflict) {
		marked, err = a.markPosting(ctx, videoID, msg.ChannelID)
	}
	if err != nil {
		logging.Log(ctx, "warn", "AUTO_REPLY", "skip: save posting record for %s: %v", msg.ChannelID, err)
		return
	}
	if !marked {
		return
	}

	req := ReplyRequest{VideoID: videoID, ChannelID: msg.ChannelID, DisplayName: msg.DisplayName}
	if a.Enqueue == nil {
		logging.Log(ctx, "info", "AUTO_REPLY", "planned reply to %s (%s) on video %s", msg.DisplayName, msg.ChannelID, videoID)
		return
	}
	a.Enqueue(ctx, req)
}

// markPosting は視聴者の記録が無ければ「投稿中」を保存して true を返す。既に記録があれば false を返す。
func (a *AutoReply) markPosting(ctx context.Context, videoID, channelID string) (bool, error) {
	doc, version, err := a.Records.Load(ctx, videoID)
	if err != nil {
		return false, fmt.Errorf("reply_records_load: %w", err)
	}
	if doc.Records == nil {
		doc.Records = map[string]port.ReplyRecord{}
	}

	// 起動後に初めて読んだ動画では、前の backend が投稿中のまま止まった記録を失敗にする
	if a.recovered == nil {
		a.recovered = map[string]bool{}
	}
	firstLoad := !a.recovered[videoID]
	if firstLoad {
		for ch, r := range doc.Records {
			if r.State == port.ReplyPosting {
				r.State = port.ReplyFailed
				doc.Records[ch] = r
			}
		}
	}

	if _, exists := doc.Records[channelID]; exists {
		if firstLoad && version != 0 {
			if err := a.Records.Save(ctx, videoID, doc, version); err != nil {
				return false, err
			}
			a.recovered[videoID] = true
		}
		return false, nil
	}

	doc.Records[channelID] = port.ReplyRecord{State: port.ReplyPosting, At: a.Clock.Now()}
	if err := a.Records.Save(ctx, videoID, doc, version); err != nil {
		if errors.Is(err, port.ErrConflict) {
			return false, err
		}
		return false, fmt.Errorf("reply_records_save: %w", err)
	}
	a.recovered[videoID] = true
	return true, nil
}

func findTarget(targets []port.ReplyTarget, channelID string) (port.ReplyTarget, bool) {
	for _, t := range targets {
		if t.ChannelID == channelID {
			return t, true
		}
	}
	return port.ReplyTarget{}, false
}
