package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/obsidian-engine/youtube-comment-user-list/backend/internal/adapter/memory"
	"github.com/obsidian-engine/youtube-comment-user-list/backend/internal/port"
	"github.com/obsidian-engine/youtube-comment-user-list/backend/internal/usecase/snapshot"
)

type restoreCoord struct {
	snapshot.NopCoordinator
	healthy bool
}

func (c *restoreCoord) RestoreHealthy() bool { return c.healthy }

type autoReplyFixture struct {
	ar      *AutoReply
	records *memory.ReplyRecordStore
	coord   *restoreCoord
	queued  []ReplyRequest
}

var flagSetAt = time.Date(2026, 11, 25, 12, 0, 0, 0, time.UTC)

func newAutoReplyFixture(t *testing.T) *autoReplyFixture {
	t.Helper()
	targets := memory.NewReplyTargetStore()
	doc := port.ReplyTargetDoc{Targets: []port.ReplyTarget{{ChannelID: "UCflag", SetAt: flagSetAt}}}
	if err := targets.Save(context.Background(), doc, 0); err != nil {
		t.Fatal(err)
	}
	f := &autoReplyFixture{records: memory.NewReplyRecordStore(), coord: &restoreCoord{healthy: true}}
	f.ar = &AutoReply{
		Enabled: true,
		Targets: targets,
		Records: f.records,
		Snap:    f.coord,
		Clock:   &fixedClock{t: flagSetAt.Add(time.Hour)},
		Enqueue: func(_ context.Context, req ReplyRequest) { f.queued = append(f.queued, req) },
	}
	return f
}

func msgAt(channelID string, at time.Time) port.ChatMessage {
	return port.ChatMessage{ID: channelID + at.String(), ChannelID: channelID, DisplayName: "Alice", PublishedAt: at}
}

func (f *autoReplyFixture) record(t *testing.T, videoID, channelID string) (port.ReplyRecord, bool) {
	t.Helper()
	doc, _, err := f.records.Load(context.Background(), videoID)
	if err != nil {
		t.Fatal(err)
	}
	r, ok := doc.Records[channelID]
	return r, ok
}

func TestAutoReply_PlansFlaggedFirstComment(t *testing.T) {
	f := newAutoReplyFixture(t)
	f.ar.Consider(context.Background(), "v1", msgAt("UCflag", flagSetAt.Add(time.Minute)))

	if len(f.queued) != 1 || f.queued[0].ChannelID != "UCflag" || f.queued[0].VideoID != "v1" {
		t.Fatalf("expected one queued request, got %+v", f.queued)
	}
	if r, ok := f.record(t, "v1", "UCflag"); !ok || r.State != port.ReplyPosting {
		t.Fatalf("expected posting record, got %+v ok=%v", r, ok)
	}
}

func TestAutoReply_SkipsWhenNotPlanned(t *testing.T) {
	cases := []struct {
		name   string
		setup  func(f *autoReplyFixture)
		msg    port.ChatMessage
		record bool
	}{
		{name: "フラグの無い視聴者", msg: msgAt("UCother", flagSetAt.Add(time.Minute))},
		{name: "フラグを設定する前のコメント", msg: msgAt("UCflag", flagSetAt.Add(-time.Minute))},
		{name: "フラグを設定した時刻と同じコメント", msg: msgAt("UCflag", flagSetAt)},
		{
			name:  "投稿時刻を読み取れなかったコメント",
			msg:   port.ChatMessage{ChannelID: "UCflag", PublishedAt: flagSetAt.Add(time.Minute), PublishedAtUnknown: true},
			setup: func(*autoReplyFixture) {},
		},
		{
			name:  "一覧の復元に失敗した動画",
			setup: func(f *autoReplyFixture) { f.coord.healthy = false },
			msg:   msgAt("UCflag", flagSetAt.Add(time.Minute)),
		},
		{
			name:  "自動返信が無効",
			setup: func(f *autoReplyFixture) { f.ar.Enabled = false },
			msg:   msgAt("UCflag", flagSetAt.Add(time.Minute)),
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			f := newAutoReplyFixture(t)
			if tc.setup != nil {
				tc.setup(f)
			}
			f.ar.Consider(context.Background(), "v1", tc.msg)
			if len(f.queued) != 0 {
				t.Fatalf("expected no request, got %+v", f.queued)
			}
			if _, ok := f.record(t, "v1", tc.msg.ChannelID); ok {
				t.Fatalf("expected no record")
			}
		})
	}
}

func TestAutoReply_DoesNotPlanTwiceInSameVideoButPlansInNextVideo(t *testing.T) {
	f := newAutoReplyFixture(t)
	ctx := context.Background()
	// リセット後の再登録のように、同じ動画で再び新規登録として渡されても投稿しない
	f.ar.Consider(ctx, "v1", msgAt("UCflag", flagSetAt.Add(time.Minute)))
	f.ar.Consider(ctx, "v1", msgAt("UCflag", flagSetAt.Add(2*time.Minute)))
	if len(f.queued) != 1 {
		t.Fatalf("expected one request in v1, got %d", len(f.queued))
	}

	f.ar.Consider(ctx, "v2", msgAt("UCflag", flagSetAt.Add(3*time.Minute)))
	if len(f.queued) != 2 || f.queued[1].VideoID != "v2" {
		t.Fatalf("expected a request in v2, got %+v", f.queued)
	}
}

func TestAutoReply_MarksPostingAsFailedAfterRestart(t *testing.T) {
	f := newAutoReplyFixture(t)
	ctx := context.Background()
	f.ar.Consider(ctx, "v1", msgAt("UCflag", flagSetAt.Add(time.Minute)))

	// backend の再起動: 記録は GCS に残り、プロセスの状態は初期化される
	restarted := &AutoReply{
		Enabled: true, Targets: f.ar.Targets, Records: f.records, Snap: f.coord,
		Clock:   f.ar.Clock,
		Enqueue: func(_ context.Context, req ReplyRequest) { f.queued = append(f.queued, req) },
	}
	restarted.Consider(ctx, "v1", msgAt("UCflag", flagSetAt.Add(2*time.Minute)))

	if len(f.queued) != 1 {
		t.Fatalf("expected no repost after restart, got %d requests", len(f.queued))
	}
	if r, _ := f.record(t, "v1", "UCflag"); r.State != port.ReplyFailed {
		t.Fatalf("expected failed record after restart, got %+v", r)
	}
}

type failingRecordStore struct{ *memory.ReplyRecordStore }

func (s *failingRecordStore) Save(context.Context, string, port.ReplyRecordDoc, int64) error {
	return context.DeadlineExceeded
}

func TestAutoReply_DoesNotPlanWhenPostingRecordCannotBeSaved(t *testing.T) {
	f := newAutoReplyFixture(t)
	f.ar.Records = &failingRecordStore{memory.NewReplyRecordStore()}
	f.ar.Consider(context.Background(), "v1", msgAt("UCflag", flagSetAt.Add(time.Minute)))
	if len(f.queued) != 0 {
		t.Fatalf("expected no request, got %+v", f.queued)
	}
}

func TestAutoReply_NilIsNoop(t *testing.T) {
	var ar *AutoReply
	ar.Consider(context.Background(), "v1", msgAt("UCflag", flagSetAt.Add(time.Minute)))
}
