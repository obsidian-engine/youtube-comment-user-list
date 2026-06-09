package usecase_test

import (
	"context"
	"sync"

	"github.com/obsidian-engine/youtube-comment-user-list/backend/internal/port"
)

// fakeSinkForUsecase は usecase_test パッケージで使用する in-memory SnapshotSink。
type fakeSinkForUsecase struct {
	mu        sync.Mutex
	snapshots map[string]*port.Snapshot
	current   *port.CurrentPointer
	loadErr   error
}

func newFakeSinkForUsecase() *fakeSinkForUsecase {
	return &fakeSinkForUsecase{
		snapshots: make(map[string]*port.Snapshot),
	}
}

func (f *fakeSinkForUsecase) Load(_ context.Context, videoID string) (*port.Snapshot, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.loadErr != nil {
		return nil, f.loadErr
	}
	snap, ok := f.snapshots[videoID]
	if !ok {
		return nil, nil
	}
	cp := *snap
	return &cp, nil
}

func (f *fakeSinkForUsecase) Save(_ context.Context, snap *port.Snapshot) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	cp := *snap
	f.snapshots[snap.VideoID] = &cp
	return nil
}

func (f *fakeSinkForUsecase) LoadCurrent(_ context.Context) (*port.CurrentPointer, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.current == nil {
		return nil, nil
	}
	cp := *f.current
	return &cp, nil
}

func (f *fakeSinkForUsecase) SaveCurrent(_ context.Context, ptr *port.CurrentPointer) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	cp := *ptr
	f.current = &cp
	return nil
}
