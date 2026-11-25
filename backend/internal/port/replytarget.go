package port

import (
	"context"
	"errors"
	"time"
)

// ErrConflict は、読んだ後に別の書き込みで記録が変わっていたため保存を拒否したことを表す。
var ErrConflict = errors.New("conflict")

// ReplyTarget は自動返信の対象としてフラグを設定した視聴者を表す。
type ReplyTarget struct {
	ChannelID string    `json:"channelId"`
	SetAt     time.Time `json:"setAt"`
}

// ReplyTargetDoc はフラグの記録全体を表す。配信をまたいで 1 つだけ存在する。
type ReplyTargetDoc struct {
	Targets []ReplyTarget `json:"targets"`
}

// ReplyTargetStore はフラグの記録の読み書きを抽象化する。
// Load が返す version を Save に渡し、その後に別の保存があれば ErrConflict を返す。
// 記録が無いときは、対象が 0 件の doc と version 0 を返す。
type ReplyTargetStore interface {
	Load(ctx context.Context) (ReplyTargetDoc, int64, error)
	Save(ctx context.Context, doc ReplyTargetDoc, version int64) error
}
