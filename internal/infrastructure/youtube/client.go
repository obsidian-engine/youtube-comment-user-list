// Package youtube YouTube APIクライアントの実装を提供します
package youtube

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/obsidian-engine/youtube-comment-user-list/internal/constants"
	"github.com/obsidian-engine/youtube-comment-user-list/internal/domain/entity"
)

// Client YouTubeClientインターフェースを実装します
type Client struct {
	apiKey     string
	httpClient *http.Client
}

// NewClient 新しいYouTubeを作成します API client
func NewClient(apiKey string) *Client {
	return &Client{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: constants.YouTubeHTTPClientTimeout,
		},
	}
}

// APIレスポンス構造体
type videosListResponse struct {
	Items []struct {
		ID                   string               `json:"id"`
		Snippet              videoSnippet         `json:"snippet"`
		LiveStreamingDetails liveStreamingDetails `json:"liveStreamingDetails"`
	} `json:"items"`
	Error *apiError `json:"error,omitempty"`
}

type videoSnippet struct {
	Title                string `json:"title"`
	ChannelTitle         string `json:"channelTitle"`
	LiveBroadcastContent string `json:"liveBroadcastContent"`
}

type liveStreamingDetails struct {
	ActiveLiveChatID   string `json:"activeLiveChatId"`
	ActualStartTime    string `json:"actualStartTime"`
	ActualEndTime      string `json:"actualEndTime"`
	ConcurrentViewers  string `json:"concurrentViewers"`
	ScheduledStartTime string `json:"scheduledStartTime"`
}

type liveChatResponse struct {
	Items                 []liveChatMessage `json:"items"`
	NextPageToken         string            `json:"nextPageToken"`
	PollingIntervalMillis int               `json:"pollingIntervalMillis"`
	Error                 *apiError         `json:"error,omitempty"`
}

type liveChatMessage struct {
	ID            string        `json:"id"`
	AuthorDetails authorDetails `json:"authorDetails"`
}

type authorDetails struct {
	DisplayName string `json:"displayName"`
	ChannelID   string `json:"channelId"`
	IsChatOwner bool   `json:"isChatOwner"`
	IsModerator bool   `json:"isChatModerator"`
	IsMember    bool   `json:"isChatSponsor"`
}

type apiError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Status  string `json:"status"`
}

// FetchVideoInfo ライブ配信詳細を含む動画情報を取得します
func (c *Client) FetchVideoInfo(ctx context.Context, videoID string) (*entity.VideoInfo, error) {
	apiURL := fmt.Sprintf("https://www.googleapis.com/youtube/v3/videos?part=snippet,liveStreamingDetails&id=%s&key=%s", videoID, c.apiKey)

	// URL検証
	if _, err := url.ParseRequestURI(apiURL); err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer func() {
		if err2 := resp.Body.Close(); err2 != nil {
			// 重要ではないためエラーをログに記録しますが返却しません
		}
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != constants.HTTPStatusOK {
		return nil, fmt.Errorf("videos.list API error: status=%d, body=%s", resp.StatusCode, string(body))
	}

	var response videosListResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("JSON parse failed: %w", err)
	}

	// APIエラーレスポンスを確認
	if response.Error != nil {
		return nil, fmt.Errorf("YouTube API error: %d %s - %s", response.Error.Code, response.Error.Status, response.Error.Message)
	}

	if len(response.Items) == 0 {
		return nil, fmt.Errorf("video not found (video ID: %s)", videoID)
	}

	item := response.Items[0]

	// ドメインエンティティに変換
	videoInfo := &entity.VideoInfo{
		ID:                   item.ID,
		Title:                item.Snippet.Title,
		ChannelTitle:         item.Snippet.ChannelTitle,
		LiveBroadcastContent: item.Snippet.LiveBroadcastContent,
		LiveStreamingDetails: entity.LiveStreamingDetails{
			ActiveLiveChatID:   item.LiveStreamingDetails.ActiveLiveChatID,
			ActualStartTime:    item.LiveStreamingDetails.ActualStartTime,
			ActualEndTime:      item.LiveStreamingDetails.ActualEndTime,
			ConcurrentViewers:  item.LiveStreamingDetails.ConcurrentViewers,
			ScheduledStartTime: item.LiveStreamingDetails.ScheduledStartTime,
		},
	}

	return videoInfo, nil
}

// FetchLiveChat ライブチャットメッセージを取得します
func (c *Client) FetchLiveChat(ctx context.Context, liveChatID string, pageToken string) (*entity.PollResult, error) {
	base := "https://www.googleapis.com/youtube/v3/liveChat/messages"
	params := []string{
		"part=authorDetails",
		fmt.Sprintf("maxResults=%d", constants.YouTubeChatMaxResults),
		"liveChatId=" + liveChatID,
		"key=" + c.apiKey,
	}
	if pageToken != "" {
		params = append(params, "pageToken="+pageToken)
	}
	apiURL := base + "?" + strings.Join(params, "&")

	// URL検証
	if _, err := url.ParseRequestURI(apiURL); err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("LiveChat HTTP request failed: %w", err)
	}
	defer func() {
		if err2 := resp.Body.Close(); err2 != nil {
			// 重要ではないためエラーをログに記録しますが返却しません
		}
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("LiveChat response read failed: %w", err)
	}

	if resp.StatusCode != constants.HTTPStatusOK {
		return nil, fmt.Errorf("liveChatMessages.list API error: status=%d, body=%s", resp.StatusCode, string(body))
	}

	var response liveChatResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("LiveChat JSON parse failed: %w", err)
	}

	// APIエラーレスポンスを確認
	if response.Error != nil {
		return nil, fmt.Errorf("LiveChat API error: %d %s - %s", response.Error.Code, response.Error.Status, response.Error.Message)
	}

	// ドメインエンティティに変換
	messages := make([]entity.ChatMessage, len(response.Items))
	for i, item := range response.Items {
		messages[i] = entity.ChatMessage{
			ID: item.ID,
			AuthorDetails: entity.AuthorDetails{
				DisplayName: item.AuthorDetails.DisplayName,
				ChannelID:   item.AuthorDetails.ChannelID,
				IsChatOwner: item.AuthorDetails.IsChatOwner,
				IsModerator: item.AuthorDetails.IsModerator,
				IsMember:    item.AuthorDetails.IsMember,
			},
			Timestamp: time.Now(), // YouTube APIは常にタイムスタンプを提供するとは限らないため、現在時刻を使用
		}
	}

	pollResult := &entity.PollResult{
		Messages:          messages,
		NextPageToken:     response.NextPageToken,
		PollingIntervalMs: response.PollingIntervalMillis,
		Success:           true,
		Error:             nil,
	}

	return pollResult, nil
}
