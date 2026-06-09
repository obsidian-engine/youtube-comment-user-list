package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/obsidian-engine/youtube-comment-user-list/backend/internal/domain"
	"github.com/obsidian-engine/youtube-comment-user-list/backend/internal/port"
	"github.com/obsidian-engine/youtube-comment-user-list/backend/internal/usecase"
)

// --- ListHistorySnapshots tests ---

func TestListHistorySnapshots_returnsItems(t *testing.T) {
	t.Helper()
	sink := newFakeSinkForUsecase()
	now := time.Now()
	_ = sink.Save(context.Background(), &port.Snapshot{
		VideoID: "vid1",
		SavedAt: now,
		Users:   make([]domain.User, 3),
	})
	_ = sink.Save(context.Background(), &port.Snapshot{
		VideoID:  "vid2",
		SavedAt:  now.Add(-time.Hour),
		Comments: make([]domain.Comment, 5),
	})

	uc := &usecase.ListHistorySnapshots{Sink: sink}
	out, err := uc.Execute(context.Background())
	if err != nil {
		t.Fatalf("Execute returned unexpected error: %v", err)
	}
	if len(out.Items) != 2 {
		t.Errorf("got %d items, want 2", len(out.Items))
	}
}

func TestListHistorySnapshots_sortedByDescSavedAt(t *testing.T) {
	t.Helper()
	sink := newFakeSinkForUsecase()
	older := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	newer := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)

	// 古い順で save (map iteration は非決定的なので 2 件のみ)
	_ = sink.Save(context.Background(), &port.Snapshot{VideoID: "old", SavedAt: older})
	_ = sink.Save(context.Background(), &port.Snapshot{VideoID: "new", SavedAt: newer})

	uc := &usecase.ListHistorySnapshots{Sink: sink}
	out, err := uc.Execute(context.Background())
	if err != nil {
		t.Fatalf("Execute returned unexpected error: %v", err)
	}
	if len(out.Items) != 2 {
		t.Fatalf("got %d items, want 2", len(out.Items))
	}
	if !out.Items[0].SavedAt.Equal(newer) {
		t.Errorf("first item savedAt = %v, want %v (newer)", out.Items[0].SavedAt, newer)
	}
	if !out.Items[1].SavedAt.Equal(older) {
		t.Errorf("second item savedAt = %v, want %v (older)", out.Items[1].SavedAt, older)
	}
}

// --- GetHistorySnapshot tests ---

func TestGetHistorySnapshot_returnsSnapshot(t *testing.T) {
	t.Helper()
	sink := newFakeSinkForUsecase()
	snap := &port.Snapshot{
		VideoID: "vid1",
		SavedAt: time.Now(),
		Users:   make([]domain.User, 2),
	}
	_ = sink.Save(context.Background(), snap)

	uc := &usecase.GetHistorySnapshot{Sink: sink}
	out, err := uc.Execute(context.Background(), "vid1")
	if err != nil {
		t.Fatalf("Execute returned unexpected error: %v", err)
	}
	if out.Snapshot == nil {
		t.Fatal("Snapshot is nil, want non-nil")
	}
	if out.Snapshot.VideoID != "vid1" {
		t.Errorf("VideoID = %q, want %q", out.Snapshot.VideoID, "vid1")
	}
}

func TestGetHistorySnapshot_notFound(t *testing.T) {
	t.Helper()
	sink := newFakeSinkForUsecase()

	uc := &usecase.GetHistorySnapshot{Sink: sink}
	_, err := uc.Execute(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("Execute returned nil error, want ErrNotFound")
	}
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("error = %v, want domain.ErrNotFound", err)
	}
}
