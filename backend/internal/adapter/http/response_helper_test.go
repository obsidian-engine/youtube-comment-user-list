package http

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRenderJSONResponse_Success(t *testing.T) {
	tests := []struct {
		name       string
		status     int
		data       interface{}
		wantStatus int
		wantBody   string
	}{
		{
			name:       "OK with simple data",
			status:     http.StatusOK,
			data:       map[string]string{"message": "success"},
			wantStatus: http.StatusOK,
			wantBody:   `{"message":"success"}`,
		},
		{
			name:       "Created with struct data",
			status:     http.StatusCreated,
			data:       struct{ ID int `json:"id"` }{ID: 123},
			wantStatus: http.StatusCreated,
			wantBody:   `{"id":123}`,
		},
		{
			name:       "Bad Request with error message",
			status:     http.StatusBadRequest,
			data:       map[string]string{"error": "invalid input"},
			wantStatus: http.StatusBadRequest,
			wantBody:   `{"error":"invalid input"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", "/", nil)

			RenderJSONResponse(w, r, tt.status, tt.data)

			if w.Code != tt.wantStatus {
				t.Errorf("RenderJSONResponse() status = %v, want %v", w.Code, tt.wantStatus)
			}

			// JSONを正規化して比較
			var gotData, wantData interface{}
			if err := json.Unmarshal(w.Body.Bytes(), &gotData); err != nil {
				t.Fatalf("Failed to unmarshal response body: %v", err)
			}
			if err := json.Unmarshal([]byte(tt.wantBody), &wantData); err != nil {
				t.Fatalf("Failed to unmarshal expected body: %v", err)
			}

			gotJSON, _ := json.Marshal(gotData)
			wantJSON, _ := json.Marshal(wantData)

			if string(gotJSON) != string(wantJSON) {
				t.Errorf("RenderJSONResponse() body = %v, want %v", string(gotJSON), string(wantJSON))
			}
		})
	}
}

func TestRenderErrorResponse_WithErrorConfig(t *testing.T) {
	tests := []struct {
		name        string
		errorType   string
		message     string
		wantStatus  int
		wantContain []string
	}{
		{
			name:        "Bad Request error",
			errorType:   "bad_request",
			message:     "Invalid input data",
			wantStatus:  http.StatusBadRequest,
			wantContain: []string{"bad_request", "Invalid input data", "400"},
		},
		{
			name:        "Internal Server error",
			errorType:   "internal_error",
			message:     "Database connection failed",
			wantStatus:  http.StatusInternalServerError,
			wantContain: []string{"internal_error", "Database connection failed", "500"},
		},
		{
			name:        "Not Found error",
			errorType:   "not_found",
			message:     "Resource not found",
			wantStatus:  http.StatusNotFound,
			wantContain: []string{"not_found", "Resource not found", "404"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", "/", nil)

			RenderErrorResponse(w, r, tt.errorType, tt.message, tt.wantStatus)

			if w.Code != tt.wantStatus {
				t.Errorf("RenderErrorResponse() status = %v, want %v", w.Code, tt.wantStatus)
			}

			body := w.Body.String()
			for _, contain := range tt.wantContain {
				if !strings.Contains(body, contain) {
					t.Errorf("RenderErrorResponse() body should contain %q, got: %s", contain, body)
				}
			}

			// Content-Type ヘッダーの確認
			contentType := w.Header().Get("Content-Type")
			if !strings.Contains(contentType, "application/json") {
				t.Errorf("RenderErrorResponse() Content-Type = %v, want application/json", contentType)
			}
		})
	}
}

func TestRenderSuccessResponse_WithStatusAndData(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/", nil)

	data := map[string]interface{}{
		"id":      123,
		"message": "User created successfully",
		"active":  true,
	}

	RenderSuccessResponse(w, r, http.StatusCreated, data)

	if w.Code != http.StatusCreated {
		t.Errorf("RenderSuccessResponse() status = %v, want %v", w.Code, http.StatusCreated)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response["id"].(float64) != 123 {
		t.Errorf("Expected id=123, got %v", response["id"])
	}
	if response["message"] != "User created successfully" {
		t.Errorf("Expected message='User created successfully', got %v", response["message"])
	}
	if response["active"] != true {
		t.Errorf("Expected active=true, got %v", response["active"])
	}
}

// 既存のrender関数との互換性テスト
func TestBackwardCompatibility_WithExistingRenderFunctions(t *testing.T) {
	t.Run("renderInternalError compatibility", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/", nil)

		// 既存の関数を呼び出し
		renderInternalError(w, r, "Test error message")

		if w.Code != StatusInternalServerError {
			t.Errorf("renderInternalError() status = %v, want %v", w.Code, StatusInternalServerError)
		}
	})

	t.Run("renderBadRequest compatibility", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/", nil)

		renderBadRequest(w, r, "Bad request message")

		if w.Code != StatusBadRequest {
			t.Errorf("renderBadRequest() status = %v, want %v", w.Code, StatusBadRequest)
		}
	})
}