package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/obsidian-engine/youtube-comment-user-list/backend/internal/adapter/memory"
	"github.com/obsidian-engine/youtube-comment-user-list/backend/internal/port"
)

type fixedClock struct{ t time.Time }

func (c *fixedClock) Now() time.Time { return c.t }

// conflictOnceStore は最初の Save だけ、別の backend が先に保存した状況を再現する。
type conflictOnceStore struct {
	*memory.ReplyTargetStore
	conflicted bool
	saves      int
}

func (s *conflictOnceStore) Save(ctx context.Context, doc port.ReplyTargetDoc, version int64) error {
	s.saves++
	if !s.conflicted {
		s.conflicted = true
		// 別の backend の保存で version が進んだ状態にする
		other := port.ReplyTargetDoc{Targets: []port.ReplyTarget{{ChannelID: "UCother"}}}
		if err := s.ReplyTargetStore.Save(ctx, other, version); err != nil {
			return err
		}
		return port.ErrConflict
	}
	return s.ReplyTargetStore.Save(ctx, doc, version)
}

func TestReplyTargets_SetAndUnset(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 11, 25, 12, 0, 0, 0, time.UTC)
	uc := &ReplyTargets{Store: memory.NewReplyTargetStore(), Clock: &fixedClock{t: now}}

	if err := uc.Set(ctx, "UCabc", true); err != nil {
		t.Fatalf("set: %v", err)
	}
	got, err := uc.List(ctx)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(got) != 1 || got[0].ChannelID != "UCabc" || !got[0].SetAt.Equal(now) {
		t.Fatalf("unexpected targets: %+v", got)
	}

	if err := uc.Set(ctx, "UCabc", false); err != nil {
		t.Fatalf("unset: %v", err)
	}
	got, _ = uc.List(ctx)
	if len(got) != 0 {
		t.Fatalf("expected no targets, got %+v", got)
	}
}

func TestReplyTargets_ResetUpdatesSetAt(t *testing.T) {
	ctx := context.Background()
	clock := &fixedClock{t: time.Date(2026, 11, 25, 12, 0, 0, 0, time.UTC)}
	uc := &ReplyTargets{Store: memory.NewReplyTargetStore(), Clock: clock}

	_ = uc.Set(ctx, "UCabc", true)
	later := clock.t.Add(time.Hour)
	clock.t = later
	_ = uc.Set(ctx, "UCabc", true)

	got, _ := uc.List(ctx)
	if len(got) != 1 || !got[0].SetAt.Equal(later) {
		t.Fatalf("expected one target with updated SetAt, got %+v", got)
	}
}

func TestReplyTargets_RetriesOnceOnConflictWithoutLosingOtherWrite(t *testing.T) {
	ctx := context.Background()
	store := &conflictOnceStore{ReplyTargetStore: memory.NewReplyTargetStore()}
	uc := &ReplyTargets{Store: store, Clock: &fixedClock{t: time.Now()}}

	if err := uc.Set(ctx, "UCabc", true); err != nil {
		t.Fatalf("set: %v", err)
	}
	if store.saves != 2 {
		t.Fatalf("expected 2 saves (conflict then retry), got %d", store.saves)
	}
	got, _ := uc.List(ctx)
	ids := map[string]bool{}
	for _, tg := range got {
		ids[tg.ChannelID] = true
	}
	if !ids["UCabc"] || !ids["UCother"] {
		t.Fatalf("expected both writes kept, got %+v", got)
	}
}

type alwaysConflictStore struct{ *memory.ReplyTargetStore }

func (s *alwaysConflictStore) Save(context.Context, port.ReplyTargetDoc, int64) error {
	return port.ErrConflict
}

func TestReplyTargets_FailsAfterSecondConflict(t *testing.T) {
	uc := &ReplyTargets{Store: &alwaysConflictStore{memory.NewReplyTargetStore()}, Clock: &fixedClock{t: time.Now()}}
	err := uc.Set(context.Background(), "UCabc", true)
	if !errors.Is(err, port.ErrConflict) {
		t.Fatalf("expected ErrConflict, got %v", err)
	}
}
