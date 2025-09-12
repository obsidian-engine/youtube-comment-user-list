package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/obsidian-engine/youtube-comment-user-list/internal/domain/service"
)

// StaticHandler 静的ファイル配信とHTMLページを処理します
type StaticHandler struct {
	logger service.Logger
}

// NewStaticHandler 新しい静的ハンドラーを作成します
func NewStaticHandler(logger service.Logger) *StaticHandler {
	return &StaticHandler{
		logger: logger,
	}
}

// ServeHome GET / を処理します
func (h *StaticHandler) ServeHome(c *gin.Context) {
	h.logger.LogAPI("INFO", "Home page request", "", "", map[string]interface{}{
		"userAgent":  c.GetHeader("User-Agent"),
		"remoteAddr": c.ClientIP(),
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
                    messageDiv.innerHTML = '<div class="success">監視を開始しました！ユーザーリストページに移動します...</div>';
                    // 1秒後にユーザーリストページに自動遷移
                    setTimeout(() => {
                        window.location.href = '/users';
                    }, 1000);
                } else {
                    messageDiv.innerHTML = '<div class="error">エラー: ' + data.error + '</div>';
                }
            } catch (error) {
                messageDiv.innerHTML = '<div class="error">通信エラー: ' + error.message + '</div>';
            }
        });
        
        // ウィンドウクローズ時にサーバー停止処理
        window.addEventListener('beforeunload', function(e) {
            // 同期的にサーバー停止をリクエスト
            navigator.sendBeacon('/api/monitoring/stop', new FormData());
        });

    </script>
</body>
</html>`

	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, html)
}

// ServeUserListPage GET /users を処理します
func (h *StaticHandler) ServeUserListPage(c *gin.Context) {
	h.logger.LogAPI("INFO", "User list page request", "", "", map[string]interface{}{
		"userAgent":  c.GetHeader("User-Agent"),
		"remoteAddr": c.ClientIP(),
	})

	html := `<!DOCTYPE html>
<html lang="ja">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>User List - YouTube Live Chat Monitor</title>
    <style>
        body { 
            font-family: Arial, sans-serif; 
            margin: 20px;
            background-color: #f5f5f5;
        }
        .container { 
            background: white; 
            padding: 20px; 
            border-radius: 8px; 
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
        }
        .user-list { 
            margin-top: 20px; 
        }
        .user { 
            border: 1px solid #ddd; 
            margin: 5px 0; 
            padding: 10px; 
            border-radius: 4px;
        }
        .status { 
            margin: 20px 0; 
            padding: 10px; 
            border-radius: 4px;
        }
        .online { 
            background-color: #d4edda; 
            color: #155724; 
        }
        .offline { 
            background-color: #f8d7da; 
            color: #721c24; 
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>ユーザーリスト</h1>
        <p><a href="/">← ホームに戻る</a></p>
        
        <div class="form-group">
            <button onclick="loadUsers()">ユーザーリスト取得</button>
            <button onclick="stopMonitoring()" style="margin-left: 10px;">監視停止</button>
        </div>
        
        <div id="status"></div>
        <div id="userList"></div>
    </div>

    <script>
        // ページ読み込み時に自動でユーザーリストを取得
        window.onload = function() {
            loadUsers();
        };

        async function loadUsers() {
            
            const statusDiv = document.getElementById('status');
            const userListDiv = document.getElementById('userList');
            
            statusDiv.innerHTML = '<div class="status">読み込み中...</div>';
            userListDiv.innerHTML = '';
            
            try {
                // 現在アクティブなvideoIDを取得
                const activeResponse = await fetch('/api/monitoring/active');
                
                if (!activeResponse.ok) {
                    if (activeResponse.status === 404) {
                        statusDiv.innerHTML = '<div class="status offline">監視セッションが開始されていません</div>';
                        return;
                    }
                    throw new Error('Failed to get active video ID');
                }
                
                const activeData = await activeResponse.json();
                const videoId = activeData.videoId;
                
                // videoIDを使ってユーザーリストを取得
                const response = await fetch('/api/monitoring/' + videoId + '/users');
                const data = await response.json();
                
                if (data.success) {
                    statusDiv.innerHTML = '<div class="status online">オンライン - ユーザー数: ' + data.count + '</div>';
                    
                    if (data.users && data.users.length > 0) {
                        let html = '<div class="user-list">';
                        data.users.forEach(user => {
                            html += '<div class="user">' +
                                '<strong>' + user.display_name + '</strong><br>' +
                                'Channel: ' + user.channel_id + '<br>' +
                                '初回参加: ' + new Date(user.first_seen).toLocaleString() +
                                '</div>';
                        });
                        html += '</div>';
                        userListDiv.innerHTML = html;
                    } else {
                        userListDiv.innerHTML = '<p>まだユーザーが参加していません。</p>';
                    }
                } else {
                    statusDiv.innerHTML = '<div class="status offline">エラー: ' + data.error + '</div>';
                }
            } catch (error) {
                statusDiv.innerHTML = '<div class="status offline">通信エラー: ' + error.message + '</div>';
            }
        }

        async function stopMonitoring() {
            if (!confirm('監視を停止しますか？')) {
                return;
            }
            
            try {
                const response = await fetch('/api/monitoring/stop', {
                    method: 'DELETE'
                });
                
                const data = await response.json();
                
                if (data.success) {
                    alert('監視を停止しました。ホームページに戻ります。');
                    window.location.href = '/';
                } else {
                    alert('エラー: ' + data.error);
                }
            } catch (error) {
                alert('通信エラー: ' + error.message);
            }
        }

        // ウィンドウクローズ時にサーバー停止処理
        window.addEventListener('beforeunload', function(e) {
            // 同期的にサーバー停止をリクエスト
            navigator.sendBeacon('/api/monitoring/stop', new FormData());
        });

        // 10秒ごとに自動更新
        setInterval(() => {
            loadUsers();
        }, 10000);
    </script>
</body>
</html>`

	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, html)
}

// ServeLogsPage GET /logs を処理します
func (h *StaticHandler) ServeLogsPage(c *gin.Context) {
	h.logger.LogAPI("INFO", "Logs page request", "", "", map[string]interface{}{
		"userAgent":  c.GetHeader("User-Agent"),
		"remoteAddr": c.ClientIP(),
	})

	html := `<!DOCTYPE html>
<html lang="ja">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>System Logs - YouTube Live Chat Monitor</title>
    <style>
        body { 
            font-family: 'Courier New', monospace; 
            margin: 20px;
            background-color: #f5f5f5;
        }
        .container { 
            background: white; 
            padding: 20px; 
            border-radius: 8px; 
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
        }
        .controls {
            margin-bottom: 20px;
            padding: 15px;
            background-color: #f8f9fa;
            border-radius: 4px;
        }
        .log-entry { 
            margin: 5px 0; 
            padding: 5px; 
            border-left: 3px solid #ddd;
            font-size: 12px;
        }
        .log-info { 
            border-left-color: #007bff; 
        }
        .log-error { 
            border-left-color: #dc3545; 
            background-color: #f8d7da;
        }
        .log-warning { 
            border-left-color: #ffc107; 
            background-color: #fff3cd;
        }
        button {
            margin: 5px;
            padding: 8px 16px;
            border: none;
            border-radius: 4px;
            cursor: pointer;
        }
        .btn-primary { 
            background-color: #007bff; 
            color: white; 
        }
        .btn-danger { 
            background-color: #dc3545; 
            color: white; 
        }
        .btn-success { 
            background-color: #28a745; 
            color: white; 
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>システムログ</h1>
        <p><a href="/">← ホームに戻る</a></p>
        
        <div class="controls">
            <button class="btn-primary" onclick="loadLogs()">ログ更新</button>
            <button class="btn-danger" onclick="clearLogs()">ログクリア</button>
            <button class="btn-success" onclick="exportLogs()">ログエクスポート</button>
            <label>
                <input type="checkbox" id="autoRefresh" checked> 自動更新 (5秒間隔)
            </label>
        </div>
        
        <div id="logStats"></div>
        <div id="logContainer"></div>
    </div>

    <script>
        let autoRefreshInterval;

        document.getElementById('autoRefresh').addEventListener('change', function(e) {
            if (e.target.checked) {
                startAutoRefresh();
            } else {
                stopAutoRefresh();
            }
        });

        function startAutoRefresh() {
            autoRefreshInterval = setInterval(loadLogs, 5000);
        }

        function stopAutoRefresh() {
            if (autoRefreshInterval) {
                clearInterval(autoRefreshInterval);
            }
        }

        async function loadLogs() {
            try {
                const response = await fetch('/api/logs');
                const data = await response.json();
                
                const logContainer = document.getElementById('logContainer');
                
                if (data.success && data.logs) {
                    let html = '';
                    data.logs.forEach(log => {
                        let logClass = 'log-info';
                        if (log.level === 'ERROR') logClass = 'log-error';
                        else if (log.level === 'WARNING') logClass = 'log-warning';
                        
                        html += '<div class="log-entry ' + logClass + '">' +
                            '<strong>' + log.timestamp + '</strong> [' + log.level + '] ' +
                            log.message + (log.video_id ? ' (Video: ' + log.video_id + ')' : '') +
                            '</div>';
                    });
                    logContainer.innerHTML = html;
                } else {
                    logContainer.innerHTML = '<p>ログがありません。</p>';
                }
                
                // 統計情報を読み込み
                loadLogStats();
            } catch (error) {
                document.getElementById('logContainer').innerHTML = 
                    '<div class="log-entry log-error">通信エラー: ' + error.message + '</div>';
            }
        }

        async function loadLogStats() {
            try {
                const response = await fetch('/api/logs/stats');
                const data = await response.json();
                
                if (data.success) {
                    document.getElementById('logStats').innerHTML = 
                        '<p><strong>ログ統計:</strong> 総数: ' + data.total + 
                        ', エラー: ' + data.errors + ', 警告: ' + data.warnings + '</p>';
                }
            } catch (error) {
                console.error('Stats loading error:', error);
            }
        }

        async function clearLogs() {
            if (!confirm('すべてのログをクリアしますか？')) {
                return;
            }
            
            try {
                const response = await fetch('/api/logs', {
                    method: 'DELETE'
                });
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
            try {
                const response = await fetch('/api/logs/export');
                const blob = await response.blob();
                
                const url = window.URL.createObjectURL(blob);
                const a = document.createElement('a');
                a.style.display = 'none';
                a.href = url;
                a.download = 'system_logs_' + new Date().toISOString().split('T')[0] + '.json';
                document.body.appendChild(a);
                a.click();
                window.URL.revokeObjectURL(url);
            } catch (error) {
                alert('エクスポートエラー: ' + error.message);
            }
        }

        // 初期化
        loadLogs();
        startAutoRefresh();
    </script>
</body>
</html>`

	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, html)
}
