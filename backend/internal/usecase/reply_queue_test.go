package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/obsidian-engine/youtube-comment-user-list/backend/internal/adapter/memory"
	"github.com/obsidian-engine/youtube-comment-user-list/backend/internal/domain"
	"github.com/obsidian-engine/youtube-comment-user-list/backend/internal/port"
)

type fakePoster struct {
	failTimes int
	calls     []string
}

func (p *fakePoster) PostMessage(_ context.Context, liveChatID, text string) error {
	p.calls = append(p.calls, liveChatID+"|"+text)
	if len(p.calls) <= p.failTimes {
		return errors.New("youtube error")
	}
	return nil
}

type queueFixture struct {
	q       *ReplyQueue
	poster  *fakePoster
	state   port.StateRepo
	targets *memory.ReplyTargetStore
	records *memory.ReplyRecordStore
}

func newQueueFixture(t *testing.T) *queueFixture {
	t.Helper()
	ctx := context.Background()
	f := &queueFixture{
		poster:  &fakePoster{},
		state:   memory.NewStateRepo(),
		targets: memory.NewReplyTargetStore(),
		records: memory.NewReplyRecordStore(),
	}
	if err := f.state.Set(ctx, domain.LiveState{Status: domain.StatusActive, VideoID: "v1", LiveChatID: "chat1"}); err != nil {
		t.Fatal(err)
	}
	if err := f.targets.Save(ctx, port.ReplyTargetDoc{Targets: []port.ReplyTarget{{ChannelID: "UCflag", SetAt: flagSetAt}}}, 0); err != nil {
		t.Fatal(err)
	}
	f.q = NewReplyQueue(ReplyQueue{
		Poster: f.poster, State: f.state, Targets: f.targets, Records: f.records,
		Clock: &fixedClock{t: flagSetAt.Add(time.Hour)}, DailyLimit: 2,
	})
	return f
}

func (f *queueFixture) req() queuedReply {
	return queuedReply{ReplyRequest: ReplyRequest{VideoID: "v1", ChannelID: "UCflag", DisplayName: "Alice"}, LiveChatID: "chat1"}
}

func (f *queueFixture) record(t *testing.T, videoID string) port.ReplyRecord {
	t.Helper()
	doc, _, _ := f.records.Load(context.Background(), videoID)
	return doc.Records["UCflag"]
}

func TestReplyQueue_PostsChannelURL(t *testing.T) {
	f := newQueueFixture(t)
	f.q.process(context.Background(), f.req())

	want := "chat1|Aliceさん https://www.youtube.com/channel/UCflag"
	if len(f.poster.calls) != 1 || f.poster.calls[0] != want {
		t.Fatalf("unexpected posts: %+v", f.poster.calls)
	}
	if r := f.record(t, "v1"); r.State != port.ReplyDone || !r.Sent {
		t.Fatalf("expected done & sent, got %+v", r)
	}
}

func TestReplyQueue_SkipsWithoutPosting(t *testing.T) {
	cases := []struct {
		name  string
		setup func(t *testing.T, f *queueFixture)
	}{
		{"待っている間に動画を切り替えた", func(t *testing.T, f *queueFixture) {
			_ = f.state.Set(context.Background(), domain.LiveState{Status: domain.StatusActive, VideoID: "v2", LiveChatID: "chat2"})
		}},
		{"待っている間にリセットした", func(t *testing.T, f *queueFixture) {
			_ = f.state.Set(context.Background(), domain.NewWaitingState())
		}},
		{"待っている間にフラグを解除した", func(t *testing.T, f *queueFixture) {
			_ = f.targets.Save(context.Background(), port.ReplyTargetDoc{}, 1)
		}},
		{"1 日の上限に達している", func(t *testing.T, f *queueFixture) {
			sent := port.ReplyRecord{State: port.ReplyDone, At: flagSetAt.Add(30 * time.Minute), Sent: true}
			doc := port.ReplyRecordDoc{Records: map[string]port.ReplyRecord{"UCa": sent, "UCb": sent}}
			_ = f.records.Save(context.Background(), "v0", doc, 0)
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			f := newQueueFixture(t)
			tc.setup(t, f)
			f.q.process(context.Background(), f.req())
			if len(f.poster.calls) != 0 {
				t.Fatalf("expected no post, got %+v", f.poster.calls)
			}
			if r := f.record(t, "v1"); r.State != port.ReplyFailed || r.Sent {
				t.Fatalf("expected failed & not sent, got %+v", r)
			}
		})
	}
}

func TestReplyQueue_RetriesOnceThenGivesUp(t *testing.T) {
	f := newQueueFixture(t)
	f.poster.failTimes = 2
	f.q.process(context.Background(), f.req())
	if len(f.poster.calls) != 2 {
		t.Fatalf("expected 2 attempts, got %d", len(f.poster.calls))
	}
	if r := f.record(t, "v1"); r.State != port.ReplyFailed || !r.Sent {
		t.Fatalf("expected failed & sent, got %+v", r)
	}
}

func TestReplyQueue_SucceedsOnRetry(t *testing.T) {
	f := newQueueFixture(t)
	f.poster.failTimes = 1
	f.q.process(context.Background(), f.req())
	if r := f.record(t, "v1"); r.State != port.ReplyDone {
		t.Fatalf("expected done after retry, got %+v", r)
	}
}

func TestReplyQueue_EnqueueFailsWhenVideoChanged(t *testing.T) {
	f := newQueueFixture(t)
	f.q.Enqueue(context.Background(), ReplyRequest{VideoID: "old", ChannelID: "UCflag"})
	if len(f.q.ch) != 0 {
		t.Fatal("expected nothing queued")
	}
	if r := f.record(t, "old"); r.State != port.ReplyFailed {
		t.Fatalf("expected failed, got %+v", r)
	}
}

func TestQuotaDayStart_UsesPacificMidnight(t *testing.T) {
	// 日本時間 2026-11-26 16:59 は太平洋時間 (冬時間) 11-25 23:59 なので、1 日の始まりは 11-25 0:00 PST
	jst := time.FixedZone("JST", 9*3600)
	got := QuotaDayStart(time.Date(2026, 11, 26, 16, 59, 0, 0, jst))
	want := time.Date(2026, 11, 25, 8, 0, 0, 0, time.UTC)
	if !got.Equal(want) {
		t.Fatalf("got %v, want %v", got.UTC(), want)
	}
}
