package port

import (
	"context"
	"time"

	"github.com/obsidian-engine/youtube-comment-user-list/backend/internal/domain"
)

// Snapshot は単一 video の状態スナップショットです。
type Snapshot struct {
	SchemaVersion int               `json:"schemaVersion"`
	VideoID       string            `json:"videoId"`
	LiveChatID    string            `json:"liveChatId"`
	SavedAt       time.Time         `json:"savedAt"`
	Users         []domain.User     `json:"users"`
	Comments      []domain.Comment  `json:"comments"`
	ProcessedMsgs []string          `json:"processedMsgs"`
	State         *domain.LiveState `json:"state,omitempty"` // nil の場合は旧 snapshot 互換として skip
}

// CurrentPointer は現在アクティブな video を指すポインタです。
type CurrentPointer struct {
	VideoID string    `json:"videoId"`
	SavedAt time.Time `json:"savedAt"`
}

// SnapshotSink はスナップショットの永続化 port です。
// 不在時は (nil, nil) を返します。
type SnapshotSink interface {
	Load(ctx context.Context, videoID string) (*Snapshot, error)
	Save(ctx context.Context, snap *Snapshot) error
	LoadCurrent(ctx context.Context) (*CurrentPointer, error)
	SaveCurrent(ctx context.Context, ptr *CurrentPointer) error
}
