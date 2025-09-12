package handler

import (
	"fmt"
	"net/http"

	"github.com/obsidian-engine/youtube-comment-user-list/internal/domain/service"
)

// StaticHandler static file serving and HTML pagesを処理します
type StaticHandler struct {
	logger service.Logger
}

// NewStaticHandler 新しいstaticを作成します handler
func NewStaticHandler(logger service.Logger) *StaticHandler {
	return &StaticHandler{
		logger: logger,
	}
}

// ServeHome handles GET /
func (h *StaticHandler) ServeHome(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	h.logger.LogAPI("INFO", "Home page request", "", "", map[string]interface{}{
		"userAgent":  r.Header.Get("User-Agent"),
		"remoteAddr": r.RemoteAddr,
	})

	html := `<!DOCTYPE html>
<html lang="ja">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>YouTube Live Chat Monitor</title>
    <style>
        body { 
            font-family: Arial, sans-serif; 
            max-width: 800px; 
            margin: 0 auto; 
            padding: 20px;
            background-color: #f5f5f5;
        }
        .container { 
            background: white; 
            padding: 30px; 
            border-radius: 8px; 
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
        }
        .form-group { 
            margin-bottom: 20px; 
        }
        label { 
            display: block; 
            margin-bottom: 8px; 
            font-weight: bold;
        }
        input[type="text"], input[type="number"] { 
            width: 100%; 
            padding: 10px; 
            border: 1px solid #ddd; 
            border-radius: 4px; 
            box-sizing: border-box;
        }
        button { 
            background: #4CAF50; 
            color: white; 
            padding: 12px 24px; 
            border: none; 
            border-radius: 4px; 
            cursor: pointer; 
            font-size: 16px;
        }
        button:hover { 
            background: #45a049; 
        }
        .links {
            margin-top: 30px;
            padding-top: 20px;
            border-top: 1px solid #ddd;
        }
        .links a {
            display: inline-block;
            margin-right: 15px;
            color: #007cba;
            text-decoration: none;
            padding: 8px 16px;
            border: 1px solid #007cba;
            border-radius: 4px;
        }
        .links a:hover {
            background-color: #007cba;
            color: white;
        }
        .error {
            color: red;
            margin-top: 10px;
        }
        .success {
            color: green;
            margin-top: 10px;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>YouTube Live Chat Monitor</h1>
        <p>YouTube Live配信のチャット参加者をリアルタイムで監視します。</p>
        
        <form id="monitoringForm">
            <div class="form-group">
                <label for="videoInput">YouTube Video URL または Video ID:</label>
                <input type="text" id="videoInput" name="videoInput" 
                       placeholder="例: https://www.youtube.com/watch?v=VIDEO_ID または VIDEO_ID"
                       required>
            </div>
            
            <div class="form-group">
                <label for="maxUsers">最大ユーザー数 (デフォルト: 1000):</label>
                <input type="number" id="maxUsers" name="maxUsers" 
                       value="1000" min="1" max="10000">
            </div>
            
            <button type="submit">監視開始</button>
        </form>
        
        <div id="message"></div>
        
        <div class="links">
            <h3>メニュー</h3>
            <a href="/logs">ログ表示</a>
            <a href="javascript:void(0)" onclick="showActiveVideos()">アクティブ動画一覧</a>
        </div>
        
        <div id="activeVideos" style="margin-top: 20px; display: none;">
            <h3>現在監視中の動画</h3>
            <div id="videoList"></div>
        </div>
    </div>

    <script>
        document.getElementById('monitoringForm').addEventListener('submit', async (e) => {
            e.preventDefault();
            
            const formData = new FormData(e.target);
            const videoInput = formData.get('videoInput');
            const maxUsers = parseInt(formData.get('maxUsers')) || 1000;
            
            const messageDiv = document.getElementById('message');
            messageDiv.innerHTML = '監視を開始しています...';
            
            try {
                const response = await fetch('/api/monitoring/start', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json',
                    },
                    body: JSON.stringify({
                        video_input: videoInput,
                        max_users: maxUsers
                    })
                });
                
                const data = await response.json();
                
                if (data.success) {
                    messageDiv.innerHTML = '<div class="success">監視を開始しました！<br>' +
                        '<a href="/users?video_id=' + data.video_id + '" target="_blank">ユーザーリストを表示</a></div>';
                } else {
                    messageDiv.innerHTML = '<div class="error">エラー: ' + data.error + '</div>';
                }
            } catch (error) {
                messageDiv.innerHTML = '<div class="error">通信エラー: ' + error.message + '</div>';
            }
        });
        
        async function showActiveVideos() {
            const activeDiv = document.getElementById('activeVideos');
            const videoListDiv = document.getElementById('videoList');
            
            try {
                const response = await fetch('/api/monitoring/active');
                const data = await response.json();
                
                if (data.success) {
                    if (data.videos.length === 0) {
                        videoListDiv.innerHTML = '<p>現在監視中の動画はありません。</p>';
                    } else {
                        let html = '<ul>';
                        data.videos.forEach(videoId => {
                            html += '<li>' +
                                '<strong>' + videoId + '</strong> - ' +
                                '<a href="/users?video_id=' + videoId + '" target="_blank">ユーザーリスト</a> | ' +
                                '<button onclick="stopMonitoring(\'' + videoId + '\')">監視停止</button>' +
                                '</li>';
                        });
                        html += '</ul>';
                        videoListDiv.innerHTML = html;
                    }
                    activeDiv.style.display = 'block';
                } else {
                    videoListDiv.innerHTML = '<div class="error">エラー: ' + data.error + '</div>';
                    activeDiv.style.display = 'block';
                }
            } catch (error) {
                videoListDiv.innerHTML = '<div class="error">通信エラー: ' + error.message + '</div>';
                activeDiv.style.display = 'block';
            }
        }
        
        async function stopMonitoring(videoId) {
            if (!confirm('動画 ' + videoId + ' の監視を停止しますか？')) {
                return;
            }
            
            try {
                const response = await fetch('/api/monitoring/stop/' + videoId, {
                    method: 'POST'
                });
                
                const data = await response.json();
                
                if (data.success) {
                    alert('監視を停止しました');
                    showActiveVideos(); // Refresh the list
                } else {
                    alert('エラー: ' + data.error);
                }
            } catch (error) {
                alert('通信エラー: ' + error.message);
            }
        }
    </script>
</body>
</html>`

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if _, err := w.Write([]byte(html)); err != nil {
		h.logger.LogError("ERROR", "Failed to write home page", "", "", err, nil)
	}
}

// ServeUserListPage handles GET /users
func (h *StaticHandler) ServeUserListPage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	videoID := r.URL.Query().Get("video_id")
	if videoID == "" {
		http.Error(w, "video_id parameter is required", http.StatusBadRequest)
		return
	}

	h.logger.LogAPI("INFO", "User list page request", videoID, "", map[string]interface{}{
		"userAgent": r.Header.Get("User-Agent"),
	})

	html := fmt.Sprintf(`<!DOCTYPE html>
<html lang="ja">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>ユーザーリスト - %s</title>
    <style>
        body { 
            font-family: Arial, sans-serif; 
            max-width: 1200px; 
            margin: 0 auto; 
            padding: 20px;
            background-color: #f5f5f5;
        }
        .container { 
            background: white; 
            padding: 30px; 
            border-radius: 8px; 
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
        }
        .header {
            display: flex;
            justify-content: space-between;
            align-items: center;
            margin-bottom: 20px;
            padding-bottom: 15px;
            border-bottom: 1px solid #ddd;
        }
        .user-grid {
            display: grid;
            grid-template-columns: repeat(auto-fill, minmax(250px, 1fr));
            gap: 15px;
            margin-top: 20px;
        }
        .user-card {
            background: #f9f9f9;
            padding: 15px;
            border-radius: 6px;
            border-left: 4px solid #4CAF50;
        }
        .user-name {
            font-weight: bold;
            margin-bottom: 5px;
        }
        .user-id {
            font-size: 0.9em;
            color: #666;
        }
        .stats {
            background: #e7f3ff;
            padding: 15px;
            border-radius: 6px;
            margin-bottom: 20px;
        }
        .nav-links a {
            color: #007cba;
            text-decoration: none;
            margin-right: 15px;
        }
        .nav-links a:hover {
            text-decoration: underline;
        }
        .status {
            display: inline-block;
            padding: 4px 8px;
            border-radius: 4px;
            font-size: 0.8em;
            margin-left: 10px;
        }
        .status.connected { background: #d4edda; color: #155724; }
        .status.disconnected { background: #f8d7da; color: #721c24; }
        .status.loading { background: #fff3cd; color: #856404; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>ユーザーリスト</h1>
            <div class="nav-links">
                <a href="/">ホーム</a>
                <a href="/logs">ログ</a>
                <a href="javascript:void(0)" onclick="refreshUsers()">更新</a>
            </div>
        </div>
        
        <div class="stats">
            <strong>Video ID:</strong> %s
            <span id="connectionStatus" class="status loading">接続中...</span>
            <br>
            <strong>ユーザー数:</strong> <span id="userCount">読み込み中...</span>
        </div>
        
        <div id="userGrid" class="user-grid">
            読み込み中...
        </div>
    </div>

    <script>
        const videoId = '%s';
        let eventSource;
        let userCount = 0;
        
        function connectSSE() {
            eventSource = new EventSource('/api/sse/' + videoId + '/users');
            
            eventSource.onopen = function() {
                document.getElementById('connectionStatus').className = 'status connected';
                document.getElementById('connectionStatus').textContent = '接続済み';
            };
            
            eventSource.onmessage = function(event) {
                try {
                    const data = JSON.parse(event.data);
                    if (data.type === 'user_list') {
                        displayUsers(data.data.users, data.data.count);
                    }
                } catch (error) {
                    console.error('SSE message parse error:', error);
                }
            };
            
            eventSource.onerror = function() {
                document.getElementById('connectionStatus').className = 'status disconnected';
                document.getElementById('connectionStatus').textContent = '接続エラー';
                
                // Retry connection after 5 seconds
                setTimeout(() => {
                    if (eventSource.readyState === EventSource.CLOSED) {
                        connectSSE();
                    }
                }, 5000);
            };
        }
        
        function displayUsers(users, count) {
            userCount = count;
            document.getElementById('userCount').textContent = count;
            
            const userGrid = document.getElementById('userGrid');
            
            if (!users || users.length === 0) {
                userGrid.innerHTML = '<p>まだユーザーが参加していません。</p>';
                return;
            }
            
            let html = '';
            users.forEach(user => {
                html += '<div class="user-card">' +
                    '<div class="user-name">' + escapeHtml(user.DisplayName) + '</div>' +
                    '<div class="user-id">' + escapeHtml(user.ChannelID) + '</div>' +
                    '</div>';
            });
            
            userGrid.innerHTML = html;
        }
        
        function escapeHtml(text) {
            const div = document.createElement('div');
            div.textContent = text;
            return div.innerHTML;
        }
        
        async function refreshUsers() {
            try {
                const response = await fetch('/api/monitoring/' + videoId + '/users');
                const data = await response.json();
                
                if (data.success) {
                    displayUsers(data.users, data.count);
                } else {
                    alert('エラー: ' + data.error);
                }
            } catch (error) {
                alert('通信エラー: ' + error.message);
            }
        }
        
        // Initialize SSE connection
        connectSSE();
        
        // Cleanup on page unload
        window.addEventListener('beforeunload', function() {
            if (eventSource) {
                eventSource.close();
            }
        });
    </script>
</body>
</html>`, videoID, videoID, videoID)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if _, err := w.Write([]byte(html)); err != nil {
		h.logger.LogError("ERROR", "Failed to write user list page", videoID, "", err, nil)
	}
}

// ServeLogsPage handles GET /logs
func (h *StaticHandler) ServeLogsPage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	h.logger.LogAPI("INFO", "Logs page request", "", "", map[string]interface{}{
		"userAgent": r.Header.Get("User-Agent"),
	})

	html := `<!DOCTYPE html>
<html lang="ja">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>ログ表示</title>
    <style>
        body { 
            font-family: Arial, sans-serif; 
            max-width: 1400px; 
            margin: 0 auto; 
            padding: 20px;
            background-color: #f5f5f5;
        }
        .container { 
            background: white; 
            padding: 30px; 
            border-radius: 8px; 
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
        }
        .header {
            display: flex;
            justify-content: space-between;
            align-items: center;
            margin-bottom: 20px;
            padding-bottom: 15px;
            border-bottom: 1px solid #ddd;
        }
        .controls {
            margin-bottom: 20px;
            background: #f9f9f9;
            padding: 15px;
            border-radius: 6px;
        }
        .controls label {
            display: inline-block;
            width: 120px;
            margin-right: 10px;
        }
        .controls select, .controls input {
            padding: 5px;
            margin-right: 15px;
        }
        .controls button {
            padding: 6px 12px;
            margin-right: 10px;
            border: 1px solid #ddd;
            background: white;
            cursor: pointer;
        }
        .controls button:hover {
            background: #f0f0f0;
        }
        .log-entry {
            margin-bottom: 10px;
            padding: 10px;
            border-left: 4px solid #ddd;
            background: #fafafa;
            font-family: monospace;
            font-size: 0.9em;
        }
        .log-entry.INFO { border-left-color: #4CAF50; }
        .log-entry.ERROR { border-left-color: #f44336; background: #fff5f5; }
        .log-entry.DEBUG { border-left-color: #2196F3; }
        .log-entry.WARN { border-left-color: #ff9800; background: #fffbf0; }
        .log-meta {
            color: #666;
            font-size: 0.8em;
            margin-bottom: 5px;
        }
        .log-context {
            background: #eee;
            padding: 5px;
            margin-top: 5px;
            border-radius: 3px;
            cursor: pointer;
        }
        .nav-links a {
            color: #007cba;
            text-decoration: none;
            margin-right: 15px;
        }
        .nav-links a:hover {
            text-decoration: underline;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>ログ表示</h1>
            <div class="nav-links">
                <a href="/">ホーム</a>
                <button onclick="refreshLogs()">更新</button>
                <button onclick="clearLogs()">ログクリア</button>
                <button onclick="exportLogs()">エクスポート</button>
            </div>
        </div>
        
        <div class="controls">
            <label>レベル:</label>
            <select id="levelFilter">
                <option value="">全て</option>
                <option value="ERROR">ERROR</option>
                <option value="WARN">WARN</option>
                <option value="INFO">INFO</option>
                <option value="DEBUG">DEBUG</option>
            </select>
            
            <label>コンポーネント:</label>
            <select id="componentFilter">
                <option value="">全て</option>
                <option value="api">API</option>
                <option value="poller">Poller</option>
                <option value="user">User</option>
                <option value="error">Error</option>
            </select>
            
            <label>Video ID:</label>
            <input type="text" id="videoIdFilter" placeholder="Video ID">
            
            <label>件数:</label>
            <select id="limitFilter">
                <option value="50">50件</option>
                <option value="100" selected>100件</option>
                <option value="200">200件</option>
                <option value="500">500件</option>
            </select>
            
            <button onclick="applyFilters()">フィルター適用</button>
        </div>
        
        <div id="logEntries">
            読み込み中...
        </div>
    </div>

    <script>
        async function loadLogs() {
            const level = document.getElementById('levelFilter').value;
            const component = document.getElementById('componentFilter').value;
            const videoId = document.getElementById('videoIdFilter').value;
            const limit = document.getElementById('limitFilter').value;
            
            let url = '/api/logs?';
            const params = new URLSearchParams();
            if (level) params.append('level', level);
            if (component) params.append('component', component);
            if (videoId) params.append('video_id', videoId);
            if (limit) params.append('limit', limit);
            
            try {
                const response = await fetch('/api/logs?' + params.toString());
                const data = await response.json();
                
                if (data.success) {
                    displayLogs(data.logs);
                } else {
                    document.getElementById('logEntries').innerHTML = 
                        '<div class="log-entry ERROR">エラー: ' + data.error + '</div>';
                }
            } catch (error) {
                document.getElementById('logEntries').innerHTML = 
                    '<div class="log-entry ERROR">通信エラー: ' + error.message + '</div>';
            }
        }
        
        function displayLogs(logs) {
            const container = document.getElementById('logEntries');
            
            if (!logs || logs.length === 0) {
                container.innerHTML = '<p>ログがありません。</p>';
                return;
            }
            
            let html = '';
            logs.forEach(log => {
                html += '<div class="log-entry ' + log.level + '">' +
                    '<div class="log-meta">' +
                    '[' + log.timestamp + '] ' + log.level + 
                    (log.component ? ' [' + log.component + ']' : '') +
                    (log.event ? ' [' + log.event + ']' : '') +
                    (log.video_id ? ' (Video: ' + log.video_id + ')' : '') +
                    (log.correlation_id ? ' (ID: ' + log.correlation_id + ')' : '') +
                    '</div>' +
                    '<div>' + escapeHtml(log.message) + '</div>';
                
                if (log.context && Object.keys(log.context).length > 0) {
                    html += '<div class="log-context" onclick="toggleContext(this)">' +
                        'Context (クリックで展開): ' + Object.keys(log.context).length + ' items' +
                        '<pre style="display: none; margin: 5px 0 0 0;">' + 
                        JSON.stringify(log.context, null, 2) + '</pre>' +
                        '</div>';
                }
                
                html += '</div>';
            });
            
            container.innerHTML = html;
        }
        
        function toggleContext(element) {
            const pre = element.querySelector('pre');
            if (pre.style.display === 'none') {
                pre.style.display = 'block';
            } else {
                pre.style.display = 'none';
            }
        }
        
        function escapeHtml(text) {
            const div = document.createElement('div');
            div.textContent = text;
            return div.innerHTML;
        }
        
        function refreshLogs() {
            loadLogs();
        }
        
        function applyFilters() {
            loadLogs();
        }
        
        async function clearLogs() {
            if (!confirm('全てのログをクリアしますか？')) {
                return;
            }
            
            try {
                const response = await fetch('/api/logs', { method: 'DELETE' });
                const data = await response.json();
                
                if (data.success) {
                    alert('ログをクリアしました');
                    loadLogs();
                } else {
                    alert('エラー: ' + data.error);
                }
            } catch (error) {
                alert('通信エラー: ' + error.message);
            }
        }
        
        async function exportLogs() {
            const level = document.getElementById('levelFilter').value;
            const component = document.getElementById('componentFilter').value;
            const videoId = document.getElementById('videoIdFilter').value;
            const limit = document.getElementById('limitFilter').value;
            
            const params = new URLSearchParams();
            if (level) params.append('level', level);
            if (component) params.append('component', component);
            if (videoId) params.append('video_id', videoId);
            if (limit) params.append('limit', limit);
            
            window.open('/api/logs/export?' + params.toString());
        }
        
        // Initial load
        loadLogs();
        
        // Auto refresh every 30 seconds
        setInterval(loadLogs, 30000);
    </script>
</body>
</html>`

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if _, err := w.Write([]byte(html)); err != nil {
		h.logger.LogError("ERROR", "Failed to write logs page", "", "", err, nil)
	}
}
