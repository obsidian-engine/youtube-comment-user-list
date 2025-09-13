// @ts-check
import { getLogs, getLogsStats, clearLogs } from '../api.js';

const $ = (sel) => document.querySelector(sel);
const esc = (s) => (window.esc ? window.esc(s) : String(s));

let autoTimer = null; let logsSeries = [];

/**
 * @typedef {Object} LogQueryParams
 * @property {string=} level
 * @property {string=} component
 * @property {string=} video_id
 * @property {string=} limit
 * @property {number=} export
 */

/**
 * @returns {LogQueryParams}
 */
function buildParams(){
  /** @type {LogQueryParams} */ const p = {};
  const lv = /** @type {HTMLSelectElement|null} */($('#level')); if(lv && lv.value) p.level = lv.value;
  const comp = /** @type {HTMLInputElement|null} */($('#component')); if(comp && comp.value.trim()) p.component = comp.value.trim();
  const vid = /** @type {HTMLInputElement|null} */($('#video')); if(vid && vid.value.trim()) p.video_id = vid.value.trim();
  const limit = /** @type {HTMLSelectElement|null} */($('#limit')); p.limit = (limit && limit.value) || '300';
  return p;
}

async function loadStats(params){
  try{ const { ok, data } = await getLogsStats(params); if(ok && data && data.success){ const el = $('#stats'); if(el) el.textContent = '統計: 総数 '+data.total+' / エラー '+data.errors+' / 警告 '+data.warnings; } }catch(_){ }
}

let loading = false;
async function loadLogs(){
  if(loading) return; loading = true;
  const meta = $('#meta'); const table = $('#logTable');
  if(table && window.buildSkeletonTable){ table.innerHTML = window.buildSkeletonTable(6,8); }
  const params = buildParams();
  try{
    const { ok, data } = await getLogs(params);
    if(ok && data && data.success){
      const logs = Array.isArray(data.logs) ? data.logs : [];
      if(meta) meta.textContent = '件数: '+logs.length+'（更新: '+new Date().toLocaleTimeString()+'）';
      if(!logs.length){ if(table) table.innerHTML='<div class="muted">ログがありません</div>'; }
      else {
        if(table) table.innerHTML = '<div class="table-scroll"><table><thead><tr><th>時刻</th><th>Level</th><th>メッセージ</th><th>component</th><th>video_id</th><th>correlation</th></tr></thead><tbody>'
          + logs.map(l=>'<tr><td>'+esc(l.timestamp||'')+'</td><td>'+window.badge(l.level||'')+'</td><td>'+esc(l.message||'')+'</td><td>'+esc(l.component||'')+'</td><td>'+esc(l.video_id||'')+'</td><td>'+esc(l.correlation_id||'')+'</td></tr>').join('')
          + '</tbody></table></div>';
      }
      try{ logsSeries.push(logs.length); if(logsSeries.length>40) logsSeries.shift(); const c = /** @type {HTMLCanvasElement|null} */($('#sparkLogs')); if(c && window.drawSparkline) window.drawSparkline(c, logsSeries, '#2563eb'); }catch(_){ }
    } else { if(meta) meta.textContent = 'エラー: '+esc((data && data.error) || 'unknown'); }
  }catch(e){ if(meta) meta.textContent='通信エラー: '+esc(e && e.message || e); }
  loadStats(params);
  loading = false;
}

function stopAuto(){ if(autoTimer){ clearInterval(autoTimer); autoTimer=null; } }
function startAuto(){ stopAuto(); const auto = /** @type {HTMLInputElement|null} */($('#auto')); if(!auto || !auto.checked) return; const iv = /** @type {HTMLSelectElement|null} */($('#interval')); const ms = parseInt((iv && iv.value)||'10000',10); autoTimer = setInterval(loadLogs, ms); }
function resetAuto(){ stopAuto(); startAuto(); }
function toggleAuto(){ const a = /** @type {HTMLInputElement|null} */($('#auto')); if(a && a.checked) startAuto(); else stopAuto(); }

// イベント委譲: data-* で 1 箇所バインド
document.addEventListener('click', async (e)=>{
  const t = /** @type {Element|null} */(e.target instanceof Element ? e.target.closest('[data-action]') : null);
  if(!t) return;
  const action = String(t.getAttribute('data-action')||'');
  switch(action){
    case 'logs:refresh':
      e.preventDefault();
      loadLogs();
      break;
    case 'logs:clear':
      e.preventDefault();
      { const ok = await window.confirmModal('ログ全削除','すべてのログを削除します。よろしいですか？','削除','キャンセル'); if(!ok) return; const { data } = await clearLogs(); if(data && data.success){ window.toast('ログをクリアしました',{type:'success'}); loadLogs(); } else { window.toast('削除失敗: '+(data && data.error || 'unknown'),{type:'error'}); } }
      break;
    case 'logs:export':
      e.preventDefault();
      try{ const p=buildParams(); p.export=1; const sp=new URLSearchParams(p); const url='/api/logs?'+sp.toString(); const a=document.createElement('a'); a.href=url; a.download='logs_'+new Date().toISOString().split('T')[0]+'.json'; document.body.appendChild(a); a.click(); a.remove(); window.toast('エクスポート開始',{type:'success'});}catch(err){ window.toast('エクスポート失敗: '+(err && err.message || err),{type:'error'}); }
      break;
    default:
      break;
  }
});

document.addEventListener('change', (e)=>{
  const el = /** @type {Element|null} */(e.target instanceof Element ? e.target.closest('[data-action]') : null);
  if(!el) return;
  const action = String(el.getAttribute('data-action')||'');
  switch(action){
    case 'logs:auto-toggle':
      toggleAuto();
      break;
    case 'logs:auto-interval':
      resetAuto();
      break;
    default:
      break;
  }
});

window.addEventListener('load', ()=>{ loadLogs(); startAuto(); });
window.addEventListener('beforeunload', stopAuto);
