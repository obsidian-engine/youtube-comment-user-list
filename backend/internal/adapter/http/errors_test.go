package http

import (
	"encoding/json"
	"net/http/httptest"
	"testing"
)

func TestRenderError(t *testing.T) {
	tests := []struct {
		name     string
		code     int
		err      string
		message  string
		expected ErrorResponse
	}{
		{
			name:    "internal server error",
			code:    500,
			err:     "internal_error",
			message: "Something went wrong",
			expected: ErrorResponse{
				Error:   "internal_error",
				Message: "Something went wrong",
				Code:    500,
			},
		},
		{
			name:    "bad request",
			code:    400,
			err:     "bad_request",
			message: "Invalid input",
			expected: ErrorResponse{
				Error:   "bad_request",
				Message: "Invalid input",
				Code:    400,
			},
		},
		{
			name:    "empty message",
			code:    404,
			err:     "not_found",
			message: "",
			expected: ErrorResponse{
				Error:   "not_found",
				Message: "",
				Code:    404,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", "/", nil)

			renderError(w, r, tt.code, tt.err, tt.message)

			if w.Code != tt.code {
				t.Errorf("Expected status code %d, got %d", tt.code, w.Code)
			}

			var response ErrorResponse
			if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
				t.Fatalf("Failed to unmarshal response: %v", err)
			}

			if response.Error != tt.expected.Error {
				t.Errorf("Expected error %q, got %q", tt.expected.Error, response.Error)
			}
			if response.Message != tt.expected.Message {
				t.Errorf("Expected message %q, got %q", tt.expected.Message, response.Message)
			}
			if response.Code != tt.expected.Code {
				t.Errorf("Expected code %d, got %d", tt.expected.Code, response.Code)
			}
		})
	}
}

func TestRenderInternalError(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)

	renderInternalError(w, r, "Test internal error")

	if w.Code != StatusInternalServerError {
		t.Errorf("Expected status code %d, got %d", StatusInternalServerError, w.Code)
	}

	var response ErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	expected := ErrorResponse{
		Error:   "internal_error",
		Message: "Test internal error",
		Code:    StatusInternalServerError,
	}

	if response.Error != expected.Error {
		t.Errorf("Expected error %q, got %q", expected.Error, response.Error)
	}
	if response.Message != expected.Message {
		t.Errorf("Expected message %q, got %q", expected.Message, response.Message)
	}
	if response.Code != expected.Code {
		t.Errorf("Expected code %d, got %d", expected.Code, response.Code)
	}
}

func TestRenderBadRequest(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)

	renderBadRequest(w, r, "Test bad request")

	if w.Code != StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", StatusBadRequest, w.Code)
	}

	var response ErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	expected := ErrorResponse{
		Error:   "bad_request",
		Message: "Test bad request",
		Code:    StatusBadRequest,
	}

	if response.Error != expected.Error {
		t.Errorf("Expected error %q, got %q", expected.Error, response.Error)
	}
	if response.Message != expected.Message {
		t.Errorf("Expected message %q, got %q", expected.Message, response.Message)
	}
	if response.Code != expected.Code {
		t.Errorf("Expected code %d, got %d", expected.Code, response.Code)
	}
}

func TestRenderBadGateway(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)

	renderBadGateway(w, r, "Test bad gateway")

	if w.Code != StatusBadGateway {
		t.Errorf("Expected status code %d, got %d", StatusBadGateway, w.Code)
	}

	var response ErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	expected := ErrorResponse{
		Error:   "bad_gateway",
		Message: "Test bad gateway",
		Code:    StatusBadGateway,
	}

	if response.Error != expected.Error {
		t.Errorf("Expected error %q, got %q", expected.Error, response.Error)
	}
	if response.Message != expected.Message {
		t.Errorf("Expected message %q, got %q", expected.Message, response.Message)
	}
	if response.Code != expected.Code {
		t.Errorf("Expected code %d, got %d", expected.Code, response.Code)
	}
}

func TestRenderErrorWithConfig(t *testing.T) {
	tests := []struct {
		name     string
		config   ErrorConfig
		expected ErrorResponse
	}{
		{
			name: "internal server error with config",
			config: ErrorConfig{
				Code:    500,
				Error:   "internal_error",
				Message: "Something went wrong",
			},
			expected: ErrorResponse{
				Error:   "internal_error",
				Message: "Something went wrong",
				Code:    500,
			},
		},
		{
			name: "bad request with config",
			config: ErrorConfig{
				Code:    400,
				Error:   "bad_request",
				Message: "Invalid input",
			},
			expected: ErrorResponse{
				Error:   "bad_request",
				Message: "Invalid input",
				Code:    400,
			},
		},
		{
			name: "empty message with config",
			config: ErrorConfig{
				Code:    404,
				Error:   "not_found",
				Message: "",
			},
			expected: ErrorResponse{
				Error:   "not_found",
				Message: "",
				Code:    404,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", "/", nil)
			
			// テスト用にResponseWriterとRequestを設定
			tt.config.ResponseWriter = w
			tt.config.Request = r

			renderErrorWithConfig(tt.config)

			if w.Code != tt.config.Code {
				t.Errorf("Expected status code %d, got %d", tt.config.Code, w.Code)
			}

			var response ErrorResponse
			if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
				t.Fatalf("Failed to unmarshal response: %v", err)
			}

			if response.Error != tt.expected.Error {
				t.Errorf("Expected error %q, got %q", tt.expected.Error, response.Error)
			}
			if response.Message != tt.expected.Message {
				t.Errorf("Expected message %q, got %q", tt.expected.Message, response.Message)
			}
			if response.Code != tt.expected.Code {
				t.Errorf("Expected code %d, got %d", tt.expected.Code, response.Code)
			}
		})
	}
}

// 既存のAPIと新しいAPIが同じ結果を返すことを確認するテスト
func TestRenderErrorBackwardCompatibility(t *testing.T) {
	tests := []struct {
		name    string
		code    int
		err     string
		message string
	}{
		{
			name:    "internal server error comparison",
			code:    500,
			err:     "internal_error",
			message: "Something went wrong",
		},
		{
			name:    "bad request comparison",
			code:    400,
			err:     "bad_request",
			message: "Invalid input",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 既存のAPI
			w1 := httptest.NewRecorder()
			r1 := httptest.NewRequest("GET", "/", nil)
			renderError(w1, r1, tt.code, tt.err, tt.message)

			// 新しいAPI
			w2 := httptest.NewRecorder()
			r2 := httptest.NewRequest("GET", "/", nil)
			renderErrorWithConfig(ErrorConfig{
				ResponseWriter: w2,
				Request:        r2,
				Code:           tt.code,
				Error:          tt.err,
				Message:        tt.message,
			})

			// 結果が同じかチェック
			if w1.Code != w2.Code {
				t.Errorf("Status codes differ: old=%d, new=%d", w1.Code, w2.Code)
			}

			var response1, response2 ErrorResponse
			if err := json.Unmarshal(w1.Body.Bytes(), &response1); err != nil {
				t.Fatalf("Failed to unmarshal old response: %v", err)
			}
			if err := json.Unmarshal(w2.Body.Bytes(), &response2); err != nil {
				t.Fatalf("Failed to unmarshal new response: %v", err)
			}

			if response1.Error != response2.Error {
				t.Errorf("Error fields differ: old=%q, new=%q", response1.Error, response2.Error)
			}
			if response1.Message != response2.Message {
				t.Errorf("Message fields differ: old=%q, new=%q", response1.Message, response2.Message)
			}
			if response1.Code != response2.Code {
				t.Errorf("Code fields differ: old=%d, new=%d", response1.Code, response2.Code)
			}
		})
	}
}