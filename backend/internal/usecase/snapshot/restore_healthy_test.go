package snapshot_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/obsidian-engine/youtube-comment-user-list/backend/internal/domain"
	"github.com/obsidian-engine/youtube-comment-user-list/backend/internal/port"
	"github.com/obsidian-engine/youtube-comment-user-list/backend/internal/usecase/snapshot"
)

func TestRestoreHealthy(t *testing.T) {
	ctx := context.Background()
	newCoord := func(sink *fakeSink) snapshot.Coordinator {
		ur, cr := newTestRepos()
		return snapshot.NewCoordinator(sink, ur, cr, nil, 30*time.Second)
	}

	t.Run("保存が無い動画は成功として扱う", func(t *testing.T) {
		c := newCoord(newFakeSink())
		if _, err := c.RestoreFor(ctx, "new"); err != nil {
			t.Fatal(err)
		}
		if !c.RestoreHealthy() {
			t.Fatal("expected healthy")
		}
	})

	t.Run("読み込みの error は失敗として扱い、次の成功で戻る", func(t *testing.T) {
		sink := newFakeSink()
		c := newCoord(sink)
		sink.setLoadError(errors.New("gcs down"))
		if _, err := c.RestoreFor(ctx, "v1"); err == nil {
			t.Fatal("expected error")
		}
		if c.RestoreHealthy() {
			t.Fatal("expected unhealthy after load error")
		}
		sink.setLoadError(nil)
		sink.snapshots["v1"] = &port.Snapshot{VideoID: "v1"}
		if restored, err := c.RestoreFor(ctx, "v1"); err != nil || !restored {
			t.Fatalf("restore: restored=%v err=%v", restored, err)
		}
		if !c.RestoreHealthy() {
			t.Fatal("expected healthy after successful restore")
		}
	})

	t.Run("別の動画の snapshot だったときは失敗として扱う", func(t *testing.T) {
		sink := newFakeSink()
		sink.snapshots["v1"] = &port.Snapshot{VideoID: "v1", State: &domain.LiveState{VideoID: "other"}}
		c := newCoord(sink)
		if restored, err := c.RestoreFor(ctx, "v1"); err != nil || restored {
			t.Fatalf("restore: restored=%v err=%v", restored, err)
		}
		if c.RestoreHealthy() {
			t.Fatal("expected unhealthy after mismatched snapshot")
		}
	})
}
