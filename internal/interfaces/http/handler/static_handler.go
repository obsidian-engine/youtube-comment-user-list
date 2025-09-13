package handler

import (
	"net/http"

	"github.com/obsidian-engine/youtube-comment-user-list/internal/constants"
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
	h.logger.LogAPI(constants.LogLevelInfo, "Home page request", "", "", map[string]interface{}{
		"userAgent":  r.Header.Get("User-Agent"),
		"remoteAddr": r.RemoteAddr,
	})

	html := `<!DOCTYPE html>
<html lang="ja" data-theme="light">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>Home - YouTube Live Chat Monitor</title>
<link href="https://fonts.googleapis.com/css2?family=Inter:wght@400;600;700&display=swap" rel="stylesheet">
<link rel="stylesheet" href="/static/ui.css">
<script src="/static/app.js"></script>
<link rel="stylesheet" href="https://fonts.googleapis.com/css2?family=Material+Symbols+Outlined" />
</head>
<body>
<div class="sidebar">
  <div class="logo">YT Monitor</div>
  <nav>
    <a href="/" class="active"><span class="icon material-symbols-outlined">home</span>ホーム</a>
    <a href="/users"><span class="icon material-symbols-outlined">group</span>ユーザー一覧</a>
    <a href="/logs"><span class="icon material-symbols-outlined">list</span>ログ</a>
  </nav>
  <div class="theme-box" style="margin:12px 12px 0">
    <button id="themeToggle" class="btn secondary" type="button" style="padding:6px 10px;font-size:12px"><span class="material-symbols-outlined" style="font-size:18px">dark_mode</span> テーマ切替</button>
    <button id="themeReset" class="btn secondary" type="button" title="システム設定に戻す" style="padding:6px 8px;font-size:12px;margin-left:6px"><span class="material-symbols-outlined" style="font-size:18px">settings_backup_restore</span></button>
  </div>
</div>
<div class="main fade-in">
  <div class="appbar" style="box-shadow:none;background:transparent;border:none;padding:0 0 24px 0">
    <div class="page-header" style="gap:18px;">
      <span class="material-symbols-outlined" aria-hidden="true" style="font-size:28px;color:#2563eb;">home</span>
      <div>
        <div class="title" style="font-size:24px;">YouTube Live Chat Monitor</div>
        <div class="sub">チャット参加者をリアルタイム収集</div>
      </div>
    </div>
  </div>
  <div class="breadcrumb" style="margin-bottom:12px"><a href="/">Home</a></div>
  <div id="runBanner" class="card" style="margin-bottom:14px;display:none"><div class="card-content" style="display:flex;gap:10px;align-items:center;flex-wrap:wrap"><span class="badge"><span class="material-symbols-outlined" style="font-size:18px;vertical-align:-4px">sensors</span> 監視中</span><span id="runInfo" class="sub"></span><a class="btn" style="margin-left:auto" href="/users"><span class="material-symbols-outlined" style="font-size:18px">group</span> ユーザー一覧を見る</a><button id="statusRefreshBtn" class="btn secondary" type="button" style="margin-left:4px;padding:8px 10px;font-size:12px"><span class="material-symbols-outlined" style="font-size:16px">refresh</span> ステータス更新</button></div></div>
  <div class="card fade-in"><div class="card-content">
    <form id="monitoringForm" class="section">
      <div><label for="videoInput">YouTube Video ID</label><input type="text" id="videoInput" name="videoInput" placeholder="例: VIDEO_ID" required></div>
      <div><label for="maxUsers">最大ユーザー数 (未指定時は 200)</label><input type="number" id="maxUsers" name="maxUsers" placeholder="未指定時は 200" min="1" max="10000"></div>
      <div><button id="startBtn" type="submit" class="btn primary"><span class="material-symbols-outlined" style="font-size:18px">play_circle</span> 監視を開始</button><button id="resumeBtn" type="button" class="btn secondary" style="margin-left:8px;display:none"><span class="material-symbols-outlined" style="font-size:18px">play_circle</span> 前回の動画で再開</button><span class="sub" style="margin-left:8px">開始後はユーザー一覧に遷移します</span></div>
    </form>
    <div id="message" class="sub" style="min-height:22px;margin-top:6px"></div>
  </div></div>
</div>
<script>
(function(){
 const form=document.getElementById('monitoringForm');
 const startBtn=document.getElementById('startBtn');
 const msg=document.getElementById('message');
 const videoInput=document.getElementById('videoInput');
 const runBanner=document.getElementById('runBanner');
 const runInfo=document.getElementById('runInfo');
 let activeVideoId=null; let activeStatus=false; let submitting=false; let detectTimer=null; let detecting=false; const resumeBtn=document.getElementById('resumeBtn');

 // Configurable via query params (?detectInterval=30000&detectTimeout=2500)
 const qs=new URLSearchParams(location.search);
 const DETECT_INTERVAL_MS = Math.max(5000, parseInt(qs.get('detectInterval')||'60000',10));
 const DETECT_TIMEOUT_MS  = Math.max(500,  parseInt(qs.get('detectTimeout') ||'2500',10));

 function setBtnState(disabled, text){
   if(!startBtn) return;
   startBtn.disabled=disabled;
   if(text){ startBtn.dataset.origText = startBtn.dataset.origText || startBtn.innerHTML; startBtn.innerHTML=text; }
   else if(!disabled && startBtn.dataset.origText){ startBtn.innerHTML=startBtn.dataset.origText; }
 }

 const detect = (videoId) => new Promise(resolve => {
   if(!videoId){ return resolve(false); }
   if(detecting) { // simple guard to avoid piling
     return resolve(activeStatus && activeVideoId===videoId);
   }
   detecting=true;
   try {
     const es = new EventSource('/api/sse/'+encodeURIComponent(videoId)+'/users');
     let decided=false;
     const finish = (val) => { if(decided) return; decided=true; es.close(); detecting=false; resolve(val); };
     const timeout = setTimeout(() => finish(false), DETECT_TIMEOUT_MS);
     es.addEventListener('connected', () => { clearTimeout(timeout); finish(true); });
     es.addEventListener('user_list', () => { clearTimeout(timeout); finish(true); });
     es.addEventListener('error', () => { clearTimeout(timeout); finish(false); });
   } catch { detecting=false; resolve(false); }
 });

 async function refreshExisting(){
   const stored=localStorage.getItem('currentVideoId')||'';
   if(!stored){
     activeVideoId=null; activeStatus=false;
     runBanner.style.display='none';
     return;
   }
   const active=await detect(stored);
   if(active){
     activeVideoId=stored; activeStatus=true;
     msg.innerHTML='<span style="color:#86efac">現在監視中: '+stored+' <a href="/users" style="color:#90caf9">ユーザー一覧へ移動</a></span>';
     if(videoInput && !videoInput.value) videoInput.value=stored;
     runInfo.textContent='videoId: '+stored+' / 状態: 起動中';
     runBanner.style.display='block';
   } else {
     if(activeStatus){ // transitioned to inactive
       msg.innerHTML='<span style="color:#fda4af">前回の監視は停止しました</span>';
     }
     activeVideoId=null; activeStatus=false;
     runBanner.style.display='none';
   }
 }

 function startDetectLoop(){
   if(detectTimer) clearInterval(detectTimer);
   detectTimer=setInterval(refreshExisting, DETECT_INTERVAL_MS);
 }

 form.addEventListener('submit',async(e)=>{
   e.preventDefault();
   if(submitting) return; // 二重送信防止
   const fd=new FormData(form);
   const video=(fd.get('videoInput')||'').trim();
   if(!video){ msg.textContent='Video ID を入力してください'; return; }
   if(activeStatus && activeVideoId===video){
      msg.innerHTML='<span style="color:#86efac">既に監視中です。遷移します…</span>';
      setTimeout(()=>location.href='/users',400);
      return;
   }
   submitting=true; setBtnState(true,'開始中…'); msg.textContent='バリデーション中...';
   const maxStr = (fd.get('maxUsers')||'').trim();
   try{
     const payload = { video_input: video };
     if (maxStr) payload.max_users = Number(maxStr);
     const res=await fetch('/api/monitoring/start',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify(payload)});
     const data=await res.json().catch(()=>({success:false,error:'invalid json'}));
     if(data.success){
       localStorage.setItem('currentVideoId', video);
       msg.innerHTML='<span style="color:#86efac">開始しました。遷移します…</span>';
       runInfo.textContent='videoId: '+video+' / 状態: 起動中';
       runBanner.style.display='block';
       setTimeout(()=>location.href='/users',600);
     } else {
       submitting=false; setBtnState(false); msg.innerHTML='<span style="color:#fda4af">エラー: '+(data.error||'unknown')+'</span>';
     }
   }catch(err){ submitting=false; setBtnState(false); msg.innerHTML='<span style="color:#fda4af">通信エラー: '+err.message+'</span>'; }
 });

 document.addEventListener('DOMContentLoaded', ()=>{ refreshExisting().then(()=>{ // 前回IDがあり未稼働なら再開ボタンを出す
  try{ const stored=localStorage.getItem('currentVideoId')||''; if(resumeBtn){ resumeBtn.style.display = (!activeStatus && stored) ? 'inline-flex' : 'none'; resumeBtn.onclick = async ()=>{ await resumeMonitoring(stored); }; } }catch(_){}
  startDetectLoop();
 }); });
 // 監視停止(ホーム) ボタンイベントをtoast/confirmModalに統一
 const appbarStop=document.getElementById('appbarStop');
 if(appbarStop){ appbarStop.addEventListener('click', async ()=>{ if(!(await confirmModal('監視停止','監視を停止しますか？','停止','キャンセル'))) return; try{ const r=await fetch('/api/monitoring/stop',{method:'DELETE'}); const d=await r.json(); if(d.success){ try{ localStorage.removeItem('currentVideoId'); }catch(_){} toast('監視を停止しました',{type:'success'}); refreshExisting(); } else { toast('停止失敗: '+(d.error||'unknown'),{type:'error'}); } }catch(e){ toast('通信エラー: '+e.message,{type:'error'}); } }); }
 document.getElementById('statusRefreshBtn').addEventListener('click', refreshExisting);
 window.addEventListener('beforeunload', ()=>{ if(detectTimer) clearInterval(detectTimer); });
})();
</script>
</div>
</body>
</html>`

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(html))
}

// ServeUserListPage GET /users を処理します
func (h *StaticHandler) ServeUserListPage(w http.ResponseWriter, r *http.Request) {
	h.logger.LogAPI(constants.LogLevelInfo, "User list page request", "", "", map[string]interface{}{
		"userAgent":  r.Header.Get("User-Agent"),
		"remoteAddr": r.RemoteAddr,
	})

	html := `<!DOCTYPE html>
<html lang="ja" data-theme="light">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>User List - YouTube Live Chat Monitor</title>
    <link href="https://fonts.googleapis.com/css2?family=Inter:wght@400;600;700&display=swap" rel="stylesheet">
    <link rel="stylesheet" href="/static/ui.css">
    <script src="/static/app.js"></script>
    <link rel="stylesheet" href="https://fonts.googleapis.com/css2?family=Material+Symbols+Outlined" />
</head>
<body>
    <div class="sidebar">
      <div class="logo">YT Monitor</div>
      <nav>
        <a href="/"><span class="icon material-symbols-outlined">home</span>ホーム</a>
        <a href="/users" class="active"><span class="icon material-symbols-outlined">group</span>ユーザー一覧</a>
        <a href="/logs"><span class="icon material-symbols-outlined">list</span>ログ</a>
      </nav>
      <div class="theme-box" style="margin:12px 12px 0">
        <button id="themeToggle" class="btn secondary" type="button" style="padding:6px 10px;font-size:12px"><span class="material-symbols-outlined" style="font-size:18px">dark_mode</span> テーマ切替</button>
        <button id="themeReset" class="btn secondary" type="button" title="システム設定に戻す" style="padding:6px 8px;font-size:12px;margin-left:6px"><span class="material-symbols-outlined" style="font-size:18px">settings_backup_restore</span></button>
      </div>
    </div>
    <div class="main fade-in">
    <div class="appbar" style="box-shadow:none;background:transparent;border:none;padding:0 0 24px 0">
        <div class="page-header" style="gap:18px;">
            <span class="material-symbols-outlined" aria-hidden="true" style="font-size:28px;color:#2563eb;">group</span>
            <div>
                <div class="title" style="font-size:24px;">ユーザーリスト</div>
                <div class="sub">YouTube Live Chat 参加者</div>
            </div>
            <div class="page-actions ml-auto"><canvas id="sparkUsers" width="120" height="28" aria-hidden="true"></canvas></div>
        </div>
    </div>
    <div class="breadcrumb" style="margin-bottom:12px"><a href="/">Home</a><span class="sep"> / </span><span>ユーザー一覧</span></div>
    <div class="card fade-in">
        <div class="toolbar">
            <div class="left controls">
                <input id="search" type="text" placeholder="名前・Channel IDで検索" oninput="renderUsers()"/>
                <select id="order" onchange="renderUsers()">
                  <option value="first_seen" selected>参加順</option>
                  <option value="kana">五十音順</option>
                  <option value="message_count">発言数</option>
                </select>
                <label style="display:inline-flex;align-items:center;gap:6px;font-size:13px"><input id="autoEnd" type="checkbox" checked> 自動終了検知</label>
                <button class="btn primary" onclick="manualRefresh()"><span class="material-symbols-outlined" style="font-size:18px">sync</span> 即時更新</button>
                <button class="btn danger" onclick="stopMonitoring()"><span class="material-symbols-outlined" style="font-size:18px">stop_circle</span> 監視停止</button>
            </div>
            <div class="right">
                <div id="status" class="status">初回取得中...</div>
            </div>
        </div>
        <div id="warn" style="display:none;padding:10px 16px;color:#fecaca;border-top:1px solid var(--line);background:rgba(244,63,94,.08)">警告: LIVEが inactive の可能性。監視は起動中です。</div>
        <div class="card-content">
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
            <div><a href="/" class="btn secondary"><span class="material-symbols-outlined" style="font-size:18px">home</span> ホーム</a></div>
        </div>
    </div>
    </div>

    <script>
        // --- 設定 ---
        const REFRESH_INTERVAL_MS = 60000; // 60秒固定
        // --- 状態 ---
        let cachedUsers = []; let isActive = false; let refreshTimer = null; let currentVideoId = ''; let fetching = false; let lastStatusFetch = 0; let frozen = false; let usersSeries = [];
        // 初期化
        window.addEventListener('DOMContentLoaded', () => { initialLoad(); loadAutoEnd(); setupNavGuard(); });
        window.addEventListener('beforeunload', () => { if (refreshTimer) clearInterval(refreshTimer); });

        // 共通確認モーダル経由でナビゲーション時に停止
        function setupNavGuard(){ const links=document.querySelectorAll('.sidebar nav a'); links.forEach(a=>{ if(a.classList.contains('active')) return; a.addEventListener('click', async (e)=>{ if(!isActive) return; e.preventDefault(); const ok=await confirmModal('監視を停止して移動','移動すると監視を停止します。続行しますか？','停止して移動','キャンセル'); if(!ok) return; await stopMonitoringCore(true); location.href=a.href; }); }); }

        async function initialLoad(){ const list=document.getElementById('userList'); if(list && window.buildSkeletonTable){ list.innerHTML=buildSkeletonTable(6,8); } await refreshUsers(true); if(!refreshTimer && !frozen){ refreshTimer=setInterval(()=>{ refreshUsers(false); }, REFRESH_INTERVAL_MS);} }
        async function manualRefresh(){ if(frozen) return; refreshUsers(false,{force:true}); }

        async function resumeMonitoring(video){ try{ const body=video?{video_input:String(video)}:{}; const r=await fetch('/api/monitoring/resume',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify(body)}); const d=await r.json(); if(d.success){ try{ if(d.videoId) localStorage.setItem('currentVideoId', d.videoId); }catch(_){} await fetch('/api/monitoring/auto-end',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({enabled:true})}).catch(()=>{}); toast('監視を再開しました',{type:'success'}); refreshUsers(false,{force:true}); } else { toast('再開失敗: '+(d.error||'unknown'),{type:'error'}); } }catch(e){ toast('通信エラー: '+e.message,{type:'error'}); } }

        async function refreshUsers(isInitial, opts={}){ if(frozen) return; if(fetching && !opts.force) return; fetching=true; const statusDiv=document.getElementById('status'); const updated=document.getElementById('updated'); statusDiv.textContent=isInitial?'初回取得中...':'更新中...'; try { const activeRes=await fetch('/api/monitoring/active',{cache:'no-store'}); if(!activeRes.ok){ if(activeRes.status===404){ statusDiv.className='status offline'; statusDiv.textContent='監視セッションがありません'; document.getElementById('userList').innerHTML='<div class="empty">監視を開始するにはホームへ戻ってください。</div>'; } else { statusDiv.className='status offline'; statusDiv.textContent='アクティブ確認失敗 ('+activeRes.status+')'; } fetching=false; return; } const activeData=await activeRes.json(); const videoId=(activeData.data && activeData.data.videoId)||activeData.videoId; isActive=(activeData.data && typeof activeData.data.isActive!=='undefined')?activeData.data.isActive:activeData.isActive; currentVideoId=videoId||currentVideoId; if(!videoId){ statusDiv.className='status offline'; statusDiv.textContent='videoId 取得不可'; fetching=false; return; } const listRes=await fetch('/api/monitoring/'+encodeURIComponent(videoId)+'/users',{cache:'no-store'}); if(!listRes.ok){ statusDiv.className='status offline'; statusDiv.textContent='ユーザー取得失敗 ('+listRes.status+')'; fetching=false; return; } const listData=await listRes.json(); if(!listData.success){ statusDiv.className='status offline'; statusDiv.textContent='エラー: '+(listData.error||'unknown'); fetching=false; return; } cachedUsers=Array.isArray(listData.users)?listData.users:[]; const cls=isActive?'status online':'status offline'; const txt=isActive?'オンライン':'停止済み'; statusDiv.className='status '+cls.split(' ').pop(); statusDiv.textContent=txt+' - コメントユーザー数: '+(listData.count ?? cachedUsers.length); updated.textContent='（更新: '+new Date().toLocaleTimeString()+'）'; updateLiveStatusWarning(videoId); renderUsers(); }catch(err){ statusDiv.className='status offline'; statusDiv.textContent='通信エラー: '+err.message; } finally { fetching=false; } }

        async function updateLiveStatusWarning(videoId){ const warn=document.getElementById('warn'); const now=Date.now(); if(now-lastStatusFetch<60000) return; try{ const sres=await fetch('/api/monitoring/'+encodeURIComponent(videoId)+'/status',{cache:'no-store'}); if(!sres.ok){ warn.style.display='none'; return; } const sdata=await sres.json(); lastStatusFetch=now; const statusVal=extractBroadcastStatus(sdata); const norm=(statusVal||'').toLowerCase(); const OK=['live','upcoming']; if(isActive && statusVal && !OK.includes(norm)){ warn.style.display='block'; warn.textContent='警告: LIVEステータスが '+statusVal+' の可能性。監視は起動中です。'; } else { warn.style.display='none'; } }catch{ warn.style.display='none'; } }
        function extractBroadcastStatus(resp){ if(!resp) return ''; if(typeof resp.status==='string') return resp.status; if(resp.data){ const d=resp.data; if(typeof d.status==='string') return d.status; } return ''; }

        function renderUsers(){ const list=document.getElementById('userList'); const q=(document.getElementById('search').value||'').toLowerCase(); const order=(document.getElementById('order').value||'first_seen'); let users=cachedUsers.slice(); if(q){ users=users.filter(u=>{ const name=(u.display_name||u.displayName||'').toLowerCase(); const cid=(u.channel_id||u.channelID||'').toLowerCase(); return name.includes(q)||cid.includes(q); }); } if(users.length>1){ if(order==='kana'){ users.sort((a,b)=> ((a.display_name||a.displayName||'')+'').localeCompare(((b.display_name||b.displayName||'')+'').toString(),'ja')); } else if(order==='message_count'){ users.sort((a,b)=>{ const am=a.message_count||a.messageCount||0; const bm=b.message_count||b.messageCount||0; if(am===bm){ return ((a.display_name||a.displayName||'')+'').localeCompare(((b.display_name||b.displayName||'')+'').toString(),'ja'); } return bm-am; }); } else { users.sort((a,b)=>{ const fa=new Date(a.first_seen||a.firstSeen||0).getTime(); const fb=new Date(b.first_seen||b.firstSeen||0).getTime(); if(fa!==fb) return fa-fb; return ((a.display_name||a.displayName||'')+'').localeCompare(((b.display_name||b.displayName||'')+'').toString(),'ja'); }); } } document.getElementById('count').textContent=users.length; if(users.length===0){ list.innerHTML='<div class="empty">該当するユーザーがいません</div>'; return; } let html='<table><thead><tr><th class="idx">#</th><th>ユーザー名</th><th>Channel ID</th><th>初回参加</th><th>発言数</th><th></th></tr></thead><tbody>'; users.forEach((u,i)=>{ const name=(u.display_name||u.displayName||''); const cid=(u.channel_id||u.channelID||''); const first=fmtDate(u.first_seen||u.firstSeen); const msgCount=(u.message_count!=null)?u.message_count:(u.messageCount||0); const url= cid ? 'https://www.youtube.com/channel/'+encodeURIComponent(cid):'#'; html+='<tr><td class="idx">'+(i+1)+'</td><td class="name">'+escapeHtml(name)+'</td><td><a href="'+url+'" target="_blank" rel="noopener">'+escapeHtml(cid)+'</a></td><td><span class="pill">'+first+'</span></td><td>'+msgCount+'</td><td class="actions"><button class="btn secondary" onclick="navigator.clipboard.writeText(\''+cid+'\').then(()=>toast(\'コピーしました\',{type:\'success\'})).catch(()=>toast(\'コピー失敗\',{type:\'error\'}))">Copy</button></td></tr>'; }); html+='</tbody></table>'; list.innerHTML=html; }
        // render後にスパークラインを更新
        (function(){ try{ var _orig = renderUsers; renderUsers = function(){ _orig.apply(this, arguments); try{ usersSeries.push((cachedUsers||[]).length); if(usersSeries.length>40) usersSeries.shift(); var c=document.getElementById('sparkUsers'); if(c && window.drawSparkline) drawSparkline(c, usersSeries, '#2563eb'); }catch(_){ } }; }catch(_){ } })();
        function fmtDate(s){ try{return new Date(s).toLocaleString();}catch(e){return '-';} }

        // 自動終了設定
        function loadAutoEnd(){ const cb=document.getElementById('autoEnd'); if(!cb) return; fetch('/api/monitoring/auto-end').then(r=>r.json()).then(d=>{ if(d.success){ cb.checked=!!d.enabled; }}).catch(()=>{}); cb.addEventListener('change',()=>{ fetch('/api/monitoring/auto-end',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({enabled:cb.checked})}).catch(()=>{ cb.checked=!cb.checked; toast('更新失敗',{type:'error'}); }); }); }

        // 停止コア
        async function stopMonitoringCore(silent){ try{ const r=await fetch('/api/monitoring/stop',{method:'DELETE'}); const d=await r.json(); if(d.success){ try{ localStorage.removeItem('currentVideoId'); }catch(_){} try{ const cb=document.getElementById('autoEnd'); if(cb){ cb.checked=false; } fetch('/api/monitoring/auto-end',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({enabled:false})}); }catch(_){} isActive=false; const statusDiv=document.getElementById('status'); if(statusDiv){ statusDiv.className='status offline'; statusDiv.textContent='停止済み'; } if(!silent) toast('監視を停止しました',{type:'success'}); renderUsers(); return true; } else { if(!silent) toast('停止失敗: '+(d.error||'unknown'),{type:'error'}); return false; } }catch(e){ if(!silent) toast('通信エラー: '+e.message,{type:'error'}); return false; }

        async function stopMonitoring(){ if(!isActive){ toast('既に停止しています'); return; } const ok=await confirmModal('監視停止','現在の監視を停止しますか？','停止','キャンセル'); if(!ok) return; await stopMonitoringCore(false); }
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
	h.logger.LogAPI(constants.LogLevelInfo, "Logs page request", "", "", map[string]interface{}{"userAgent": r.Header.Get("User-Agent"), "remoteAddr": r.RemoteAddr})

	// Avoid JS template literals inside Go raw string to prevent syntax issues
 html := `<!DOCTYPE html><html lang="ja" data-theme="light"><head><meta charset="UTF-8"><meta name="viewport" content="width=device-width,initial-scale=1.0"><title>System Logs</title>
<link href="https://fonts.googleapis.com/css2?family=Inter:wght@400;600;700&display=swap" rel="stylesheet">
<link rel="stylesheet" href="https://fonts.googleapis.com/css2?family=Material+Symbols+Outlined" />
<link rel="stylesheet" href="/static/ui.css">
<script src="/static/app.js"></script>
</head><body>
<div class="sidebar"><div class="logo">YT Monitor</div><nav><a href="/"><span class="icon material-symbols-outlined">home</span>ホーム</a><a href="/users"><span class="icon material-symbols-outlined">group</span>ユーザー一覧</a><a href="/logs" class="active"><span class="icon material-symbols-outlined">list</span>ログ</a></nav><div class="theme-box" style="margin:12px 12px 0"><button id="themeToggle" class="btn secondary" type="button" style="padding:6px 10px;font-size:12px"><span class="material-symbols-outlined" style="font-size:18px">dark_mode</span> テーマ切替</button><button id="themeReset" class="btn secondary" type="button" title="システム設定に戻す" style="padding:6px 8px;font-size:12px;margin-left:6px"><span class="material-symbols-outlined" style="font-size:18px">settings_backup_restore</span></button></div></div>
<div class="main fade-in">
  <div class="appbar" style="box-shadow:none;background:transparent;border:none;padding:0 0 24px 0"><div class="page-header" style="gap:18px;"><span class="material-symbols-outlined" style="font-size:28px;color:#2563eb;">list</span><div><div class="title" style="font-size:24px;">システムログ</div><div class="sub">アプリケーションイベント</div></div><div class="page-actions ml-auto"><canvas id="sparkLogs" width="120" height="28" aria-hidden="true"></canvas></div></div></div>
  <div class="breadcrumb" style="margin-bottom:12px"><a href="/">Home</a><span class="sep"> / </span><span>ログ</span></div>
  <div class="card fade-in"><div class="toolbar"><div class="controls" style="display:flex;flex-wrap:wrap;gap:8px;align-items:center"><select id="level"><option value="">全レベル</option><option value="INFO">INFO</option><option value="WARNING">WARNING</option><option value="ERROR">ERROR</option></select><input id="component" type="text" placeholder="component で絞り込み"/><input id="video" type="text" placeholder="video_id で絞り込み"/><select id="limit"><option value="100">100件</option><option value="300" selected>300件</option><option value="1000">1000件</option></select><button class="btn primary" onclick="loadLogs()"><span class="material-symbols-outlined" style="font-size:18px">sync</span> 更新</button><button class="btn danger" onclick="clearLogs()"><span class="material-symbols-outlined" style="font-size:18px">delete</span> 全クリア</button><button class="btn secondary" onclick="exportLogs()"><span class="material-symbols-outlined" style="font-size:18px">download</span> エクスポート</button></div><div class="status"><label style="display:inline-flex;align-items:center;gap:6px"><input id="auto" type="checkbox" checked onchange="toggleAuto()"> 自動更新</label><select id="interval" onchange="resetAuto()"><option value="10000" selected>10秒</option><option value="30000">30秒</option></select></div></div><div class="card-content" style="padding:14px 16px"><div class="muted" id="meta">読み込み中…</div><div id="logTable" style="margin-top:12px"></div><div class="muted" id="stats"></div></div></div>
</div>
<script>
var autoTimer=null; var logsSeries=[];
function loadStats(params){fetch('/api/logs'+q(Object.assign({},params,{stats:1}))).then(r=>r.json()).then(d=>{if(d.success){document.getElementById('stats').textContent='統計: 総数 '+d.total+' / エラー '+d.errors+' / 警告 '+d.warnings;}}).catch(()=>{});} 
function loadLogs(){const meta=document.getElementById('meta');const table=document.getElementById('logTable');if(table && window.buildSkeletonTable){ table.innerHTML=buildSkeletonTable(6,8); }var params={};var lv=document.getElementById('level').value;if(lv)params.level=lv;var comp=document.getElementById('component').value.trim();if(comp)params.component=comp;var vid=document.getElementById('video').value.trim();if(vid)params.video_id=vid;params.limit=document.getElementById('limit').value||'300';fetch('/api/logs'+q(params)).then(r=>r.json()).then(d=>{if(d.success){var logs=Array.isArray(d.logs)?d.logs:[];meta.textContent='件数: '+logs.length+'（更新: '+new Date().toLocaleTimeString()+'）';if(!logs.length){table.innerHTML='<div class="muted">ログがありません</div>';}else{table.innerHTML='<table><thead><tr><th>時刻</th><th>Level</th><th>メッセージ</th><th>component</th><th>video_id</th><th>correlation</th></tr></thead><tbody>'+logs.map(l=>'<tr><td>'+esc(l.timestamp||'')+'</td><td>'+badge(l.level||'')+'</td><td>'+esc(l.message||'')+'</td><td>'+esc(l.component||'')+'</td><td>'+esc(l.video_id||'')+'</td><td>'+esc(l.correlation_id||'')+'</td></tr>').join('')+'</tbody></table>';}}else meta.textContent='エラー: '+(d.error||'unknown');try{ logsSeries.push((d && d.logs && d.logs.length)||0); if(logsSeries.length>40) logsSeries.shift(); var c=document.getElementById('sparkLogs'); if(c && window.drawSparkline) drawSparkline(c, logsSeries, '#2563eb'); }catch(_){ } loadStats(params);}).catch(e=>{meta.textContent='通信エラー: '+e.message;});}
function startAuto(){stopAuto();if(!document.getElementById('auto').checked)return;var ms=parseInt(document.getElementById('interval').value||'10000',10);autoTimer=setInterval(loadLogs,ms);}function stopAuto(){if(autoTimer){clearInterval(autoTimer);autoTimer=null;}}function resetAuto(){stopAuto();startAuto();}function toggleAuto(){document.getElementById('auto').checked?startAuto():stopAuto();}
async function clearLogs(){
  const ok = await confirmModal('ログ全削除','すべてのログを削除します。よろしいですか？','削除','キャンセル');
  if(!ok) return;
  try {
    const res = await fetch('/api/logs',{method:'DELETE'});
    const data = await res.json();
    if(data.success){ toast('ログをクリアしました',{type:'success'}); loadLogs(); }
    else { toast('削除失敗: '+(data.error||'unknown'),{type:'error'}); }
  } catch(e){ toast('通信エラー: '+e.message,{type:'error'}); }
}
// exportLogs をトースト付きで安全化
function exportLogs(){
  try {
    var params={};
    var lv=document.getElementById('level').value; if(lv) params.level=lv;
    var comp=document.getElementById('component').value.trim(); if(comp) params.component=comp;
    var vid=document.getElementById('video').value.trim(); if(vid) params.video_id=vid;
    params.limit=document.getElementById('limit').value||'300';
    params.export=1;
    var url='/api/logs'+q(params);
    var a=document.createElement('a'); a.href=url; a.download='logs_'+new Date().toISOString().split('T')[0]+'.json'; document.body.appendChild(a); a.click(); a.remove();
    toast('エクスポート開始',{type:'success'});
  } catch(e) {
    toast('エクスポート失敗: '+e.message,{type:'error'});
  }
}
window.addEventListener('load',function(){loadLogs();startAuto();});window.addEventListener('beforeunload',stopAuto);
</script>
</body></html>`

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(html))
}
