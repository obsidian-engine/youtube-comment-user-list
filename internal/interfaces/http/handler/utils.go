package handler

import (
	"fmt"
	"net/http"
	"strings"
	"time"
)

// correlationIDFrom はリクエストヘッダ (X-Request-ID, requestId) から相関IDを取得し、
// 無い場合は現在時刻ナノ秒を用いた簡易IDを生成して prefix を付与します。
func correlationIDFrom(r *http.Request, prefix string) string {
	id := r.Header.Get("X-Request-ID")
	if id == "" {
		id = r.Header.Get("requestId")
	}
	id = strings.TrimSpace(id)
	if id == "" {
		id = fmt.Sprintf("%d", time.Now().UnixNano())
	}
	if prefix == "" {
		return id
	}
	return prefix + "-" + id
}
