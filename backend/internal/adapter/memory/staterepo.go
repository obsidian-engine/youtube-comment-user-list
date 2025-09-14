package memory

import (
    "context"
    "sync"

    "github.com/obsidian-engine/youtube-comment-user-list/backend/internal/domain"
)

type StateRepo struct {
    mu  sync.RWMutex
    cur domain.LiveState
}

func NewStateRepo() *StateRepo { return &StateRepo{} }

func (r *StateRepo) Get(ctx context.Context) (domain.LiveState, error) {
    r.mu.RLock()
    st := r.cur
    r.mu.RUnlock()
    return st, nil
}

func (r *StateRepo) Set(ctx context.Context, st domain.LiveState) error {
    r.mu.Lock()
    r.cur = st
    r.mu.Unlock()
    return nil
}
