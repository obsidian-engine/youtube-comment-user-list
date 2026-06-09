package usecase

import (
	"context"
	"fmt"
	"sort"

	"github.com/obsidian-engine/youtube-comment-user-list/backend/internal/domain"
	"github.com/obsidian-engine/youtube-comment-user-list/backend/internal/port"
)

// ListHistorySnapshots は GCS 上の全 snapshot サマリーを savedAt 降順で返します。
type ListHistorySnapshots struct {
	Sink port.SnapshotSink
}

// ListHistorySnapshotsOutput は ListHistorySnapshots の出力です。
type ListHistorySnapshotsOutput struct {
	Items []port.SnapshotSummary
}

// Execute は snapshot サマリー一覧を取得して savedAt 降順にソートして返します。
func (uc *ListHistorySnapshots) Execute(ctx context.Context) (ListHistorySnapshotsOutput, error) {
	items, err := uc.Sink.List(ctx)
	if err != nil {
		return ListHistorySnapshotsOutput{}, fmt.Errorf("snapshot_list: %w", err)
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].SavedAt.After(items[j].SavedAt)
	})
	return ListHistorySnapshotsOutput{Items: items}, nil
}

// GetHistorySnapshot は指定 videoID の snapshot を返します。
// snapshot が存在しない場合は domain.ErrNotFound を返します。
type GetHistorySnapshot struct {
	Sink port.SnapshotSink
}

// GetHistorySnapshotOutput は GetHistorySnapshot の出力です。
type GetHistorySnapshotOutput struct {
	Snapshot *port.Snapshot
}

// Execute は指定 videoID の snapshot を取得して返します。
func (uc *GetHistorySnapshot) Execute(ctx context.Context, videoID string) (GetHistorySnapshotOutput, error) {
	snap, err := uc.Sink.Load(ctx, videoID)
	if err != nil {
		return GetHistorySnapshotOutput{}, fmt.Errorf("snapshot_load: %w", err)
	}
	if snap == nil {
		return GetHistorySnapshotOutput{}, domain.ErrNotFound
	}
	return GetHistorySnapshotOutput{Snapshot: snap}, nil
}
