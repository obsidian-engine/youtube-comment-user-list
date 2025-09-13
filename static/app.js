(function(w){
  'use strict';
  // Utility: query builder
  if(typeof w.q !== 'function'){
    w.q = function(params){
      try {
        var sp = new URLSearchParams(params||{});
        var qs = sp.toString();
        return qs ? ('?' + qs) : '';
      } catch(e) { return ''; }
    };
  }
  // Utility: HTML escape
  if(typeof w.esc !== 'function'){
    w.esc = function(s){
      return (s==null?'' : String(s)).replace(/[&<>"']/g, function(c){
        return ({'&':'&amp;','<':'&lt;','>':'&gt;','"':'&quot;','\'':'&#039;'}[c]);
      });
    };
  }
  if(typeof w.escapeHtml !== 'function') w.escapeHtml = w.esc;
  // Utility: level badge
  if(typeof w.badge !== 'function'){
    w.badge = function(level){
      var l = (level||'').toUpperCase();
      var cls = l==='ERROR' ? 'lvl-error' : (l==='WARNING' ? 'lvl-warn' : 'lvl-info');
      return '<span class="badge '+cls+'">'+l+'</span>';
    };
  }
  if(!w.Toast){
    var container=null;
    function ensure(){
      if(container && document.body.contains(container)) return container;
      container=document.createElement('div');
      container.className='toast-container';
      document.body.appendChild(container);
      return container;
    }
    function show(message, opts){
      opts=opts||{};
      var dur = typeof opts.duration==='number'? opts.duration : 1800;
      var type = opts.type||'info'; // info | success | error | warn
      var c = ensure();
      var item=document.createElement('div');
      item.className='toast toast-'+type;
      item.setAttribute('role','status');
      item.setAttribute('aria-live','polite');
      item.textContent=message;
      c.appendChild(item);
      requestAnimationFrame(()=>{ item.classList.add('show'); });
      setTimeout(()=>{
        item.classList.remove('show');
        item.classList.add('hide');
        setTimeout(()=>{ try{ item.remove(); }catch(_){} }, 280);
      }, dur);
    }
    w.Toast = { show };
    if(!w.toast){ w.toast = function(m,o){ show(m,o); }; }
  }
})(window);

// Simple confirm modal (Promise-based)
(function(w){
  if(typeof w.confirmModal === 'function') return;
  function build(title, message, okText, cancelText){
    var overlay=document.createElement('div'); overlay.className='modal-overlay';
    var modal=document.createElement('div'); modal.className='modal';
    var h=document.createElement('h3'); h.textContent=title||'確認';
    var p=document.createElement('p'); p.textContent=message||'';
    var actions=document.createElement('div'); actions.className='actions';
    var cancel=document.createElement('button'); cancel.type='button'; cancel.className='btn secondary'; cancel.textContent=cancelText||'キャンセル';
    var ok=document.createElement('button'); ok.type='button'; ok.className='btn'; ok.textContent=okText||'OK';
    actions.appendChild(cancel); actions.appendChild(ok);
    modal.appendChild(h); modal.appendChild(p); modal.appendChild(actions);
    overlay.appendChild(modal);
    return {overlay, modal, ok, cancel};
  }
  w.confirmModal=function(title, message, okText, cancelText){
    return new Promise(function(resolve){
      try{
        var ui=build(title,message,okText,cancelText);
        document.body.appendChild(ui.overlay);
        var cleanup=function(){ try{ ui.overlay.remove(); }catch(_){} window.removeEventListener('keydown',onKey); };
        var done=function(v){ cleanup(); resolve(!!v); };
        var onKey=function(e){ if(e.key==='Escape'){ e.preventDefault(); done(false);} if(e.key==='Enter'){ e.preventDefault(); done(true);} };
        window.addEventListener('keydown', onKey);
        ui.cancel.addEventListener('click', function(){ done(false); });
        ui.ok.addEventListener('click', function(){ done(true); });
        setTimeout(function(){ try{ ui.ok.focus(); }catch(_){} }, 0);
      }catch(_){ resolve(window.confirm(message||title||'確認しますか？')); }
    });
  };
})(window);

// Theme toggle & persistence
(function(w){
  try{
    var KEY='ui_theme_pref_v1'; // 'light' | 'dark' | 'system'
    function applyTheme(pref){
      if(pref==='light' || pref==='dark'){
        document.documentElement.setAttribute('data-theme', pref);
      } else {
        document.documentElement.removeAttribute('data-theme'); // system (media query任せ)
      }
      updateToggleLabel();
    }
    function currentPref(){
      try { return localStorage.getItem(KEY)||'system'; } catch(_) { return 'system'; }
    }
    function savePref(p){ try { if(p==='system') localStorage.removeItem(KEY); else localStorage.setItem(KEY,p); }catch(_){} }
    function toggle(){
      var p=currentPref();
      var next;
      if(p==='system'){
        // システム設定を反転する方向へ
        var sysDark = w.matchMedia && w.matchMedia('(prefers-color-scheme: dark)').matches;
        next = sysDark ? 'light' : 'dark';
      } else if(p==='light') next='dark'; else if(p==='dark') next='light'; else next='dark';
      savePref(next); applyTheme(next); if(w.toast) toast('テーマ: '+(next==='system'?'システム':next),{type:'success'});
    }
    function reset(){ savePref('system'); applyTheme('system'); if(w.toast) toast('テーマをシステム設定に戻しました',{type:'info'}); }
    function updateToggleLabel(){
      var btn=document.getElementById('themeToggle'); if(!btn) return;
      var pref=currentPref();
      var now = (pref==='system') ? (w.matchMedia && w.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark':'light') : pref;
      // 次に切り替わるターゲットを表示
      var next = now==='dark' ? 'ライト' : 'ダーク';
      btn.innerHTML='<span class="material-symbols-outlined" style="font-size:18px">'+(now==='dark'?'light_mode':'dark_mode')+'</span> '+next+'モード';
    }
    function init(){
      applyTheme(currentPref());
      var tgl=document.getElementById('themeToggle');
      var rst=document.getElementById('themeReset');
      if(tgl){ tgl.addEventListener('click', toggle); }
      if(rst){ rst.addEventListener('click', reset); }
      if(w.matchMedia){
        try { w.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', function(){ if(currentPref()==='system') applyTheme('system'); }); }catch(_){ /* Safari <14 fallback */ }
      }
    }
    if(document.readyState==='loading') document.addEventListener('DOMContentLoaded', init); else init();
  }catch(_){ /* no-op */ }
})(window);

// ライトテーマ固定（トグルは通知のみ）
(function(w){
  try{
    var FORCE=true;
    if(!FORCE) return;
    document.documentElement.setAttribute('data-theme','light');
    try{ localStorage.setItem('ui_theme_pref_v1','light'); }catch(_){ }
    function lock(e){ e && e.preventDefault && e.preventDefault(); if(w.toast) toast('ライトテーマ固定です'); }
    if(document.readyState==='loading') document.addEventListener('DOMContentLoaded', function(){
      var t=document.getElementById('themeToggle'); var r=document.getElementById('themeReset');
      if(t){ t.onclick=lock; }
      if(r){ r.onclick=lock; }
    }); else {
      var t=document.getElementById('themeToggle'); var r=document.getElementById('themeReset');
      if(t){ t.onclick=lock; }
      if(r){ r.onclick=lock; }
    }
  }catch(_){ }
})(window);

// Sparkline: values[] を小さなcanvasに描画
window.drawSparkline = function(canvas, values, color){
  try{
    if(!canvas || !values || !values.length) return;
    var ctx = canvas.getContext('2d'); if(!ctx) return;
    var w = canvas.width, h = canvas.height; ctx.clearRect(0,0,w,h);
    var min = Math.min.apply(null, values); var max = Math.max.apply(null, values);
    if(min===max){ min = 0; }
    var pad = 2; var n = values.length; var dx = (w - pad*2) / Math.max(1, n-1);
    ctx.lineWidth = 1.5; ctx.strokeStyle = color||'#2563eb'; ctx.beginPath();
    for(var i=0;i<n;i++){
      var v = values[i]; var y = h - pad - ((v - min) / (max - min || 1)) * (h - pad*2);
      var x = pad + dx * i; if(i===0) ctx.moveTo(x,y); else ctx.lineTo(x,y);
    }
    ctx.stroke();
    var last = values[n-1]; var y2 = h - pad - ((last - min) / (max - min || 1)) * (h - pad*2); var x2 = pad + dx*(n-1);
    ctx.fillStyle = ctx.strokeStyle; ctx.beginPath(); ctx.arc(x2,y2,2,0,Math.PI*2); ctx.fill();
  }catch(_){ }
};

// Skeleton table (cols, rows)
window.buildSkeletonTable = function(cols, rows){
  cols = Math.max(1, parseInt(cols||3,10)); rows = Math.max(1, parseInt(rows||6,10));
  var head = '<tr>'+Array(cols).fill('<th></th>').join('')+'</tr>';
  var body = '';
  for(var r=0;r<rows;r++){
    body += '<tr>'+Array(cols).fill('<td><span class="skeleton-line"></span></td>').join('')+'</tr>';
  }
  return '<div class="table-scroll"><table><thead>'+head+'</thead><tbody>'+body+'</tbody></table></div>';
};
