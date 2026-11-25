package memory

import (
	"context"
	"sync"
	"time"

	"github.com/obsidian-engine/youtube-comment-user-list/backend/internal/port"
)

// ReplyRecordStore はメモリだけに保存する port.ReplyRecordStore 実装。
// GCS_BUCKET が未設定の環境 (ローカル開発) と test で使う。
type ReplyRecordStore struct {
	mu       sync.Mutex
	docs     map[string]port.ReplyRecordDoc
	versions map[string]int64
}

func NewReplyRecordStore() *ReplyRecordStore {
	return &ReplyRecordStore{docs: map[string]port.ReplyRecordDoc{}, versions: map[string]int64{}}
}

func (s *ReplyRecordStore) Load(_ context.Context, videoID string) (port.ReplyRecordDoc, int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return copyRecords(s.docs[videoID]), s.versions[videoID], nil
}

func (s *ReplyRecordStore) Save(_ context.Context, videoID string, doc port.ReplyRecordDoc, version int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if version != s.versions[videoID] {
		return port.ErrConflict
	}
	s.docs[videoID] = copyRecords(doc)
	s.versions[videoID]++
	return nil
}

func (s *ReplyRecordStore) CountSentSince(_ context.Context, since time.Time) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	n := 0
	for _, doc := range s.docs {
		n += port.CountSent(doc, since)
	}
	return n, nil
}

func copyRecords(doc port.ReplyRecordDoc) port.ReplyRecordDoc {
	records := make(map[string]port.ReplyRecord, len(doc.Records))
	for k, v := range doc.Records {
		records[k] = v
	}
	return port.ReplyRecordDoc{Records: records}
}
