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
  if(typeof w.escapeHtml !== 'function'){
    w.escapeHtml = w.esc;
  }
  // Utility: level badge
  if(typeof w.badge !== 'function'){
    w.badge = function(level){
      var l = (level||'').toUpperCase();
      var cls = l==='ERROR' ? 'lvl-error' : (l==='WARNING' ? 'lvl-warn' : 'lvl-info');
      return '<span class="badge '+cls+'">'+l+'</span>';
    };
  }
})(window);

// UI preferences (theme/color/radius/density)
(function(w){
  'use strict';
  var LS_KEY = 'ui_prefs_v1';
  function def(){
    return { theme: 'dark', primary: 'blue', radius: 'md', density: 'comfortable' };
  }
  function load(){
    try { var v = JSON.parse(localStorage.getItem(LS_KEY)||'null'); if(!v) return def(); return Object.assign(def(), v); }
    catch(_) { return def(); }
  }
  function save(p){ try { localStorage.setItem(LS_KEY, JSON.stringify(p)); } catch(_){} }
  function apply(p){
    var root = document.documentElement;
    root.setAttribute('data-theme', p.theme);
    root.setAttribute('data-primary', p.primary);
    root.setAttribute('data-radius', p.radius);
    root.setAttribute('data-density', p.density);
  }
  function init(){
    var prefs = load();
    apply(prefs);
    var tgl = document.getElementById('uiSettingsToggle');
    var pnl = document.getElementById('uiSettingsPanel');
    if(!tgl || !pnl) return;
    var themeSel = document.getElementById('uiTheme');
    var primarySel = document.getElementById('uiPrimary');
    var radiusSel = document.getElementById('uiRadius');
    var densitySel = document.getElementById('uiDensity');
    if(themeSel) themeSel.value = prefs.theme;
    if(primarySel) primarySel.value = prefs.primary;
    if(radiusSel) radiusSel.value = prefs.radius;
    if(densitySel) densitySel.value = prefs.density;

    function update(){
      var p = {
        theme: themeSel ? themeSel.value : prefs.theme,
        primary: primarySel ? primarySel.value : prefs.primary,
        radius: radiusSel ? radiusSel.value : prefs.radius,
        density: densitySel ? densitySel.value : prefs.density,
      };
      apply(p); save(p);
    }
    function open(){ pnl.classList.add('open'); }
    function close(){ pnl.classList.remove('open'); }
    function toggle(){ pnl.classList.toggle('open'); }
    tgl.addEventListener('click', function(e){ e.preventDefault(); toggle(); });
    [themeSel, primarySel, radiusSel, densitySel].forEach(function(el){ if(!el) return; el.addEventListener('change', update); });
    document.addEventListener('click', function(e){ if(!pnl.classList.contains('open')) return; if(pnl.contains(e.target) || e.target===tgl) return; close(); });
    document.addEventListener('keydown', function(e){ if(e.key==='Escape') close(); });
  }
  w.UI = { load: load, save: save, apply: apply, init: init };
  if(document.readyState === 'loading') document.addEventListener('DOMContentLoaded', function(){ try{ init(); }catch(_){}});
  else { try{ init(); }catch(_){} }
})(window);
