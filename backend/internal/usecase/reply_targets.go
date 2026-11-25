package usecase

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/obsidian-engine/youtube-comment-user-list/backend/internal/port"
)

// ReplyTargets は自動返信のフラグの取得と設定・解除を扱う。
// 同じ backend の中の書き込みは mu で 1 件ずつ順に行い、
// 別の backend と重なったときは Store の ErrConflict を受けて、最新の記録を読み込んで 1 回だけやり直す。
type ReplyTargets struct {
	Store port.ReplyTargetStore
	Clock port.Clock

	mu sync.Mutex
}

// List はフラグを設定した視聴者の一覧を返す。
func (uc *ReplyTargets) List(ctx context.Context) ([]port.ReplyTarget, error) {
	doc, _, err := uc.Store.Load(ctx)
	if err != nil {
		return nil, fmt.Errorf("reply_targets_load: %w", err)
	}
	return doc.Targets, nil
}

// Set は channelID のフラグを設定 (enabled=true) または解除する。
// 再び設定したときは設定時刻を今の時刻に更新する (初めてのコメントの判定で使うため)。
func (uc *ReplyTargets) Set(ctx context.Context, channelID string, enabled bool) error {
	uc.mu.Lock()
	defer uc.mu.Unlock()

	err := uc.setOnce(ctx, channelID, enabled)
	if errors.Is(err, port.ErrConflict) {
		err = uc.setOnce(ctx, channelID, enabled)
	}
	return err
}

func (uc *ReplyTargets) setOnce(ctx context.Context, channelID string, enabled bool) error {
	doc, version, err := uc.Store.Load(ctx)
	if err != nil {
		return fmt.Errorf("reply_targets_load: %w", err)
	}

	targets := make([]port.ReplyTarget, 0, len(doc.Targets)+1)
	for _, t := range doc.Targets {
		if t.ChannelID != channelID {
			targets = append(targets, t)
		}
	}
	if enabled {
		targets = append(targets, port.ReplyTarget{ChannelID: channelID, SetAt: uc.Clock.Now()})
	}

	if err := uc.Store.Save(ctx, port.ReplyTargetDoc{Targets: targets}, version); err != nil {
		if errors.Is(err, port.ErrConflict) {
			return err
		}
		return fmt.Errorf("reply_targets_save: %w", err)
	}
	return nil
}
