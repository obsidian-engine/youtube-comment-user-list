package http

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCORSMiddleware(t *testing.T) {
	tests := []struct {
		name            string
		frontendOrigin  string
		requestMethod   string
		requestHeaders  map[string]string
		expectedOrigin  string
		expectedStatus  int
		expectedHeaders map[string]string
	}{
		{
			name:           "OPTIONS preflight request with frontend origin",
			frontendOrigin: "https://example.com",
			requestMethod:  "OPTIONS",
			requestHeaders: map[string]string{
				"Origin": "https://example.com",
			},
			expectedOrigin: "https://example.com",
			expectedStatus: StatusNoContent,
			expectedHeaders: map[string]string{
				"Access-Control-Allow-Origin":  "https://example.com",
				"Access-Control-Allow-Methods": "GET,POST,PUT,DELETE,OPTIONS",
				"Access-Control-Allow-Headers": "Content-Type,Authorization,X-Requested-With",
				"Access-Control-Max-Age":       "86400",
				"Vary":                         "Origin",
			},
		},
		{
			name:           "GET request with frontend origin",
			frontendOrigin: "https://example.com",
			requestMethod:  "GET",
			requestHeaders: map[string]string{
				"Origin": "https://example.com",
			},
			expectedOrigin: "https://example.com",
			expectedStatus: 200,
			expectedHeaders: map[string]string{
				"Access-Control-Allow-Origin":  "https://example.com",
				"Access-Control-Allow-Methods": "GET,POST,PUT,DELETE,OPTIONS",
				"Access-Control-Allow-Headers": "Content-Type,Authorization,X-Requested-With",
				"Access-Control-Max-Age":       "86400",
				"Vary":                         "Origin",
			},
		},
		{
			name:           "request without frontend origin set",
			frontendOrigin: "",
			requestMethod:  "GET",
			requestHeaders: map[string]string{},
			expectedOrigin: "",
			expectedStatus: 200,
			expectedHeaders: map[string]string{
				"Access-Control-Allow-Methods": "GET,POST,PUT,DELETE,OPTIONS",
				"Access-Control-Allow-Headers": "Content-Type,Authorization,X-Requested-With",
				"Access-Control-Max-Age":       "86400",
			},
		},
		{
			name:           "POST request with frontend origin",
			frontendOrigin: "https://app.example.com",
			requestMethod:  "POST",
			requestHeaders: map[string]string{
				"Origin":        "https://app.example.com",
				"Content-Type":  "application/json",
				"Authorization": "Bearer token123",
			},
			expectedOrigin: "https://app.example.com",
			expectedStatus: 200,
			expectedHeaders: map[string]string{
				"Access-Control-Allow-Origin":  "https://app.example.com",
				"Access-Control-Allow-Methods": "GET,POST,PUT,DELETE,OPTIONS",
				"Access-Control-Allow-Headers": "Content-Type,Authorization,X-Requested-With",
				"Access-Control-Max-Age":       "86400",
				"Vary":                         "Origin",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// ダミーハンドラーを作成
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(200)
				_, _ = w.Write([]byte("OK"))
			})

			// CORSミドルウェアを適用
			corsHandler := CORSMiddleware(tt.frontendOrigin)(handler)

			// テスト用のリクエストを作成
			req := httptest.NewRequest(tt.requestMethod, "/test", nil)
			for key, value := range tt.requestHeaders {
				req.Header.Set(key, value)
			}

			// レスポンスレコーダーを作成
			w := httptest.NewRecorder()

			// リクエストを実行
			corsHandler.ServeHTTP(w, req)

			// ステータスコードを確認
			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status code %d, got %d", tt.expectedStatus, w.Code)
			}

			// 期待されるヘッダーを確認
			for expectedHeader, expectedValue := range tt.expectedHeaders {
				actualValue := w.Header().Get(expectedHeader)
				if actualValue != expectedValue {
					t.Errorf("Expected header %s: %q, got %q", expectedHeader, expectedValue, actualValue)
				}
			}

			// Access-Control-Allow-Originが期待通りかチェック
			actualOrigin := w.Header().Get("Access-Control-Allow-Origin")
			if actualOrigin != tt.expectedOrigin {
				t.Errorf("Expected Access-Control-Allow-Origin: %q, got %q", tt.expectedOrigin, actualOrigin)
			}
		})
	}
}

func TestLoggingMiddleware(t *testing.T) {
	tests := []struct {
		name           string
		handlerStatus  int
		expectedStatus int
		method         string
		path           string
	}{
		{
			name:           "successful GET request",
			handlerStatus:  200,
			expectedStatus: 200,
			method:         "GET",
			path:           "/test",
		},
		{
			name:           "not found GET request",
			handlerStatus:  404,
			expectedStatus: 404,
			method:         "GET",
			path:           "/notfound",
		},
		{
			name:           "internal server error POST request",
			handlerStatus:  500,
			expectedStatus: 500,
			method:         "POST",
			path:           "/error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// ダミーハンドラーを作成
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.handlerStatus)
				_, _ = w.Write([]byte("response"))
			})

			// ロギングミドルウェアを適用
			loggingHandler := LoggingMiddleware(handler)

			// テスト用のリクエストを作成
			req := httptest.NewRequest(tt.method, tt.path, nil)
			req.RemoteAddr = "127.0.0.1:12345"

			// レスポンスレコーダーを作成
			w := httptest.NewRecorder()

			// リクエストを実行
			loggingHandler.ServeHTTP(w, req)

			// ステータスコードを確認
			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status code %d, got %d", tt.expectedStatus, w.Code)
			}

			// レスポンスボディを確認
			expectedBody := "response"
			if w.Body.String() != expectedBody {
				t.Errorf("Expected body %q, got %q", expectedBody, w.Body.String())
			}
		})
	}
}

func TestCollectorMiddleware_InjectsCollector(t *testing.T) {
	var capturedCollector interface{}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedCollector = r.Context().Value(collectorCtxKey{})
		w.WriteHeader(200)
	})

	mw := CollectorMiddleware(handler)
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	mw.ServeHTTP(w, req)

	if capturedCollector == nil {
		t.Fatal("Expected collector to be injected into context, got nil")
	}
}

func TestCollectorFromRequest_ReturnsCollector(t *testing.T) {
	var got interface{}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = collectorFromRequest(r)
		w.WriteHeader(200)
	})

	mw := CollectorMiddleware(handler)
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	mw.ServeHTTP(w, req)

	if got == nil {
		t.Fatal("Expected collectorFromRequest to return non-nil collector")
	}
}

func TestCollectorFromRequest_NilWithoutMiddleware(t *testing.T) {
	req := httptest.NewRequest("GET", "/test", nil)
	c := collectorFromRequest(req)
	if c != nil {
		t.Errorf("Expected nil when middleware not applied, got %v", c)
	}
}

func TestRecoverMiddleware_PanicReturns500WithLogs(t *testing.T) {
	// CollectorMiddleware → RecoverMiddleware の順で積む
	panicHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("test panic message")
	})

	// CollectorMiddleware を先に適用してから RecoverMiddleware を適用
	chain := CollectorMiddleware(RecoverMiddleware(panicHandler))

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	chain.ServeHTTP(w, req)

	if w.Code != StatusInternalServerError {
		t.Errorf("Expected 500, got %d", w.Code)
	}

	var resp ErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to unmarshal panic response: %v", err)
	}

	if resp.HTTPCode != StatusInternalServerError {
		t.Errorf("Expected httpCode 500, got %d", resp.HTTPCode)
	}

	if !strings.Contains(resp.Message, "test panic message") {
		t.Errorf("Expected message to contain panic value, got %q", resp.Message)
	}

	// PANIC log entry が含まれること
	if len(resp.Logs) == 0 {
		t.Fatal("Expected logs to be non-empty after panic recovery")
	}

	foundPanic := false
	for _, l := range resp.Logs {
		if l.Source == "PANIC" {
			foundPanic = true
			break
		}
	}
	if !foundPanic {
		t.Errorf("Expected PANIC source in logs, got %+v", resp.Logs)
	}
}

func TestRecoverMiddleware_NoPanicPassesThrough(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_, _ = w.Write([]byte("ok"))
	})

	chain := CollectorMiddleware(RecoverMiddleware(handler))

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	chain.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("Expected 200, got %d", w.Code)
	}
	if w.Body.String() != "ok" {
		t.Errorf("Expected body 'ok', got %q", w.Body.String())
	}
}
