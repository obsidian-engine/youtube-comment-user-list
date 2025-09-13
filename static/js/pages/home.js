// @ts-check
// Home page module
import { fetchJSON, startMonitoring } from '../api.js';

const $ = (sel) => document.querySelector(sel);

function setBtnState(btn, disabled, html) {
  if (!btn) return;
  btn.disabled = disabled;
  if (html) {
    btn.dataset.origText = btn.dataset.origText || btn.innerHTML;
    btn.innerHTML = html;
  } else if (!disabled && btn.dataset.origText) {
    btn.innerHTML = btn.dataset.origText;
  }
}

function esc(s) { return (window.esc ? window.esc(s) : String(s)); }

// Detect active by SSE connect
function detect(videoId, timeoutMs) {
  return new Promise((resolve) => {
    if (!videoId) return resolve(false);
    let decided = false;
    const finish = (v) => { if (decided) return; decided = true; try { es && es.close(); } catch(_){} resolve(v); };
    let es = null;
    try {
      es = new EventSource('/api/sse/' + encodeURIComponent(videoId) + '/users');
      const t = setTimeout(() => finish(false), Math.max(500, timeoutMs||2500));
      es.addEventListener('connected', () => { clearTimeout(t); finish(true); });
      es.addEventListener('user_list', () => { clearTimeout(t); finish(true); });
      es.addEventListener('error', () => { clearTimeout(t); finish(false); });
    } catch { finish(false); }
  });
}

async function main() {
  const form = $('#monitoringForm');
  const startBtn = $('#startBtn');
  const msg = $('#message');
  const videoInput = /** @type {HTMLInputElement|null} */($('#videoInput'));
  const runBanner = $('#runBanner');
  const runInfo = $('#runInfo');
  const resumeBtn = $('#resumeBtn');
  const appbarStop = $('#appbarStop');

  let activeVideoId = null; let activeStatus = false; let submitting = false; let detectTimer = null; let detecting = false;
  const qs = new URLSearchParams(location.search);
  const DETECT_INTERVAL_MS = Math.max(5000, parseInt(qs.get('detectInterval')||'60000',10));
  const DETECT_TIMEOUT_MS  = Math.max(500,  parseInt(qs.get('detectTimeout') ||'2500',10));

  async function refreshExisting(){
    const stored = (localStorage.getItem('currentVideoId')||'');
    if(!stored){ activeVideoId=null; activeStatus=false; if(runBanner) runBanner.style.display='none'; return; }
    const active = await detect(stored, DETECT_TIMEOUT_MS);
    if(active){
      activeVideoId = stored; activeStatus = true;
      if(msg) msg.innerHTML = '<span style="color:#16a34a">現在監視中: '+esc(stored)+' <a href="/users" style="color:#2563eb">ユーザー一覧へ移動</a></span>';
      if(videoInput && !videoInput.value) videoInput.value = stored;
      if(runInfo) runInfo.textContent = 'videoId: '+stored+' / 状態: 起動中';
      if(runBanner) runBanner.style.display = 'block';
    } else {
      if(activeStatus && msg) msg.innerHTML = '<span style="color:#dc2626">前回の監視は停止しました</span>';
      activeVideoId=null; activeStatus=false; if(runBanner) runBanner.style.display='none';
    }
  }

  function startDetectLoop(){ if(detectTimer) clearInterval(detectTimer); detectTimer=setInterval(refreshExisting, DETECT_INTERVAL_MS); }

  if(form){
    form.addEventListener('submit', async (e)=>{
      e.preventDefault(); if(submitting) return;
      const fd = new FormData(form);
      const video = String(fd.get('videoInput')||'').trim();
      if(!video){ if(msg) msg.textContent='Video ID を入力してください'; return; }
      if(activeStatus && activeVideoId===video){ if(msg) msg.innerHTML='<span style="color:#16a34a">既に監視中です。遷移します…</span>'; setTimeout(()=>location.href='/users',400); return; }
      submitting = true; setBtnState(startBtn, true, '開始中…'); if(msg) msg.textContent='バリデーション中...';
      const maxStr = String(fd.get('maxUsers')||'').trim();
      let maxNum = undefined;
      if (maxStr) {
        const n = Number(maxStr);
        if (!Number.isFinite(n) || n < 1) { submitting=false; setBtnState(startBtn,false); if(msg) msg.textContent='最大ユーザー数は 1 以上の数値を入力してください'; return; }
        maxNum = n;
      }
      try{
        const { ok, data } = await startMonitoring(video, maxNum);
        if(ok && data && data.success){
          try{ localStorage.setItem('currentVideoId', video); }catch(_){ }
          if(msg) msg.innerHTML='<span style="color:#16a34a">開始しました。遷移します…</span>';
          if(runInfo) runInfo.textContent='videoId: '+video+' / 状態: 起動中';
          if(runBanner) runBanner.style.display='block';
          setTimeout(()=>location.href='/users',600);
        } else { submitting=false; setBtnState(startBtn,false); if(msg) msg.innerHTML='<span style="color:#dc2626">エラー: '+esc(data && data.error || 'unknown')+'</span>'; }
      }catch(err){ submitting=false; setBtnState(startBtn,false); if(msg) msg.innerHTML='<span style="color:#dc2626">通信エラー: '+esc(err.message||err)+'</span>'; }
    });
  }

  if(appbarStop){
    appbarStop.addEventListener('click', async ()=>{
      const ok = await window.confirmModal('監視停止','監視を停止しますか？','停止','キャンセル');
      if(!ok) return;
      try{
        const { data } = await fetchJSON('/api/monitoring/stop', { method:'DELETE' });
        if(data && data.success){ try{ localStorage.removeItem('currentVideoId'); }catch(_){ } window.toast('監視を停止しました',{type:'success'}); refreshExisting(); }
        else { window.toast('停止失敗: '+(data && data.error || 'unknown'),{type:'error'}); }
      }catch(e){ window.toast('通信エラー: '+(e && e.message || e),{type:'error'}); }
    });
  }

  // イベント委譲: data-action で 1 箇所バインド
  document.addEventListener('click', (e)=>{
    const t = /** @type {Element|null} */(e.target instanceof Element ? e.target.closest('[data-action]') : null);
    if(!t) return;
    const action = String(t.getAttribute('data-action')||'');
    switch(action){
      case 'home:status-refresh':
        e.preventDefault();
        refreshExisting();
        break;
      default:
        break;
    }
  });

  document.addEventListener('DOMContentLoaded', async ()=>{
    await refreshExisting();
    try{
      const stored=localStorage.getItem('currentVideoId')||'';
      if(resumeBtn){
        resumeBtn.style.display = (!activeStatus && stored) ? 'inline-flex' : 'none';
        resumeBtn.addEventListener('click', async ()=>{ try{ await window.Monitoring.resume(stored); }catch(_){ } });
      }
    }catch(_){ }
    startDetectLoop();
  });

  window.addEventListener('beforeunload', ()=>{ if(detectTimer) clearInterval(detectTimer); });
}

main().catch(()=>{});
