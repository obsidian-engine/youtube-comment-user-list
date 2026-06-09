package http

import (
	"context"
	"fmt"
	"log"
	stdhttp "net/http"
	"runtime/debug"
	"time"

	"github.com/obsidian-engine/youtube-comment-user-list/backend/internal/adapter/logging"
)

// LoggingMiddleware はリクエスト/レスポンスをログ出力するミドルウェア
func LoggingMiddleware(next stdhttp.Handler) stdhttp.Handler {
	return stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
		start := time.Now()

		// レスポンスライターをラップしてステータスコードを取得
		wrapped := &responseWriter{
			ResponseWriter: w,
			statusCode:     200,
		}

		// リクエスト開始ログ
		log.Printf("[REQUEST] %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)

		// 次のハンドラーを実行
		next.ServeHTTP(wrapped, r)

		// レスポンス完了ログ
		duration := time.Since(start)
		log.Printf("[RESPONSE] %s %s -> %d (%v)",
			r.Method, r.URL.Path, wrapped.statusCode, duration)
	})
}

// responseWriter はステータスコードを記録するためのラッパー
type responseWriter struct {
	stdhttp.ResponseWriter
	statusCode int
}

// WriteHeader はステータスコードを記録する
func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// CORSMiddleware はCORS設定を処理するミドルウェア
func CORSMiddleware(frontendOrigin string) func(stdhttp.Handler) stdhttp.Handler {
	return func(next stdhttp.Handler) stdhttp.Handler {
		return stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
			if frontendOrigin != "" {
				w.Header().Set("Access-Control-Allow-Origin", frontendOrigin)
				w.Header().Set("Vary", "Origin")
				log.Printf("[CORS] Set Allow-Origin: %s", frontendOrigin)
			}

			// すべてのリクエストに対して基本的なCORSヘッダーを設定
			w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type,Authorization,X-Requested-With")
			w.Header().Set("Access-Control-Max-Age", "86400")

			if r.Method == stdhttp.MethodOptions {
				w.WriteHeader(StatusNoContent)
				log.Printf("[CORS] Handled preflight request for %s", r.URL.Path)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// collectorCtxKey はCollectorをcontextに格納するためのキー型
type collectorCtxKey struct{}

// CollectorMiddleware は全リクエストのcontextにlogging.CollectorをInjectするミドルウェア
func CollectorMiddleware(next stdhttp.Handler) stdhttp.Handler {
	return stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
		collector := logging.NewCollector()
		// logging パッケージの WithCollector でも格納し、usecase 層の Log() が使える状態にする
		ctx := logging.WithCollector(r.Context(), collector)
		// handler 層から collectorFromRequest() で取り出せるよう独自 key でも格納
		ctx = context.WithValue(ctx, collectorCtxKey{}, collector)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// collectorFromRequest はリクエストのcontextからCollectorを取り出す
func collectorFromRequest(r *stdhttp.Request) *logging.Collector {
	c, _ := r.Context().Value(collectorCtxKey{}).(*logging.Collector)
	return c
}

// RecoverMiddleware はpanicをrecoverして500レスポンスを返すミドルウェア
func RecoverMiddleware(next stdhttp.Handler) stdhttp.Handler {
	return stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				stack := debug.Stack()
				log.Printf("[PANIC] %v\n%s", rec, stack)
				if c := collectorFromRequest(r); c != nil {
					c.Add("error", "PANIC", fmt.Sprintf("%v", rec))
					c.Add("error", "PANIC_STACK", string(stack))
				}
				renderInternalErrorWithCollector(w, r, fmt.Sprintf("internal panic: %v", rec), collectorFromRequest(r))
			}
		}()
		next.ServeHTTP(w, r)
	})
}
