package port

import (
    "context"
    "github.com/obsidian-engine/youtube-comment-user-list/backend/internal/domain"
)

// StateRepo は配信状態を永続化（または InMemory 保持）します。
type StateRepo interface {
    Get(ctx context.Context) (domain.LiveState, error)
    Set(ctx context.Context, st domain.LiveState) error
}
