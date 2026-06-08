package port

import (
	"github.com/obsidian-engine/youtube-comment-user-list/backend/internal/domain"
)

// UserSnapshot は UserRepo の snapshot を表す値型です。
// tuple 戻り値の代わりに struct を使うことで、将来フィールド追加時の呼出し側変更を局所化します。
type UserSnapshot struct {
	Users         []domain.User
	ProcessedMsgs []string
}

// UserSnapshotSource は in-memory UserRepo の snapshot dump/restore port です。
type UserSnapshotSource interface {
	Dump() UserSnapshot
	LoadFrom(snap UserSnapshot)
}

// CommentSnapshotSource は in-memory CommentRepo の snapshot dump/restore port です。
type CommentSnapshotSource interface {
	Dump() []domain.Comment
	LoadFrom(comments []domain.Comment)
}
