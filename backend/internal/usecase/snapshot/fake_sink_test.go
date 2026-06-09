package snapshot_test

import (
	"context"
	"sync"

	"github.com/obsidian-engine/youtube-comment-user-list/backend/internal/port"
)

// fakeSink は in-memory で port.SnapshotSink を実装するテスト用 fake です。
type fakeSink struct {
	mu         sync.Mutex
	snapshots  map[string]*port.Snapshot
	current    *port.CurrentPointer
	saveCount  int
	forceError error // nil でない場合、Save / SaveCurrent でこのエラーを返す
	loadError  error // nil でない場合、Load でこのエラーを返す
}

func newFakeSink() *fakeSink {
	return &fakeSink{
		snapshots: make(map[string]*port.Snapshot),
	}
}

func (f *fakeSink) Load(_ context.Context, videoID string) (*port.Snapshot, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.loadError != nil {
		return nil, f.loadError
	}
	snap, ok := f.snapshots[videoID]
	if !ok {
		return nil, nil
	}
	cp := *snap
	return &cp, nil
}

func (f *fakeSink) Save(_ context.Context, snap *port.Snapshot) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.forceError != nil {
		return f.forceError
	}
	cp := *snap
	f.snapshots[snap.VideoID] = &cp
	f.saveCount++
	return nil
}

func (f *fakeSink) LoadCurrent(_ context.Context) (*port.CurrentPointer, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.current == nil {
		return nil, nil
	}
	cp := *f.current
	return &cp, nil
}

func (f *fakeSink) SaveCurrent(_ context.Context, ptr *port.CurrentPointer) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.forceError != nil {
		return f.forceError
	}
	cp := *ptr
	f.current = &cp
	return nil
}

func (f *fakeSink) getSaveCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.saveCount
}

func (f *fakeSink) setForceError(err error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.forceError = err
}

func (f *fakeSink) setLoadError(err error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.loadError = err
}
