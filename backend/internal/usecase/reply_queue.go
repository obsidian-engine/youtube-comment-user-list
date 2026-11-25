package usecase

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"
	_ "time/tzdata" // 実行環境に tz の情報が無くても太平洋時間で 1 日を区切れるようにする

	"github.com/obsidian-engine/youtube-comment-user-list/backend/internal/domain"
	"github.com/obsidian-engine/youtube-comment-user-list/backend/internal/port"
)

// quotaLocation は YouTube の API 使用量が 0 に戻る時刻 (太平洋時間の 0 時) の基準。
var quotaLocation = mustLoadLocation("America/Los_Angeles")

func mustLoadLocation(name string) *time.Location {
	loc, err := time.LoadLocation(name)
	if err != nil {
		panic(err)
	}
	return loc
}

// QuotaDayStart は t を含む 1 日 (太平洋時間の 0 時から翌日の 0 時まで) の始まりを返す。
func QuotaDayStart(t time.Time) time.Time {
	lt := t.In(quotaLocation)
	return time.Date(lt.Year(), lt.Month(), lt.Day(), 0, 0, 0, 0, quotaLocation)
}

// queuedReply は待ち行列に追加した時点の投稿先の配信を固定した依頼。
type queuedReply struct {
	ReplyRequest
	LiveChatID string
}

// ReplyQueue は自動返信の依頼を取得の処理と別に 1 件ずつ投稿する。
// 投稿の直前に、配信が変わっていないか・フラグが解除されていないか・1 日の上限に達していないかを確かめ、
// どれかに当たれば投稿せずに返信の記録を「失敗」にする。
type ReplyQueue struct {
	Poster     port.ChatPoster
	State      port.StateRepo
	Targets    port.ReplyTargetStore
	Records    port.ReplyRecordStore
	Clock      port.Clock
	Interval   time.Duration // 投稿と投稿の間隔
	RetryWait  time.Duration // 失敗した投稿をやり直すまでの待ち時間
	DailyLimit int

	ch chan queuedReply
}

func NewReplyQueue(q ReplyQueue) *ReplyQueue {
	q.ch = make(chan queuedReply, 100)
	return &q
}

// Enqueue は投稿の依頼を待ち行列に追加する。投稿先の配信はこの時点の配信に固定する。
// 待ち行列が満杯のときや配信が取得できないときは投稿せず、返信の記録を「失敗」にする。
func (q *ReplyQueue) Enqueue(ctx context.Context, req ReplyRequest) {
	state, err := q.State.Get(ctx)
	if err != nil || state.VideoID != req.VideoID || state.LiveChatID == "" {
		q.finish(ctx, req.VideoID, req.ChannelID, port.ReplyFailed, false, "video changed before enqueue")
		return
	}
	select {
	case q.ch <- queuedReply{ReplyRequest: req, LiveChatID: state.LiveChatID}:
	default:
		q.finish(ctx, req.VideoID, req.ChannelID, port.ReplyFailed, false, "queue is full")
	}
}

// Run は ctx が終わるまで待ち行列の依頼を 1 件ずつ投稿する。
func (q *ReplyQueue) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case req := <-q.ch:
			q.process(ctx, req)
			select {
			case <-ctx.Done():
				return
			case <-time.After(q.Interval):
			}
		}
	}
}

func (q *ReplyQueue) process(ctx context.Context, req queuedReply) {
	if reason := q.skipReason(ctx, req); reason != "" {
		q.finish(ctx, req.VideoID, req.ChannelID, port.ReplyFailed, false, reason)
		return
	}

	text := fmt.Sprintf("%sさん https://www.youtube.com/channel/%s", req.DisplayName, req.ChannelID)
	err := q.Poster.PostMessage(ctx, req.LiveChatID, text)
	if err != nil {
		log.Printf("[AUTO_REPLY] post to %s failed, retrying once: %v", req.ChannelID, err)
		select {
		case <-ctx.Done():
		case <-time.After(q.RetryWait):
		}
		err = q.Poster.PostMessage(ctx, req.LiveChatID, text)
	}
	if err != nil {
		q.finish(ctx, req.VideoID, req.ChannelID, port.ReplyFailed, true, fmt.Sprintf("post failed: %v", err))
		return
	}
	q.finish(ctx, req.VideoID, req.ChannelID, port.ReplyDone, true, "")
}

// skipReason は投稿しない理由を返す。投稿してよければ空文字を返す。
func (q *ReplyQueue) skipReason(ctx context.Context, req queuedReply) string {
	state, err := q.State.Get(ctx)
	if err != nil {
		return fmt.Sprintf("state unavailable: %v", err)
	}
	if state.Status != domain.StatusActive || state.VideoID != req.VideoID || state.LiveChatID != req.LiveChatID {
		return "video changed or reset while waiting"
	}

	doc, _, err := q.Targets.Load(ctx)
	if err != nil {
		return fmt.Sprintf("reply targets unavailable: %v", err)
	}
	if _, ok := findTarget(doc.Targets, req.ChannelID); !ok {
		return "flag removed while waiting"
	}

	sent, err := q.Records.CountSentSince(ctx, QuotaDayStart(q.Clock.Now()))
	if err != nil {
		return fmt.Sprintf("daily count unavailable: %v", err)
	}
	if sent >= q.DailyLimit {
		return fmt.Sprintf("daily limit reached (%d)", q.DailyLimit)
	}
	return ""
}

// finish は返信の記録を state にする。別の書き込みと重なったら最新の記録を読み込んで 1 回だけやり直す。
// 保存できなかった記録は「投稿中」のまま残り、次の起動で失敗として扱われる。
func (q *ReplyQueue) finish(ctx context.Context, videoID, channelID string, state port.ReplyState, sent bool, reason string) {
	if reason != "" {
		log.Printf("[AUTO_REPLY] %s -> %s on video %s: %s", channelID, state, videoID, reason)
	} else {
		log.Printf("[AUTO_REPLY] %s -> %s on video %s", channelID, state, videoID)
	}
	err := q.saveRecord(ctx, videoID, channelID, state, sent)
	if errors.Is(err, port.ErrConflict) {
		err = q.saveRecord(ctx, videoID, channelID, state, sent)
	}
	if err != nil {
		log.Printf("[AUTO_REPLY] save record for %s failed: %v", channelID, err)
	}
}

func (q *ReplyQueue) saveRecord(ctx context.Context, videoID, channelID string, state port.ReplyState, sent bool) error {
	doc, version, err := q.Records.Load(ctx, videoID)
	if err != nil {
		return err
	}
	if doc.Records == nil {
		doc.Records = map[string]port.ReplyRecord{}
	}
	doc.Records[channelID] = port.ReplyRecord{State: state, At: q.Clock.Now(), Sent: sent}
	return q.Records.Save(ctx, videoID, doc, version)
}
