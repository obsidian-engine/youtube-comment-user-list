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
