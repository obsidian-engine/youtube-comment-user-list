package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"slices"
	"strings"
	"sync"
	"time"
	// 標準net/httpで十分。外部SDKなし。
)

type LiveStreamingDetails struct {
	ActiveLiveChatID string `json:"activeLiveChatId"`
}
type VideosListResp struct {
	Items []struct {
		LiveStreamingDetails LiveStreamingDetails `json:"liveStreamingDetails"`
	} `json:"items"`
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
	u := fmt.Sprintf("https://www.googleapis.com/youtube/v3/videos?part=liveStreamingDetails&id=%s&key=%s", videoID, apiKey)
	resp, err := http.Get(u)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("videos.list status=%d", resp.StatusCode)
	}
	var v VideosListResp
	if err := json.NewDecoder(resp.Body).Decode(&v); err != nil {
		return "", err
	}
	if len(v.Items) == 0 || v.Items[0].LiveStreamingDetails.ActiveLiveChatID == "" {
		return "", fmt.Errorf("activeLiveChatId not found (配信が開始前/終了後の可能性)")
	}
	return v.Items[0].LiveStreamingDetails.ActiveLiveChatID, nil
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
	resp, err := http.Get(url)
	if err != nil {
		return LiveChatResp{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return LiveChatResp{}, fmt.Errorf("liveChatMessages.list status=%d", resp.StatusCode)
	}
	var v LiveChatResp
	if err := json.NewDecoder(resp.Body).Decode(&v); err != nil {
		return LiveChatResp{}, err
	}
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
		for {
			resp, err := fetchLiveChatOnce(apiKey, liveChatID, pageToken)
			if err != nil {
				log.Printf("fetch error: %v", err)
				time.Sleep(time.Duration(lastWait) * time.Millisecond)
				continue
			}
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
