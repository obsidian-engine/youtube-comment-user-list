// @ts-check
import { getActive, getUsers, getStatus, stopMonitoring } from '../api.js';

/**
 * @typedef {Object} ChatUser
 * @property {string=} display_name
 * @property {string=} displayName
 * @property {string=} channel_id
 * @property {string=} channelID
 * @property {string|number|Date=} first_seen
 * @property {string|number|Date=} firstSeen
 * @property {number=} message_count
 * @property {number=} messageCount
 */

const $ = (sel) => document.querySelector(sel);
const $$ = (sel) => Array.from(document.querySelectorAll(sel));
const esc = (s) => (window.esc ? window.esc(s) : String(s));

/** @type {ChatUser[]} */
let cachedUsers = [];
let isActive = false;
let refreshTimer = null;
let currentVideoId = '';
let fetching = false;
let lastStatusFetch = 0;
let usersSeries = [];

const REFRESH_INTERVAL_MS = 60000;

function setupNavGuard(){
  $$('.sidebar nav a').forEach((a)=>{
    if(a.classList.contains('active')) return;
    a.addEventListener('click', async (e)=>{
      if(!isActive) return;
      e.preventDefault();
      const ok = await window.confirmModal('監視を停止して移動','移動すると監視を停止します。続行しますか？','停止して移動','キャンセル');
      if(!ok) return;
      const stopped = await stopMonitoringCore(true);
      if(stopped) location.href = a.href; else window.toast('停止できませんでした',{type:'error'});
    });
  });
}

async function stopMonitoringCore(silent){
  try{ const { data } = await stopMonitoring();
    if(data && data.success){
      try{ localStorage.removeItem('currentVideoId'); }catch(_){ }
      try{ const cb = /** @type {HTMLInputElement|null} */($('#autoEnd')); if(cb){ cb.checked=false; fetch('/api/monitoring/auto-end',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({enabled:false})}); } }catch(_){ }
      isActive=false; const statusDiv=$('#status'); if(statusDiv){ statusDiv.className='status offline'; statusDiv.textContent='停止済み'; }
      if(!silent) window.toast('監視を停止しました',{type:'success'});
      renderUsers(); return true;
    } else { if(!silent) window.toast('停止失敗: '+(data && data.error || 'unknown'),{type:'error'}); return false; }
  }catch(e){ if(!silent) window.toast('通信エラー: '+(e && e.message || e),{type:'error'}); return false; }
}

async function refreshUsers(isInitial, opts={}){
  if(fetching && !opts.force) return; fetching=true;
  const statusDiv = $('#status'); const updated = $('#updated'); if(statusDiv) statusDiv.textContent = isInitial?'初回取得中...':'更新中...';
  try{
    const act = await getActive();
    if(!act.ok){
      if(act.status===404){ if(statusDiv){ statusDiv.className='status offline'; statusDiv.textContent='監視セッションがありません'; }
        const list=$('#userList'); if(list) list.innerHTML='<div class="empty">監視を開始するにはホームへ戻ってください。</div>';
      } else { if(statusDiv){ statusDiv.className='status offline'; statusDiv.textContent='アクティブ確認失敗 ('+act.status+')'; } }
      return;
    }
    const a = act.data || {};
    const videoId = (a.data && a.data.videoId) || a.videoId;
    isActive = (a.data && typeof a.data.isActive !== 'undefined') ? a.data.isActive : a.isActive;
    currentVideoId = videoId || currentVideoId;
    if(!videoId){ if(statusDiv){ statusDiv.className='status offline'; statusDiv.textContent='videoId 取得不可'; } return; }
    const ul = await getUsers(videoId);
    if(!ul.ok){ if(statusDiv){ statusDiv.className='status offline'; statusDiv.textContent='ユーザー取得失敗 ('+ul.status+')'; } return; }
    const d = ul.data || {}; if(!d.success){ if(statusDiv){ statusDiv.className='status offline'; statusDiv.textContent='エラー: '+(d.error||'unknown'); } return; }
    cachedUsers = Array.isArray(d.users) ? d.users : [];
    const cls = isActive ? 'online' : 'offline';
    const txt = isActive ? 'オンライン' : '停止済み';
    if(statusDiv){ statusDiv.className='status '+cls; statusDiv.textContent = txt+' - コメントユーザー数: '+(d.count ?? cachedUsers.length); }
    if(updated){ updated.textContent='（更新: '+new Date().toLocaleTimeString()+'）'; }
    updateLiveStatusWarning(videoId);
    renderUsers();
    try{ usersSeries.push(cachedUsers.length); if(usersSeries.length>40) usersSeries.shift(); const c = /** @type {HTMLCanvasElement|null} */($('#sparkUsers')); if(c && window.drawSparkline) window.drawSparkline(c, usersSeries, '#2563eb'); }catch(_){ }
  }finally{ fetching=false; }
}

/**
 * @param {ChatUser[]} users
 * @returns {ChatUser[]}
 */
function filterUsers(users) {
  const q = String((/** @type {HTMLInputElement} */($('#search'))||{value:''}).value || '').toLowerCase();
  if(!q) return users;
  return users.filter(u=>{
    const name = String(u.display_name||u.displayName||'').toLowerCase();
    const cid  = String(u.channel_id||u.channelID||'').toLowerCase();
    return name.includes(q)||cid.includes(q);
  });
}

/**
 * @param {ChatUser[]} users
 * @returns {ChatUser[]}
 */
function sortUsers(users) {
  const order = String((/** @type {HTMLSelectElement} */($('#order'))||{value:'first_seen'}).value || 'first_seen');
  const copy = users.slice();
  if(copy.length<=1) return copy;
  if(order==='kana')
    copy.sort((a,b)=> String(a.display_name||a.displayName||'').localeCompare(String(b.display_name||b.displayName||''),'ja'));
  else if(order==='message_count')
    copy.sort((a,b)=>{ const am=a.message_count||a.messageCount||0; const bm=b.message_count||b.messageCount||0; if(am===bm) return String(a.display_name||a.displayName||'').localeCompare(String(b.display_name||b.displayName||''),'ja'); return bm-am; });
  else
    copy.sort((a,b)=>{ const fa=new Date(a.first_seen||a.firstSeen||0).getTime(); const fb=new Date(b.first_seen||b.firstSeen||0).getTime(); if(fa!==fb) return fa-fb; return String(a.display_name||a.displayName||'').localeCompare(String(b.display_name||b.displayName||''),'ja'); });
  return copy;
}

/**
 * @param {ChatUser[]} users
 * @returns {string}
 */
function renderUsersTable(users){
  let html = '<div class="table-scroll"><table><thead><tr><th class="idx">#</th><th>ユーザー名</th><th>Channel ID</th><th>初回参加</th><th>発言数</th><th></th></tr></thead><tbody>';
  users.forEach((u,i)=>{
    const name=(u.display_name||u.displayName||'');
    const cid=(u.channel_id||u.channelID||'');
    const first=fmtDate(u.first_seen||u.firstSeen);
    const msgCount=(u.message_count!=null)?u.message_count:(u.messageCount||0);
    const url= cid ? 'https://www.youtube.com/channel/'+encodeURIComponent(cid):'#';
    html += '<tr><td class="idx">'+(i+1)+'</td><td class="name">'+esc(name)+'</td><td><a href="'+url+'" target="_blank" rel="noopener">'+esc(cid)+'</a></td><td><span class="pill">'+first+'</span></td><td>'+msgCount+'</td><td class="actions"><button class="btn secondary" data-action="users:copy" data-cid="'+esc(cid)+'">Copy</button></td></tr>';
  });
  html += '</tbody></table></div>';
  return html;
}

function renderUsers(){
  const list = $('#userList'); if(!list) return;
  let users = sortUsers(filterUsers(cachedUsers));
  const countEl = $('#count'); if(countEl) countEl.textContent = String(users.length);
  if(users.length===0){ list.innerHTML='<div class="empty">該当するユーザーがいません</div>'; return; }
  list.innerHTML = renderUsersTable(users);
}

function fmtDate(s){ try{ return new Date(s).toLocaleString(); }catch(_){ return '-'; } }

async function updateLiveStatusWarning(videoId){
  const warn = $('#warn'); if(!warn) return;
  const now = Date.now(); if(now - lastStatusFetch < 60000) return;
  try{ const { ok, data } = await getStatus(videoId); if(!ok){ warn.style.display='none'; return; } lastStatusFetch = now; const statusVal = (data && (data.status || (data.data && data.data.status))) || ''; const norm = String(statusVal).toLowerCase(); const OK = ['live','upcoming']; if(isActive && statusVal && !OK.includes(norm)){ warn.style.display='block'; warn.textContent='警告: LIVEステータスが '+statusVal+' の可能性。監視は起動中です。'; } else { warn.style.display='none'; } }catch{ warn.style.display='none'; }
}

function loadAutoEnd(){
  const cb = /** @type {HTMLInputElement|null} */($('#autoEnd')); if(!cb) return;
  fetch('/api/monitoring/auto-end').then(r=>r.json()).then(d=>{ if(d.success){ cb.checked=!!d.enabled; } }).catch(()=>{});
  cb.addEventListener('change',()=>{
    fetch('/api/monitoring/auto-end',{ method:'POST', headers:{'Content-Type':'application/json'}, body:JSON.stringify({enabled:cb.checked}) })
      .catch(()=>{ cb.checked=!cb.checked; window.toast('更新失敗',{type:'error'}); });
  });
}

async function initialLoad(){
  const list=$('#userList'); if(list && window.buildSkeletonTable){ list.innerHTML = window.buildSkeletonTable(6,8); }
  await refreshUsers(true);
  if(!refreshTimer){ refreshTimer = setInterval(()=>{ refreshUsers(false); }, REFRESH_INTERVAL_MS); }
}

// イベント委譲: data-* で 1 箇所バインド
document.addEventListener('click', async (e)=>{
  const t = /** @type {Element|null} */(e.target instanceof Element ? e.target.closest('[data-action]') : null);
  if(!t) return;
  const action = String(t.getAttribute('data-action')||'');
  switch(action){
    case 'users:manual-refresh':
      e.preventDefault();
      refreshUsers(false,{force:true});
      break;
    case 'users:stop':
      e.preventDefault();
      if(!isActive){ window.toast('既に停止しています'); return; }
      { const ok=await window.confirmModal('監視停止','現在の監視を停止しますか？','停止','キャンセル'); if(!ok) return; await stopMonitoringCore(false); }
      break;
    case 'users:copy':
      e.preventDefault();
      try{ const cid = String(t.getAttribute('data-cid')||''); await navigator.clipboard.writeText(cid); window.toast('コピーしました',{type:'success'});}catch{ window.toast('コピー失敗',{type:'error'}); }
      break;
    default:
      break;
  }
});

document.addEventListener('input', (e)=>{
  const el = /** @type {Element|null} */(e.target instanceof Element ? e.target.closest('[data-action]') : null);
  if(!el) return;
  const action = String(el.getAttribute('data-action')||'');
  if(action==='users:search'){ renderUsers(); }
});

document.addEventListener('change', (e)=>{
  const el = /** @type {Element|null} */(e.target instanceof Element ? e.target.closest('[data-action]') : null);
  if(!el) return;
  const action = String(el.getAttribute('data-action')||'');
  if(action==='users:order'){ renderUsers(); }
});

document.addEventListener('DOMContentLoaded', ()=>{
  initialLoad(); loadAutoEnd(); setupNavGuard();
  window.addEventListener('beforeunload', ()=>{ if(refreshTimer) clearInterval(refreshTimer); });
});
