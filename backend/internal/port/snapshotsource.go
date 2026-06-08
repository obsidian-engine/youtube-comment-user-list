package port

import (
	"github.com/obsidian-engine/youtube-comment-user-list/backend/internal/domain"
)

// UserSnapshotSource は in-memory UserRepo の snapshot dump/restore port です。
type UserSnapshotSource interface {
	Dump() ([]domain.User, []string) // users + processedMsgs
	LoadFrom(users []domain.User, processedMsgs []string)
}

// CommentSnapshotSource は in-memory CommentRepo の snapshot dump/restore port です。
type CommentSnapshotSource interface {
	Dump() []domain.Comment
	LoadFrom(comments []domain.Comment)
}
