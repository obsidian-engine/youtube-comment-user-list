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

  // Toast notification system
  if(!w.Toast){
    var container=null;
    function ensure(){
      if(container && document.body.contains(container)) return container;
      container=document.createElement('div');
      container.className='toast-container';
      container.setAttribute('aria-live', 'polite');
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
      item.textContent=message;
      c.appendChild(item);
      requestAnimationFrame(function(){ item.classList.add('show'); });
      setTimeout(function(){
        item.classList.remove('show');
        setTimeout(function(){ try{ c.removeChild(item); }catch(_){} }, 300);
      }, dur);
    }
    w.Toast = { show: show };
    w.toast = show;
  }

  // Modal confirmation system
  if(!w.confirmModal){
    function build(title, message, okText, cancelText){
      var overlay=document.createElement('div');
      overlay.className='modal-overlay';
      overlay.setAttribute('role', 'dialog');
      overlay.setAttribute('aria-modal', 'true');

      var modal=document.createElement('div');
      modal.className='modal';

      var h=document.createElement('h3');
      h.textContent=title||'確認';

      var p=document.createElement('p');
      p.textContent=message||'';

      var actions=document.createElement('div');
      actions.className='actions';

      var cancel=document.createElement('button');
      cancel.type='button';
      cancel.className='btn secondary';
      cancel.textContent=cancelText||'キャンセル';

      var ok=document.createElement('button');
      ok.type='button';
      ok.className='btn primary';
      ok.textContent=okText||'OK';

      actions.appendChild(cancel);
      actions.appendChild(ok);
      modal.appendChild(h);
      modal.appendChild(p);
      modal.appendChild(actions);
      overlay.appendChild(modal);

      return {overlay, modal, ok, cancel};
    }

    w.confirmModal=function(title, message, okText, cancelText){
      return new Promise(function(resolve){
        try{
          var ui=build(title,message,okText,cancelText);
          document.body.appendChild(ui.overlay);

          var cleanup=function(){
            try{ ui.overlay.remove(); }catch(_){}
            window.removeEventListener('keydown',onKey);
          };

          var done=function(v){ cleanup(); resolve(!!v); };

          var onKey=function(e){
            if(e.key==='Escape'){ e.preventDefault(); done(false);}
            if(e.key==='Enter'){ e.preventDefault(); done(true);}
          };

          window.addEventListener('keydown', onKey);
          ui.cancel.addEventListener('click', function(){ done(false); });
          ui.ok.addEventListener('click', function(){ done(true); });

          // Focus management
          setTimeout(function(){ try{ ui.ok.focus(); }catch(_){} }, 0);

        }catch(_){
          resolve(window.confirm(message||title||'確認しますか？'));
        }
      });
    };
  }
})(window);

// Sparkline: values[] を小さなcanvasに描画
window.drawSparkline = function(canvas, values, color){
  try{
    if(!canvas || !values || !values.length) return;
    var ctx = canvas.getContext('2d');
    if(!ctx) return;

    var w = canvas.width, h = canvas.height;
    ctx.clearRect(0,0,w,h);

    var min = Math.min.apply(null, values);
    var max = Math.max.apply(null, values);
    if(min===max){ min = 0; }

    var pad = 2;
    var n = values.length;
    var dx = (w - pad*2) / Math.max(1, n-1);

    ctx.lineWidth = 1.5;
    ctx.strokeStyle = color||'#2563eb';
    ctx.beginPath();

    for(var i=0;i<n;i++){
      var v = values[i];
      var y = h - pad - ((v - min) / (max - min || 1)) * (h - pad*2);
      var x = pad + dx * i;
      if(i===0) ctx.moveTo(x,y);
      else ctx.lineTo(x,y);
    }
    ctx.stroke();

    // 最後の点をハイライト
    var last = values[n-1];
    var y2 = h - pad - ((last - min) / (max - min || 1)) * (h - pad*2);
    var x2 = pad + dx*(n-1);
    ctx.fillStyle = ctx.strokeStyle;
    ctx.beginPath();
    ctx.arc(x2,y2,2,0,Math.PI*2);
    ctx.fill();
  }catch(_){ }
};

// Skeleton table (cols, rows)
window.buildSkeletonTable = function(cols, rows){
  cols = Math.max(1, parseInt(cols||3,10));
  rows = Math.max(1, parseInt(rows||6,10));

  var head = '<tr>'+Array(cols).fill('<th></th>').join('')+'</tr>';
  var body = '';
  for(var r=0;r<rows;r++){
    body += '<tr>'+Array(cols).fill('<td><span class="skeleton-line"></span></td>').join('')+'</tr>';
  }
  return '<div class="table-scroll"><table><thead>'+head+'</thead><tbody>'+body+'</tbody></table></div>';
};

// Monitoring: 共通の再開API呼び出し
(function(w){
  if(!w.Monitoring){ w.Monitoring = {}; }

  if(typeof w.Monitoring.resume !== 'function'){
    w.Monitoring.resume = async function(video){
      try{
        const body = video ? { video_input: String(video) } : {};
        const r = await fetch('/api/monitoring/resume', {
          method: 'POST',
          headers: {'Content-Type':'application/json'},
          body: JSON.stringify(body)
        });
        const d = await r.json();

        if(d.success){
          try{
            if(d.videoId) localStorage.setItem('currentVideoId', d.videoId);
          }catch(_){ }

          try{
            await fetch('/api/monitoring/auto-end', {
              method:'POST',
              headers:{'Content-Type':'application/json'},
              body: JSON.stringify({enabled:true})
            });
          }catch(_){ }

          if(w.toast) toast('監視を再開しました',{type:'success'});
          return true;
        } else {
          if(w.toast) toast('再開失敗: '+(d.error||'unknown'),{type:'error'});
          return false;
        }
      }catch(e){
        if(w.toast) toast('通信エラー: '+e.message,{type:'error'});
        return false;
      }
    };
  }
})(window);
