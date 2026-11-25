package gcs

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"cloud.google.com/go/storage"
	"github.com/obsidian-engine/youtube-comment-user-list/backend/internal/port"
	"google.golang.org/api/googleapi"
)

const replyTargetsObject = "reply-targets.json"

// ReplyTargetStore は GCS を使った port.ReplyTargetStore 実装。
// version には GCS の object の generation を使い、Save は generation が一致するときだけ上書きする。
type ReplyTargetStore struct {
	client *storage.Client
	bucket string
}

func NewReplyTargetStore(client *storage.Client, bucket string) *ReplyTargetStore {
	return &ReplyTargetStore{client: client, bucket: bucket}
}

func (s *ReplyTargetStore) Load(ctx context.Context) (port.ReplyTargetDoc, int64, error) {
	rc, err := s.client.Bucket(s.bucket).Object(replyTargetsObject).NewReader(ctx)
	if err != nil {
		if errors.Is(err, storage.ErrObjectNotExist) {
			return port.ReplyTargetDoc{}, 0, nil
		}
		return port.ReplyTargetDoc{}, 0, fmt.Errorf("gcs: open %s: %w", replyTargetsObject, err)
	}
	defer func() { _ = rc.Close() }()

	data, err := io.ReadAll(rc)
	if err != nil {
		return port.ReplyTargetDoc{}, 0, fmt.Errorf("gcs: read %s: %w", replyTargetsObject, err)
	}
	var doc port.ReplyTargetDoc
	if err := json.Unmarshal(data, &doc); err != nil {
		return port.ReplyTargetDoc{}, 0, fmt.Errorf("gcs: unmarshal %s: %w", replyTargetsObject, err)
	}
	return doc, rc.Attrs.Generation, nil
}

func (s *ReplyTargetStore) Save(ctx context.Context, doc port.ReplyTargetDoc, version int64) error {
	data, err := json.Marshal(&doc)
	if err != nil {
		return fmt.Errorf("gcs: marshal %s: %w", replyTargetsObject, err)
	}

	obj := s.client.Bucket(s.bucket).Object(replyTargetsObject)
	// version 0 は「まだ記録が無い」と読んだ状態なので、作成のみを許す
	if version == 0 {
		obj = obj.If(storage.Conditions{DoesNotExist: true})
	} else {
		obj = obj.If(storage.Conditions{GenerationMatch: version})
	}

	wc := obj.NewWriter(ctx)
	wc.ContentType = "application/json"
	if _, err := wc.Write(data); err != nil {
		_ = wc.Close()
		return fmt.Errorf("gcs: write %s: %w", replyTargetsObject, err)
	}
	if err := wc.Close(); err != nil {
		var gerr *googleapi.Error
		if errors.As(err, &gerr) && gerr.Code == http.StatusPreconditionFailed {
			return port.ErrConflict
		}
		return fmt.Errorf("gcs: close %s writer: %w", replyTargetsObject, err)
	}
	return nil
}
