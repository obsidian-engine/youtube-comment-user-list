package port

import (
	"context"
	"time"
)

// ReplyState は 1 人の視聴者への自動返信の状態。
type ReplyState string

const (
	// ReplyPosting は投稿しようとしていて、成功か失敗かがまだ決まっていない状態。
	ReplyPosting ReplyState = "posting"
	// ReplyDone は投稿に成功した状態。
	ReplyDone ReplyState = "done"
	// ReplyFailed は投稿に失敗したか、投稿せずに待ちを削除した状態。再投稿しない。
	ReplyFailed ReplyState = "failed"
)

// ReplyRecord は 1 人の視聴者への自動返信の記録。
// Sent は YouTube へ投稿を送ったかどうかで、1 日の上限の集計に使う。
type ReplyRecord struct {
	State ReplyState `json:"state"`
	At    time.Time  `json:"at"`
	Sent  bool       `json:"sent"`
}

// ReplyRecordDoc は 1 つの動画の返信の記録。key はチャンネル ID。
type ReplyRecordDoc struct {
	Records map[string]ReplyRecord `json:"records"`
}

// ReplyRecordStore は動画ごとの返信の記録の読み書きを抽象化する。
// version の扱いは ReplyTargetStore と同じで、記録が無いときは version 0 を返す。
type ReplyRecordStore interface {
	Load(ctx context.Context, videoID string) (ReplyRecordDoc, int64, error)
	Save(ctx context.Context, videoID string, doc ReplyRecordDoc, version int64) error
}
