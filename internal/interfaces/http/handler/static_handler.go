package handler

import (
	"net/http"

	"github.com/obsidian-engine/youtube-comment-user-list/internal/domain/repository"
)

// StaticHandler 静的ファイル配信とHTMLページを処理します
type StaticHandler struct {
	logger repository.Logger
}

// NewStaticHandler 新しい静的ハンドラーを作成します
func NewStaticHandler(logger repository.Logger) *StaticHandler {
	return &StaticHandler{
		logger: logger,
	}
}

// ServeHome GET / を処理します
func (h *StaticHandler) ServeHome(w http.ResponseWriter, r *http.Request) {
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

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte(html)); err != nil {
		h.logger.LogError("ERROR", "Failed to write HTML response", "", "", err, nil)
	}
}

// ServeUserListPage GET /users を処理します
func (h *StaticHandler) ServeUserListPage(w http.ResponseWriter, r *http.Request) {
	h.logger.LogAPI("INFO", "User list page request", "", "", map[string]interface{}{
		"userAgent":  r.Header.Get("User-Agent"),
		"remoteAddr": r.RemoteAddr,
	})

	html := `<!DOCTYPE html>
<html lang="ja">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>User List - YouTube Live Chat Monitor</title>
    <style>
        :root{
            --bg:#0f172a; /* slate-900 */
            --panel:#111827; /* gray-900 */
            --panel-2:#0b1222;
            --text:#e5e7eb; /* gray-200 */
            --muted:#94a3b8; /* slate-400 */
            --accent:#22d3ee; /* cyan-400 */
            --accent-2:#38bdf8; /* sky-400 */
            --danger:#fda4af; /* rose-300 */
            --ok:#86efac; /* green-300 */
            --border:#1f2937; /* gray-800 */
            --row:#0b1222;
            --row-alt:#0d1426;
            --badge:#1f2937;
        }
        *{box-sizing:border-box}
        body{margin:0;background:var(--bg);color:var(--text);font-family: ui-sans-serif, system-ui, -apple-system, Segoe UI, Roboto, "Helvetica Neue", Arial, "Noto Sans", "Apple Color Emoji", "Segoe UI Emoji"}
        a{color:var(--accent)}
        .wrap{max-width:1100px;margin:24px auto;padding:0 16px}
        header{display:flex;align-items:center;gap:12px;margin-bottom:16px}
        .title{font-size:22px;font-weight:700}
        .sub{color:var(--muted);font-size:13px}
        .card{background:linear-gradient(180deg,var(--panel),var(--panel-2));border:1px solid var(--border);border-radius:12px;box-shadow:0 6px 30px rgba(0,0,0,.25)}
        .toolbar{display:flex;flex-wrap:wrap;gap:12px;align-items:center;justify-content:space-between;padding:14px 16px;border-bottom:1px solid var(--border)}
        .left, .right{display:flex;gap:12px;align-items:center}
        .controls input[type="text"]{background:#0b1222;border:1px solid var(--border);border-radius:8px;color:var(--text);padding:10px 12px;min-width:220px}
        .controls select, .controls button, .controls .toggle{background:#0b1222;border:1px solid var(--border);border-radius:8px;color:var(--text);padding:10px 12px}
        .controls button{cursor:pointer}
        .status{padding:10px 12px;border-radius:8px;font-size:13px}
        .status.online{background:rgba(34,211,238,.08);color:#67e8f9;border:1px solid rgba(103,232,249,.25)}
        .status.offline{background:rgba(244,63,94,.08);color:#fda4af;border:1px solid rgba(253,164,175,.25)}
        .meta{color:var(--muted);font-size:12px}
        .content{padding:12px 16px}
        table{width:100%;border-collapse:separate;border-spacing:0 8px}
        thead th{font-size:12px;text-transform:uppercase;letter-spacing:.06em;color:var(--muted);text-align:left;padding:0 10px}
        tbody tr{background:var(--row);border:1px solid var(--border)}
        tbody tr:nth-child(even){background:var(--row-alt)}
        td{padding:12px 10px;vertical-align:middle}
        .idx{color:var(--muted);width:56px}
        .name{display:flex;align-items:center;gap:10px}
        .avatar{width:28px;height:28px;border-radius:50%;background:#1f2937;display:inline-flex;align-items:center;justify-content:center;color:#cbd5e1;font-weight:700}
        .badge{font-size:11px;color:#a5b4fc;background:rgba(99,102,241,.12);border:1px solid rgba(99,102,241,.3);padding:2px 8px;border-radius:999px}
        .pill{font-size:12px;color:#bae6fd;background:rgba(56,189,248,.12);border:1px solid rgba(56,189,248,.3);padding:4px 10px;border-radius:999px}
        .actions button{background:#0b1222;border:1px solid var(--border);color:var(--text);border-radius:8px;padding:6px 10px;cursor:pointer}
        .empty{padding:24px;text-align:center;color:var(--muted)}
        .footer{display:flex;justify-content:space-between;align-items:center;padding:12px 16px;border-top:1px solid var(--border)}
        .btn{background:linear-gradient(180deg,#0ea5e9,#0891b2);border:0;color:#e6faff;padding:10px 14px;border-radius:10px;cursor:pointer}
        @media (max-width:720px){
            thead{display:none}
            table, tbody, tr, td{display:block;width:100%}
            tbody tr{margin:8px 0;border-radius:10px}
            td{padding:8px 12px}
            .idx{display:none}
        }
    </style>
</head>
<body>
    <div class="wrap">
        <header>
            <div class="title">ユーザーリスト</div>
            <div class="sub">YouTube Live Chat 参加者</div>
            <div style="margin-left:auto"><a href="/">ホームに戻る →</a></div>
        </header>

        <div class="card">
            <div class="toolbar">
                <div class="left controls">
                    <input id="search" type="text" placeholder="名前・Channel IDで検索" oninput="renderUsers()"/>
                    <select id="sort" onchange="renderUsers()">
                        <option value="first_seen_desc">新しい順</option>
                        <option value="first_seen_asc">古い順</option>
                        <option value="name_asc">名前 A→Z</option>
                        <option value="name_desc">名前 Z→A</option>
                    </select>
                    <button class="btn" onclick="loadUsers()">更新</button>
                    <button onclick="stopMonitoring()">監視停止</button>
                </div>
                <div class="right">
                    <div id="status" class="status">読み込み中...</div>
                </div>
            </div>
            <div class="content">
                <div class="meta" style="margin-bottom:8px">
                    <span id="count">0</span> 名 <span id="updated"></span>
                    <span style="margin-left:12px">自動更新: 
                        <label class="toggle"><input id="auto" type="checkbox" checked onchange="toggleAuto()"> ON</label>
                        <select id="interval" onchange="resetAuto()">
                            <option value="5000">5秒</option>
                            <option value="10000" selected>10秒</option>
                            <option value="30000">30秒</option>
                        </select>
                    </span>
                </div>
                <div id="userList">
                    <div class="empty">データを読み込んでいます…</div>
                </div>
            </div>
            <div class="footer">
                <div class="meta">ChannelをクリックするとYouTubeチャンネルを開きます</div>
                <div><a href="/">← ホーム</a></div>
            </div>
        </div>
    </div>

    <script>
        let consecutiveErrors = 0;
        const MAX_CONSECUTIVE_ERRORS = 3;
        let autoUpdateInterval = null;
        let cachedUsers = [];
        let isActive = false;

        window.onload = function() {
            loadUsers();
            startAutoUpdate();
        };

        function fmtDate(s){
            try{ return new Date(s).toLocaleString(); }catch(e){ return '-'; }
        }
        function initials(name){
            if(!name) return '?';
            return name.trim().split(/\s+/).map(p=>p[0]).join('').substring(0,2).toUpperCase();
        }

        async function loadUsers() {
            const statusDiv = document.getElementById('status');
            const updated = document.getElementById('updated');
            statusDiv.textContent = '読み込み中...';
            try {
                const activeResponse = await fetch('/api/monitoring/active');
                if (!activeResponse.ok) {
                    if (activeResponse.status === 404) {
                        consecutiveErrors++;
                        isActive = false;
                        statusDiv.className = 'status offline';
                        statusDiv.textContent = '監視セッションがありません';
                        document.getElementById('userList').innerHTML = '<div class="empty">監視を開始するにはホームに戻ってください。</div>';
                        if (consecutiveErrors >= MAX_CONSECUTIVE_ERRORS) {
                            stopAutoUpdate();
                            statusDiv.innerHTML += '  自動更新を停止しました';
                        }
                        return;
                    }
                    throw new Error('Failed to get active video info (status ' + activeResponse.status + ')');
                }
                const activeData = await activeResponse.json();
                const videoId = (activeData.data && activeData.data.videoId) || activeData.videoId;
                isActive = (activeData.data && typeof activeData.data.isActive !== 'undefined') ? activeData.data.isActive : activeData.isActive;
                if (!videoId) {
                    throw new Error('Active videoId not found in response');
                }
                const response = await fetch('/api/monitoring/' + encodeURIComponent(videoId) + '/users');
                if (!response.ok) {
                    throw new Error('Failed to get user list (status ' + response.status + ')');
                }
                const data = await response.json();
                if (data.success) {
                    consecutiveErrors = 0;
                    cachedUsers = Array.isArray(data.users) ? data.users : [];
                    const cls = isActive ? 'status online' : 'status offline';
                    const txt = isActive ? 'オンライン' : '停止済み';
                    statusDiv.className = cls; statusDiv.textContent = txt + ' - ユーザー数: ' + (data.count ?? cachedUsers.length);
                    updated.textContent = '（更新: ' + new Date().toLocaleTimeString() + '）';
                    renderUsers();
                } else {
                    statusDiv.className = 'status offline';
                    statusDiv.textContent = 'エラー: ' + (data.error || 'unknown');
                }
            } catch (error) {
                statusDiv.className = 'status offline';
                statusDiv.textContent = '通信エラー: ' + error.message;
            }
        }

        function renderUsers(){
            const list = document.getElementById('userList');
            const q = (document.getElementById('search').value || '').toLowerCase();
            const sort = document.getElementById('sort').value;
            let users = cachedUsers.slice();
            if(q){
                users = users.filter(u => {
                    const name = (u.display_name || u.displayName || '').toLowerCase();
                    const cid = (u.channel_id || u.channelID || '').toLowerCase();
                    return name.includes(q) || cid.includes(q);
                });
            }
            users.sort((a,b)=>{
                const nameA = (a.display_name || a.displayName || '').toLowerCase();
                const nameB = (b.display_name || b.displayName || '').toLowerCase();
                const tA = new Date(a.first_seen || a.firstSeen || 0).getTime();
                const tB = new Date(b.first_seen || b.firstSeen || 0).getTime();
                switch(sort){
                    case 'first_seen_asc': return tA - tB;
                    case 'name_asc': return nameA.localeCompare(nameB);
                    case 'name_desc': return nameB.localeCompare(nameA);
                    default: return tB - tA; // first_seen_desc
                }
            });
            document.getElementById('count').textContent = users.length;
            if(users.length === 0){
                list.innerHTML = '<div class="empty">該当するユーザーがいません</div>';
                return;
            }
            let html = '';
            html += '<table><thead><tr>'+
                    '<th class="idx">#</th>'+
                    '<th>ユーザー名</th>'+
                    '<th>Channel ID</th>'+
                    '<th>初回参加</th>'+
                    '<th></th>'+
                '</tr></thead><tbody>';
            users.forEach((u,i)=>{
                const name = (u.display_name || u.displayName || '');
                const cid = (u.channel_id || u.channelID || '');
                const first = fmtDate(u.first_seen || u.firstSeen);
                const init = initials(name);
                const url = cid ? 'https://www.youtube.com/channel/' + encodeURIComponent(cid) : '#';
                html += '<tr>'+
                    '<td class="idx">'+(i+1)+'</td>'+
                    '<td class="name"><span class="avatar">'+init+'</span><div><div>'+escapeHtml(name)+'</div><div class="badge">Live</div></div></td>'+
                    '<td><a href="'+url+'" target="_blank" rel="noopener">'+escapeHtml(cid)+'</a></td>'+
                    '<td><span class="pill">'+first+'</span></td>'+
                    '<td class="actions"><button onclick="copy(\''+cid+'\')">Copy</button></td>'+
                '</tr>';
            });
            html += '</tbody></table>';
            list.innerHTML = html;
        }

        function escapeHtml(s){
            return (s||'').replace(/[&<>"']/g, c=>({
                '&':'&amp;','<':'&lt;','>':'&gt;','"':'&quot;','\'':'&#039;'
            })[c]);
        }

        function copy(text){
            if(!text) return;
            navigator.clipboard?.writeText(text).then(()=>{
                toast('Channel IDをコピーしました');
            }).catch(()=>{
                const ta = document.createElement('textarea');
                ta.value = text; document.body.appendChild(ta); ta.select(); document.execCommand('copy'); document.body.removeChild(ta);
                toast('Channel IDをコピーしました');
            });
        }

        let toastTimer=null;
        function toast(msg){
            clearTimeout(toastTimer);
            let el = document.getElementById('toast');
            if(!el){
                el = document.createElement('div');
                el.id='toast';
                el.style.position='fixed';el.style.right='16px';el.style.bottom='16px';el.style.padding='10px 14px';
                el.style.background='#0ea5e9';el.style.color='#082f49';el.style.borderRadius='10px';el.style.boxShadow='0 10px 30px rgba(0,0,0,.35)';
                document.body.appendChild(el);
            }
            el.textContent = msg; el.style.opacity='1';
            toastTimer = setTimeout(()=>{ el.style.opacity='0'; }, 1600);
        }

        function startAutoUpdate() {
            if (autoUpdateInterval) clearInterval(autoUpdateInterval);
            if (!document.getElementById('auto').checked) return;
            const ms = parseInt(document.getElementById('interval').value || '10000', 10);
            autoUpdateInterval = setInterval(loadUsers, ms);
        }
        function stopAutoUpdate() { if (autoUpdateInterval) { clearInterval(autoUpdateInterval); autoUpdateInterval = null; } }
        function resetAuto(){ stopAutoUpdate(); startAutoUpdate(); }
        function toggleAuto(){ if(document.getElementById('auto').checked){ startAutoUpdate(); } else { stopAutoUpdate(); } }

        async function stopMonitoring() {
            if (!confirm('監視を停止しますか？')) { return; }
            try {
                const response = await fetch('/api/monitoring/stop', { method: 'DELETE' });
                const data = await response.json();
                if (data.success) { alert('監視を停止しました。ホームに戻ります。'); window.location.href = '/'; }
                else { alert('エラー: ' + data.error); }
            } catch (error) { alert('通信エラー: ' + error.message); }
        }

        window.addEventListener('beforeunload', function(){ stopAutoUpdate(); navigator.sendBeacon('/api/monitoring/stop', new FormData()); });
    </script>
</body>
</html>`

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte(html)); err != nil {
		h.logger.LogError("ERROR", "Failed to write HTML response", "", "", err, nil)
	}
}

// ServeLogsPage GET /logs を処理します
func (h *StaticHandler) ServeLogsPage(w http.ResponseWriter, r *http.Request) {
	h.logger.LogAPI("INFO", "Logs page request", "", "", map[string]interface{}{
		"userAgent":  r.Header.Get("User-Agent"),
		"remoteAddr": r.RemoteAddr,
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

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte(html)); err != nil {
		h.logger.LogError("ERROR", "Failed to write HTML response", "", "", err, nil)
	}
}
