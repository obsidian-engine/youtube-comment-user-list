package gcs

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"cloud.google.com/go/storage"
	"github.com/obsidian-engine/youtube-comment-user-list/backend/internal/port"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/iterator"
)

// CountSentSince は reply-records/ の object のうち、since 以降に更新されたものだけを読んで集計する。
// 1 日の上限の判定に使い、読めなかった object があれば error を返す (上限を超えて投稿しないため)。
func (s *ReplyRecordStore) CountSentSince(ctx context.Context, since time.Time) (int, error) {
	n := 0
	it := s.client.Bucket(s.bucket).Objects(ctx, &storage.Query{Prefix: replyRecordPrefix})
	for {
		attrs, err := it.Next()
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			return 0, fmt.Errorf("gcs: list reply records: %w", err)
		}
		if !strings.HasSuffix(attrs.Name, ".json") || attrs.Updated.Before(since) {
			continue
		}
		videoID := strings.TrimSuffix(strings.TrimPrefix(attrs.Name, replyRecordPrefix), ".json")
		doc, _, err := s.Load(ctx, videoID)
		if err != nil {
			return 0, err
		}
		n += port.CountSent(doc, since)
	}
	return n, nil
}

const replyRecordPrefix = "reply-records/"

// ReplyRecordStore は GCS を使った port.ReplyRecordStore 実装。動画ごとに 1 つの object に保存する。
// version には object の generation を使い、Save は generation が一致するときだけ上書きする。
type ReplyRecordStore struct {
	client *storage.Client
	bucket string
}

func NewReplyRecordStore(client *storage.Client, bucket string) *ReplyRecordStore {
	return &ReplyRecordStore{client: client, bucket: bucket}
}

func (s *ReplyRecordStore) Load(ctx context.Context, videoID string) (port.ReplyRecordDoc, int64, error) {
	objName := replyRecordPrefix + videoID + ".json"
	rc, err := s.client.Bucket(s.bucket).Object(objName).NewReader(ctx)
	if err != nil {
		if errors.Is(err, storage.ErrObjectNotExist) {
			return port.ReplyRecordDoc{Records: map[string]port.ReplyRecord{}}, 0, nil
		}
		return port.ReplyRecordDoc{}, 0, fmt.Errorf("gcs: open %s: %w", objName, err)
	}
	defer func() { _ = rc.Close() }()

	data, err := io.ReadAll(rc)
	if err != nil {
		return port.ReplyRecordDoc{}, 0, fmt.Errorf("gcs: read %s: %w", objName, err)
	}
	var doc port.ReplyRecordDoc
	if err := json.Unmarshal(data, &doc); err != nil {
		return port.ReplyRecordDoc{}, 0, fmt.Errorf("gcs: unmarshal %s: %w", objName, err)
	}
	if doc.Records == nil {
		doc.Records = map[string]port.ReplyRecord{}
	}
	return doc, rc.Attrs.Generation, nil
}

func (s *ReplyRecordStore) Save(ctx context.Context, videoID string, doc port.ReplyRecordDoc, version int64) error {
	objName := replyRecordPrefix + videoID + ".json"
	data, err := json.Marshal(&doc)
	if err != nil {
		return fmt.Errorf("gcs: marshal %s: %w", objName, err)
	}

	obj := s.client.Bucket(s.bucket).Object(objName)
	if version == 0 {
		obj = obj.If(storage.Conditions{DoesNotExist: true})
	} else {
		obj = obj.If(storage.Conditions{GenerationMatch: version})
	}

	wc := obj.NewWriter(ctx)
	wc.ContentType = "application/json"
	if _, err := wc.Write(data); err != nil {
		_ = wc.Close()
		return fmt.Errorf("gcs: write %s: %w", objName, err)
	}
	if err := wc.Close(); err != nil {
		var gerr *googleapi.Error
		if errors.As(err, &gerr) && gerr.Code == http.StatusPreconditionFailed {
			return port.ErrConflict
		}
		return fmt.Errorf("gcs: close %s writer: %w", objName, err)
	}
	return nil
}
