// Package youtube YouTube APIクライアントの実装を提供します
package youtube

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"strconv"
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

// NewClient 新しいYouTube APIクライアントを作成します
func NewClient(apiKey string) *Client {
	// HTTPクライアントの設定を最適化
	transport := &http.Transport{
		MaxIdleConns:        10,
		MaxIdleConnsPerHost: 5,
		IdleConnTimeout:     90 * time.Second,
		DisableKeepAlives:   false,
	}

	return &Client{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout:   constants.YouTubeHTTPClientTimeout,
			Transport: transport,
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

// doRequestWithRetry 指数バックオフ＋ジッタ付きリトライロジックでHTTPリクエストを実行します
func (c *Client) doRequestWithRetry(ctx context.Context, url string) (*http.Response, error) {
	var lastErr error

	for attempt := 0; attempt < constants.YouTubeAPIMaxRetries; attempt++ {
		if attempt > 0 {
			// 指数バックオフ + ジッタで遅延計算
			delay := c.calculateRetryDelay(attempt)

			// リトライログ出力（構造化）
			// Note: 実際の本番環境では適切なロガーを注入する
			fmt.Printf("[RETRY] attempt %d/%d, delay: %v, last_error: %v\n", attempt+1, constants.YouTubeAPIMaxRetries, delay, lastErr)

			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(delay):
			}
		}

		// 各試行で個別のタイムアウトを設定
		attemptCtx, cancel := context.WithTimeout(ctx, 15*time.Second)

		req, err := http.NewRequestWithContext(attemptCtx, "GET", url, nil)
		if err != nil {
			cancel()
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		resp, err := c.httpClient.Do(req)
		cancel() // すぐにキャンセル

		if err != nil {
			lastErr = err
			// コンテキストキャンセルは即座に終了
			if ctx.Err() != nil {
				return nil, ctx.Err()
			}
			// 一時的なネットワークエラーやタイムアウトかチェック
			if c.isRetryableError(err) {
				continue // リトライ
			}
			return nil, fmt.Errorf("HTTP request failed: %w", err)
		}

		// レスポンスボディを読み取り（ステータス判定に必要）
		body, readErr := io.ReadAll(resp.Body)
		_ = resp.Body.Close() // エラーは無視（レスポンス処理済み）

		if readErr != nil {
			// コンテキストキャンセル/タイムアウトは即座に終了
			if ctx.Err() != nil || errors.Is(readErr, context.Canceled) || errors.Is(readErr, context.DeadlineExceeded) {
				return nil, ctx.Err()
			}
			lastErr = fmt.Errorf("failed to read response body: %w", readErr)
			continue
		}

		// ステータスコードによる詳細判定
		if c.isRetryableHTTPStatus(resp.StatusCode, body) {
			// レート制限の場合はRetry-Afterヘッダーを確認
			if resp.StatusCode == 429 || resp.StatusCode == 403 {
				retryAfter := c.parseRetryAfter(resp.Header.Get("Retry-After"))

				if retryAfter > 0 && retryAfter < 60*time.Second {
					select {
					case <-ctx.Done():
						return nil, ctx.Err()
					case <-time.After(retryAfter):
					}
				}
			}
			lastErr = fmt.Errorf("retryable HTTP error: status %d, body: %s", resp.StatusCode, string(body))
			continue
		}

		// 成功またはリトライ不可能なエラー
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			// 成功時はボディを再構築してレスポンスを返す
			resp.Body = io.NopCloser(strings.NewReader(string(body)))
			return resp, nil
		}

		// リトライ不可能なクライアントエラー
		return nil, fmt.Errorf("non-retryable HTTP error: status %d, body: %s", resp.StatusCode, string(body))
	}

	return nil, fmt.Errorf("max retries exceeded, last error: %w", lastErr)
}

// calculateRetryDelay 指数バックオフ + ジッタで遅延時間を計算します
func (c *Client) calculateRetryDelay(attempt int) time.Duration {
	base := constants.YouTubeAPIInitialRetryDelay

	// 指数バックオフ
	for i := 1; i < attempt; i++ {
		base = base * constants.YouTubeAPIRetryMultiplier
		if base > constants.YouTubeAPIMaxRetryDelay {
			base = constants.YouTubeAPIMaxRetryDelay
			break
		}
	}

	// ジッタを追加 (±25%) - リトライ用なので弱い乱数で十分
	jitter := float64(base) * 0.25 * (rand.Float64()*2 - 1) //nolint:gosec
	finalDelay := time.Duration(float64(base) + jitter)

	if finalDelay < 0 {
		finalDelay = constants.YouTubeAPIInitialRetryDelay
	}
	if finalDelay > constants.YouTubeAPIMaxRetryDelay {
		finalDelay = constants.YouTubeAPIMaxRetryDelay
	}

	return finalDelay
}

// parseRetryAfter Retry-Afterヘッダーをパースします
func (c *Client) parseRetryAfter(header string) time.Duration {
	if header == "" {
		return 0
	}

	// 秒数での指定
	if seconds, err := strconv.Atoi(header); err == nil {
		return time.Duration(seconds) * time.Second
	}

	return 0
}

// isRetryableError エラーがリトライ可能かどうかを判定します
func (c *Client) isRetryableError(err error) bool {
	if err == nil {
		return false
	}

	// コンテキストエラーはリトライしない
	if err == context.Canceled || err == context.DeadlineExceeded {
		return false
	}

	// ネットワークエラー
	if netErr, ok := err.(net.Error); ok {
		return netErr.Timeout()
	}

	// DNS解決エラー
	if dnsErr, ok := err.(*net.DNSError); ok {
		return dnsErr.Temporary()
	}

	// 接続エラー
	errStr := err.Error()
	retryableErrors := []string{
		"connection reset by peer",
		"connection refused",
		"no such host",
		"timeout",
		"network is unreachable",
		"connection timeout",
		"i/o timeout",
		"broken pipe",
	}

	for _, retryableErr := range retryableErrors {
		if strings.Contains(strings.ToLower(errStr), retryableErr) {
			return true
		}
	}

	return false
}

// isRetryableHTTPStatus HTTPステータスコードがリトライ可能かどうかを判定します
func (c *Client) isRetryableHTTPStatus(statusCode int, body []byte) bool {
	switch statusCode {
	case 500, 502, 503, 504: // サーバーエラー
		return true
	case 429: // レート制限
		return true
	case 403: // 認証・認可エラー（一部はリトライ不可）
		// YouTube APIの403エラーには複数パターンがある
		bodyStr := strings.ToLower(string(body))
		// リトライ不可能なエラー
		nonRetryableErrors := []string{
			"keyinvalid",
			"invalidcredentials",
			"dailylimitexceeded",
			"quotaexceeded",
			"accessnotconfigured",
		}

		for _, nonRetryable := range nonRetryableErrors {
			if strings.Contains(bodyStr, nonRetryable) {
				return false
			}
		}

		// その他の403はリトライ可能
		return true
	case 401: // 認証エラー（基本的にリトライ不可）
		return false
	case 400: // リクエストエラー（基本的にリトライ不可）
		// ただし一部の400は一時的な場合がある
		bodyStr := strings.ToLower(string(body))
		if strings.Contains(bodyStr, "backenderror") ||
			strings.Contains(bodyStr, "internalerror") {
			return true
		}
		return false
	default:
		return false
	}
}

// FetchVideoInfo ライブ配信詳細を含む動画情報を取得します
func (c *Client) FetchVideoInfo(ctx context.Context, videoID string) (*entity.VideoInfo, error) {
	apiURL := fmt.Sprintf("https://www.googleapis.com/youtube/v3/videos?part=snippet,liveStreamingDetails&id=%s&key=%s", videoID, c.apiKey)

	// URL検証
	if _, err := url.ParseRequestURI(apiURL); err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	resp, err := c.doRequestWithRetry(ctx, apiURL)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err2 := resp.Body.Close(); err2 != nil {
			// 重要ではないためエラーをログに記録しますが返却しません
			fmt.Printf("Warning: failed to close response body: %v\n", err2)
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

	resp, err := c.doRequestWithRetry(ctx, apiURL)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err2 := resp.Body.Close(); err2 != nil {
			// 重要ではないためエラーをログに記録しますが返却しません
			fmt.Printf("Warning: failed to close response body: %v\n", err2)
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
