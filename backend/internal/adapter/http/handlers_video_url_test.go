package http

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/obsidian-engine/youtube-comment-user-list/backend/internal/adapter/memory"
	"github.com/obsidian-engine/youtube-comment-user-list/backend/internal/domain"
	"github.com/obsidian-engine/youtube-comment-user-list/backend/internal/port"
	"github.com/obsidian-engine/youtube-comment-user-list/backend/internal/usecase"
)

type fakeYTForURL struct{}

func (f *fakeYTForURL) GetActiveLiveChatID(ctx context.Context, videoID string) (string, error) {
	return "live:chat:" + videoID, nil
}

func (f *fakeYTForURL) ListLiveChatMessages(ctx context.Context, liveChatID string, pageToken string) ([]port.ChatMessage, string, bool, error) {
	return nil, "", false, nil
}

type fakeClockForURL struct{}

func (f *fakeClockForURL) Now() time.Time {
	return time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
}

func TestSwitchVideoWithURL(t *testing.T) {
	// セットアップ
	users := memory.NewUserRepo()
	state := memory.NewStateRepo()
	yt := &fakeYTForURL{}

	clock := &fakeClockForURL{}
	handlers := &Handlers{
		Users:       users,
		SwitchVideo: &usecase.SwitchVideo{YT: yt, Users: users, State: state, Clock: clock},
	}

	router := NewRouter(handlers, "*")

	tests := []struct {
		name           string
		input          string
		expectedVideoID string
		shouldSucceed  bool
	}{
		{
			name:           "ライブチャットURL",
			input:          "https://www.youtube.com/live_chat?is_popout=1&v=Qw3tyIFqKrg",
			expectedVideoID: "Qw3tyIFqKrg",
			shouldSucceed:  true,
		},
		{
			name:           "通常の動画URL",
			input:          "https://www.youtube.com/watch?v=Qw3tyIFqKrg",
			expectedVideoID: "Qw3tyIFqKrg",
			shouldSucceed:  true,
		},
		{
			name:           "短縮URL",
			input:          "https://youtu.be/Qw3tyIFqKrg",
			expectedVideoID: "Qw3tyIFqKrg",
			shouldSucceed:  true,
		},
		{
			name:           "video_idのみ（既存動作）",
			input:          "Qw3tyIFqKrg",
			expectedVideoID: "Qw3tyIFqKrg",
			shouldSucceed:  true,
		},
		{
			name:          "無効なURL",
			input:         "invalid-url",
			shouldSucceed: false,
		},
		{
			name:          "YouTube以外のURL",
			input:         "https://example.com/watch?v=Qw3tyIFqKrg",
			shouldSucceed: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// リクエスト作成
			reqBody := map[string]string{
				"videoId": tt.input,
			}
			bodyBytes, _ := json.Marshal(reqBody)
			req := httptest.NewRequest("POST", "/switch-video", bytes.NewBuffer(bodyBytes))
			req.Header.Set("Content-Type", "application/json")

			// レスポンス記録
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if tt.shouldSucceed {
				if w.Code != http.StatusOK {
					t.Errorf("期待するステータス: 200, 実際: %d, レスポンス: %s", w.Code, w.Body.String())
					return
				}

				// レスポンスの確認
				var response map[string]interface{}
				if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
					t.Errorf("レスポンスのJSONパースエラー: %v", err)
					return
				}

				if videoID, ok := response["videoId"].(string); !ok || videoID != tt.expectedVideoID {
					t.Errorf("期待するvideoId: %s, 実際: %v", tt.expectedVideoID, response["videoId"])
				}

				// StateにACTIVEが設定されているか確認
				currentState, _ := state.Get(context.Background())
				if currentState.Status != domain.StatusActive {
					t.Errorf("期待するStatus: %v, 実際: %v", domain.StatusActive, currentState.Status)
				}
				if currentState.VideoID != tt.expectedVideoID {
					t.Errorf("期待するVideoID: %s, 実際: %s", tt.expectedVideoID, currentState.VideoID)
				}
			} else {
				if w.Code != http.StatusBadRequest {
					t.Errorf("期待するステータス: 400, 実際: %d", w.Code)
				}
			}
		})
	}
}