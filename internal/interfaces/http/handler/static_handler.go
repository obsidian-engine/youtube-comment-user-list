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
	    <title>Home - YouTube Live Chat Monitor</title>
	    <link href="https://fonts.googleapis.com/css2?family=Roboto:wght@400;500;700&display=swap" rel="stylesheet">
	    <link href="https://fonts.googleapis.com/css2?family=Material+Symbols+Outlined:wght@400;700" rel="stylesheet" />
	    <style>
	        :root{
	            --md-sys-color-primary:#90caf9; --md-sys-color-secondary:#80cbc4; --md-sys-color-surface:#121212; --md-sys-color-background:#0e1116; --md-sys-color-on-surface:#e2e7ee; --md-sys-color-outline:#2a2f3a; --md-sys-color-error:#ffb4ab;
	        }
	        *{box-sizing:border-box}
	        body{margin:0;background:var(--md-sys-color-background);color:var(--md-sys-color-on-surface);font-family: Roboto, ui-sans-serif, system-ui, -apple-system, Segoe UI, "Helvetica Neue", Arial}
	        a{color:var(--md-sys-color-primary);text-decoration:none}
	        a:hover{text-decoration:underline}
	        .appbar{position:sticky;top:0;background:#1d252f;border-bottom:1px solid var(--md-sys-color-outline);box-shadow:0 2px 6px rgba(0,0,0,.35);padding:12px 16px;margin-bottom:16px}
	        .wrap{max-width:960px;margin:0 auto 24px auto;padding:0 16px}
	        .row{display:flex;align-items:center;gap:12px}
	        .title{font-size:20px;font-weight:700}
	        .sub{opacity:.75;font-size:13px}
	        .card{background:linear-gradient(180deg,#161b22,#12161c);border:1px solid var(--md-sys-color-outline);border-radius:12px;box-shadow:0 8px 30px rgba(0,0,0,.35)}
	        .content{padding:16px}
	        .section{display:flex;flex-direction:column;gap:12px}
	        label{font-size:13px;opacity:.8}
	        input[type="text"],input[type="number"]{background:#0f141b;border:1px solid var(--md-sys-color-outline);border-radius:10px;color:var(--md-sys-color-on-surface);padding:12px 14px;width:100%;font-size:14px}
	        .md-btn{--elev:0 2px 4px rgba(0,0,0,.35);--elevH:0 6px 16px rgba(0,0,0,.45);display:inline-flex;align-items:center;gap:8px;padding:12px 16px;border-radius:10px;border:1px solid rgba(255,255,255,.06);background:linear-gradient(180deg,#2196f3,#1976d2);color:#e8f2ff;font-weight:600;letter-spacing:.3px;cursor:pointer;box-shadow:var(--elev);transition:box-shadow .25s,transform .1s,filter .2s}
	        .md-btn:hover{box-shadow:var(--elevH);filter:saturate(1.05)}
	        .md-btn:active{transform:translateY(1px)}
	        .md-btn.outlined{background:transparent;border:1px solid var(--md-sys-color-outline);color:var(--md-sys-color-primary)}
	        .pill{display:inline-block;font-size:12px;color:#cfe8ff;background:rgba(144,202,249,.12);border:1px solid rgba(144,202,249,.28);padding:4px 10px;border-radius:999px}
         .warn{color:#fecaca}
         #runBanner{display:none!important}
     </style>
	</head>
	<body>
	    <div class="appbar">
	        <div class="wrap">
	            <div class="row">
                 <span class="material-symbols-outlined" aria-hidden="true">home</span>
                 <div class="title">YouTube Live Chat Monitor</div>
                 <div class="sub">チャット参加者をリアルタイム収集</div>
                 <div style="margin-left:auto;display:flex;gap:8px;align-items:center">
                     <span id="appbarMon" class="pill" style="display:none;align-items:center;gap:6px"><span class="material-symbols-outlined" style="font-size:16px;vertical-align:-3px">sensors</span> 監視中 <button id="appbarStop" class="md-btn outlined" style="padding:4px 8px;font-size:12px;line-height:1.2;margin-left:6px"><span class="material-symbols-outlined" style="font-size:16px">stop_circle</span> 停止</button></span>
                     <a href="/users" class="md-btn outlined"><span class="material-symbols-outlined" style="font-size:18px">group</span> ユーザー一覧</a>
                     <a href="/logs" class="md-btn outlined"><span class="material-symbols-outlined" style="font-size:18px">list</span> ログ</a>
                 </div>
	            </div>
	        </div>
	    </div>
	    <div class="wrap">
	        <div id="runBanner" class="card" style="margin-bottom:14px; display:none">
	            <div class="content" style="display:flex;gap:10px;align-items:center;flex-wrap:wrap">
	                <span class="pill"><span class="material-symbols-outlined" style="font-size:18px;vertical-align:-4px">sensors</span> 監視中</span>
	                <span id="runInfo" class="sub"></span>
	                <a class="md-btn" style="margin-left:auto" href="/users"><span class="material-symbols-outlined" style="font-size:18px">group</span> ユーザー一覧を見る</a>
                 <button id="statusRefreshBtn" class="md-btn outlined" type="button" style="margin-left:4px;padding:8px 10px;font-size:12px"><span class="material-symbols-outlined" style="font-size:16px">refresh</span> ステータス更新</button>
	            </div>
	        </div>
	        <div id="warnBanner" class="card" style="margin-bottom:14px; display:none;border-color:#7f1d1d">
	            <div class="content warn">
	                <strong>警告:</strong> LIVEステータスが inactive ですが、監視は起動中です。配信の状態を確認してください。
	                <div id="warnDetail" class="sub" style="margin-top:6px;opacity:.9"></div>
	            </div>
	        </div>
	        <div class="card">
	            <div class="content">
	                <form id="monitoringForm" class="section">
	                    <div>
	                        <label for="videoInput">YouTube Video ID</label>
	                        <input type="text" id="videoInput" name="videoInput" placeholder="例: VIDEO_ID" required>
	                    </div>
	                    <div>
	                        <label for="maxUsers">最大ユーザー数 (デフォルト: 1000)</label>
	                        <input type="number" id="maxUsers" name="maxUsers" value="1000" min="1" max="10000">
	                    </div>
	                    <div>
	                        <button type="submit" class="md-btn"><span class="material-symbols-outlined" style="font-size:18px">play_circle</span> 監視を開始</button>
	                        <span class="sub" style="margin-left:8px">開始後はユーザー一覧に遷移します</span>
	                    </div>
	                </form>
	                <div id="message" class="sub" style="min-height:22px;margin-top:6px"></div>
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
	                    messageDiv.innerHTML = '<span style="color:#86efac">メンバーリスト取得を開始しました。ユーザーリストへ遷移します…</span>';
	                    setTimeout(()=>{ window.location.href = '/users'; }, 800);
	                } else {
	                    messageDiv.innerHTML = '<span style="color:#fda4af">エラー: ' + (data.error||'unknown') + '</span>';
	                }
	            } catch (error) {
	                messageDiv.innerHTML = '<span style="color:#fda4af">通信エラー: ' + error.message + '</span>';
	            }
	        });
	       
	        // サーバー（メンバーリスト取得）ステータスの可視化
	        async function refreshStatus(){
	            if(window.__ACTIVE_SESSION_LOCK){return;} // prevent concurrent
	            window.__ACTIVE_SESSION_LOCK=true;
	            const runBanner = document.getElementById('runBanner');
	            const runInfo = document.getElementById('runInfo');
	            const warnBanner = document.getElementById('warnBanner');
	            const warnDetail = document.getElementById('warnDetail');
	            try{
	                // 既に videoId がキャッシュされていれば再取得しない（明示リフレッシュ時のみ再取得）
	                if(!window.__ACTIVE_VIDEO_ID || window.__FORCE_ACTIVE_REFRESH){
	                    const res = await fetch('/api/monitoring/active');
	                    if(!res.ok){ runBanner.style.display='none'; warnBanner.style.display='none'; window.__ACTIVE_VIDEO_ID=''; return; }
	                    const data = await res.json();
	                    window.__ACTIVE_VIDEO_ID = (data.data && data.data.videoId) || data.videoId || '';
	                    window.__ACTIVE_IS_ACTIVE = (data.data && typeof data.data.isActive !== 'undefined') ? data.data.isActive : data.isActive;
	                    window.__FORCE_ACTIVE_REFRESH=false;
	                }
	                const videoId = window.__ACTIVE_VIDEO_ID;
	                const isActive = window.__ACTIVE_IS_ACTIVE;
	                if(!videoId){ runBanner.style.display='none'; warnBanner.style.display='none'; return; }
	                runBanner.style.display='block';
	                runInfo.textContent = 'videoId: '+videoId+' / 状態: ' + (isActive? '起動中' : '停止');
	                // LIVE ステータスは負荷軽減のため抑制（従来の詳細チェック削除）
	            }catch(_){
	                runBanner.style.display='none'; warnBanner.style.display='none';
	            } finally { window.__ACTIVE_SESSION_LOCK=false; }
	        }
	        document.addEventListener('DOMContentLoaded', ()=>{ refreshStatus(); /* 自動繰り返し廃止 */ });
	        const statusBtn = document.getElementById('statusRefreshBtn');
	        if(statusBtn){ statusBtn.addEventListener('click', ()=>{ window.__FORCE_ACTIVE_REFRESH=true; refreshStatus(); }); }
	    </script>
	    <style id="activeBannerStyles">
	        .active-banner{position:fixed;left:0;right:0;bottom:0;background:#7f1d1d;color:#fee2e2;border-top:2px solid #ef4444;z-index:9999;box-shadow:0 -8px 30px rgba(0,0,0,.45)}
	        .active-banner .ab-wrap{max-width:1100px;margin:0 auto;padding:12px 16px;display:flex;align-items:center;gap:12px}
	        .md-btn.danger{background:linear-gradient(180deg,#ef4444,#b91c1c);color:#fff;border-color:rgba(255,255,255,.08)}
	        @keyframes abPulse{0%{filter:saturate(1)}50%{filter:saturate(1.25)}100%{filter:saturate(1)}}
	        .active-banner{animation:abPulse 2.5s ease-in-out infinite}
	    </style>
	    <div id="activeBanner" class="active-banner" style="display:none" role="alert" aria-live="assertive">
	        <div class="ab-wrap">
	            <span class="material-symbols-outlined" aria-hidden="true">warning</span>
	            <div class="ab-text"><strong>監視実行中</strong> — 作業後は必ず停止してください</div>
	            <button id="abStop" class="md-btn danger" style="margin-left:auto"><span class="material-symbols-outlined" style="font-size:18px">stop_circle</span> 停止する</button>
	        </div>
	    </div>
	    <script>
	        async function updateActiveBanner(){
	            // 高頻度ポーリング廃止：初回のみ
	            const el=document.getElementById('activeBanner');
	            if(window.__ACTIVE_BANNER_INIT){ return; }
	            window.__ACTIVE_BANNER_INIT=true;
	            try{
	                const res=await fetch('/api/monitoring/active');
	                if(!res.ok){ if(el) el.style.display='none'; return; }
	                const data=await res.json();
	                const active=(data.data&&typeof data.data.isActive!=='undefined')?data.data.isActive:data.isActive;
	                if(el){ el.style.display=active?'block':'none'; }
	                if(active){ window.__ACTIVE_VIDEO_ID=(data.data&&data.data.videoId)||data.videoId||window.__ACTIVE_VIDEO_ID; }
	            }catch(_){ if(el) el.style.display='none'; }
	        }
	        document.addEventListener('DOMContentLoaded',()=>{ updateActiveBanner(); /* setInterval removed */ const btn=document.getElementById('abStop'); if(btn) btn.addEventListener('click', stopFromBanner); });
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
    <link href="https://fonts.googleapis.com/css2?family=Roboto:wght@400;500;700&display=swap" rel="stylesheet">
    <link href="https://fonts.googleapis.com/css2?family=Material+Symbols+Outlined:wght@400;700" rel="stylesheet" />
    <style>
        :root{
            --md-sys-color-primary:#90caf9; --md-sys-color-secondary:#80cbc4; --md-sys-color-surface:#121212; --md-sys-color-background:#0e1116; --md-sys-color-on-surface:#e2e7ee; --md-sys-color-outline:#2a2f3a; --md-sys-color-error:#ffb4ab;
            --row:#0f141b; --row-alt:#0d131a;
        }
        *{box-sizing:border-box}
        body{margin:0;background:var(--md-sys-color-background);color:var(--md-sys-color-on-surface);font-family: Roboto, ui-sans-serif, system-ui, -apple-system, Segoe UI, "Helvetica Neue", Arial}
        a{color:var(--md-sys-color-primary);text-decoration:none}
        a:hover{text-decoration:underline}
        .appbar{position:sticky;top:0;background:#1d252f;border-bottom:1px solid var(--md-sys-color-outline);box-shadow:0 2px 6px rgba(0,0,0,.35);padding:12px 16px;margin-bottom:16px}
        .row{display:flex;align-items:center;gap:12px}
        .title{font-size:20px;font-weight:700}
        .sub{opacity:.75;font-size:13px}
        .wrap{max-width:1100px;margin:0 auto 24px auto;padding:0 16px}
        .card{background:linear-gradient(180deg,#161b22,#12161c);border:1px solid var(--md-sys-color-outline);border-radius:12px;box-shadow:0 8px 30px rgba(0,0,0,.35)}
        .toolbar{display:flex;flex-wrap:wrap;gap:12px;align-items:center;justify-content:space-between;padding:14px 16px;border-bottom:1px solid var(--md-sys-color-outline)}
        .left, .right{display:flex;gap:12px;align-items:center}
        .controls input[type="text"]{background:#0f141b;border:1px solid var(--md-sys-color-outline);border-radius:8px;color:var(--md-sys-color-on-surface);padding:10px 12px;min-width:220px}
        .controls select{background:#0f141b;border:1px solid var(--md-sys-color-outline);border-radius:8px;color:var(--md-sys-color-on-surface);padding:10px 12px}
        .status{padding:10px 12px;border-radius:8px;font-size:13px}
        .status.online{background:rgba(144,202,249,.12);color:#cfe8ff;border:1px solid rgba(144,202,249,.28)}
        .status.offline{background:rgba(244,63,94,.08);color:#fda4af;border:1px solid rgba(253,164,175,.25)}
        .meta{opacity:.75;font-size:12px}
        .content{padding:12px 16px}
        table{width:100%;border-collapse:separate;border-spacing:0 8px}
        thead th{font-size:12px;text-transform:uppercase;letter-spacing:.06em;opacity:.75;text-align:left;padding:0 10px}
        tbody tr{background:var(--row);border:1px solid var(--md-sys-color-outline)}
        tbody tr:nth-child(even){background:var(--row-alt)}
        td{padding:12px 10px;vertical-align:middle}
        .idx{opacity:.7;width:56px}
        .name{display:flex;align-items:center;gap:10px}
        .badge{font-size:11px;color:#c5c6ff;background:rgba(129,140,248,.14);border:1px solid rgba(129,140,248,.35);padding:2px 8px;border-radius:999px}
        .pill{font-size:12px;color:#cfe8ff;background:rgba(144,202,249,.12);border:1px solid rgba(144,202,249,.28);padding:4px 10px;border-radius:999px}
        .actions button{background:#0f141b;border:1px solid var(--md-sys-color-outline);color:var(--md-sys-color-on-surface);border-radius:8px;padding:6px 10px;cursor:pointer}
        .empty{padding:24px;text-align:center;opacity:.75}
        .footer{display:flex;justify-content:space-between;align-items:center;padding:12px 16px;border-top:1px solid var(--md-sys-color-outline)}
        .md-btn{--elev:0 2px 4px rgba(0,0,0,.35);--elevH:0 6px 16px rgba(0,0,0,.45);display:inline-flex;align-items:center;gap:8px;padding:10px 14px;border-radius:10px;border:1px solid rgba(255,255,255,.06);background:linear-gradient(180deg,#2196f3,#1976d2);color:#e8f2ff;font-weight:600;letter-spacing:.3px;cursor:pointer;box-shadow:var(--elev);transition:box-shadow .25s,transform .1s,filter .2s}
        .md-btn:hover{box-shadow:var(--elevH);filter:saturate(1.05)}
        .md-btn:active{transform:translateY(1px)}
        .md-btn.outlined{background:transparent;border:1px solid var(--md-sys-color-outline);color:var(--md-sys-color-primary}
        @media (max-width:720px){
            thead{display:none}
            table, tbody, tr, td{display:block;width:100%}
            tbody tr{margin:8px 0;border-radius:10px}
            td{padding:8px 12px}
            .idx{display:none}
        }
        /* Readable numbered list styles for shoutout */
        #numberedList{list-style:none;padding:0;margin:0;counter-reset:n}
        #numberedList li{counter-increment:n;margin:.35em 0;line-height:1.45;font-size:clamp(18px,3vw,28px)}
        #numberedList li::before{content:counter(n)'. ';font-weight:700;color:#FFD700;margin-right:.35em}
        .count{margin-top:12px;color:#9aa0a6;font-size:clamp(12px,2vw,16px)}
    </style>
</head>
<body>
    <div class="appbar">
        <div class="wrap">
            <div class="row">
                <span class="material-symbols-outlined" aria-hidden="true">group</span>
                <div class="title">ユーザーリスト</div>
                <div class="sub">YouTube Live Chat 参加者</div>
                <div style="margin-left:auto;display:flex;gap:8px;align-items:center">
                    <span id="appbarMon" class="pill" style="display:none;align-items:center;gap:6px"><span class="material-symbols-outlined" style="font-size:16px;vertical-align:-3px">sensors</span> 監視中 <button id="appbarStop" class="md-btn outlined" style="padding:4px 8px;font-size:12px;line-height:1.2;margin-left:6px"><span class="material-symbols-outlined" style="font-size:16px">stop_circle</span> 停止</button></span>
                    <a href="/" class="md-btn outlined"><span class="material-symbols-outlined" style="font-size:18px">home</span> ホーム</a>
                </div>
            </div>
        </div>
    </div>

    <div class="wrap">
        <div class="card">
            <div class="toolbar">
                <div class="left controls">
                    <input id="search" type="text" placeholder="名前・Channel IDで検索" oninput="renderUsers()"/>
                    <select id="sort" onchange="renderUsers()" style="display:none"></select>
                    <button class="md-btn" onclick="manualRefresh()"><span class="material-symbols-outlined" style="font-size:18px">sync</span> 即時更新</button>
                    <button class="md-btn outlined" onclick="stopMonitoring()"><span class="material-symbols-outlined" style="font-size:18px">stop_circle</span> 監視停止</button>
                </div>
                <div class="right">
                    <div id="status" class="status">初回取得中...</div>
                </div>
            </div>
            <div id="warn" style="display:none;padding:10px 16px;color:#fecaca;border-top:1px solid var(--md-sys-color-outline);background:rgba(244,63,94,.08)">警告: LIVEが inactive の可能性。監視は起動中です。</div>
            <div class="content">
                <div class="meta" style="margin-bottom:8px">
                    <span id="count">0</span> 名 <span id="updated"></span>
                    <span style="margin-left:12px;font-size:11px;opacity:.75">自動更新: 60秒間隔</span>
                </div>
                <div id="userList">
                    <div class="empty">データを読み込んでいます…</div>
                </div>
            </div>
            <div class="footer">
                <div class="meta">ChannelをクリックするとYouTubeチャンネルを開きます</div>
                <div><a href="/" class="md-btn outlined"><span class="material-symbols-outlined" style="font-size:18px">home</span> ホーム</a></div>
            </div>
        </div>
    </div>

    <script>
        // --- 設定 ---
        const REFRESH_INTERVAL_MS = 60000; // 60秒(1分)固定
        
        // --- 状態 ---
        let cachedUsers = [];
        let isActive = false;
        let refreshTimer = null;
        let currentVideoId = '';
        let fetching = false;
        let lastStatusFetch = 0;

        // 初期化
        window.addEventListener('DOMContentLoaded', () => {
            initialLoad();
        });
        window.addEventListener('beforeunload', () => { if (refreshTimer) clearInterval(refreshTimer); });

        async function initialLoad(){
            await refreshUsers(true);
            if(!refreshTimer){
                refreshTimer = setInterval(() => { refreshUsers(false); }, REFRESH_INTERVAL_MS);
            }
        }

        async function manualRefresh(){
            refreshUsers(false, { force: true });
        }

        async function refreshUsers(isInitial, opts={}){
            if(fetching && !opts.force) return;
            fetching = true;
            const statusDiv = document.getElementById('status');
            const updated = document.getElementById('updated');
            statusDiv.textContent = isInitial ? '初回取得中...' : '更新中...';
            try {
                // active セッションを初回のみ取得
                if(!currentVideoId){
                    const activeRes = await fetch('/api/monitoring/active', { cache:'no-store' });
                    if(!activeRes.ok){
                        if(activeRes.status === 404){
                            statusDiv.className = 'status offline';
                            statusDiv.textContent = '監視セッションがありません';
                            document.getElementById('userList').innerHTML = '<div class="empty">監視を開始するにはホームへ戻ってください。</div>';
                            fetching = false; return; }
                        statusDiv.className = 'status offline';
                        statusDiv.textContent = 'アクティブ確認失敗 (' + activeRes.status + ')';
                        fetching=false; return;
                    }
                    const activeData = await activeRes.json();
                    currentVideoId = (activeData.data && activeData.data.videoId) || activeData.videoId || '';
                    isActive = (activeData.data && typeof activeData.data.isActive !== 'undefined') ? activeData.data.isActive : activeData.isActive;
                }
                if(!currentVideoId){
                    statusDiv.className='status offline'; statusDiv.textContent='videoId 取得不可'; fetching=false; return;
                }
                // ユーザーリストのみ取得（active 再問い合わせしない）
                const listRes = await fetch('/api/monitoring/' + encodeURIComponent(currentVideoId) + '/users', { cache:'no-store' });
                if(!listRes.ok){ statusDiv.className='status offline'; statusDiv.textContent='ユーザー取得失敗 ('+listRes.status+')'; fetching=false; return; }
                const listData = await listRes.json();
                if(!listData.success){ statusDiv.className='status offline'; statusDiv.textContent='エラー: '+(listData.error||'unknown'); fetching=false; return; }
                cachedUsers = Array.isArray(listData.users) ? listData.users : [];
                const cls = isActive ? 'status online' : 'status offline';
                const txt = isActive ? 'オンライン' : '停止済み';
                statusDiv.className = 'status ' + cls.split(' ').pop();
                statusDiv.textContent = txt + ' - コメントユーザー数: ' + (listData.count ?? cachedUsers.length);
                updated.textContent = '（更新: ' + new Date().toLocaleTimeString() + '）';
                renderUsers();
            } catch(err){
                statusDiv.className='status offline'; statusDiv.textContent='通信エラー: '+err.message;
            } finally { fetching=false; }
        }

        function renderUsers(){
            const list = document.getElementById('userList');
            const q = (document.getElementById('search').value || '').toLowerCase();
            let users = cachedUsers.slice();

            if(q){
                users = users.filter(u => {
                    const name = (u.display_name || u.displayName || '').toLowerCase();
                    const cid = (u.channel_id || u.channelID || '').toLowerCase();
                    return name.includes(q) || cid.includes(q);
                });
            }

            // ソート
            const sort = document.getElementById('sort').value;
            users.sort((a,b)=>{
                const firstA = new Date(a.first_seen || a.firstSeen || 0).getTime();
                const firstB = new Date(b.first_seen || b.firstSeen || 0).getTime();
                const nameA = (a.display_name || a.displayName || '').toLowerCase();
                const nameB = (b.display_name || b.displayName || '').toLowerCase();
                switch(sort){
                    case 'first_seen_desc':
                        if (firstA !== firstB) return firstB - firstA; // 時刻降順
                        return nameA.localeCompare(nameB);             // 同時刻は名前昇順
                    case 'name_asc':
                        return nameA.localeCompare(nameB);
                    case 'name_desc':
                        return nameB.localeCompare(nameA);
                    default: // first_seen_asc
                        if (firstA !== firstB) return firstA - firstB; // 時刻昇順
                        return nameA.localeCompare(nameB);             // 同時刻は名前昇順
                }
            });

            document.getElementById('count').textContent = users.length;
            if(users.length === 0){
                list.innerHTML = '<div class="empty">該当するユーザーがいません</div>';
                return;
            }

            let html = '';
            html += '<ol id="numberedList">';
            users.forEach((u,i)=>{
                const name = (u.display_name || u.displayName || '');
                html += '<li>'+escapeHtml(name)+'</li>';
            });
            html += '</ol>';
            list.innerHTML = html;
        }

        function fmtDate(s){
            try{ return new Date(s).toLocaleString(); }catch(e){ return '-'; }
        }
        function escapeHtml(s){
            return (s||'').replace(/[&<>"']/g, c=>({ '&':'&amp;','<':'&lt;','>':'&gt;','"':'&quot;','\'':'&#039;' }[c]));
        }
        function copy(text){
            if(!text) return;
            navigator.clipboard?.writeText(text).then(()=>{ toast('Channel IDをコピーしました'); }).catch(()=>{
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

        async function stopMonitoring() {
            if (!confirm('監視を停止しますか？')) { return; }
            try {
                const response = await fetch('/api/monitoring/stop', { method: 'DELETE' });
                const data = await response.json();
                if (data.success) { alert('監視を停止しました。ホームに戻ります。'); window.location.href = '/'; }
                else { alert('エラー: ' + data.error); }
            } catch (error) { alert('通信エラー: ' + error.message); }
        }
    </script>
    <script>
        async function updateAppbarMon(){
            // 繰り返し廃止：初回のみ
            if(window.__APPBAR_MON_INIT){ return; }
            window.__APPBAR_MON_INIT=true;
            const pill=document.getElementById('appbarMon');
            try{
                const res=await fetch('/api/monitoring/active');
                if(!res.ok){ if(pill) pill.style.display='none'; return; }
                const data=await res.json();
                const active=(data.data&&typeof data.data.isActive!=='undefined')?data.data.isActive:data.isActive;
                if(pill){ pill.style.display=active?'inline-flex':'none'; }
            }catch(_){ if(pill) pill.style.display='none'; }
        }
        document.addEventListener('DOMContentLoaded',()=>{ updateAppbarMon(); const btn=document.getElementById('appbarStop'); if(btn) btn.addEventListener('click', stopFromAppbar); });
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
	    <link href="https://fonts.googleapis.com/css2?family=Roboto:wght@400;500;700&display=swap" rel="stylesheet">
	    <link href="https://fonts.googleapis.com/css2?family=Material+Symbols+Outlined:wght@400;700" rel="stylesheet" />
	    <style>
	        :root{ --primary:#90caf9; --secondary:#80cbc4; --surface:#121212; --bg:#0e1116; --on:#e2e7ee; --outline:#2a2f3a; --row:#0f141b; --row-alt:#0d131a }
	        *{box-sizing:border-box}
	        body{margin:0;background:var(--bg);color:var(--on);font-family: Roboto, ui-sans-serif, system-ui, -apple-system, Segoe UI, "Helvetica Neue", Arial}
	        a{color:var(--primary);text-decoration:none}
	        a:hover{text-decoration:underline}
	        .appbar{position:sticky;top:0;background:#1d252f;border-bottom:1px solid var(--outline);box-shadow:0 2px 6px rgba(0,0,0,.35);padding:12px 16px;margin-bottom:16px}
	        .wrap{max-width:1100px;margin:0 auto 24px auto;padding:0 16px}
	        .row{display:flex;align-items:center;gap:12px}
	        .title{font-size:20px;font-weight:700}
	        .sub{opacity:.75;font-size:13px}
	        .card{background:linear-gradient(180deg,#161b22,#12161c);border:1px solid var(--outline);border-radius:12px;box-shadow:0 8px 30px rgba(0,0,0,.35)}
	        .toolbar{display:flex;flex-wrap:wrap;gap:12px;align-items:center;justify-content:space-between;padding:14px 16px;border-bottom:1px solid var(--outline)}
	        .controls input[type="text"], .controls select{background:#0f141b;border:1px solid var(--outline);border-radius:8px;color:var(--on);padding:10px 12px}
	        .status{padding:10px 12px;border-radius:8px;font-size:13px;background:#0f141b;border:1px solid var(--outline)}
	        .content{padding:12px 16px}
	        table{width:100%;border-collapse:separate;border-spacing:0 8px}
	        thead th{font-size:12px;text-transform:uppercase;letter-spacing:.06em;opacity:.75;text-align:left;padding:0 10px}
	        tbody tr{background:var(--row);border:1px solid var(--outline)}
	        tbody tr:nth-child(even){background:var(--row-alt)}
	        td{padding:10px 10px;vertical-align:middle;font-size:13px}
	        .badge{font-size:11px;padding:2px 8px;border-radius:999px;border:1px solid rgba(255,255,255,.12)}
	        .lvl-info{color:#cfe8ff;background:rgba(144,202,249,.12);border-color:rgba(144,202,249,.28)}
	        .lvl-warn{color:#fde68a;background:rgba(253,230,138,.12);border-color:rgba(253,230,138,.3)}
	        .lvl-error{color:#fda4af;background:rgba(253,164,175,.12);border-color:rgba(253,164,175,.3)}
	        .footer{display:flex;justify-content:space-between;align-items:center;padding:12px 16px;border-top:1px solid var(--outline)}
	        .muted{opacity:.75;font-size:12px}
	        .md-btn{--elev:0 2px 4px rgba(0,0,0,.35);--elevH:0 6px 16px rgba(0,0,0,.45);display:inline-flex;align-items:center;gap:8px;padding:10px 14px;border-radius:10px;border:1px solid rgba(255,255,255,.06);background:linear-gradient(180deg,#2196f3,#1976d2);color:#e8f2ff;font-weight:600;letter-spacing:.3px;cursor:pointer;box-shadow:var(--elev);transition:box-shadow .25s,transform .1s,filter .2s}
	        .md-btn:hover{box-shadow:var(--elevH);filter:saturate(1.05)}
	        .md-btn:active{transform:translateY(1px)}
	        .md-btn.outlined{background:transparent;border:1px solid var(--outline);color:var(--primary)}
	        @media (max-width:720px){ thead{display:none} table, tbody, tr, td{display:block;width:100%} tbody tr{margin:8px 0;border-radius:10px} td{padding:8px 12px} }
	    </style>
	</head>
	<body>
	    <div class="appbar">
	        <div class="wrap">
	            <div class="row">
	                <span class="material-symbols-outlined" aria-hidden="true">list</span>
	                <div class="title">システムログ</div>
	                <div class="sub">アプリケーションイベントの一覧</div>
	                <div style="margin-left:auto"><a href="/" class="md-btn outlined"><span class="material-symbols-outlined" style="font-size:18px">home</span> ホーム</a></div>
	            </div>
	        </div>
	    </div>
	    <div class="wrap">
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
	                    <button class="md-btn" onclick="loadLogs()"><span class="material-symbols-outlined" style="font-size:18px">sync</span> 更新</button>
	                    <button class="md-btn outlined" onclick="clearLogs()"><span class="material-symbols-outlined" style="font-size:18px">delete</span> 全クリア</button>
	                    <button class="md-btn outlined" onclick="exportLogs()"><span class="material-symbols-outlined" style="font-size:18px">download</span> エクスポート</button>
	                </div>
	                <div class="status">
	                    <label style="display:inline-flex;align-items:center;gap:6px">
	                        <input id="auto" type="checkbox" checked onchange="toggleAuto()"> 自動更新
	                    </label>
	                    <select id="interval" onchange="resetAuto()">
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
	    <style id="activeBannerStyles">
	        .active-banner{position:fixed;left:0;right:0;bottom:0;background:#7f1d1d;color:#fee2e2;border-top:2px solid #ef4444;z-index:9999;box-shadow:0 -8px 30px rgba(0,0,0,.45)}
	        .active-banner .ab-wrap{max-width:1100px;margin:0 auto;padding:12px 16px;display:flex;align-items:center;gap:12px}
	        .md-btn.danger{background:linear-gradient(180deg,#ef4444,#b91c1c);color:#fff;border-color:rgba(255,255,255,.08)}
	        @keyframes abPulse{0%{filter:saturate(1)}50%{filter:saturate(1.25)}100%{filter:saturate(1)}}
	        .active-banner{animation:abPulse 2.5s ease-in-out infinite}
	    </style>
	    <div id="activeBanner" class="active-banner" style="display:none" role="alert" aria-live="assertive">
	        <div class="ab-wrap">
	            <span class="material-symbols-outlined" aria-hidden="true">warning</span>
	            <div class="ab-text"><strong>監視実行中</strong> — 作業後は必ず停止してください</div>
	            <button id="abStop" class="md-btn danger" style="margin-left:auto"><span class="material-symbols-outlined" style="font-size:18px">stop_circle</span> 停止する</button>
	        </div>
	    </div>
	    <script>
	        async function updateActiveBanner(){
	            // logs ページも初回のみ
	            if(window.__LOG_ACTIVE_BANNER_INIT){ return; }
	            window.__LOG_ACTIVE_BANNER_INIT=true;
	            const el=document.getElementById('activeBanner');
	            try{ const res=await fetch('/api/monitoring/active'); if(!res.ok){ if(el) el.style.display='none'; return;} const data=await res.json(); const active=(data.data&&typeof data.data.isActive!=='undefined')?data.data.isActive:data.isActive; if(el){ el.style.display=active?'block':'none'; } }catch(_){ if(el) el.style.display='none'; }
	        }
	        document.addEventListener('DOMContentLoaded',()=>{ updateActiveBanner(); const btn=document.getElementById('abStop'); if(btn) btn.addEventListener('click', stopFromBanner); });
	    </script>
	</body>
	</html>`

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte(html)); err != nil {
		h.logger.LogError("ERROR", "Failed to write HTML response", "", "", err, nil)
	}
}
