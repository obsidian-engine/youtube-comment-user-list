package port

import "time"

// Clock は現在時刻の取得を抽象化します。
// 実装例: systemClock など。
type Clock interface {
    Now() time.Time
}
