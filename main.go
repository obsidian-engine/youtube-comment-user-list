package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/joho/godotenv"
)

type LiveStreamingDetails struct {
	ActiveLiveChatID      string `json:"activeLiveChatId"`
	ActualStartTime       string `json:"actualStartTime"`
	ActualEndTime         string `json:"actualEndTime"`
	ConcurrentViewers     string `json:"concurrentViewers"`
	ScheduledStartTime    string `json:"scheduledStartTime"`
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
	Title           string `json:"title"`
	ChannelTitle    string `json:"channelTitle"`
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
	url := base + "?" + strings.Join(params, "&")
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

func main() {
	apiKey := os.Getenv("YT_API_KEY")
	videoID := os.Getenv("YT_VIDEO_ID")
	port := getEnv("PORT", "8080")

	if apiKey == "" || videoID == "" {
		log.Fatal("環境変数 YT_API_KEY と YT_VIDEO_ID を設定してください")
	}

	liveChatID, err := fetchActiveLiveChatID(apiKey, videoID)
	if err != nil {
		log.Fatalf("activeLiveChatId取得に失敗: %v", err)
	}
	log.Printf("activeLiveChatId: %s", liveChatID)

	users := NewUserList()

	// ポーリング・ループ
	go func() {
		var pageToken string
		var lastWait = 2000 // ms フォールバック
		var consecutiveErrors int
		for {
			resp, err := fetchLiveChatOnce(apiKey, liveChatID, pageToken)
			if err != nil {
				consecutiveErrors++
				log.Printf("LiveChat fetch error (%d回目): %v", consecutiveErrors, err)
				
				// エラー回数に応じた待機時間の調整
				errorWait := lastWait
				if consecutiveErrors > 5 {
					errorWait = lastWait * 3 // 3倍待機
					log.Printf("連続エラーが多いため、待機時間を %dms に延長", errorWait)
				}
				
				time.Sleep(time.Duration(errorWait) * time.Millisecond)
				continue
			}
			
			// 成功した場合はエラーカウントをリセット
			consecutiveErrors = 0
			
			for _, item := range resp.Items {
				// 重複排除はchannelIDで行う
				users.Add(item.AuthorDetails.ChannelID, item.AuthorDetails.DisplayName)
			}
			pageToken = resp.NextPageToken
			wait := resp.PollingIntervalMillis
			if wait <= 0 {
				wait = lastWait
			} else {
				lastWait = wait
			}
			time.Sleep(time.Duration(wait) * time.Millisecond)
		}
	}()

	// JSONエンドポイント
	http.HandleFunc("/users.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		type out struct {
			Count int      `json:"count"`
			Users []string `json:"users"`
		}
		list := users.Snapshot()
		_ = json.NewEncoder(w).Encode(out{Count: len(list), Users: list})
	})

	// OBS向けシンプル・オーバーレイ
	http.HandleFunc("/overlay", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, `<!doctype html>
<html>
<head>
<meta charset="utf-8" />
<meta name="viewport" content="width=device-width,initial-scale=1" />
<title>Commenters Overlay</title>
<style>
  :root {
    --bg: rgba(0,0,0,0.0);      /* 透過背景（OBS用） */
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
    <div class="title">Commented Users <span class="count" id="count">0</span></div>
    <ul id="list"></ul>
  </div>
</div>
<script>
async function refresh(){
  try{
    const r = await fetch('/users.json', {cache:'no-store'});
    const j = await r.json();
    const ul = document.getElementById('list');
    const cnt = document.getElementById('count');
    ul.innerHTML = '';
    (j.users||[]).forEach(name=>{
      const li = document.createElement('li');
      const dot = document.createElement('span');
      dot.className='dot';
      const span = document.createElement('span');
      span.textContent = name;
      li.appendChild(dot);
      li.appendChild(span);
      ul.appendChild(li);
    });
    cnt.textContent = j.count||0;
  }catch(e){}
}
refresh();
setInterval(refresh, 5000); // 5秒ごと更新
</script>
</body>
</html>`)
	})

	log.Printf("server listening on :%s  (overlay: http://localhost:%s/overlay)", port, port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
