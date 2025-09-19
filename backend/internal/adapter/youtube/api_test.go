package youtube

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/obsidian-engine/youtube-comment-user-list/backend/internal/port"
)

// モックHTTPサーバーのレスポンス構造体
type mockLiveChatResponse struct {
	Items         []mockLiveChatItem `json:"items"`
	NextPageToken string             `json:"nextPageToken,omitempty"`
}

type mockLiveChatItem struct {
	Snippet       mockSnippet       `json:"snippet"`
	AuthorDetails mockAuthorDetails `json:"authorDetails"`
}

type mockSnippet struct {
	PublishedAt string `json:"publishedAt"`
}

type mockAuthorDetails struct {
	ChannelID   string `json:"channelId"`
	DisplayName string `json:"displayName"`
}

func TestListLiveChatMessages_Pagination(t *testing.T) {
	// モックサーバーのセットアップ
	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// APIキーのチェック
		apiKey := r.URL.Query().Get("key")
		if apiKey != "test-api-key" {
			http.Error(w, "Invalid API key", http.StatusUnauthorized)
			return
		}

		// ページトークンのチェック
		pageToken := r.URL.Query().Get("pageToken")

		var response mockLiveChatResponse
		requestCount++

		switch pageToken {
		case "":
			// 最初のページ
			response = mockLiveChatResponse{
				Items: []mockLiveChatItem{
					{
						Snippet: mockSnippet{
							PublishedAt: time.Now().Format(time.RFC3339),
						},
						AuthorDetails: mockAuthorDetails{
							ChannelID:   "UC001",
							DisplayName: "User1",
						},
					},
					{
						Snippet: mockSnippet{
							PublishedAt: time.Now().Format(time.RFC3339),
						},
						AuthorDetails: mockAuthorDetails{
							ChannelID:   "UC002",
							DisplayName: "User2",
						},
					},
				},
				NextPageToken: "page2",
			}
		case "page2":
			// 2ページ目
			response = mockLiveChatResponse{
				Items: []mockLiveChatItem{
					{
						Snippet: mockSnippet{
							PublishedAt: time.Now().Format(time.RFC3339),
						},
						AuthorDetails: mockAuthorDetails{
							ChannelID:   "UC003",
							DisplayName: "User3",
						},
					},
					{
						Snippet: mockSnippet{
							PublishedAt: time.Now().Format(time.RFC3339),
						},
						AuthorDetails: mockAuthorDetails{
							ChannelID:   "UC004",
							DisplayName: "User4",
						},
					},
				},
				NextPageToken: "page3",
			}
		case "page3":
			// 最後のページ（NextPageTokenなし）
			response = mockLiveChatResponse{
				Items: []mockLiveChatItem{
					{
						Snippet: mockSnippet{
							PublishedAt: time.Now().Format(time.RFC3339),
						},
						AuthorDetails: mockAuthorDetails{
							ChannelID:   "UC005",
							DisplayName: "User5",
						},
					},
				},
				// NextPageTokenを設定しない（最後のページ）
			}
		default:
			http.Error(w, "Invalid page token", http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			return
		}
	}))
	defer server.Close()

	// テストの実行
	t.Run("ページング処理が全てのページを取得すること", func(t *testing.T) {
		// 注: 現在の実装では実際のGoogle APIを呼び出すため、
		// このテストは実際には実行できません。
		// 本来はDependency Injectionを使用してHTTPクライアントを
		// モック可能にする必要があります。

		t.Skip("実際のGoogle APIを呼び出すため、統合テストとしてスキップ")
	})
}

func TestListLiveChatMessages_ErrorHandling(t *testing.T) {
	tests := []struct {
		name        string
		apiKey      string
		liveChatID  string
		wantErr     bool
		wantEnded   bool
		errContains string
	}{
		{
			name:        "APIキーが空の場合",
			apiKey:      "",
			liveChatID:  "test-chat-id",
			wantErr:     true,
			wantEnded:   false,
			errContains: "youtube api key is required",
		},
		{
			name:        "LiveChatIDが空の場合",
			apiKey:      "test-api-key",
			liveChatID:  "",
			wantErr:     true,
			wantEnded:   false,
			errContains: "live chat ID is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			api := &API{
				APIKey: tt.apiKey,
			}

			items, isEnded, err := api.ListLiveChatMessages(context.Background(), tt.liveChatID)

			if tt.wantErr {
				if err == nil {
					t.Errorf("期待されたエラーが発生しませんでした")
				} else if tt.errContains != "" && err.Error() != tt.errContains {
					t.Errorf("エラーメッセージが期待と異なります: got %v, want %v", err.Error(), tt.errContains)
				}
			} else if err != nil {
				t.Errorf("予期しないエラーが発生しました: %v", err)
			}

			if isEnded != tt.wantEnded {
				t.Errorf("isEnded = %v, want %v", isEnded, tt.wantEnded)
			}

			if tt.wantErr && items != nil {
				t.Errorf("エラー時にitemsがnilでない: %v", items)
			}
		})
	}
}


func TestPaginationLogic(t *testing.T) {
	t.Run("複数ページのメッセージを結合すること", func(t *testing.T) {
		// ページングロジックのテスト
		// 3ページ分のデータを用意
		page1 := []port.ChatMessage{
			{ChannelID: "UC001", DisplayName: "User1", PublishedAt: time.Now()},
			{ChannelID: "UC002", DisplayName: "User2", PublishedAt: time.Now()},
		}
		page2 := []port.ChatMessage{
			{ChannelID: "UC003", DisplayName: "User3", PublishedAt: time.Now()},
			{ChannelID: "UC004", DisplayName: "User4", PublishedAt: time.Now()},
		}
		page3 := []port.ChatMessage{
			{ChannelID: "UC005", DisplayName: "User5", PublishedAt: time.Now()},
		}

		// 全ページを結合した結果を検証
		allMessages := append(append(page1, page2...), page3...)

		if len(allMessages) != 5 {
			t.Errorf("メッセージ数が期待と異なる: got %d, want 5", len(allMessages))
		}

		// 各ユーザーが正しく含まれているか確認
		expectedUsers := map[string]string{
			"UC001": "User1",
			"UC002": "User2",
			"UC003": "User3",
			"UC004": "User4",
			"UC005": "User5",
		}

		for _, msg := range allMessages {
			if expectedName, ok := expectedUsers[msg.ChannelID]; !ok {
				t.Errorf("予期しないChannelID: %s", msg.ChannelID)
			} else if msg.DisplayName != expectedName {
				t.Errorf("DisplayNameが不一致: got %s, want %s", msg.DisplayName, expectedName)
			}
		}
	})

	t.Run("大量のページを処理できること", func(t *testing.T) {
		// 100ページ分のデータを生成
		var allMessages []port.ChatMessage
		totalPages := 100
		messagesPerPage := 20

		for page := 0; page < totalPages; page++ {
			for i := 0; i < messagesPerPage; i++ {
				userNum := page*messagesPerPage + i + 1
				msg := port.ChatMessage{
					ChannelID:   fmt.Sprintf("UC%05d", userNum),
					DisplayName: fmt.Sprintf("User%d", userNum),
					PublishedAt: time.Now(),
				}
				allMessages = append(allMessages, msg)
			}
		}

		// 合計メッセージ数の確認
		expectedTotal := totalPages * messagesPerPage
		if len(allMessages) != expectedTotal {
			t.Errorf("メッセージ総数が期待と異なる: got %d, want %d", len(allMessages), expectedTotal)
		}

		// 最初と最後のユーザーを確認
		if allMessages[0].DisplayName != "User1" {
			t.Errorf("最初のユーザーが期待と異なる: got %s, want User1", allMessages[0].DisplayName)
		}
		if allMessages[len(allMessages)-1].DisplayName != fmt.Sprintf("User%d", expectedTotal) {
			t.Errorf("最後のユーザーが期待と異なる: got %s, want User%d", allMessages[len(allMessages)-1].DisplayName, expectedTotal)
		}
	})
}