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
<html lang="ja">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>Home - YouTube Live Chat Monitor</title>
<link href="https://fonts.googleapis.com/css2?family=Roboto:wght@400;500;700&display=swap" rel="stylesheet">
<link href="https://fonts.googleapis.com/css2?family=Material+Symbols+Outlined:wght@400;700" rel="stylesheet" />
<link rel="stylesheet" href="/static/material.css">
<script src="/static/app.js"></script>
</head>
<body>
<div class="appbar"><div class="wrap"><div class="row"><span class="material-symbols-outlined" aria-hidden="true">home</span><div class="title">YouTube Live Chat Monitor</div><div class="sub">チャット参加者をリアルタイム収集</div><div style="margin-left:auto;display:flex;gap:8px;align-items:center"><span id="appbarMon" class="pill" style="display:none;align-items:center;gap:6px"><span class="material-symbols-outlined" style="font-size:16px;vertical-align:-3px">sensors</span> 監視中 <button id="appbarStop" class="md-btn outlined" style="padding:4px 8px;font-size:12px;margin-left:6px"><span class="material-symbols-outlined" style="font-size:16px">stop_circle</span> 停止</button></span><a href="/users" class="md-btn outlined"><span class="material-symbols-outlined" style="font-size:18px">group</span> ユーザー一覧</a><a href="/logs" class="md-btn outlined"><span class="material-symbols-outlined" style="font-size:18px">list</span> ログ</a><button id="uiSettingsToggle" class="md-btn outlined" title="表示設定"><span class="material-symbols-outlined" style="font-size:18px">tune</span></button></div></div></div></div>
<div id="uiSettingsPanel" class="settings-panel" role="dialog" aria-modal="false" aria-labelledby="uiSettingsTitle"><h3 id="uiSettingsTitle">表示設定</h3><div class="section"><label>テーマ<select id="uiTheme"><option value="dark">ダーク</option><option value="light">ライト</option></select></label><label>プライマリカラー<select id="uiPrimary"><option value="blue">Blue</option><option value="indigo">Indigo</option><option value="teal">Teal</option><option value="green">Green</option><option value="orange">Orange</option><option value="pink">Pink</option><option value="purple">Purple</option></select></label><label>角丸<select id="uiRadius"><option value="sm">小</option><option value="md" selected>中</option><option value="lg">大</option></select></label><label>密度<select id="uiDensity"><option value="compact">コンパクト</option><option value="comfortable" selected>標準</option><option value="loose">ゆったり</option></select></label></div></div>
<div class="wrap">
  <div id="runBanner" class="card" style="margin-bottom:14px;display:none"><div class="content" style="display:flex;gap:10px;align-items:center;flex-wrap:wrap"><span class="pill"><span class="material-symbols-outlined" style="font-size:18px;vertical-align:-4px">sensors</span> 監視中</span><span id="runInfo" class="sub"></span><a class="md-btn" style="margin-left:auto" href="/users"><span class="material-symbols-outlined" style="font-size:18px">group</span> ユーザー一覧を見る</a><button id="statusRefreshBtn" class="md-btn outlined" type="button" style="margin-left:4px;padding:8px 10px;font-size:12px"><span class="material-symbols-outlined" style="font-size:16px">refresh</span> ステータス更新</button></div></div>
  <div class="card"><div class="content">
    <form id="monitoringForm" class="section">
      <div><label for="videoInput">YouTube Video ID</label><input type="text" id="videoInput" name="videoInput" placeholder="例: VIDEO_ID" required></div>
      <div><label for="maxUsers">最大ユーザー数 (デフォルト: 1000)</label><input type="number" id="maxUsers" name="maxUsers" value="1000" min="1" max="10000"></div>
      <div><button id="startBtn" type="submit" class="md-btn"><span class="material-symbols-outlined" style="font-size:18px">play_circle</span> 監視を開始</button><span class="sub" style="margin-left:8px">開始後はユーザー一覧に遷移します</span></div>
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
 let activeVideoId=null; let activeStatus=false; let submitting=false; let detectTimer=null; let detecting=false;

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
   const maxUsers=parseInt(fd.get('maxUsers'))||1000;
   try{
     const res=await fetch('/api/monitoring/start',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({video_input:video,max_users:maxUsers})});
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

 document.addEventListener('DOMContentLoaded', ()=>{ refreshExisting().then(startDetectLoop); });
 document.getElementById('statusRefreshBtn').addEventListener('click', refreshExisting);
 window.addEventListener('beforeunload', ()=>{ if(detectTimer) clearInterval(detectTimer); });
})();
</script>
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
<html lang="ja"><head><meta charset="UTF-8"><meta name="viewport" content="width=device-width,initial-scale=1.0"><title>User List</title>
<link href="https://fonts.googleapis.com/css2?family=Roboto:wght@400;500;700&display=swap" rel="stylesheet"><link href="https://fonts.googleapis.com/css2?family=Material+Symbols+Outlined:wght@400;700" rel="stylesheet" />
<link rel="stylesheet" href="/static/material.css">
<script src="/static/app.js"></script>
</head><body>
<div class="appbar"><div class="wrap"><div class="row"><span class="material-symbols-outlined">group</span><div class="title">ユーザーリスト</div><div class="sub">YouTube Live Chat 参加者</div><div style="margin-left:auto;display:flex;gap:8px;align-items:center"><span id="appbarMon" style="display:none;font-size:12px;background:rgba(144,202,249,.12);color:#cfe8ff;padding:4px 10px;border-radius:999px">監視中</span><a href="/" class="md-btn outlined">ホーム</a><button id="uiSettingsToggle" class="md-btn outlined" title="表示設定"><span class="material-symbols-outlined" style="font-size:18px">tune</span></button></div></div></div></div>
<div id="uiSettingsPanel" class="settings-panel" role="dialog" aria-modal="false" aria-labelledby="uiSettingsTitle"><h3 id="uiSettingsTitle">表示設定</h3><div class="section"><label>テーマ<select id="uiTheme"><option value="dark">ダーク</option><option value="light">ライト</option></select></label><label>プライマリカラー<select id="uiPrimary"><option value="blue">Blue</option><option value="indigo">Indigo</option><option value="teal">Teal</option><option value="green">Green</option><option value="orange">Orange</option><option value="pink">Pink</option><option value="purple">Purple</option></select></label><label>角丸<select id="uiRadius"><option value="sm">小</option><option value="md" selected>中</option><option value="lg">大</option></select></label><label>密度<select id="uiDensity"><option value="compact">コンパクト</option><option value="comfortable" selected>標準</option><option value="loose">ゆったり</option></select></label></div></div>
<div class="wrap"><div class="card"><div class="toolbar"><div style="display:flex;gap:8px;flex-wrap:wrap;align-items:center"><input id="search" type="text" placeholder="名前/Channel IDで検索" oninput="renderUsers()" />
<select id="order" onchange="renderUsers()">
  <option value="first_seen" selected>参加順</option>
  <option value="kana">五十音順</option>
  <option value="message_count">発言数</option>
</select>
<label style="display:inline-flex;align-items:center;gap:6px;font-size:13px"><input id="autoEnd" type="checkbox"> 自動終了検知</label>
<button class="md-btn" onclick="manualRefresh()"><span class="material-symbols-outlined" style="font-size:18px">sync</span> 即時更新</button>
<button class="md-btn danger" onclick="stopMonitoring()"><span class="material-symbols-outlined" style="font-size:18px">stop_circle</span> 監視停止</button>
<button class="md-btn neutral" onclick="freezeAndStop()"><span class="material-symbols-outlined" style="font-size:18px">task_alt</span> 終了して固定</button>
</div><div class="right"><div id="status" class="status">初回取得中...</div></div></div><div class="content" style="padding:12px 16px"><div class="meta" style="margin-bottom:8px"><span id="count">0</span> 名 <span id="updated"></span> <span style="margin-left:12px;font-size:11px;opacity:.75">SSE更新</span></div><div id="userList"><div class="empty">データを読み込んでいます…</div></div></div></div></div>
<script>
let cachedUsers=[];let currentVideoId=localStorage.getItem('currentVideoId')||'';let es=null;let sseConnected=false;let retries=0;const maxBackoff=30000;let frozen=false;
function initUserList(){const st=document.getElementById('status');if(!currentVideoId){st.className='status offline';st.textContent='videoId が保存されていません';document.getElementById('userList').innerHTML='<div class="empty">監視セッションなし</div>';return;}connectSSE();}
function connectSSE(){if(frozen) return;const st=document.getElementById('status');const upd=document.getElementById('updated');if(es) es.close();st.textContent='SSE接続中...';try{es=new EventSource('/api/sse/'+encodeURIComponent(currentVideoId)+'/users');es.addEventListener('connected',()=>{if(frozen){es.close();return;}sseConnected=true;retries=0;st.className='status online';st.textContent='オンライン - 接続済み';});es.addEventListener('user_list',ev=>{if(frozen){es.close();return;}const data=JSON.parse(ev.data);cachedUsers=Array.isArray(data.users)?data.users:[];st.className='status online';st.textContent='オンライン - コメントユーザー数: '+(data.count||cachedUsers.length);upd.textContent='（更新: '+new Date().toLocaleTimeString()+'）';renderUsers();});es.addEventListener('timeout',()=>{if(frozen){es.close();return;}st.className='status offline';st.textContent='タイムアウト';scheduleReconnect();});es.addEventListener('monitoring_stopped',()=>{if(frozen){es.close();return;}st.className='status offline';st.textContent='停止済み';});es.addEventListener('error',()=>{if(frozen){es.close();return;}if(!sseConnected){st.className='status offline';st.textContent='接続失敗';}scheduleReconnect();});}catch(e){st.className='status offline';st.textContent='SSEエラー: '+e.message;scheduleReconnect();}}
function scheduleReconnect(){if(retries>8) return;const delay=Math.min(1000*Math.pow(2,retries),maxBackoff);retries++;setTimeout(connectSSE,delay);}
function manualRefresh(){if(frozen) return;connectSSE();}
function renderUsers(){const list=document.getElementById('userList');const q=(document.getElementById('search').value||'').toLowerCase();const order=(document.getElementById('order').value||'first_seen');let users=cachedUsers.slice();if(q){users=users.filter(u=>{const n=(u.display_name||u.displayName||'').toLowerCase();const c=(u.channel_id||u.channelID||'').toLowerCase();return n.includes(q)||c.includes(q);});}
  if(users.length>1){if(order==='kana'){users.sort((a,b)=>((a.display_name||a.displayName||'').toString()).toLocaleCompare((b.display_name||b.displayName||'').toString(),'ja'));}else if(order==='message_count'){users.sort((a,b)=>{const am=a.message_count||a.messageCount||0;const bm=b.message_count||b.messageCount||0; if(am===bm){return ((a.display_name||a.displayName||'').toString()).toLocaleCompare((b.display_name||b.displayName||'').toString(),'ja');} return bm-am;});}}
  document.getElementById('count').textContent=users.length;if(users.length===0){list.innerHTML='<div class="empty">該当するユーザーがいません</div>';return;}list.innerHTML='<ol>'+users.map(u=>'<li>'+escapeHtml(u.display_name||u.displayName||'')+'</li>').join('')+'</ol>'; }
function escapeHtml(s){return (s||'').replace(/[&<>"']/g,c=>({'&':'&amp;','<':'&lt;','>':'&gt;','"':'&quot;','\'':'&#039;'}[c]));}
function stopMonitoring(){if(!confirm('監視を停止しますか？'))return;fetch('/api/monitoring/stop',{method:'DELETE'}).then(r=>r.json()).then(d=>{if(d.success){localStorage.removeItem('currentVideoId');alert('監視を停止しました。ホームへ戻ります。');location.href='/';}else alert('エラー: '+(d.error||'unknown'));}).catch(e=>alert('通信エラー: '+e.message));}
function freezeAndStop(){if(!confirm('現在のリストで固定表示します。監視を停止しますか？'))return;try{frozen=true;if(es){es.close();}const st=document.getElementById('status');st.className='status offline';st.textContent='固定表示（監視停止済み）';fetch('/api/monitoring/stop',{method:'DELETE'}).catch(()=>{});}catch(_){} }
function loadAutoEnd(){const cb=document.getElementById('autoEnd');if(!cb)return;fetch('/api/monitoring/auto-end').then(r=>r.json()).then(d=>{if(d.success){cb.checked=!!d.enabled;}}).catch(()=>{});cb.addEventListener('change',()=>{fetch('/api/monitoring/auto-end',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({enabled:cb.checked})}).catch(()=>{cb.checked=!cb.checked;});});}
function detectActive(videoId){return new Promise(res=>{if(!videoId)return res(false);try{const ev=new EventSource('/api/sse/'+encodeURIComponent(videoId)+'/users');let decided=false;const done=v=>{if(decided)return;decided=true;ev.close();res(v)};const to=setTimeout(()=>done(false),2000);ev.addEventListener('connected',()=>{clearTimeout(to);done(true)});ev.addEventListener('user_list',()=>{clearTimeout(to);done(true)});ev.addEventListener('error',()=>{clearTimeout(to);done(false)});}catch{res(false);}});} 
function updateAppbar(){const pill=document.getElementById('appbarMon');const vid=localStorage.getItem('currentVideoId')||'';if(!vid){pill.style.display='none';return;}detectActive(vid).then(a=>{pill.style.display=a?'inline-flex':'none';});}
window.addEventListener('DOMContentLoaded',()=>{initUserList();updateAppbar();loadAutoEnd();});
</script>
</body></html>`

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(html))
}

// ServeLogsPage GET /logs を処理します
func (h *StaticHandler) ServeLogsPage(w http.ResponseWriter, r *http.Request) {
    h.logger.LogAPI(constants.LogLevelInfo, "Logs page request", "", "", map[string]interface{}{"userAgent": r.Header.Get("User-Agent"), "remoteAddr": r.RemoteAddr})

	// Avoid JS template literals inside Go raw string to prevent syntax issues
    html := `<!DOCTYPE html><html lang="ja"><head><meta charset="UTF-8"><meta name="viewport" content="width=device-width,initial-scale=1.0"><title>System Logs</title>
<link href="https://fonts.googleapis.com/css2?family=Roboto:wght@400;500;700&display=swap" rel="stylesheet"><link href="https://fonts.googleapis.com/css2?family=Material+Symbols+Outlined:wght@400;700" rel="stylesheet" />
<link rel="stylesheet" href="/static/material.css">
<script src="/static/app.js"></script>
</head><body>
<div class="appbar"><div class="wrap"><div class="row"><span class="material-symbols-outlined">list</span><div class="title">システムログ</div><div class="sub">アプリケーションイベント</div><div style="margin-left:auto;display:flex;gap:8px;align-items:center"><a href="/" class="md-btn outlined"><span class="material-symbols-outlined" style="font-size:18px">home</span> ホーム</a><button id="uiSettingsToggle" class="md-btn outlined" title="表示設定"><span class="material-symbols-outlined" style="font-size:18px">tune</span></button></div></div></div></div>
<div id="uiSettingsPanel" class="settings-panel" role="dialog" aria-modal="false" aria-labelledby="uiSettingsTitle"><h3 id="uiSettingsTitle">表示設定</h3><div class="section"><label>テーマ<select id="uiTheme"><option value="dark">ダーク</option><option value="light">ライト</option></select></label><label>プライマリカラー<select id="uiPrimary"><option value="blue">Blue</option><option value="indigo">Indigo</option><option value="teal">Teal</option><option value="green">Green</option><option value="orange">Orange</option><option value="pink">Pink</option><option value="purple">Purple</option></select></label><label>角丸<select id="uiRadius"><option value="sm">小</option><option value="md" selected>中</option><option value="lg">大</option></select></label><label>密度<select id="uiDensity"><option value="compact">コンパクト</option><option value="comfortable" selected>標準</option><option value="loose">ゆったり</option></select></label></div></div>
<div class="wrap"><div class="card"><div class="toolbar"><div class="controls" style="display:flex;flex-wrap:wrap;gap:8px;align-items:center"><select id="level"><option value="">全レベル</option><option value="INFO">INFO</option><option value="WARNING">WARNING</option><option value="ERROR">ERROR</option></select><input id="component" type="text" placeholder="component で絞り込み"/><input id="video" type="text" placeholder="video_id で絞り込み"/><select id="limit"><option value="100">100件</option><option value="300" selected>300件</option><option value="1000">1000件</option></select><button class="md-btn" onclick="loadLogs()"><span class="material-symbols-outlined" style="font-size:18px">sync</span> 更新</button><button class="md-btn outlined" onclick="clearLogs()"><span class="material-symbols-outlined" style="font-size:18px">delete</span> 全クリア</button><button class="md-btn outlined" onclick="exportLogs()"><span class="material-symbols-outlined" style="font-size:18px">download</span> エクスポート</button></div><div class="status"><label style="display:inline-flex;align-items:center;gap:6px"><input id="auto" type="checkbox" checked onchange="toggleAuto()"> 自動更新</label><select id="interval" onchange="resetAuto()"><option value="10000" selected>10秒</option><option value="30000">30秒</option></select></div></div><div class="content" style="padding:14px 16px"><div class="muted" id="meta">読み込み中…</div><div id="logTable" style="margin-top:12px"></div><div class="muted" id="stats"></div></div></div></div>
<script>
var autoTimer=null;
function loadStats(params){fetch('/api/logs'+q(Object.assign({},params,{stats:1}))).then(r=>r.json()).then(d=>{if(d.success){document.getElementById('stats').textContent='統計: 総数 '+d.total+' / エラー '+d.errors+' / 警告 '+d.warnings;}}).catch(()=>{});} 
function loadLogs(){const meta=document.getElementById('meta');const table=document.getElementById('logTable');var params={};var lv=document.getElementById('level').value;if(lv)params.level=lv;var comp=document.getElementById('component').value.trim();if(comp)params.component=comp;var vid=document.getElementById('video').value.trim();if(vid)params.video_id=vid;params.limit=document.getElementById('limit').value||'300';fetch('/api/logs'+q(params)).then(r=>r.json()).then(d=>{if(d.success){var logs=Array.isArray(d.logs)?d.logs:[];meta.textContent='件数: '+logs.length+'（更新: '+new Date().toLocaleTimeString()+'）';if(!logs.length){table.innerHTML='<div class="muted">ログがありません</div>';}else{table.innerHTML='<table><thead><tr><th>時刻</th><th>Level</th><th>メッセージ</th><th>component</th><th>video_id</th><th>correlation</th></tr></thead><tbody>'+logs.map(l=>'<tr><td>'+esc(l.timestamp||'')+'</td><td>'+badge(l.level||'')+'</td><td>'+esc(l.message||'')+'</td><td>'+esc(l.component||'')+'</td><td>'+esc(l.video_id||'')+'</td><td>'+esc(l.correlation_id||'')+'</td></tr>').join('')+'</tbody></table>';}}else meta.textContent='エラー: '+(d.error||'unknown');loadStats(params);}).catch(e=>{meta.textContent='通信エラー: '+e.message;});}
function startAuto(){stopAuto();if(!document.getElementById('auto').checked)return;var ms=parseInt(document.getElementById('interval').value||'10000',10);autoTimer=setInterval(loadLogs,ms);}function stopAuto(){if(autoTimer){clearInterval(autoTimer);autoTimer=null;}}function resetAuto(){stopAuto();startAuto();}function toggleAuto(){document.getElementById('auto').checked?startAuto():stopAuto();}
function clearLogs(){if(!confirm('すべてのログをクリアしますか？'))return;fetch('/api/logs',{method:'DELETE'}).then(r=>r.json()).then(d=>{if(d.success){loadLogs();}else alert('エラー: '+(d.error||'unknown'));}).catch(e=>alert('通信エラー: '+e.message));}
function exportLogs(){var params={};var lv=document.getElementById('level').value;if(lv)params.level=lv;var comp=document.getElementById('component').value.trim();if(comp)params.component=comp;var vid=document.getElementById('video').value.trim();if(vid)params.video_id=vid;params.limit=document.getElementById('limit').value||'300';params.export=1;var url='/api/logs'+q(params);var a=document.createElement('a');a.href=url;a.download='logs_'+new Date().toISOString().split('T')[0]+'.json';document.body.appendChild(a);a.click();a.remove();}
window.addEventListener('load',function(){loadLogs();startAuto();});window.addEventListener('beforeunload',stopAuto);
</script>
</body></html>`

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(html))
}
