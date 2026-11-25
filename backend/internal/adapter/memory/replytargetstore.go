package memory

import (
	"context"
	"sync"

	"github.com/obsidian-engine/youtube-comment-user-list/backend/internal/port"
)

// ReplyTargetStore はメモリだけに保存する port.ReplyTargetStore 実装。
// GCS_BUCKET が未設定の環境 (ローカル開発) と test で使う。
type ReplyTargetStore struct {
	mu      sync.Mutex
	doc     port.ReplyTargetDoc
	version int64
}

func NewReplyTargetStore() *ReplyTargetStore {
	return &ReplyTargetStore{}
}

func (s *ReplyTargetStore) Load(_ context.Context) (port.ReplyTargetDoc, int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return copyDoc(s.doc), s.version, nil
}

func (s *ReplyTargetStore) Save(_ context.Context, doc port.ReplyTargetDoc, version int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if version != s.version {
		return port.ErrConflict
	}
	s.doc = copyDoc(doc)
	s.version++
	return nil
}

func copyDoc(doc port.ReplyTargetDoc) port.ReplyTargetDoc {
	targets := make([]port.ReplyTarget, len(doc.Targets))
	copy(targets, doc.Targets)
	return port.ReplyTargetDoc{Targets: targets}
}
