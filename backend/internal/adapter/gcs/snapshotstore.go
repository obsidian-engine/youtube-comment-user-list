package gcs

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"time"

	"cloud.google.com/go/storage"
	"github.com/obsidian-engine/youtube-comment-user-list/backend/internal/port"
)

const (
	snapshotPrefix = "snapshots/"
	currentObject  = "snapshots/current.json"
)

// SnapshotStore は GCS を使った port.SnapshotSink 実装です。
type SnapshotStore struct {
	client *storage.Client
	bucket string
}

// NewSnapshotStore は SnapshotStore を生成します。
// client の Close は呼び出し元で管理してください（SIGTERM hook 推奨）。
func NewSnapshotStore(client *storage.Client, bucket string) *SnapshotStore {
	return &SnapshotStore{
		client: client,
		bucket: bucket,
	}
}

// Load は videoID に対応するスナップショットを GCS から読み込みます。
// オブジェクトが存在しない場合は (nil, nil) を返します。
func (s *SnapshotStore) Load(ctx context.Context, videoID string) (*port.Snapshot, error) {
	objName := snapshotPrefix + videoID + ".json"
	rc, err := s.client.Bucket(s.bucket).Object(objName).NewReader(ctx)
	if err != nil {
		if errors.Is(err, storage.ErrObjectNotExist) {
			return nil, nil
		}
		return nil, fmt.Errorf("gcs: open snapshot %s: %w", objName, err)
	}
	defer rc.Close()

	data, err := io.ReadAll(rc)
	if err != nil {
		return nil, fmt.Errorf("gcs: read snapshot %s: %w", objName, err)
	}

	var snap port.Snapshot
	if err := json.Unmarshal(data, &snap); err != nil {
		return nil, fmt.Errorf("gcs: unmarshal snapshot %s: %w", objName, err)
	}

	return &snap, nil
}

// Save はスナップショットを GCS に書き込みます（上書き）。
func (s *SnapshotStore) Save(ctx context.Context, snap *port.Snapshot) error {
	snap.SavedAt = time.Now()

	data, err := json.Marshal(snap)
	if err != nil {
		return fmt.Errorf("gcs: marshal snapshot %s: %w", snap.VideoID, err)
	}

	objName := snapshotPrefix + snap.VideoID + ".json"
	wc := s.client.Bucket(s.bucket).Object(objName).NewWriter(ctx)
	wc.ContentType = "application/json"

	if _, err := wc.Write(data); err != nil {
		_ = wc.Close()
		return fmt.Errorf("gcs: write snapshot %s: %w", objName, err)
	}

	if err := wc.Close(); err != nil {
		return fmt.Errorf("gcs: close snapshot writer %s: %w", objName, err)
	}

	return nil
}

// LoadCurrent は current.json を読み込みます。
// オブジェクトが存在しない場合は (nil, nil) を返します。
func (s *SnapshotStore) LoadCurrent(ctx context.Context) (*port.CurrentPointer, error) {
	rc, err := s.client.Bucket(s.bucket).Object(currentObject).NewReader(ctx)
	if err != nil {
		if errors.Is(err, storage.ErrObjectNotExist) {
			return nil, nil
		}
		return nil, fmt.Errorf("gcs: open current.json: %w", err)
	}
	defer rc.Close()

	data, err := io.ReadAll(rc)
	if err != nil {
		return nil, fmt.Errorf("gcs: read current.json: %w", err)
	}

	var ptr port.CurrentPointer
	if err := json.Unmarshal(data, &ptr); err != nil {
		return nil, fmt.Errorf("gcs: unmarshal current.json: %w", err)
	}

	return &ptr, nil
}

// SaveCurrent は current.json を書き込みます（上書き）。
func (s *SnapshotStore) SaveCurrent(ctx context.Context, ptr *port.CurrentPointer) error {
	ptr.SavedAt = time.Now()

	data, err := json.Marshal(ptr)
	if err != nil {
		return fmt.Errorf("gcs: marshal current.json: %w", err)
	}

	wc := s.client.Bucket(s.bucket).Object(currentObject).NewWriter(ctx)
	wc.ContentType = "application/json"

	if _, err := wc.Write(data); err != nil {
		_ = wc.Close()
		return fmt.Errorf("gcs: write current.json: %w", err)
	}

	if err := wc.Close(); err != nil {
		return fmt.Errorf("gcs: close current.json writer: %w", err)
	}

	return nil
}
