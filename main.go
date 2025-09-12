// Package main provides YouTube Live Chat polling functionality
// with a web interface for monitoring live chat messages.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/joho/godotenv"
)

type LiveStreamingDetails struct {
	ActiveLiveChatID   string `json:"activeLiveChatId"`
	ActualStartTime    string `json:"actualStartTime"`
	ActualEndTime      string `json:"actualEndTime"`
	ConcurrentViewers  string `json:"concurrentViewers"`
	ScheduledStartTime string `json:"scheduledStartTime"`
}
type VideosListResp struct {
	Items []struct {
		ID                   string               `json:"id"`
		Snippet              VideoSnippet         `json:"snippet"`
		LiveStreamingDetails LiveStreamingDetails `json:"liveStreamingDetails"`
	} `json:"items"`
	Error *APIError `json:"error,omitempty"`
}

type VideoSnippet struct {
	Title                string `json:"title"`
	ChannelTitle         string `json:"channelTitle"`
	LiveBroadcastContent string `json:"liveBroadcastContent"`
}

type APIError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Status  string `json:"status"`
}

type AuthorDetails struct {
	DisplayName string `json:"displayName"`
	ChannelID   string `json:"channelId"`
	IsChatOwner bool   `json:"isChatOwner"`
	IsModerator bool   `json:"isChatModerator"`
	IsMember    bool   `json:"isChatSponsor"`
}
type ChatMessage struct {
	AuthorDetails AuthorDetails `json:"authorDetails"`
}
type LiveChatResp struct {
	Items                 []ChatMessage `json:"items"`
	NextPageToken         string        `json:"nextPageToken"`
	PollingIntervalMillis int           `json:"pollingIntervalMillis"`
	Error                 *APIError     `json:"error,omitempty"`
}

type UserList struct {
	mu     sync.RWMutex
	users  map[string]string // channelID -> displayName（displayName重複対策）
	sorted []string          // 表示用キャッシュ
}

// PollerRegistry - 動画ID毎のポーリング管理
type Poller struct {
	videoID  string
	cancel   context.CancelFunc
	msgsChan chan ChatMessage
	subs     int
	userList *UserList
}

type PollerRegistry struct {
	mu      sync.RWMutex
	pollers map[string]*Poller
	apiKey  string
}

// LogBuffer - アプリケーションログのリングバッファ
type LogEntry struct {
	Timestamp string `json:"timestamp"`
	Level     string `json:"level"`
	Message   string `json:"message"`
	VideoID   string `json:"video_id,omitempty"`
}

type LogBuffer struct {
	mu      sync.RWMutex
	entries []LogEntry
	maxSize int
}

var globalLogBuffer = NewLogBuffer(1000) // 最新1000件のログを保持

func NewLogBuffer(maxSize int) *LogBuffer {
	return &LogBuffer{
		entries: make([]LogEntry, 0, maxSize),
		maxSize: maxSize,
	}
}

func (lb *LogBuffer) Add(level, message, videoID string) {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	entry := LogEntry{
		Timestamp: time.Now().Format("2006-01-02 15:04:05"),
		Level:     level,
		Message:   message,
		VideoID:   videoID,
	}

	lb.entries = append(lb.entries, entry)
	if len(lb.entries) > lb.maxSize {
		lb.entries = lb.entries[1:] // 古いエントリを削除
	}
}

func (lb *LogBuffer) GetAll() []LogEntry {
	lb.mu.RLock()
	defer lb.mu.RUnlock()

	result := make([]LogEntry, len(lb.entries))
	copy(result, lb.entries)
	return result
}

// カスタムログ関数
func logInfo(message string, videoID ...string) {
	log.Printf("[INFO] %s", message)
	vid := ""
	if len(videoID) > 0 {
		vid = videoID[0]
	}
	globalLogBuffer.Add("INFO", message, vid)
}

func NewPollerRegistry(apiKey string) *PollerRegistry {
	return &PollerRegistry{
		pollers: make(map[string]*Poller),
		apiKey:  apiKey,
	}
}

func init() {
	err := godotenv.Load(".env")
	if err != nil {
		fmt.Printf("can not read env file.: %v", err)
		return
	}
}

func NewUserList() *UserList {
	return &UserList{
		users:  make(map[string]string),
		sorted: make([]string, 0),
	}
}

// Attach - ポーリングに接続、新規の場合は開始
func (pr *PollerRegistry) Attach(videoID string) <-chan ChatMessage {
	pr.mu.Lock()
	defer pr.mu.Unlock()

	poller, exists := pr.pollers[videoID]
	if !exists {
		// 新規ポーリング開始
		ctx, cancel := context.WithCancel(context.Background())
		msgsChan := make(chan ChatMessage, 100)
		userList := NewUserList()

		poller = &Poller{
			videoID:  videoID,
			cancel:   cancel,
			msgsChan: msgsChan,
			subs:     0,
			userList: userList,
		}
		pr.pollers[videoID] = poller

		// バックグラウンドでポーリング開始
		go pr.startPolling(ctx, videoID, poller)
	}

	poller.subs++
	log.Printf("Attached to %s, subscribers: %d", videoID, poller.subs)
	return poller.msgsChan
}

// Detach - ポーリングから切断、参照数0で停止
func (pr *PollerRegistry) Detach(videoID string) {
	pr.mu.Lock()
	defer pr.mu.Unlock()

	poller, exists := pr.pollers[videoID]
	if !exists {
		return
	}

	poller.subs--
	log.Printf("Detached from %s, subscribers: %d", videoID, poller.subs)

	if poller.subs <= 0 {
		// ポーリング停止
		poller.cancel()
		close(poller.msgsChan)
		delete(pr.pollers, videoID)
		log.Printf("Stopped polling for %s", videoID)
	}
}

// GetUserList - 指定動画のユーザーリストを取得
func (pr *PollerRegistry) GetUserList(videoID string) []string {
	pr.mu.RLock()
	defer pr.mu.RUnlock()

	poller, exists := pr.pollers[videoID]
	if !exists {
		return []string{}
	}

	return poller.userList.Snapshot()
}

func (ul *UserList) Add(channelID, displayName string) {
	ul.mu.Lock()
	defer ul.mu.Unlock()
	if _, ok := ul.users[channelID]; !ok {
		ul.users[channelID] = displayName
		ul.rebuild()
	}
}

func (ul *UserList) rebuild() {
	names := make([]string, 0, len(ul.users))
	for _, n := range ul.users {
		names = append(names, n)
	}
	slices.Sort(names)
	ul.sorted = names
}

func (ul *UserList) Snapshot() []string {
	ul.mu.RLock()
	defer ul.mu.RUnlock()
	out := make([]string, len(ul.sorted))
	copy(out, ul.sorted)
	return out
}

func getEnv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func fetchActiveLiveChatID(apiKey, videoID string) (string, error) {
	u := fmt.Sprintf("https://www.googleapis.com/youtube/v3/videos?part=snippet,liveStreamingDetails&id=%s&key=%s", videoID, apiKey)
	// セキュリティ: APIキーを含むURLは完全にログに出力しない
	log.Printf("Videos API リクエスト: videoID=%s", videoID)

	resp, err := http.Get(u)
	if err != nil {
		return "", fmt.Errorf("HTTP リクエスト失敗: %v", err)
	}
	defer resp.Body.Close()

	// レスポンスボディを読み取って詳細ログ出力
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("レスポンス読み取り失敗: %v", err)
	}
	log.Printf("APIレスポンス (ステータス: %d): %s", resp.StatusCode, string(body))

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("videos.list API エラー: status=%d, body=%s", resp.StatusCode, string(body))
	}

	var v VideosListResp
	if err := json.Unmarshal(body, &v); err != nil {
		return "", fmt.Errorf("JSONパース失敗: %v", err)
	}

	// API エラーレスポンスのチェック
	if v.Error != nil {
		return "", fmt.Errorf("YouTube API エラー: %d %s - %s", v.Error.Code, v.Error.Status, v.Error.Message)
	}

	if len(v.Items) == 0 {
		return "", fmt.Errorf("動画が見つかりません (動画ID: %s). 動画IDが正しいか、動画が存在するか確認してください", videoID)
	}

	item := v.Items[0]
	log.Printf("動画情報:")
	log.Printf("  タイトル: %s", item.Snippet.Title)
	log.Printf("  チャンネル: %s", item.Snippet.ChannelTitle)
	log.Printf("  ライブ配信種別: %s", item.Snippet.LiveBroadcastContent)
	log.Printf("  activeLiveChatId: %s", item.LiveStreamingDetails.ActiveLiveChatID)
	log.Printf("  actualStartTime: %s", item.LiveStreamingDetails.ActualStartTime)
	log.Printf("  actualEndTime: %s", item.LiveStreamingDetails.ActualEndTime)
	log.Printf("  concurrentViewers: %s", item.LiveStreamingDetails.ConcurrentViewers)

	// ライブ配信状態の詳細チェック（YouTube API仕様: upcoming, active, none + 実際には live も存在）
	switch item.Snippet.LiveBroadcastContent {
	case "none":
		return "", fmt.Errorf("この動画はライブ配信ではありません")
	case "upcoming":
		return "", fmt.Errorf("ライブ配信は予定されていますが、まだ開始されていません (開始予定: %s)", item.LiveStreamingDetails.ScheduledStartTime)
	case "active", "live":
		log.Printf("ライブ配信中を確認 (状態: %s)", item.Snippet.LiveBroadcastContent)
		// 実際に開始されているかも確認
		if item.LiveStreamingDetails.ActualStartTime == "" {
			log.Printf("警告: liveBroadcastContent=%s ですが actualStartTime が設定されていません", item.Snippet.LiveBroadcastContent)
		}
	default:
		log.Printf("不明なライブ配信状態: %s", item.Snippet.LiveBroadcastContent)
		return "", fmt.Errorf("サポートされていないライブ配信状態: %s", item.Snippet.LiveBroadcastContent)
	}

	if item.LiveStreamingDetails.ActiveLiveChatID == "" {
		return "", fmt.Errorf("activeLiveChatId が空です。チャット機能が無効化されているか、メンバー限定チャットの可能性があります")
	}

	return item.LiveStreamingDetails.ActiveLiveChatID, nil
}

func fetchLiveChatOnce(apiKey, liveChatID, pageToken string) (LiveChatResp, error) {
	base := "https://www.googleapis.com/youtube/v3/liveChat/messages"
	params := []string{
		"part=authorDetails",
		"maxResults=2000",
		"liveChatId=" + liveChatID,
		"key=" + apiKey,
	}
	if pageToken != "" {
		params = append(params, "pageToken="+pageToken)
	}
	apiURL := base + "?" + strings.Join(params, "&")
	// セキュリティ: APIキーを含むURLは完全にログに出力しない
	log.Printf("LiveChat API リクエスト: liveChatId=%s, pageToken=%s", liveChatID, pageToken)

	resp, err := http.Get(url)
	if err != nil {
		return LiveChatResp{}, fmt.Errorf("LiveChat HTTP リクエスト失敗: %v", err)
	}
	defer resp.Body.Close()

	// レスポンスボディを読み取り
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return LiveChatResp{}, fmt.Errorf("LiveChat レスポンス読み取り失敗: %v", err)
	}

	if resp.StatusCode != 200 {
		log.Printf("LiveChat API エラーレスポンス (ステータス: %d): %s", resp.StatusCode, string(body))
		return LiveChatResp{}, fmt.Errorf("liveChatMessages.list API エラー: status=%d, body=%s", resp.StatusCode, string(body))
	}

	var v LiveChatResp
	if err := json.Unmarshal(body, &v); err != nil {
		return LiveChatResp{}, fmt.Errorf("LiveChat JSONパース失敗: %v", err)
	}

	// API エラーレスポンスのチェック
	if v.Error != nil {
		return LiveChatResp{}, fmt.Errorf("LiveChat API エラー: %d %s - %s", v.Error.Code, v.Error.Status, v.Error.Message)
	}

	log.Printf("LiveChat API 成功: %d 件のメッセージを取得, NextPageToken: %s, PollingInterval: %dms",
		len(v.Items), v.NextPageToken, v.PollingIntervalMillis)

	return v, nil
}

// startPolling - バックグラウンドでポーリングを実行
func (pr *PollerRegistry) startPolling(ctx context.Context, videoID string, poller *Poller) {
	log.Printf("Starting polling for video: %s", videoID)

	liveChatID, err := fetchActiveLiveChatID(pr.apiKey, videoID)
	if err != nil {
		log.Printf("Failed to get activeLiveChatId for %s: %v", videoID, err)
		return
	}

	var pageToken string
	var lastWait = 2000 // ms
	var consecutiveErrors int

	// ポーリング・ループ
	go func() {
		var pageToken string
		var lastWait = 8000 // ms フォールバック
		var consecutiveErrors int
		for {
			resp, err := fetchLiveChatOnce(apiKey, liveChatID, pageToken)
			if err != nil {
				consecutiveErrors++
				log.Printf("LiveChat fetch error (%d回目): %v", consecutiveErrors, err)

				// エラー回数に応じた待機時間の調整
				errorWait := lastWait
				if consecutiveErrors > 5 {
					errorWait = lastWait * 3
				}

				select {
				case <-ctx.Done():
					return
				case <-time.After(time.Duration(errorWait) * time.Millisecond):
					continue
				}

				time.Sleep(time.Duration(errorWait) * time.Millisecond)
				continue
			}

			// 成功した場合はエラーカウントをリセット
			consecutiveErrors = 0

			for _, item := range resp.Items {
				// ユーザーリストに追加
				poller.userList.Add(item.AuthorDetails.ChannelID, item.AuthorDetails.DisplayName)

				// SSE用にメッセージを送信
				select {
				case poller.msgsChan <- item:
				default:
					// チャンネルがフルの場合はスキップ
				}
			}

			pageToken = resp.NextPageToken
			wait := resp.PollingIntervalMillis
			if wait <= 0 {
				wait = lastWait
			} else {
				lastWait = wait
			}

			select {
			case <-ctx.Done():
				return
			case <-time.After(time.Duration(wait) * time.Millisecond):
				continue
			}
		}
	}
}

func main() {
	apiKey := os.Getenv("YT_API_KEY")
	port := getEnv("PORT", "8080")

	if apiKey == "" {
		log.Fatal("環境変数 YT_API_KEY を設定してください")
	}

	// PollerRegistry初期化
	registry := NewPollerRegistry(apiKey)

	// トップページ - 動画ID入力フォーム
	http.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if _, err := fmt.Fprintf(w, homePageHTML, ""); err != nil {
			log.Printf("ホームページ出力エラー: %v", err)
		}
	})

	// JSONエンドポイント - 動画ID指定必須
	http.HandleFunc("/users.json", func(w http.ResponseWriter, r *http.Request) {
		videoID := r.URL.Query().Get("video_id")
		if videoID == "" {
			http.Error(w, "video_id parameter required", http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		type out struct {
			Count int      `json:"count"`
			Users []string `json:"users"`
		}
		list := registry.GetUserList(videoID)
		_ = json.NewEncoder(w).Encode(out{Count: len(list), Users: list})
	})

	// Server-Sent Events エンドポイント
	http.HandleFunc("/events", func(w http.ResponseWriter, r *http.Request) {
		videoID := r.URL.Query().Get("video_id")
		if videoID == "" {
			http.Error(w, "video_id parameter required", http.StatusBadRequest)
			return
		}

		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		msgCh := registry.Attach(videoID)
		defer registry.Detach(videoID)

		ctx := r.Context()
		for {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-msgCh:
				if !ok {
					return // チャンネルが閉じられた
				}
				msgJSON, _ := json.Marshal(msg)
				if _, err := fmt.Fprintf(w, "data: %s\n\n", msgJSON); err != nil {
					log.Printf("SSEデータ出力エラー: %v", err)
				}
				flusher.Flush()
			}
		}
	})

	// OBS向けオーバーレイ - 動画ID指定必須
	http.HandleFunc("/overlay", func(w http.ResponseWriter, r *http.Request) {
		videoID := r.URL.Query().Get("video_id")
		if videoID == "" {
			http.Error(w, "video_id parameter required. Use: /overlay?video_id=YOUR_VIDEO_ID", http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if _, err := fmt.Fprintf(w, overlayPageHTML, videoID); err != nil {
			log.Printf("オーバーレイページ出力エラー: %v", err)
		}
	})

	// ログ表示ページ
	http.HandleFunc("/logs", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if _, err := fmt.Fprint(w, logsPageHTML); err != nil {
			log.Printf("ログページ出力エラー: %v", err)
		}
	})

	// ログAPI エンドポイント
	http.HandleFunc("/api/logs", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		logs := globalLogBuffer.GetAll()
		_ = json.NewEncoder(w).Encode(logs)
	})

	logInfo("Server starting", "")
	log.Printf("server listening on :%s", port)
	log.Printf("  Home page: http://localhost:%s/", port)
	log.Printf("  Overlay: http://localhost:%s/overlay?video_id=YOUR_VIDEO_ID", port)
	log.Printf("  Logs: http://localhost:%s/logs", port)

	// セキュアなHTTPサーバー設定
	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           nil,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	log.Fatal(srv.ListenAndServe())
}

// HTMLテンプレート定数
const homePageHTML = `<!doctype html>
<html>
<head>
<meta charset="utf-8" />
<meta name="viewport" content="width=device-width,initial-scale=1" />
<title>YouTube Live Chat Monitor</title>
<style>
  :root {
    --primary: #1976d2;
    --primary-dark: #1565c0;
    --bg: #f5f5f5;
    --card-bg: #fff;
    --text: #212121;
    --text-secondary: #757575;
  }
  * { box-sizing: border-box; }
  html, body { margin: 0; padding: 0; font-family: 'Segoe UI', system-ui, sans-serif; background: var(--bg); }
  .container { max-width: 600px; margin: 50px auto; padding: 20px; }
  .card {
    background: var(--card-bg);
    border-radius: 12px;
    padding: 32px;
    box-shadow: 0 4px 20px rgba(0,0,0,0.1);
  }
  h1 {
    color: var(--text);
    margin-bottom: 8px;
    font-size: 28px;
    font-weight: 600;
  }
  .subtitle {
    color: var(--text-secondary);
    margin-bottom: 32px;
    font-size: 16px;
  }
  .form-group {
    margin-bottom: 24px;
  }
  label {
    display: block;
    margin-bottom: 8px;
    color: var(--text);
    font-weight: 500;
  }
  input {
    width: 100%%;
    padding: 12px 16px;
    border: 2px solid #e0e0e0;
    border-radius: 8px;
    font-size: 16px;
    transition: border-color 0.2s;
  }
  input:focus {
    outline: none;
    border-color: var(--primary);
  }
  button {
    background: var(--primary);
    color: white;
    border: none;
    padding: 12px 32px;
    border-radius: 8px;
    font-size: 16px;
    font-weight: 600;
    cursor: pointer;
    transition: background-color 0.2s;
  }
  button:hover {
    background: var(--primary-dark);
  }
  .example {
    margin-top: 16px;
    padding: 12px;
    background: #f8f9fa;
    border-radius: 6px;
    font-size: 14px;
    color: var(--text-secondary);
  }
</style>
</head>
<body>
<div class="container">
  <div class="card">
    <h1>🎥 YouTube Live Chat Monitor</h1>
    <p class="subtitle">ライブ配信のチャットに参加したユーザーを表示します</p>

    <form onsubmit="handleSubmit(event)">
      <div class="form-group">
        <label for="videoId">YouTube動画ID</label>
        <input
          type="text"
          id="videoId"
          placeholder="https://www.youtube.com/watch?v=kXpv3asP0Qw"
          value="%s"
          required
        />
        <div class="example">
          対応形式:<br>
          • https://www.youtube.com/watch?v=<strong>kXpv3asP0Qw</strong><br>
          • https://youtu.be/<strong>kXpv3asP0Qw</strong><br>
          • <strong>kXpv3asP0Qw</strong> (動画IDのみ)
        </div>
      </div>

      <button type="submit">チャット監視を開始</button>
    </form>
  </div>
</div>

<script>
function extractVideoId(input) {
  // 入力値をトリム
  const trimmed = input.trim();

  // 空文字チェック
  if (!trimmed) return '';

  // パターン1: https://www.youtube.com/watch?v=VIDEO_ID
  const watchPattern = /(?:https?:\/\/)?(?:www\.)?youtube\.com\/watch\?v=([a-zA-Z0-9_-]{11})/;
  let match = trimmed.match(watchPattern);
  if (match) return match[1];

  // パターン2: https://youtu.be/VIDEO_ID
  const shortPattern = /(?:https?:\/\/)?youtu\.be\/([a-zA-Z0-9_-]{11})/;
  match = trimmed.match(shortPattern);
  if (match) return match[1];

  // パターン3: 11文字の動画IDのみ（英数字、-、_）
  const idPattern = /^[a-zA-Z0-9_-]{11}$/;
  if (idPattern.test(trimmed)) return trimmed;

  // どのパターンにも一致しない場合
  return '';
}

function handleSubmit(event) {
  event.preventDefault();
  const input = document.getElementById('videoId').value;
  const videoId = extractVideoId(input);

  if (!videoId) {
    alert('有効なYouTube URLまたは動画IDを入力してください。\\n\\n対応形式:\\n• https://www.youtube.com/watch?v=VIDEO_ID\\n• https://youtu.be/VIDEO_ID\\n• VIDEO_ID（11文字の英数字）');
    return;
  }

  // 成功時はオーバーレイページに遷移
  window.location.href = '/overlay?video_id=' + encodeURIComponent(videoId);
}
</script>
</body>
</html>`

const overlayPageHTML = `<!doctype html>
<html>
<head>
<meta charset="utf-8" />
<meta name="viewport" content="width=device-width,initial-scale=1" />
<title>Commenters Overlay</title>
<style>
  :root {
    --bg: rgba(0,0,0,0.0);
    --card-bg: rgba(0,0,0,0.55);
    --text: #fff;
    --accent: #81E3DD;
  }
  html,body { margin:0; padding:0; background:var(--bg); font-family: system-ui, -apple-system, "Noto Sans JP", sans-serif; }
  .wrap { padding: 12px 14px; }
  .card {
    background: var(--card-bg);
    color: var(--text);
    border-radius: 14px;
    padding: 12px 14px;
    box-shadow: 0 8px 24px rgba(0,0,0,0.25);
    border: 1px solid rgba(255,255,255,0.08);
    min-width: 260px;
  }
  .title {
    font-weight: 700;
    font-size: 14px;
    letter-spacing: .02em;
    margin-bottom: 8px;
  }
  .count {
    font-size: 12px; opacity: .8; margin-left: 6px;
  }
  .status {
    font-size: 11px; opacity: .6; margin-left: 8px;
  }
  ul {
    list-style: none; padding: 0; margin: 0; display: grid; gap: 6px;
  }
  li {
    display: inline-flex;
    align-items: center;
    gap: 8px;
    padding: 6px 8px;
    border-radius: 10px;
    background: linear-gradient(104deg, rgba(83,183,255,.18), rgba(129,227,221,.18));
    border: 1px solid rgba(129,227,221,.25);
    backdrop-filter: blur(4px);
    font-size: 13px;
  }
  .dot {
    width: 6px; height: 6px; border-radius: 999px; background: var(--accent);
  }
</style>
</head>
<body>
<div class="wrap">
  <div class="card">
    <div class="title">
      Commented Users
      <span class="count" id="count">0</span>
      <span class="status" id="status">接続中...</span>
    </div>
    <ul id="list"></ul>
  </div>
</div>

<script>
const videoId = '%s';
let eventSource;
let users = new Set();

function updateUI() {
  const ul = document.getElementById('list');
  const cnt = document.getElementById('count');

  ul.innerHTML = '';
  Array.from(users).forEach(name => {
    const li = document.createElement('li');
    const dot = document.createElement('span');
    dot.className = 'dot';
    const span = document.createElement('span');
    span.textContent = name;
    li.appendChild(dot);
    li.appendChild(span);
    ul.appendChild(li);
  });

  cnt.textContent = users.size;
}

function connectSSE() {
  const status = document.getElementById('status');

  eventSource = new EventSource('/events?video_id=' + encodeURIComponent(videoId));

  eventSource.onopen = function() {
    status.textContent = '接続済み';
    status.style.color = '#4caf50';
  };

  eventSource.onmessage = function(event) {
    try {
      const msg = JSON.parse(event.data);
      if (msg.authorDetails && msg.authorDetails.displayName) {
        users.add(msg.authorDetails.displayName);
        updateUI();
      }
    } catch (e) {
      console.error('Failed to parse message:', e);
    }
  };

  eventSource.onerror = function() {
    status.textContent = '再接続中...';
    status.style.color = '#ff9800';

    setTimeout(() => {
      if (eventSource.readyState === EventSource.CLOSED) {
        connectSSE();
      }
    }, 3000);
  };
}

// 初期接続
connectSSE();

// ページ離脱時にクリーンアップ
window.addEventListener('beforeunload', () => {
  if (eventSource) {
    eventSource.close();
  }
});
</script>
</body>
</html>`

const logsPageHTML = `<!doctype html>
<html>
<head>
<meta charset="utf-8" />
<meta name="viewport" content="width=device-width,initial-scale=1" />
<title>System Logs</title>
<style>
  :root {
    --bg: #1a1a1a;
    --card-bg: #2a2a2a;
    --text: #e0e0e0;
    --text-dim: #999;
    --info: #4fc3f7;
    --error: #f44336;
    --border: #3a3a3a;
  }
  * { box-sizing: border-box; }
  html, body {
    margin: 0;
    padding: 0;
    font-family: 'Monaco', 'Consolas', monospace;
    background: var(--bg);
    color: var(--text);
  }
  .container {
    max-width: 1400px;
    margin: 0 auto;
    padding: 20px;
  }
  .header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 20px;
    padding-bottom: 10px;
    border-bottom: 1px solid var(--border);
  }
  h1 {
    margin: 0;
    font-size: 24px;
  }
  .controls {
    display: flex;
    gap: 10px;
  }
  .filter-btn {
    padding: 6px 12px;
    background: var(--card-bg);
    border: 1px solid var(--border);
    color: var(--text);
    border-radius: 4px;
    cursor: pointer;
    font-size: 12px;
  }
  .filter-btn.active {
    background: var(--info);
    color: #000;
  }
  .log-container {
    background: var(--card-bg);
    border: 1px solid var(--border);
    border-radius: 8px;
    padding: 15px;
    height: calc(100vh - 140px);
    overflow-y: auto;
  }
  .log-entry {
    display: flex;
    gap: 10px;
    padding: 8px 10px;
    margin-bottom: 4px;
    background: #222;
    border-radius: 4px;
    font-size: 13px;
    border-left: 3px solid transparent;
  }
  .log-entry.info {
    border-left-color: var(--info);
  }
  .log-entry.error {
    border-left-color: var(--error);
    background: rgba(244, 67, 54, 0.1);
  }
  .timestamp {
    color: var(--text-dim);
    font-size: 12px;
  }
  .level {
    font-weight: 600;
    width: 50px;
    text-align: center;
    padding: 2px 4px;
    border-radius: 3px;
    font-size: 11px;
  }
  .level.info {
    background: var(--info);
    color: #000;
  }
  .level.error {
    background: var(--error);
    color: #fff;
  }
  .message {
    flex: 1;
  }
  .video-id {
    color: var(--info);
    font-size: 11px;
    padding: 2px 6px;
    background: rgba(79, 195, 247, 0.1);
    border-radius: 3px;
  }
</style>
</head>
<body>
<div class="container">
  <div class="header">
    <h1>📊 System Logs</h1>
    <div class="controls">
      <button class="filter-btn active" data-filter="all">All</button>
      <button class="filter-btn" data-filter="info">Info</button>
      <button class="filter-btn" data-filter="error">Error</button>
    </div>
  </div>

  <div class="log-container" id="log-container"></div>
</div>

<script>
let logs = [];
let filter = 'all';

document.querySelectorAll('.filter-btn').forEach(btn => {
  btn.addEventListener('click', () => {
    document.querySelectorAll('.filter-btn').forEach(b => b.classList.remove('active'));
    btn.classList.add('active');
    filter = btn.dataset.filter;
    renderLogs();
  });
});

async function fetchLogs() {
  try {
    const response = await fetch('/api/logs');
    logs = await response.json();
    renderLogs();
  } catch (error) {
    console.error('Failed to fetch logs:', error);
  }
}

function renderLogs() {
  const container = document.getElementById('log-container');
  const filteredLogs = filter === 'all'
    ? logs
    : logs.filter(log => log.level.toLowerCase() === filter);

  container.innerHTML = filteredLogs.map(log => {
    const levelClass = log.level.toLowerCase();
    const videoId = log.video_id ? '<span class="video-id">' + log.video_id + '</span>' : '';
    return '<div class="log-entry ' + levelClass + '">' +
      '<span class="timestamp">' + log.timestamp + '</span>' +
      '<span class="level ' + levelClass + '">' + log.level + '</span>' +
      '<span class="message">' + log.message + '</span>' +
      videoId +
    '</div>';
  }).join('');

  container.scrollTop = container.scrollHeight;
}

fetchLogs();
setInterval(fetchLogs, 2000);
</script>
</body>
</html>`
