package http

import (
	"log"
	stdhttp "net/http"
	"time"
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

			if r.Method == stdhttp.MethodOptions {
				w.Header().Set("Access-Control-Allow-Methods", "GET,POST,OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
				w.WriteHeader(StatusNoContent)
				log.Printf("[CORS] Handled preflight request for %s", r.URL.Path)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
