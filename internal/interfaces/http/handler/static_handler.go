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
	        :root{
	            --bg:#0f172a; --panel:#111827; --panel-2:#0b1222; --text:#e5e7eb; --muted:#94a3b8; --accent:#22d3ee; --accent-2:#38bdf8; --danger:#fda4af; --ok:#86efac; --border:#1f2937;
	            --btn-grad-from:#0ea5e9; --btn-grad-to:#0369a1; --btn-hover-from:#38bdf8; --btn-hover-to:#0ea5e9; --btn-active:#075985;
	        }
	        *{box-sizing:border-box}
	        body{margin:0;background:var(--bg);color:var(--text);font-family: ui-sans-serif, system-ui, -apple-system, Segoe UI, Roboto, "Helvetica Neue", Arial, "Noto Sans", "Apple Color Emoji", "Segoe UI Emoji"}
	        a{color:var(--accent)}
	        .wrap{max-width:900px;margin:24px auto;padding:0 16px}
	        header{display:flex;align-items:center;gap:12px;margin-bottom:16px}
	        .title{font-size:22px;font-weight:700}
	        .sub{color:var(--muted);font-size:13px}
	        .card{background:linear-gradient(180deg,var(--panel),var(--panel-2));border:1px solid var(--border);border-radius:12px;box-shadow:0 6px 30px rgba(0,0,0,.25)}
	        .content{padding:16px}
	        .row{display:flex;flex-direction:column;gap:12px}
	        label{font-size:13px;color:var(--muted)}
	        input[type="text"], input[type="number"]{background:#0b1222;border:1px solid var(--border);border-radius:10px;color:var(--text);padding:12px 14px;width:100%;font-size:14px;transition:border .18s, background .18s}
	        input[type="text"]:focus, input[type="number"]:focus{outline:none;border-color:var(--accent);background:#0f1b30;box-shadow:0 0 0 2px rgba(34,211,238,.25)}
	        .btn-primary{--shadow:0 4px 16px -4px rgba(14,165,233,.45),0 2px 4px -1px rgba(0,0,0,.35);cursor:pointer;position:relative;display:inline-flex;align-items:center;justify-content:center;gap:8px;min-width:160px;padding:14px 22px;border-radius:12px;font-size:15px;font-weight:600;letter-spacing:.5px;border:1px solid rgba(255,255,255,.08);color:#ecfeff;background:linear-gradient(135deg,var(--btn-grad-from),var(--btn-grad-to));box-shadow:var(--shadow);transition:background .35s,transform .15s,box-shadow .35s,border-color .25s}
	        .btn-primary:hover{background:linear-gradient(135deg,var(--btn-hover-from),var(--btn-hover-to));box-shadow:0 6px 22px -4px rgba(56,189,248,.55),0 3px 6px -2px rgba(0,0,0,.45)}
	        .btn-primary:active{transform:translateY(2px);background:var(--btn-active);box-shadow:0 2px 8px -2px rgba(0,0,0,.5)}
	        .btn-primary:focus-visible{outline:none;box-shadow:0 0 0 3px rgba(56,189,248,.5),0 4px 18px -4px rgba(14,165,233,.5)}
	        .btn-primary:disabled{opacity:.55;cursor:not-allowed;filter:grayscale(.4);box-shadow:none}
	        .btn-primary .pulse{position:absolute;inset:0;border-radius:inherit;pointer-events:none;overflow:hidden}
	        .btn-primary .pulse:before{content:"";position:absolute;inset:0;background:radial-gradient(circle at 30% 30%,rgba(255,255,255,.35),transparent 65%);opacity:.25;mix-blend-mode:overlay}
	        .links{display:flex;gap:12px;margin-top:14px;flex-wrap:wrap}
	        .links a{background:#0b1222;padding:10px 14px;border:1px solid var(--border);border-radius:10px;font-size:13px;text-decoration:none;color:var(--accent-2);transition:background .25s,border-color .25s,transform .15s}
	        .links a:hover{background:#102037;border-color:var(--accent-2)}
	        .message{margin-top:14px;min-height:24px;font-size:14px}
	        .success{color:var(--ok);font-weight:600}
	        .error{color:var(--danger);font-weight:600}
	    </style>
	</head>
	<body>
	    <div class="wrap">
         <header>
             <div class="title">YouTube Live Chat Monitor</div>
             <div class="sub">チャット参加者をリアルタイム収集</div>
             <div style="margin-left:auto"><a href="/users">ユーザー一覧 →</a></div>
         </header>
         <div id="runBanner" class="card" style="margin-bottom:14px; display:none">
             <div class="content" style="display:flex;gap:10px;align-items:center;flex-wrap:wrap">
                 <span style="background:rgba(34,211,238,.12);color:#67e8f9;border:1px solid rgba(103,232,249,.3);padding:6px 10px;border-radius:999px;font-weight:600">メンバーリスト取得中</span>
                 <span id="runInfo" class="sub"></span>
                 <a class="btn" style="margin-left:auto" href="/users">ユーザー一覧を見る</a>
             </div>
         </div>
         <div id="warnBanner" class="card" style="margin-bottom:14px; display:none;border-color:#7f1d1d">
             <div class="content" style="color:#fecaca">
                 <strong>警告:</strong> LIVEステータスが inactive ですが、メンバーリスト取得は起動中です。配信の状態を確認してください。
                 <div id="warnDetail" class="sub" style="margin-top:6px;color:#fca5a5"></div>
             </div>
         </div>
         <div class="card">
	            <div class="content">
	                <form id="monitoringForm">
	                    <div class="form-group">
	                        <label for="videoInput">YouTube Video ID:</label>
	                        <input type="text" id="videoInput" name="videoInput" 
	                               placeholder="例: VIDEO_ID" required>
	                    </div>
	                    
	                    <div class="form-group">
	                        <label for="maxUsers">最大ユーザー数 (デフォルト: 1000):</label>
	                        <input type="number" id="maxUsers" name="maxUsers" 
	                               value="1000" min="1" max="10000">
	                    </div>
	                    
	                    <button type="submit" class="btn-primary"><span>メンバーリスト取得開始</span><span class="pulse"></span></button>
	                </form>
	                <div id="message" class="message"></div>
	                <div class="links"><a href="/logs">システムログ →</a></div>
	            </div>
	        </div>
	    </div>
	
     <script>
         document.getElementById('monitoringForm').addEventListener('submit', async (e) => {
             e.preventDefault();
             const formData = new FormData(e.target);
             const videoInput = formData.get('videoInput');
             const maxUsers = parseInt(formData.get('maxUsers')) || 1000;
             const messageDiv = document.getElementById('message');
             messageDiv.textContent = 'メンバーリスト取得を開始しています...';
             try {
                 const response = await fetch('/api/monitoring/start', {
                     method: 'POST',
                     headers: { 'Content-Type': 'application/json' },
                     body: JSON.stringify({ video_input: videoInput, max_users: maxUsers })
                 });
                 const data = await response.json();
                 if (data.success) {
                     messageDiv.innerHTML = '<span class="success">メンバーリスト取得を開始しました。ユーザーリストへ遷移します…</span>';
                     setTimeout(()=>{ window.location.href = '/users'; }, 800);
                 } else {
                     messageDiv.innerHTML = '<span class="error">エラー: ' + (data.error||'unknown') + '</span>';
                 }
             } catch (error) {
                 messageDiv.innerHTML = '<span class="error">通信エラー: ' + error.message + '</span>';
             }
         });
        
         // サーバー（メンバーリスト取得）ステータスの可視化
         async function refreshStatus(){
             const runBanner = document.getElementById('runBanner');
             const runInfo = document.getElementById('runInfo');
             const warnBanner = document.getElementById('warnBanner');
             const warnDetail = document.getElementById('warnDetail');
             try{
                 const res = await fetch('/api/monitoring/active');
                 if(!res.ok){
                     // 404など: メンバーリスト取得セッションなし
                     runBanner.style.display='none';
                     warnBanner.style.display='none';
                     return;
                 }
                 const data = await res.json();
                 const videoId = (data.data && data.data.videoId) || data.videoId;
                 const isActive = (data.data && typeof data.data.isActive !== 'undefined') ? data.data.isActive : data.isActive;
                 if(!videoId){
                     runBanner.style.display='none';
                     warnBanner.style.display='none';
                     return;
                 }
                 runBanner.style.display='block';
                 runInfo.textContent = 'videoId: '+videoId+' / 状態: ' + (isActive? '起動中' : '停止');
                 // LIVEステータス取得
                 try{
                     const sres = await fetch('/api/monitoring/'+encodeURIComponent(videoId)+'/status');
                     if(sres.ok){
                         const sdata = await sres.json();
                         const status = (sdata.data && sdata.data.status) || sdata.status || sdata.data?.Status || '';
                         const st = String(status||'').toLowerCase();
                         if(isActive && st && st !== 'live'){
                             warnBanner.style.display='block';
                             warnDetail.textContent = 'videoId='+videoId+' / LIVEステータス: '+status;
                         }else{
                             warnBanner.style.display='none';
                         }
                     }else{
                         // ステータスが取得できない場合は警告は出さない
                         warnBanner.style.display='none';
                     }
                 }catch(_){ warnBanner.style.display='none'; }
             }catch(_){
                 runBanner.style.display='none';
                 warnBanner.style.display='none';
             }
         }
         document.addEventListener('DOMContentLoaded', ()=>{ refreshStatus(); setInterval(refreshStatus, 10000); });
         // 自動停止は行いません（手動で停止してください）
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
                        <option value="first_seen_asc" selected>古い順</option>
                        <option value="first_seen_desc">新しい順</option>
                        <option value="name_asc">名前 A→Z</option>
                        <option value="name_desc">名前 Z→A</option>
                    </select>
                    <button class="btn" onclick="loadUsers()">更新</button>
                    <button onclick="stopMonitoring()">メンバーリスト取得停止</button>
                </div>
                <div class="right">
                    <div id="status" class="status">読み込み中...</div>
                </div>
            </div>
            <div id="warn" style="display:none;padding:10px 16px;color:#fecaca;border-top:1px solid var(--border);background:rgba(244,63,94,.08)">警告: LIVEが inactive の可能性。メンバーリスト取得は起動中です。</div>
            <div class="content">
                <div class="meta" style="margin-bottom:8px">
                    <span id="count">0</span> 名 <span id="updated"></span>
                    <span style="margin-left:12px">自動更新: 
                        <label class="toggle"><input id="auto" type="checkbox" checked onchange="toggleAuto()"> ON</label>
                        <select id="interval" onchange="resetAuto()">
                            <!-- 5秒オプション削除し最小10秒 -->
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

        async function loadUsers() {
            const statusDiv = document.getElementById('status');
            const updated = document.getElementById('updated');
            const warn = document.getElementById('warn');
            statusDiv.textContent = '読み込み中...';
            try {
                // 高速化: まず /api/monitoring/users でユーザー一覧だけ先に取得（アクティブセッションがある場合）
                const quickRes = await fetch('/api/monitoring/users');
                if (quickRes.ok) {
                    const data = await quickRes.json();
                    if (data.success) {
                        consecutiveErrors = 0;
                        cachedUsers = Array.isArray(data.users) ? data.users : [];
                        document.getElementById('count').textContent = cachedUsers.length;
                        updated.textContent = '（更新: ' + new Date().toLocaleTimeString() + '）';
                        statusDiv.className = 'status online';
                        statusDiv.textContent = '取得完了 - ユーザー数: ' + (data.count ?? cachedUsers.length);
                        renderUsers();
                        // バックグラウンドで videoId / isActive / LIVE status を取得
                        try {
                            const activeResponse = await fetch('/api/monitoring/active');
                            if (activeResponse.ok) {
                                const ad = await activeResponse.json();
                                const videoId = (ad.data && ad.data.videoId) || ad.videoId;
                                isActive = (ad.data && typeof ad.data.isActive !== 'undefined') ? ad.data.isActive : ad.isActive;
                                const cls = isActive ? 'status online' : 'status offline';
                                const txt = isActive ? 'オンライン' : '停止済み';
                                statusDiv.className = cls;
                                statusDiv.textContent = txt + ' - ユーザー数: ' + (data.count ?? cachedUsers.length);
                                if (videoId) {
                                    try {
                                        const sres = await fetch('/api/monitoring/' + encodeURIComponent(videoId) + '/status');
                                        if (sres.ok) {
                                            const sdata = await sres.json();
                                            const status = (sdata.data && sdata.data.status) || sdata.status || '';
                                            const st = String(status || '').toLowerCase();
                                            if (isActive && st && st !== 'live') {
                                                warn.style.display = 'block';
                                                warn.textContent = '警告: LIVEステータスが ' + status + ' の可能性。メンバーリスト取得は起動中です。';
                                            } else { warn.style.display = 'none'; }
                                        } else { warn.style.display = 'none'; }
                                    } catch (_) { warn.style.display = 'none'; }
                                } else { warn.style.display = 'none'; }
                            } else {
                                isActive = false;
                                statusDiv.className = 'status offline';
                                statusDiv.textContent = 'メンバーリスト取得セッションがありません';
                                warn.style.display = 'none';
                            }
                        } catch (_) { /* ignore */ }
                        return;
                    }
                } else if (quickRes.status === 404) {
                    // セッションなし
                    consecutiveErrors++;
                    isActive = false;
                    statusDiv.className = 'status offline';
                    statusDiv.textContent = 'メンバーリスト取得セッションがありません';
                    document.getElementById('userList').innerHTML = '<div class="empty">メンバーリスト取得を開始するにはホームに戻ってください。</div>';
                    warn.style.display = 'none';
                    if (consecutiveErrors >= MAX_CONSECUTIVE_ERRORS) {
                        stopAutoUpdate();
                        statusDiv.innerHTML += '  自動更新を停止しました';
                    }
                    return;
                }

                // フォールバック: 旧フロー（/active → /{videoId}/users）
                const activeResponse = await fetch('/api/monitoring/active');
                if (!activeResponse.ok) {
                    if (activeResponse.status === 404) {
                        consecutiveErrors++;
                        isActive = false;
                        statusDiv.className = 'status offline';
                        statusDiv.textContent = 'メンバーリスト取得セッションがありません';
                        document.getElementById('userList').innerHTML = '<div class="empty">メンバーリスト取得を開始するにはホームに戻ってください。</div>';
                        warn.style.display = 'none';
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
                if (!videoId) { throw new Error('Active videoId not found in response'); }
                const response = await fetch('/api/monitoring/' + encodeURIComponent(videoId) + '/users');
                if (!response.ok) { throw new Error('Failed to get user list (status ' + response.status + ')'); }
                const data = await response.json();
                if (data.success) {
                    consecutiveErrors = 0;
                    cachedUsers = Array.isArray(data.users) ? data.users : [];
                    const cls = isActive ? 'status online' : 'status offline';
                    const txt = isActive ? 'オンライン' : '停止済み';
                    statusDiv.className = cls; statusDiv.textContent = txt + ' - ユーザー数: ' + (data.count ?? cachedUsers.length);
                    updated.textContent = '（更新: ' + new Date().toLocaleTimeString() + '）';
                    try {
                        const sres = await fetch('/api/monitoring/' + encodeURIComponent(videoId) + '/status');
                        if (sres.ok) {
                            const sdata = await sres.json();
                            const status = (sdata.data && sdata.data.status) || sdata.status || '';
                            const st = String(status || '').toLowerCase();
                            if (isActive && st && st !== 'live') { warn.style.display='block'; warn.textContent='警告: LIVEステータスが ' + status + ' の可能性。メンバーリスト取得は起動中です。'; }
                            else { warn.style.display='none'; }
                        } else { warn.style.display='none'; }
                    } catch (_) { warn.style.display='none'; }
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
            // 並び替えは行わず、初回確定順を維持
            let users = cachedUsers.slice();
            if(q){
                users = users.filter(u => {
                    const name = (u.display_name || u.displayName || '').toLowerCase();
                    const cid = (u.channel_id || u.channelID || '').toLowerCase();
                    return name.includes(q) || cid.includes(q);
                });
            }
            document.getElementById('count').textContent = users.length;
            if(users.length === 0){
                list.innerHTML = '<div class="empty">該当するユーザーがいません</div>';
                return;
            }
            let html = '';
            html += '<table><thead><tr>'+\n                    '<th class="idx">#</th>'+\n                    '<th>ユーザー名</th>'+\n                    '<th>Channel ID</th>'+\n                    '<th>初回参加</th>'+\n                    '<th>最終発言</th>'+\n                    '<th>発言数</th>'+\n                    '<th></th>'+\n                '</tr></thead><tbody>';
            users.forEach((u,i)=>{
                const name = (u.display_name || u.displayName || '');
                const cid = (u.channel_id || u.channelID || '');
                const first = fmtDate(u.first_seen || u.firstSeen);
                const last = fmtDate(u.last_seen || u.lastSeen);
                const msgCount = (u.message_count != null) ? u.message_count : (u.messageCount || 0);
                const url = cid ? 'https://www.youtube.com/channel/' + encodeURIComponent(cid) : '#';
                html += '<tr>'+\n                    '<td class="idx">'+(i+1)+'</td>'+\n                    '<td class="name">'+escapeHtml(name)+'</td>'+ \n                    '<td><a href="'+url+'" target="_blank" rel="noopener">'+escapeHtml(cid)+'</a></td>'+\n                    '<td><span class="pill">'+first+'</span></td>'+\n                    '<td><span class="pill">'+last+'</span></td>'+\n                    '<td>'+msgCount+'</td>'+\n                    '<td class="actions"><button onclick="copy(\''+cid+'\')">Copy</button></td>'+\n                '</tr>';
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
            if (!confirm('メンバーリスト取得を停止しますか？')) { return; }
            try {
                const response = await fetch('/api/monitoring/stop', { method: 'DELETE' });
                const data = await response.json();
                if (data.success) { alert('メンバーリスト取得を停止しました。ホームに戻ります。'); window.location.href = '/'; }
                else { alert('エラー: ' + data.error); }
            } catch (error) { alert('通信エラー: ' + error.message); }
        }

        window.addEventListener('beforeunload', function(){ stopAutoUpdate(); });
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
	        :root{ --bg:#0f172a; --panel:#111827; --panel-2:#0b1222; --text:#e5e7eb; --muted:#94a3b8; --accent:#22d3ee; --accent-2:#38bdf8; --danger:#fda4af; --ok:#86efac; --border:#1f2937; --row:#0b1222; --row-alt:#0d1426; }
	        *{box-sizing:border-box}
	        body{margin:0;background:var(--bg);color:var(--text);font-family: ui-sans-serif, system-ui, -apple-system, Segoe UI, Roboto, "Helvetica Neue", Arial}
	        a{color:var(--accent)}
	        .wrap{max-width:1100px;margin:24px auto;padding:0 16px}
	        header{display:flex;align-items:center;gap:12px;margin-bottom:16px}
	        .title{font-size:22px;font-weight:700}
	        .sub{color:var(--muted);font-size:13px}
	        .card{background:linear-gradient(180deg,var(--panel),var(--panel-2));border:1px solid var(--border);border-radius:12px;box-shadow:0 6px 30px rgba(0,0,0,.25)}
	        .toolbar{display:flex;flex-wrap:wrap;gap:12px;align-items:center;justify-content:space-between;padding:14px 16px;border-bottom:1px solid var(--border)}
	        .controls input[type="text"], .controls select{background:#0b1222;border:1px solid var(--border);border-radius:8px;color:var(--text);padding:10px 12px}
	        .controls .btn, .controls button{background:#0b1222;border:1px solid var(--border);border-radius:8px;color:var(--text);padding:10px 12px;cursor:pointer}
	        .controls .btn-primary{background:linear-gradient(180deg,#0ea5e9,#0891b2);border:0;color:#e6faff}
	        .status{padding:10px 12px;border-radius:8px;font-size:13px;background:#0b1222;border:1px solid var(--border)}
	        .content{padding:12px 16px}
	        table{width:100%;border-collapse:separate;border-spacing:0 8px}
	        thead th{font-size:12px;text-transform:uppercase;letter-spacing:.06em;color:var(--muted);text-align:left;padding:0 10px}
	        tbody tr{background:var(--row);border:1px solid var(--border)}
	        tbody tr:nth-child(even){background:var(--row-alt)}
	        td{padding:10px 10px;vertical-align:middle;font-size:13px}
	        .badge{font-size:11px;padding:2px 8px;border-radius:999px;border:1px solid rgba(255,255,255,.12)}
	        .lvl-info{color:#67e8f9;background:rgba(103,232,249,.12);border-color:rgba(103,232,249,.3)}
	        .lvl-warn{color:#fde68a;background:rgba(253,230,138,.12);border-color:rgba(253,230,138,.3)}
	        .lvl-error{color:#fda4af;background:rgba(253,164,175,.12);border-color:rgba(253,164,175,.3)}
	        .footer{display:flex;justify-content:space-between;align-items:center;padding:12px 16px;border-top:1px solid var(--border)}
	        .muted{color:var(--muted);font-size:12px}
	        @media (max-width:720px){ thead{display:none} table, tbody, tr, td{display:block;width:100%} tbody tr{margin:8px 0;border-radius:10px} td{padding:8px 12px} }
	    </style>
	</head>
	<body>
	    <div class="wrap">
	        <header>
	            <div class="title">システムログ</div>
	            <div class="sub">アプリケーションイベントの一覧</div>
	            <div style="margin-left:auto"><a href="/">← ホーム</a></div>
	        </header>
	        <div class="card">
	            <div class="toolbar">
	                <div class="controls" style="display:flex;gap:8px;flex-wrap:wrap;align-items:center">
	                    <select id="level">
	                        <option value="">全レベル</option>
	                        <option value="INFO">INFO</option>
	                        <option value="WARNING">WARNING</option>
	                        <option value="ERROR">ERROR</option>
	                    </select>
	                    <input id="component" type="text" placeholder="component で絞り込み">
	                    <input id="video" type="text" placeholder="video_id で絞り込み">
	                    <select id="limit">
	                        <option value="100">100件</option>
	                        <option value="300" selected>300件</option>
	                        <option value="1000">1000件</option>
	                    </select>
	                    <button class="btn-primary" onclick="loadLogs()">更新</button>
	                    <button onclick="clearLogs()">全クリア</button>
	                    <button onclick="exportLogs()">エクスポート</button>
	                </div>
	                <div class="status">
	                    <label style="display:inline-flex;align-items:center;gap:6px">
	                        <input id="auto" type="checkbox" checked onchange="toggleAuto()"> 自動更新
	                    </label>
	                    <select id="interval" onchange="resetAuto()">
	                        <!-- 5秒オプション削除し最小10秒 -->
	                        <option value="10000" selected>10秒</option>
	                        <option value="30000">30秒</option>
	                    </select>
	                </div>
	            </div>
	            <div class="content">
	                <div class="muted" id="meta">読み込み中…</div>
	                <div id="logTable"></div>
	            </div>
	            <div class="footer">
	                <div class="muted" id="stats"></div>
	                <div class="muted"><a href="/users">ユーザー一覧</a></div>
	            </div>
	        </div>
	    </div>
	    <script>
	        let autoTimer=null; let lastCount=0;
	        window.onload = function(){ loadLogs(); startAuto(); };
	        function startAuto(){ if(autoTimer) clearInterval(autoTimer); if(!document.getElementById('auto').checked) return; const ms=parseInt(document.getElementById('interval').value||'10000',10); autoTimer=setInterval(loadLogs, ms); }
	        function stopAuto(){ if(autoTimer){ clearInterval(autoTimer); autoTimer=null; } }
	        function resetAuto(){ stopAuto(); startAuto(); }
	        function toggleAuto(){ if(document.getElementById('auto').checked) startAuto(); else stopAuto(); }
	        function q(params){ const sp=new URLSearchParams(params); return sp.toString() ? ('?'+sp.toString()) : ''; }
	        function badge(level){ level=(level||'').toUpperCase(); const cls= level==='ERROR'?'lvl-error':(level==='WARNING'?'lvl-warn':'lvl-info'); return '<span class="badge '+cls+'">'+level+'</span>'; }
	        function esc(s){ return (s||'').replace(/[&<>"']/g, c=>({"&":"&amp;","<":"&lt;",">":"&gt;","\"":"&quot;","'":"&#039;"}[c])); }
	        async function loadLogs(){
	            const meta=document.getElementById('meta'); const table=document.getElementById('logTable');
	            const params={};
	            const lv=document.getElementById('level').value; if(lv) params.level=lv;
	            const comp=document.getElementById('component').value.trim(); if(comp) params.component=comp;
	            const vid=document.getElementById('video').value.trim(); if(vid) params.video_id=vid;
	            const lim=document.getElementById('limit').value || '300'; params.limit=lim;
	            try{
	                const res=await fetch('/api/logs'+q(params)); const data=await res.json();
	                if(data.success){ const logs=Array.isArray(data.logs)?data.logs:[]; lastCount=logs.length; meta.textContent='件数: '+logs.length+'（更新: '+new Date().toLocaleTimeString()+'）';
	                    if(logs.length===0){ table.innerHTML='<div class="muted">ログがありません</div>'; }
	                    else{
	                        let html='<table><thead><tr><th>時刻</th><th>Level</th><th>メッセージ</th><th>component</th><th>video_id</th><th>correlation</th></tr></thead><tbody>';
	                        logs.forEach(l=>{
	                            const ts=esc(l.timestamp||'');
	                            const lv=badge(l.level||'');
	                            const msg=esc(l.message||'');
	                            const comp=esc(l.component||'');
	                            const vid=esc(l.video_id||'');
	                            const corr=esc(l.correlation_id||'');
	                            html+='<tr><td>'+ts+'</td><td>'+lv+'</td><td>'+msg+'</td><td>'+comp+'</td><td>'+vid+'</td><td>'+corr+'</td></tr>';
	                        });
	                        html+='</tbody></table>';
	                        table.innerHTML=html;
	                    }
	                }else{ meta.textContent='エラー: '+(data.error||'unknown'); }
	            }catch(e){ meta.textContent='通信エラー: '+e.message; }
	            loadStats(params);
	        }
	        async function loadStats(params){
	            try{
	                const res=await fetch('/api/logs'+q(Object.assign({}, params, {stats:1}))); const data=await res.json();
	                if(data.success){ document.getElementById('stats').textContent='統計: 総数 '+data.total+' / エラー '+data.errors+' / 警告 '+data.warnings; }
	            }catch(e){ /* noop */ }
	        }
	        async function clearLogs(){ if(!confirm('すべてのログをクリアしますか？')) return; try{ const res=await fetch('/api/logs',{method:'DELETE'}); const data=await res.json(); if(data.success){ loadLogs(); } else { alert('エラー: '+(data.error||'unknown')); } }catch(e){ alert('通信エラー: '+e.message); } }
	        async function exportLogs(){
	            const params={}; const lv=document.getElementById('level').value; if(lv) params.level=lv; const comp=document.getElementById('component').value.trim(); if(comp) params.component=comp; const vid=document.getElementById('video').value.trim(); if(vid) params.video_id=vid; const lim=document.getElementById('limit').value||'300'; params.limit=lim; params.export=1;
	            const url='/api/logs'+q(params); const a=document.createElement('a'); a.href=url; a.download='logs_'+new Date().toISOString().split('T')[0]+'.json'; document.body.appendChild(a); a.click(); document.body.removeChild(a);
	        }
	        window.addEventListener('beforeunload', function(){ stopAuto(); });
	    </script>
	</body>
	</html>`

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte(html)); err != nil {
		h.logger.LogError("ERROR", "Failed to write HTML response", "", "", err, nil)
	}
}
