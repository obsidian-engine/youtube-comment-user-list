package youtube

import (
	"context"
	"errors"
	"log"
	"strings"
	"time"

	"github.com/obsidian-engine/youtube-comment-user-list/backend/internal/adapter/logging"
	"github.com/obsidian-engine/youtube-comment-user-list/backend/internal/port"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

type API struct {
	APIKey string
}

func New(apiKey string) *API { return &API{APIKey: apiKey} }

// isLiveChatEnded はエラーがライブチャットの終了または無効化を示すかどうかを判定します。
// "forbidden" は rate limit や quota 超過でも返されるため、終了判定には使用しない。
func isLiveChatEnded(err error) bool {
	if err == nil {
		return false
	}

	errMsg := strings.ToLower(err.Error())
	endedKeywords := []string{
		"livechatdisabled",
		"livechatended",
		"livechatnotfound",
		"chatdisabled",
		"livechatnotactive",
	}

	for _, keyword := range endedKeywords {
		if strings.Contains(errMsg, keyword) {
			return true
		}
	}

	// "notfound" + "livechat" の組み合わせもチェック
	return strings.Contains(errMsg, "notfound") && strings.Contains(errMsg, "livechat")
}

// isTransientError は一時的なエラー（リトライで解消される可能性があるもの）かどうかを判定します。
// 永続的な認証エラー（APIキー不正等）はリトライしない。
func isTransientError(err error) bool {
	if err == nil {
		return false
	}

	// googleapi.Error の構造体でHTTPステータスコードを判定
	var apiErr *googleapi.Error
	if errors.As(err, &apiErr) {
		switch apiErr.Code {
		case 429, 500, 503:
			return true
		}
	}

	errMsg := strings.ToLower(err.Error())
	transientKeywords := []string{
		"ratelimit",
		"quota",
		"backend error",
		"service unavailable",
	}

	for _, keyword := range transientKeywords {
		if strings.Contains(errMsg, keyword) {
			return true
		}
	}

	return false
}

func (a *API) GetActiveLiveChatID(ctx context.Context, videoID string) (string, error) {
	log.Printf("[YOUTUBE_API] GetActiveLiveChatID called with videoID: %s", videoID)

	if a.APIKey == "" {
		log.Printf("[YOUTUBE_API] Error: API key is empty")
		return "", errors.New("youtube api key is required")
	}

	if videoID == "" {
		log.Printf("[YOUTUBE_API] Error: videoID is empty")
		return "", errors.New("video ID is required")
	}

	// YouTube Data API v3を使用して動画情報を取得
	service, err := youtube.NewService(ctx, option.WithAPIKey(a.APIKey))
	if err != nil {
		log.Printf("[YOUTUBE_API] Failed to create YouTube service: %v", err)
		return "", err
	}

	// 動画情報を取得してライブチャットIDを取得
	call := service.Videos.List([]string{"liveStreamingDetails"}).Id(videoID)
	response, err := call.Do()
	if err != nil {
		log.Printf("[YOUTUBE_API] Failed to get video details: %v", err)
		return "", err
	}

	if len(response.Items) == 0 {
		log.Printf("[YOUTUBE_API] Video not found: %s", videoID)
		return "", errors.New("video not found")
	}

	video := response.Items[0]
	if video.LiveStreamingDetails == nil {
		log.Printf("[YOUTUBE_API] Video is not a live stream: %s", videoID)
		return "", errors.New("video is not a live stream")
	}

	if video.LiveStreamingDetails.ActiveLiveChatId == "" {
		log.Printf("[YOUTUBE_API] Live chat is not active for video: %s", videoID)
		return "", errors.New("live chat is not active")
	}

	liveChatID := video.LiveStreamingDetails.ActiveLiveChatId
	log.Printf("[YOUTUBE_API] Found active liveChatID: %s", liveChatID)
	return liveChatID, nil
}

func (a *API) ListLiveChatMessages(ctx context.Context, liveChatID string, pageToken string) (items []port.ChatMessage, nextPageToken string, pollingIntervalMillis int64, skippedCount int, isEnded bool, err error) {
	log.Printf("[YOUTUBE_API] ListLiveChatMessages called with liveChatID: %s", liveChatID)

	if a.APIKey == "" {
		log.Printf("[YOUTUBE_API] Error: API key is empty")
		return nil, "", 0, 0, false, errors.New("youtube api key is required")
	}

	if liveChatID == "" {
		log.Printf("[YOUTUBE_API] Error: liveChatID is empty")
		return nil, "", 0, 0, false, errors.New("live chat ID is required")
	}

	// YouTube Data API v3を使用してライブチャットメッセージを取得
	service, err := youtube.NewService(ctx, option.WithAPIKey(a.APIKey))
	if err != nil {
		log.Printf("[YOUTUBE_API] Failed to create YouTube service: %v", err)
		return nil, "", 0, 0, false, err
	}

	// Live Chat APIは1回の呼び出しで増分取得を行う設計
	// 無限ループを避けるため、1ページのみ取得
	call := service.LiveChatMessages.List(liveChatID, []string{"snippet", "authorDetails"})

	// ページトークンを設定（初回は空）
	if pageToken != "" {
		call = call.PageToken(pageToken)
	}

	// Live Chat API仕様: デフォルト200、最大2000
	// 最大値に設定してより多くのコメントを一度に取得
	call = call.MaxResults(2000)

	const maxAttempts = 4
	var response *youtube.LiveChatMessageListResponse
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		response, err = call.Do()
		if err == nil {
			break
		}

		logging.Log(ctx, "warn", "YOUTUBE_API", "API call failed (attempt %d/%d): %v", attempt, maxAttempts, err)

		if isLiveChatEnded(err) {
			logging.Log(ctx, "warn", "YOUTUBE_API", "Live chat ended or disabled")
			return nil, "", 0, 0, true, nil
		}

		if attempt < maxAttempts && isTransientError(err) {
			backoff := time.Duration(1<<uint(attempt-1)) * time.Second // 1s, 2s, 4s
			logging.Log(ctx, "warn", "YOUTUBE_API", "Retrying after %v...", backoff)
			select {
			case <-time.After(backoff):
				continue
			case <-ctx.Done():
				log.Printf("[YOUTUBE_API] Context cancelled during retry backoff: %v", ctx.Err())
				return nil, "", 0, 0, false, ctx.Err()
			}
		}

		return nil, "", 0, 0, false, err
	}

	if response == nil {
		return nil, "", 0, 0, false, errors.New("youtube API returned nil response")
	}

	// レスポンスからメッセージを変換
	var messages []port.ChatMessage
	skippedCount = 0
	for _, item := range response.Items {
		if item.AuthorDetails != nil && item.Snippet != nil {
			// publishedAtを解析
			publishedAt, err := time.Parse(time.RFC3339, item.Snippet.PublishedAt)
			if err != nil {
				log.Printf("[YOUTUBE_API] Failed to parse publishedAt for message %s: raw=%q err=%v", item.Id, item.Snippet.PublishedAt, err)
				publishedAt = time.Now()
			}

			messages = append(messages, port.ChatMessage{
				ID:          item.Id,
				ChannelID:   item.AuthorDetails.ChannelId,
				DisplayName: item.AuthorDetails.DisplayName,
				Message:     item.Snippet.DisplayMessage,
				PublishedAt: publishedAt,
			})
		} else {
			skippedCount++
			logging.Log(ctx, "warn", "YOUTUBE_API", "Skipped message in chat %s (AuthorDetails=%v, Snippet=%v, ID=%s)", liveChatID, item.AuthorDetails != nil, item.Snippet != nil, item.Id)
		}
	}

	logging.Log(ctx, "info", "YOUTUBE_API", "Retrieved %d messages, skipped %d (total=%d, pageToken=%s, next=%s, pollingIntervalMillis=%d)", len(messages), skippedCount, len(response.Items), pageToken, response.NextPageToken, response.PollingIntervalMillis)

	return messages, response.NextPageToken, int64(response.PollingIntervalMillis), skippedCount, false, nil
}
