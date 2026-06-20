package youtube

import (
	"context"
	"errors"
	"log"
	"strings"
	"time"

	"github.com/obsidian-engine/youtube-comment-user-list/backend/internal/adapter/logging"
	"github.com/obsidian-engine/youtube-comment-user-list/backend/internal/domain"
	"github.com/obsidian-engine/youtube-comment-user-list/backend/internal/port"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

type API struct {
	APIKey             string
	channelNameCache   map[string]string // channelID -> チャンネル名キャッシュ
	channelHandleCache map[string]string // channelID -> ハンドル(@username)キャッシュ
}

func New(apiKey string) *API {
	return &API{
		APIKey:             apiKey,
		channelNameCache:   make(map[string]string),
		channelHandleCache: make(map[string]string),
	}
}

// reasonToCode は googleapi.Error の Errors[0].Reason を domain.APIErrorCode にマッピングする。
// YouTube Data API v3 が返す reason 文字列を機械可読コードに変換する。
var reasonToCode = map[string]domain.APIErrorCode{
	"quotaExceeded":         domain.ErrCodeQuotaExceeded,
	"userRateLimitExceeded": domain.ErrCodeRateLimited,
	"videoNotFound":         domain.ErrCodeVideoNotFound,
	"liveChatEnded":         domain.ErrCodeLiveChatEnded,
	"liveChatDisabled":      domain.ErrCodeLiveChatEnded,
	"liveChatNotActive":     domain.ErrCodeLiveChatEnded,
}

// classifyAPIError は googleapi.Error を domain.APIError に変換する。
// HTTP status + Errors[0].Reason に基づいて分類し、既存の logging.Log 呼出は維持する。
// 分類できない場合は元のエラーをそのまま返す。
func classifyAPIError(err error) error {
	if err == nil {
		return nil
	}

	var gErr *googleapi.Error
	if !errors.As(err, &gErr) {
		return err
	}

	// 401/403 で reason 不明な場合は auth_failed にフォールバック
	if gErr.Code == 401 || gErr.Code == 403 {
		reason := ""
		if len(gErr.Errors) > 0 {
			reason = gErr.Errors[0].Reason
		}
		if code, ok := reasonToCode[reason]; ok {
			return &domain.APIError{Code: code, Message: gErr.Message, Wrapped: err}
		}
		return &domain.APIError{Code: domain.ErrCodeAuthFailed, Message: gErr.Message, Wrapped: err}
	}

	// reason マップで変換を試みる
	reason := ""
	if len(gErr.Errors) > 0 {
		reason = gErr.Errors[0].Reason
	}
	if code, ok := reasonToCode[reason]; ok {
		return &domain.APIError{Code: code, Message: gErr.Message, Wrapped: err}
	}

	// 429 は reason なしでも rate_limited
	if gErr.Code == 429 {
		return &domain.APIError{Code: domain.ErrCodeRateLimited, Message: gErr.Message, Wrapped: err}
	}

	return err
}

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

func (a *API) GetActiveLiveChatID(ctx context.Context, videoID string) (port.VideoMeta, error) {
	log.Printf("[YOUTUBE_API] GetActiveLiveChatID called with videoID: %s", videoID)

	if a.APIKey == "" {
		log.Printf("[YOUTUBE_API] Error: API key is empty")
		return port.VideoMeta{}, errors.New("youtube api key is required")
	}

	if videoID == "" {
		log.Printf("[YOUTUBE_API] Error: videoID is empty")
		return port.VideoMeta{}, errors.New("video ID is required")
	}

	// YouTube Data API v3を使用して動画情報を取得
	service, err := youtube.NewService(ctx, option.WithAPIKey(a.APIKey))
	if err != nil {
		log.Printf("[YOUTUBE_API] Failed to create YouTube service: %v", err)
		return port.VideoMeta{}, err
	}

	// snippet を追加して title / channelTitle を同一 call で取得する
	call := service.Videos.List([]string{"liveStreamingDetails", "snippet"}).Id(videoID)
	response, err := call.Do()
	if err != nil {
		log.Printf("[YOUTUBE_API] Failed to get video details: %v", err)
		return port.VideoMeta{}, classifyAPIError(err)
	}

	if len(response.Items) == 0 {
		log.Printf("[YOUTUBE_API] Video not found: %s", videoID)
		return port.VideoMeta{}, &domain.APIError{Code: domain.ErrCodeVideoNotFound, Message: "video not found"}
	}

	video := response.Items[0]
	if video.LiveStreamingDetails == nil {
		log.Printf("[YOUTUBE_API] Video is not a live stream: %s", videoID)
		return port.VideoMeta{}, &domain.APIError{Code: domain.ErrCodeVideoNotFound, Message: "video is not a live stream"}
	}

	if video.LiveStreamingDetails.ActiveLiveChatId == "" {
		log.Printf("[YOUTUBE_API] Live chat is not active for video: %s", videoID)
		return port.VideoMeta{}, &domain.APIError{Code: domain.ErrCodeLiveChatEnded, Message: "live chat is not active"}
	}

	liveChatID := video.LiveStreamingDetails.ActiveLiveChatId

	// snippet から title / channelTitle を抽出 (nil ガード: 空文字で続行)
	var title, channelTitle string
	if video.Snippet != nil {
		title = video.Snippet.Title
		channelTitle = video.Snippet.ChannelTitle
	}

	log.Printf("[YOUTUBE_API] Found active liveChatID: %s title: %q channel: %q", liveChatID, title, channelTitle)
	return port.VideoMeta{
		LiveChatID:   liveChatID,
		Title:        title,
		ChannelTitle: channelTitle,
	}, nil
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

		return nil, "", 0, 0, false, classifyAPIError(err)
	}

	if response == nil {
		return nil, "", 0, 0, false, errors.New("youtube API returned nil response")
	}

	// レスポンスからメッセージを変換
	messages := []port.ChatMessage{}
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

func (a *API) GetChannelDisplayNames(ctx context.Context, channelIDs []string) (map[string]string, error) {
	result := make(map[string]string)
	if len(channelIDs) == 0 {
		return result, nil
	}

	// キャッシュ済みを返し、未キャッシュを収集
	var uncached []string
	for _, id := range channelIDs {
		if name, ok := a.channelNameCache[id]; ok {
			result[id] = name
		} else {
			uncached = append(uncached, id)
		}
	}

	if len(uncached) == 0 || a.APIKey == "" {
		return result, nil
	}

	// YouTube Channels API: 1リクエストあたり最大50件
	const batchSize = 50
	for i := 0; i < len(uncached); i += batchSize {
		end := min(i+batchSize, len(uncached))
		batch := uncached[i:end]

		service, err := youtube.NewService(ctx, option.WithAPIKey(a.APIKey))
		if err != nil {
			logging.Log(ctx, "warn", "YOUTUBE_API", "Failed to create service for channel names: %v", err)
			continue
		}

		call := service.Channels.List([]string{"snippet"}).Id(strings.Join(batch, ","))
		response, err := call.Do()
		if err != nil {
			logging.Log(ctx, "warn", "YOUTUBE_API", "Failed to get channel names: %v", err)
			continue
		}

		for _, item := range response.Items {
			name := item.Snippet.Title
			a.channelNameCache[item.Id] = name
			result[item.Id] = name

			// ハンドル(CustomUrl)も同時にキャッシュ
			a.channelHandleCache[item.Id] = item.Snippet.CustomUrl
		}
	}

	logging.Log(ctx, "info", "YOUTUBE_API", "Resolved %d/%d channel display names (cache size: %d)", len(result), len(channelIDs), len(a.channelNameCache))
	return result, nil
}

// GetChannelHandles は channelIDs に対応するハンドル(@username)マップを返す。
// GetChannelDisplayNames で既に解決済みの @ チャンネルはキャッシュから返す。
// 未キャッシュのチャンネルは Channels API を1回呼び出し、以降はキャッシュする。
// ハンドルが存在しない/空のチャンネルはマップに含まれない。
func (a *API) GetChannelHandles(ctx context.Context, channelIDs []string) (map[string]string, error) {
	result := make(map[string]string)
	if len(channelIDs) == 0 {
		return result, nil
	}

	// キャッシュ済みを返し、未キャッシュを収集
	var uncached []string
	for _, id := range channelIDs {
		if handle, ok := a.channelHandleCache[id]; ok {
			if handle != "" {
				result[id] = handle
			}
		} else {
			uncached = append(uncached, id)
		}
	}

	if len(uncached) == 0 || a.APIKey == "" {
		return result, nil
	}

	// YouTube Channels API: 1リクエストあたり最大50件
	const batchSize = 50
	for i := 0; i < len(uncached); i += batchSize {
		end := min(i+batchSize, len(uncached))
		batch := uncached[i:end]

		service, err := youtube.NewService(ctx, option.WithAPIKey(a.APIKey))
		if err != nil {
			logging.Log(ctx, "warn", "YOUTUBE_API", "Failed to create service for channel handles: %v", err)
			continue
		}

		call := service.Channels.List([]string{"snippet"}).Id(strings.Join(batch, ","))
		response, err := call.Do()
		if err != nil {
			logging.Log(ctx, "warn", "YOUTUBE_API", "Failed to get channel handles: %v", err)
			continue
		}

		for _, item := range response.Items {
			handle := item.Snippet.CustomUrl
			// 名前もキャッシュ（GetChannelDisplayNames との重複呼び出しを避けるため）
			a.channelNameCache[item.Id] = item.Snippet.Title
			a.channelHandleCache[item.Id] = handle
			if handle != "" {
				result[item.Id] = handle
			}
		}
	}

	logging.Log(ctx, "info", "YOUTUBE_API", "Resolved %d/%d channel handles (cache size: %d)", len(result), len(channelIDs), len(a.channelHandleCache))
	return result, nil
}
