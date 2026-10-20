(function () {
  'use strict';
  var root = document.querySelector('.fn-market'), config = window.fieldnotesMarket;
  if (!root || !config || root.dataset.initialized) return;
  root.dataset.initialized = '1';
  var form = root.querySelector('.fn-market-search'), input = root.querySelector('#fn-market-code');
  var chart = root.querySelector('.fn-market-chart'), svg = root.querySelector('.fn-market-svg'), empty = root.querySelector('.fn-market-chart-empty');
  var status = root.querySelector('.fn-market-status'), retry = root.querySelector('.fn-market-retry'), tooltip = root.querySelector('.fn-market-tooltip');
  var title = root.querySelector('.fn-market-stock-title'), subtitle = root.querySelector('.fn-market-stock-subtitle'), source = root.querySelector('.fn-market-source');
  var adjustSelect = root.querySelector('#fn-market-adjust'), table = root.querySelector('.fn-market-data-table');
  var symbol = root.dataset.symbol || '', period = 'minute', adjust = 'qfq', data = null, count = 80, offset = 0, cursor = -1;
  var quoteSequence = 0, noteSequence = 0, quoteController = null, noteController = null, plot = null;
  var live = root.querySelector('.fn-market-live'), intervalSelect = root.querySelector('#fn-market-interval'), autoState = root.querySelector('.fn-market-auto-state');
  var liveController = null, liveSequence = 0, liveBusy = false, liveFailures = 0, lastRefresh = 0, refreshSeconds = 5, ignoreSavedSymbol = '';
  var storageKey = 'fieldnotes:markets:v1:' + config.ajaxUrl, watchlist = [];
  var NS = 'http://www.w3.org/2000/svg', colors = {up:'#b66a62',down:'#7f929a',ma5:'#9b8874',ma10:'#85879b',ma20:'#789494'};
  function node(tag, attrs, text) { var el = document.createElementNS(NS,tag); Object.keys(attrs || {}).forEach(function (key) { el.setAttribute(key,attrs[key]); }); if (text != null) el.textContent = text; return el; }
  function html(tag, cls, text) { var el = document.createElement(tag); if (cls) el.className=cls; if (text!=null) el.textContent=text; return el; }
  function normalize(value) {
    var v = value.trim().toUpperCase(), match, market, code;
    if ((match=/^(SH|SZ|BJ)[:.]?(\d{6})$/.exec(v))) { market=match[1]; code=match[2]; }
    else if ((match=/^(\d{6})\.(SH|SZ|BJ)$/.exec(v))) { market=match[2]; code=match[1]; }
    else if (/^\d{6}$/.test(v)) { code=v; market=/^(4|8|92)/.test(code)?'BJ':/^[69]/.test(code)?'SH':'SZ'; }
    else return '';
    if (market==='SH'&&!/^[69]/.test(code)||market==='SZ'&&!/^[023]/.test(code)||market==='BJ'&&!/^(4|8|92)/.test(code)) return '';
    return market+code;
  }
  function paramsFromURL() {
    var url = new URL(window.location.href);
    return {symbol:normalize(url.searchParams.get('stock')||''),period:['minute','day','week','month'].includes(url.searchParams.get('period'))?url.searchParams.get('period'):'minute',adjust:url.searchParams.get('adjust')==='none'?'none':'qfq',page:Math.max(1,parseInt(url.searchParams.get('notes_page')||'1',10)||1)};
  }
  function updateURL(page, replace) {
    var url=new URL(window.location.href); if(symbol)url.searchParams.set('stock',symbol);else url.searchParams.delete('stock');
    url.searchParams.set('period',period);url.searchParams.set('adjust',adjust);
    if(page>1)url.searchParams.set('notes_page',page);else url.searchParams.delete('notes_page');
    window.history[replace?'replaceState':'pushState']({},'',url.pathname+url.search+url.hash);
  }
  function request(action, fields, controller) {
    var url=new URL(config.ajaxUrl,window.location.href);url.searchParams.set('action',action);
    Object.keys(fields).forEach(function(key){url.searchParams.set(key,fields[key]);});
    return fetch(url.href,{credentials:'same-origin',cache:'no-store',signal:controller?controller.signal:undefined}).then(function(response){
      return response.json().catch(function(){throw new Error('站点暂时没有返回有效数据，请稍后重试。');});
    }).then(function(result){if(!result||!result.success)throw new Error(result&&result.data&&result.data.message||'读取未完成，请重试。');return result.data;});
  }
  function message(text,error) { status.textContent=text||'';status.classList.toggle('is-error',!!error); }
  function updateControls() {
    root.classList.toggle('has-stock',!!symbol);
    root.querySelectorAll('[data-period]').forEach(function(button){button.setAttribute('aria-pressed',String(button.dataset.period===period));button.disabled=!symbol;});
    adjustSelect.value=adjust;adjustSelect.disabled=!symbol;adjustSelect.hidden=period==='minute';
    root.querySelector('.fn-market-ma-legend').hidden=period==='minute';
    root.querySelectorAll('.fn-market-zoom-out,.fn-market-zoom-in,.fn-market-zoom-reset').forEach(function(button){button.disabled=!data;button.hidden=period==='minute';});
    root.querySelector('.fn-market-refresh').disabled=!symbol;
  }
  function resetChart(text) {
    data=null;plot=null;svg.replaceChildren();tooltip.hidden=true;empty.hidden=false;empty.querySelector('p').textContent=text||'输入名称或代码，展开一段价格记录。';table.hidden=true;
    svg.setAttribute('aria-label','等待查询股票历史行情');updateControls();
  }
  function validData(result, expected) {
    if(period==='minute') return result&&result.symbol===expected&&result.period==='minute'&&/^\d{4}-\d{2}-\d{2}$/.test(result.date)&&Array.isArray(result.points)&&result.points.length>0&&result.points.length<=300&&(result.previous_close===null||typeof result.previous_close==='number'&&Number.isFinite(result.previous_close)&&result.previous_close>0)&&result.points.every(function(point,index,points){
      return /^\d{2}:\d{2}$/.test(point.time)&&typeof point.price==='number'&&Number.isFinite(point.price)&&point.price>0&&typeof point.volume==='number'&&Number.isFinite(point.volume)&&point.volume>=0&&(index===0||point.time>points[index-1].time);
    });
    return result&&result.symbol===expected&&Array.isArray(result.bars)&&result.bars.length>0&&result.bars.every(function(bar){
      return /^\d{4}-\d{2}-\d{2}$/.test(bar.date)&&['open','close','high','low','volume'].every(function(key){return typeof bar[key]==='number'&&Number.isFinite(bar[key]);})&&bar.low>0&&bar.volume>=0&&bar.high>=Math.max(bar.open,bar.close)&&bar.low<=Math.min(bar.open,bar.close);
    });
  }
  function format(value) { return Number(value).toLocaleString('zh-CN',{minimumFractionDigits:2,maximumFractionDigits:2}); }
  function average(index,length) { if(index<length-1)return null;var sum=0;for(var i=index-length+1;i<=index;i++)sum+=data.bars[i].close;return sum/length; }
  function draw() {
    if(!data)return;
    if(data.period==='minute'){drawMinute();return;}
    var width=Math.max(280,chart.clientWidth||1000),height=chart.clientHeight||480;
    var bars=data.bars,end=Math.max(1,bars.length-offset),start=Math.max(0,end-count),visible=bars.slice(start,end);
    var left=12,right=62,top=30,bottom=height-120,volumeTop=height-93,volumeBottom=height-28;
    var plotWidth=width-left-right,step=plotWidth/visible.length;
    var max=Math.max.apply(null,visible.map(function(bar){return bar.high;})),min=Math.min.apply(null,visible.map(function(bar){return bar.low;}));
    var pad=(max-min)*.12||max*.04;max+=pad;min=Math.max(0,min-pad);
    var y=function(value){return bottom-(value-min)/(max-min)*(bottom-top);};
    var x=function(index){return left+step*(index+.5);};
    var volumeMax=Math.max.apply(null,visible.map(function(bar){return bar.volume;}))||1;
    svg.replaceChildren();svg.setAttribute('viewBox','0 0 '+width+' '+height);
    svg.setAttribute('aria-label',data.name+' '+symbol+' 历史'+({day:'日',week:'周',month:'月'}[period])+' K 线，'+visible[0].date+' 至 '+visible[visible.length-1].date);
    var defs=node('defs'),clip=node('clipPath',{id:'fn-market-clip'});clip.appendChild(node('rect',{x:left,y:top,width:plotWidth,height:bottom-top}));defs.appendChild(clip);svg.appendChild(defs);
    for(var grid=0;grid<5;grid++){
      var gy=top+(bottom-top)*grid/4,value=max-(max-min)*grid/4;
      svg.appendChild(node('line',{x1:left,y1:gy,x2:width-right,y2:gy,stroke:'#e8e8ed','stroke-width':.7}));
      svg.appendChild(node('text',{x:width-right+9,y:gy+4,fill:'#929298','font-size':10,'font-family':'-apple-system, sans-serif'},format(value)));
    }
    visible.forEach(function(bar,index){
      var color=bar.close>=bar.open?colors.up:colors.down,bw=Math.max(1,Math.min(10,step*.65));
      svg.appendChild(node('line',{x1:x(index),x2:x(index),y1:y(bar.high),y2:y(bar.low),stroke:color,'stroke-width':1}));
      svg.appendChild(node('rect',{x:x(index)-bw/2,y:Math.min(y(bar.open),y(bar.close)),width:bw,height:Math.max(1,Math.abs(y(bar.open)-y(bar.close))),fill:bar.close>=bar.open?'#fafafb':color,stroke:color,'stroke-width':1}));
      var vh=bar.volume/volumeMax*(volumeBottom-volumeTop);
      svg.appendChild(node('rect',{x:x(index)-bw/2,y:volumeBottom-vh,width:bw,height:Math.max(.5,vh),fill:color,opacity:.6}));
    });
    [5,10,20].forEach(function(length){
      var points=[];visible.forEach(function(bar,index){var value=average(start+index,length);if(value!=null)points.push(x(index)+','+y(value));});
      if(points.length>1)svg.appendChild(node('polyline',{points:points.join(' '),fill:'none',stroke:colors['ma'+length],'stroke-width':1.1,'clip-path':'url(#fn-market-clip)'}));
    });
    svg.appendChild(node('text',{x:left,y:volumeTop-8,fill:'#a1a1a6','font-size':9,'font-family':'-apple-system, sans-serif'},'成交量（手）'));
    [0,Math.floor(visible.length/2),visible.length-1].forEach(function(index,order){
      svg.appendChild(node('text',{x:order===0?left:order===2?width-right:x(index),y:height-6,fill:'#a1a1a6','font-size':9,'text-anchor':order===0?'start':order===2?'end':'middle','font-family':'-apple-system, sans-serif'},visible[index].date));
    });
    var line=node('line',{x1:0,x2:0,y1:top,y2:volumeBottom,stroke:'#a1a1a6','stroke-width':.8,'stroke-dasharray':'3 3',visibility:'hidden'});svg.appendChild(line);
    plot={bars:visible,start:start,left:left,right:width-right,width:width,step:step,line:line,x:x};
    empty.hidden=true;cursor=Math.min(cursor,visible.length-1);if(cursor>=0)showBar(cursor);
  }

  function minutePosition(time) {
    var parts=time.split(':'),clock=Number(parts[0])*60+Number(parts[1]);
    return Math.max(0,Math.min(240,clock<=690?clock-570:120+clock-780));
  }
  function drawMinute() {
    var points=data.points,width=Math.max(280,chart.clientWidth||1000),height=chart.clientHeight||480;
    var left=12,right=65,top=28,bottom=height-120,volumeTop=height-90,volumeBottom=height-29,pw=width-left-right;
    var prices=points.map(function(p){return p.price;}),min=Math.min.apply(null,prices),max=Math.max.apply(null,prices),previous=data.previous_close;
    if(previous){var delta=Math.max(Math.abs(max-previous),Math.abs(min-previous),previous*.003)*1.15;min=Math.max(0,previous-delta);max=previous+delta;}
    else {var pad=(max-min)*.12||max*.005;max+=pad;min-=pad;}
    var y=function(v){return bottom-(v-min)/(max-min)*(bottom-top);},x=function(i){return left+minutePosition(points[i].time)/240*pw;};
    var volumeMax=Math.max.apply(null,points.map(function(p){return p.volume;}))||1;
    svg.replaceChildren();svg.setAttribute('viewBox','0 0 '+width+' '+height);svg.setAttribute('aria-label',data.name+' '+symbol+' '+data.date+' 分时图，已取得 '+points[0].time+' 至 '+points[points.length-1].time+' 的数据');
    for(var grid=0;grid<5;grid++){
      var price=max-(max-min)*grid/4,gy=y(price);svg.appendChild(node('line',{x1:left,x2:width-right,y1:gy,y2:gy,stroke:'#e8e8ed','stroke-width':.7}));
      svg.appendChild(node('text',{x:width-right+8,y:gy+4,fill:'#929298','font-size':10},format(price)));
    }
    if(previous){svg.appendChild(node('line',{x1:left,x2:width-right,y1:y(previous),y2:y(previous),stroke:'#b5b5bc','stroke-width':1,'stroke-dasharray':'4 4'}));svg.appendChild(node('text',{x:left+5,y:y(previous)-6,fill:'#a1a1a6','font-size':9},'昨收'));}
    var coordinates=points.map(function(p,i){return x(i)+','+y(p.price);}).join(' ');
    if(points.length>1){svg.appendChild(node('polygon',{points:x(0)+','+bottom+' '+coordinates+' '+x(points.length-1)+','+bottom,fill:'#8e9ba50d'}));svg.appendChild(node('polyline',{points:coordinates,fill:'none',stroke:'#778b99','stroke-width':1.6,'stroke-linejoin':'round'}));}
    svg.appendChild(node('circle',{cx:x(points.length-1),cy:y(points[points.length-1].price),r:2.8,fill:'#778b99'}));
    points.forEach(function(p,i){var vh=p.volume/volumeMax*(volumeBottom-volumeTop),bw=Math.max(.8,Math.min(5,pw/241*.7));svg.appendChild(node('rect',{x:x(i)-bw/2,y:volumeBottom-vh,width:bw,height:Math.max(.5,vh),fill:'#9aa8b1',opacity:.7}));});
    svg.appendChild(node('text',{x:left,y:volumeTop-8,fill:'#a1a1a6','font-size':9},'成交量（手）'));
    [[0,'09:30'],[120,'11:30 / 13:00'],[240,'15:00']].forEach(function(mark,i){svg.appendChild(node('text',{x:left+mark[0]/240*pw,y:height-6,fill:'#a1a1a6','font-size':9,'text-anchor':i===0?'start':i===2?'end':'middle'},mark[1]));});
    var line=node('line',{x1:0,x2:0,y1:top,y2:volumeBottom,stroke:'#a1a1a6','stroke-width':.8,'stroke-dasharray':'3 3',visibility:'hidden'});svg.appendChild(line);
    plot={minute:true,bars:points,start:0,left:left,right:width-right,width:width,step:pw/241,line:line,x:x};empty.hidden=true;cursor=Math.min(cursor,points.length-1);if(cursor>=0)showBar(cursor);
  }

  function showBar(index) {
    if(!plot)return;cursor=Math.max(0,Math.min(plot.bars.length-1,index));var bar=plot.bars[cursor];
    plot.line.setAttribute('x1',plot.x(cursor));plot.line.setAttribute('x2',plot.x(cursor));plot.line.setAttribute('visibility','visible');
    tooltip.textContent=period==='minute'?data.date+' '+bar.time+'\n价格 '+format(bar.price)+(data.previous_close?'   较昨收 '+((bar.price/data.previous_close-1)*100).toFixed(2)+'%':'')+'\n本分钟成交量 '+Number(bar.volume).toLocaleString('zh-CN')+' 手':bar.date+'  '+({day:'日',week:'周',month:'月'}[period])+' K\n开 '+format(bar.open)+'   收 '+format(bar.close)+'   高 '+format(bar.high)+'   低 '+format(bar.low)+'\n成交量 '+Number(bar.volume).toLocaleString('zh-CN')+' 手';tooltip.hidden=false;
  }
  function hideBar(){tooltip.hidden=true;cursor=-1;if(plot)plot.line.setAttribute('visibility','hidden');}
  function tableData() {
    var body=table.querySelector('tbody'),head=table.querySelector('thead tr'),minute=period==='minute';body.replaceChildren();head.replaceChildren();
    (minute?['时间','价格','较昨收','成交量（手）']:['日期','开盘','收盘','最高','最低','成交量（手）']).forEach(function(label){head.appendChild(html('th','',label));});
    (minute?data.points:data.bars).slice(-20).reverse().forEach(function(bar){
      var row=html('tr');(minute?['time','price','change','volume']:['date','open','close','high','low','volume']).forEach(function(key){var value=key==='date'||key==='time'?bar[key]:key==='change'?(data.previous_close?((bar.price/data.previous_close-1)*100).toFixed(2)+'%':'—'):key==='volume'?Number(bar[key]).toLocaleString('zh-CN'):format(bar[key]);row.appendChild(html('td','',value));});body.appendChild(row);
    });table.hidden=false;
  }
  function loadQuotes(quiet) {
    if(!symbol)return;
    quoteSequence++;var sequence=quoteSequence,target=symbol;
    if(quoteController)quoteController.abort();quoteController=typeof AbortController!=='undefined'?new AbortController():null;
    var keep=quiet===true&&data&&data.period===period;
    if(!keep){resetChart('正在读取行情…');message('正在读取 '+target+' 的行情。');title.textContent=target;subtitle.textContent=period==='minute'?'当日分时':'历史 K 线';source.textContent='';}retry.hidden=true;chart.setAttribute('aria-busy','true');
    request('fieldnotes_market',{symbol:target,period:period,adjust:adjust},quoteController).then(function(result){
      if(sequence!==quoteSequence)return;if(!validData(result,target))throw new Error('行情代码或价格字段不匹配，未显示本次数据。');
      data=result;if(!keep){offset=0;cursor=-1;}title.textContent=result.name+' · '+target;rememberStock(target,result.name);
      var bars=period==='minute'?result.points:result.bars,last=bars[bars.length-1],prior=bars[bars.length-2],change=period==='minute'?(result.previous_close?(last.price/result.previous_close-1)*100:null):(prior?(last.close/prior.close-1)*100:null);
      subtitle.textContent=(period==='minute'?'分时末价 '+format(last.price):'K 线末价 '+format(last.close))+(change==null?'':(period==='minute'?'  较昨收 ':'  较前根 ')+(change>=0?'+':'')+change.toFixed(2)+'%')+' · '+(period==='minute'?result.date+' '+last.time:last.date);
      var fetched=new Date(result.fetched_at*1000).toLocaleString('zh-CN',{timeZone:'Asia/Shanghai',hour12:false});
      source.textContent=result.source+' · 数据截至 '+(period==='minute'?result.date+' '+last.time:last.date)+' · 获取于 '+fetched+' · '+(period==='minute'?'分时数据可能延迟':adjust==='qfq'?'前复权':'不复权')+(result.stale?' · 上次缓存':'');
      message(result.stale?'行情服务暂不可达，当前展示上次缓存，请留意数据时间。':period==='minute'?'分时价格与每分钟成交量 · 轻触或使用方向键查看 · 按所选间隔尝试更新，数据可能延迟':'红涨蓝跌 · 移动指针或轻触查看价格 · Ctrl + 滚轮缩放',!!result.stale);tableData();draw();updateControls();
    }).catch(function(error){
      if(sequence!==quoteSequence||error.name==='AbortError')return;if(keep){message('更新暂未完成，当前保留上次读取的图表，请留意数据时间。',true);}else{resetChart('暂时没有取得这只股票的行情。');message(error.message||'行情读取失败，请重试。',true);source.textContent='本次未显示价格数据。';}retry.hidden=false;
    }).finally(function(){if(sequence===quoteSequence)chart.removeAttribute('aria-busy');});
  }
  function renderNotes(result) {
    var list=root.querySelector('.fn-market-note-list');list.replaceChildren();
    (result.posts||[]).forEach(function(note){
      var article=html('article','fn-market-note'),link=html('a','fn-market-note-link'),time=html('time','',note.date),body=html('div'),arrow=html('span','','↗');
      var url=new URL(note.url,window.location.href);if(!/^https?:$/.test(url.protocol))return;link.href=url.href;time.dateTime=note.iso_date;arrow.setAttribute('aria-hidden','true');
      body.append(html('h3','',note.title),html('p','',note.excerpt));link.append(time,body,arrow);article.appendChild(link);list.appendChild(article);
    });
    if(!result.posts||!result.posts.length){var emptyNote=html('div','fn-market-empty');emptyNote.append(html('span','','＋'),html('h3','','下一次判断，从这里留下。'),html('p','','写下看见的变化，保留当时的依据。回看时，才有迹可循。'));list.appendChild(emptyNote);}
    root.querySelector('[data-stat=all]').textContent=result.all_total;root.querySelector('[data-stat=matched]').textContent=result.total;root.querySelector('[data-stat=latest]').textContent=result.latest||'待续';
    root.querySelector('.fn-market-match-label').textContent=symbol?'这只股票的记录':'已发表记录';
    root.querySelector('.fn-market-notes-caption').textContent=symbol?symbol+' 的记录':'市场的变化，和自己的思考。';
    var pagination=root.querySelector('.fn-market-note-pagination');pagination.replaceChildren();
    for(var page=Math.max(1,result.page-2);page<=Math.min(result.pages,result.page+2);page++){
      var link=html('a','',page),url=new URL(window.location.href);url.searchParams.set('notes_page',page);link.href=url.href;link.dataset.notesPage=page;if(page===result.page)link.setAttribute('aria-current','page');pagination.appendChild(link);
    }
  }
  function loadNotes(page) {
    noteSequence++;var sequence=noteSequence;
    if(noteController)noteController.abort();noteController=typeof AbortController!=='undefined'?new AbortController():null;
    var list=root.querySelector('.fn-market-note-list');list.setAttribute('aria-busy','true');
    return request('fieldnotes_market_notes',{symbol:symbol,notes_page:page},noteController).then(function(result){
      if(sequence===noteSequence)renderNotes(result);
    }).catch(function(error){if(sequence!==noteSequence||error.name==='AbortError')return;list.replaceChildren(html('p','boughs-empty',error.message||'记录未能读取，请刷新后重试。'));}).finally(function(){if(sequence===noteSequence)list.removeAttribute('aria-busy');});
  }
  function setStock(value, page, updateHistory) {
    var target=normalize(value);if(!target){input.setAttribute('aria-invalid','true');message('请输入有效的沪深北股票代码，例如 600519。',true);input.focus();return;}
    var changed=symbol!==target;ignoreSavedSymbol='';symbol=target;input.value=target;input.removeAttribute('aria-invalid');closeMatches();
    if(changed){clearLive();liveFailures=0;}lastRefresh=Date.now();
    if(updateHistory)updateURL(page,false);
    if(changed){root.querySelector('.fn-market-note-list').replaceChildren(html('p','boughs-empty','正在读取这只股票的观察记录…'));root.querySelector('[data-stat=matched]').textContent='…';root.querySelector('[data-stat=latest]').textContent='…';}
    var write=root.querySelector('.fn-market-write form');
    if(write&&!write.querySelector('[name=note_title]').value&&!write.querySelector('[name=note_content]').value)write.querySelector('[name=stock]').value=target;
    loadQuotes();loadLive();loadNotes(page);updateAutoState();
  }

  // The watchlist belongs to this browser tab, never to the public WordPress database.
  var watch=root.querySelector('.fn-market-watchlist'),watchItems=root.querySelector('.fn-market-watch-items');
  function readSession() {
    try {
      var saved=JSON.parse(window.sessionStorage.getItem(storageKey)||'{}');
      watchlist=Array.isArray(saved.items)?saved.items.filter(function(item){return item&&normalize(item.symbol)===item.symbol&&typeof item.name==='string';}).slice(0,12):[];
      if([0,2,5,10,30].includes(saved.interval))refreshSeconds=saved.interval;
    } catch(ignore){watchlist=[];}
    intervalSelect.value=String(refreshSeconds);renderWatchlist();
  }
  function writeSession(){try{window.sessionStorage.setItem(storageKey,JSON.stringify({items:watchlist,interval:refreshSeconds}));}catch(ignore){watch.querySelector('small').textContent='浏览器未允许临时保存';}}
  function renderWatchlist(){
    watchItems.replaceChildren();watch.hidden=!watchlist.length;
    watchlist.forEach(function(item){
      var chip=html('div','fn-market-watch-chip'),choose=html('button','fn-market-watch-choose'),remove=html('button','fn-market-watch-remove','×');
      choose.type='button';choose.dataset.stock=item.symbol;choose.append(html('span','',item.name),html('small','',item.symbol));choose.setAttribute('aria-pressed',String(item.symbol===symbol));
      remove.type='button';remove.dataset.removeStock=item.symbol;remove.setAttribute('aria-label','从临时自选移除 '+item.name);chip.append(choose,remove);watchItems.appendChild(chip);
    });
  }
  function rememberStock(code,name){
    if(!normalize(code)||typeof name!=='string'||ignoreSavedSymbol===code)return;
    var current=watchlist.find(function(item){return item.symbol===code;});
    if(current&&current.name===name){renderWatchlist();return;}
    watchlist=watchlist.filter(function(item){return item.symbol!==code;});watchlist.unshift({symbol:code,name:name.slice(0,60)});watchlist=watchlist.slice(0,12);writeSession();renderWatchlist();
  }
  watchItems.addEventListener('click',function(event){
    var remove=event.target.closest('[data-remove-stock]'),choose=event.target.closest('[data-stock]');
    if(remove){if(remove.dataset.removeStock===symbol)ignoreSavedSymbol=symbol;watchlist=watchlist.filter(function(item){return item.symbol!==remove.dataset.removeStock;});writeSession();renderWatchlist();}
    else if(choose){setStock(choose.dataset.stock,1,true);renderWatchlist();}
  });
  root.querySelector('.fn-market-watch-clear').addEventListener('click',function(){ignoreSavedSymbol=symbol;watchlist=[];writeSession();renderWatchlist();});

  // Accessible name search: IME composition, disambiguation and stale-response cancellation.
  var matches=root.querySelector('.fn-market-matches'),searchHint=root.querySelector('#fn-market-search-hint');
  var searchController=null,searchSequence=0,searchTimer=0,composing=false,candidates=[],activeMatch=-1;
  function closeMatches(){input.removeAttribute('aria-busy');searchSequence++;if(searchController)searchController.abort();window.clearTimeout(searchTimer);matches.hidden=true;input.setAttribute('aria-expanded','false');input.removeAttribute('aria-activedescendant');candidates=[];activeMatch=-1;}
  function highlightMatch(index){
    activeMatch=index;matches.querySelectorAll('[role=option]').forEach(function(option,i){option.setAttribute('aria-selected',String(i===index));});
    if(index>=0){var option=matches.children[index];input.setAttribute('aria-activedescendant',option.id);if(option.scrollIntoView)option.scrollIntoView({block:'nearest'});}else input.removeAttribute('aria-activedescendant');
  }
  function renderMatches(items){
    candidates=items.filter(function(item){return item&&normalize(item.symbol)===item.symbol&&typeof item.name==='string';}).slice(0,12);activeMatch=-1;matches.replaceChildren();
    candidates.forEach(function(item,i){var option=html('button','fn-market-match'),name=html('span','',item.name),code=html('small','',item.symbol);option.type='button';option.id='fn-stock-option-'+i;option.dataset.match=i;option.setAttribute('role','option');option.setAttribute('aria-selected','false');option.tabIndex=-1;option.append(name,code);matches.appendChild(option);});
    matches.hidden=!candidates.length;input.setAttribute('aria-expanded',String(candidates.length>0));input.removeAttribute('aria-activedescendant');
  }
  function searchStock(submit){
    window.clearTimeout(searchTimer);var query=input.value.trim();if(!query)return;
    var direct=normalize(query);if(direct){if(submit)setStock(direct,1,true);return;}
    if (/^\d{6}$/.test(query)||/^(?:(?:SH|SZ|BJ)[:.]?\d+|\d+\.(?:SH|SZ|BJ))$/i.test(query)) { input.setAttribute('aria-invalid','true');searchHint.textContent='代码与交易市场不匹配，请核对后重试。';return; }
    if(Array.from(query).length<2){searchHint.textContent='再输入一个字，或使用六位股票代码。';return;}
    searchSequence++;var sequence=searchSequence;if(searchController)searchController.abort();searchController=typeof AbortController!=='undefined'?new AbortController():null;
    searchHint.textContent='正在检索…';input.setAttribute('aria-busy','true');
    request('fieldnotes_market_search',{q:query},searchController).then(function(result){
      if(sequence!==searchSequence||input.value.trim()!==query)return;
      renderMatches(Array.isArray(result.items)?result.items:[]);
      if(submit&&candidates.length===1){var item=candidates[0];setStock(item.symbol,1,true);searchHint.textContent=item.name+' · '+item.symbol;}
      else searchHint.textContent=candidates.length?'选择匹配的股票，再展开行情。':'没有找到沪深北 A 股，请补全名称或输入六位代码。';
    }).catch(function(error){if(sequence===searchSequence&&error.name!=='AbortError'){renderMatches([]);searchHint.textContent=error.message||'检索暂不可用，可直接输入六位代码。';}}).finally(function(){if(sequence===searchSequence)input.removeAttribute('aria-busy');});
  }
  function scheduleSearch(){closeMatches();input.removeAttribute('aria-invalid');input.removeAttribute('aria-busy');if(!composing&&input.value.trim().length>=2&&!normalize(input.value))searchTimer=window.setTimeout(function(){searchStock(false);},350);else searchHint.textContent='沪深北 A 股 · 中文名称、简拼或六位代码';}
  input.addEventListener('compositionstart',function(){composing=true;closeMatches();});
  input.addEventListener('compositionend',function(){composing=false;scheduleSearch();});
  input.addEventListener('input',scheduleSearch);
  input.addEventListener('keydown',function(event){
    if(composing||event.isComposing||event.keyCode===229)return;
    if(event.key==='Escape'){closeMatches();return;}
    if(!matches.hidden&&candidates.length&&(event.key==='ArrowDown'||event.key==='ArrowUp')){event.preventDefault();highlightMatch(activeMatch<0?(event.key==='ArrowDown'?0:candidates.length-1):(activeMatch+(event.key==='ArrowDown'?1:-1)+candidates.length)%candidates.length);}
    else if(event.key==='Enter'&&activeMatch>=0&&!matches.hidden){event.preventDefault();var item=candidates[activeMatch];setStock(item.symbol,1,true);searchHint.textContent=item.name+' · '+item.symbol;}
  });
  matches.addEventListener('pointerdown',function(event){event.preventDefault();});
  matches.addEventListener('click',function(event){var option=event.target.closest('[data-match]');if(!option)return;var item=candidates[Number(option.dataset.match)];if(item){setStock(item.symbol,1,true);searchHint.textContent=item.name+' · '+item.symbol;}});
  form.addEventListener('focusout',function(){window.setTimeout(function(){if(!form.contains(document.activeElement))closeMatches();},0);});
  form.addEventListener('submit',function(event){event.preventDefault();if(!composing)searchStock(true);});

  // A quote has its own exchange timestamp; polling never invents new ticks or candles.
  function clearLive(){liveSequence++;if(liveController)liveController.abort();liveBusy=false;live.hidden=true;delete live.dataset.asOf;live.classList.remove('is-stale');}
  function loadLive(){
    if(!symbol||liveBusy)return;liveBusy=true;var sequence=++liveSequence,target=symbol;
    liveController=typeof AbortController!=='undefined'?new AbortController():null;
    request('fieldnotes_market_quote',{symbol:target},liveController).then(function(result){
      if(sequence!==liveSequence||target!==symbol)return;
      if(result.symbol!==target||typeof result.price!=='number'||!Number.isFinite(result.price)||result.price<=0||typeof result.percent!=='number'||!Number.isFinite(result.percent)||!/^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\+08:00$/.test(result.as_of))throw new Error('报价格式无效');
      var old=result.stale||Date.now()-Date.parse(result.as_of)>90000;
      live.hidden=false;live.dataset.asOf=result.as_of;live.classList.toggle('is-stale',!!old);live.querySelector('.fn-market-live-price').textContent=format(result.price);
      live.querySelector('.fn-market-live-change').textContent=(result.percent>=0?'+':'')+result.percent.toFixed(2)+'%';live.querySelector('.fn-market-live-change').dataset.direction=result.percent>=0?'up':'down';
      live.querySelector('.fn-market-live-time').textContent=result.source+' · 报价 '+result.as_of.slice(0,10)+' '+result.as_of.slice(11,19)+(old?' · 请留意报价时间':'');
      rememberStock(target,result.name);liveFailures=result.stale?Math.min(liveFailures+1,4):0;
    }).catch(function(error){if(sequence!==liveSequence||error.name==='AbortError')return;liveFailures=Math.min(liveFailures+1,4);if(live.hidden){live.hidden=false;live.querySelector('.fn-market-live-price').textContent='—';live.querySelector('.fn-market-live-change').textContent='';}live.classList.add('is-stale');live.querySelector('.fn-market-live-time').textContent=(live.dataset.asOf?'上次报价 '+live.dataset.asOf.slice(0,10)+' '+live.dataset.asOf.slice(11,19)+' · ':'')+'最新报价暂未连接，稍后重试。';}).finally(function(){if(sequence===liveSequence){liveBusy=false;updateAutoState();}});
  }
  function inTradingHours(){
    var parts=new Intl.DateTimeFormat('en-GB',{timeZone:'Asia/Shanghai',weekday:'short',hour:'2-digit',minute:'2-digit',hourCycle:'h23'}).formatToParts(new Date()),clock={};parts.forEach(function(p){clock[p.type]=p.value;});
    var minutes=Number(clock.hour)*60+Number(clock.minute);return !['Sat','Sun'].includes(clock.weekday)&&((minutes>=570&&minutes<690)||(minutes>=780&&minutes<900));
  }
  function updateAutoState(){
    var stateText=!symbol?'选择股票后开始':!refreshSeconds?'自动刷新已暂停':document.visibilityState==='hidden'?'页面离开时暂停':!inTradingHours()?'非交易时段 · 可手动刷新':liveFailures?'连接稍慢 · 自动降低频率重试':'每 '+refreshSeconds+' 秒尝试更新 · 行情可能延迟';if(autoState.textContent!==stateText)autoState.textContent=stateText;
  }
  function poll(){
    updateAutoState();if(!symbol||!refreshSeconds||document.visibilityState==='hidden'||!inTradingHours()||liveBusy||chart.getAttribute('aria-busy')==='true')return;
    var delay=Math.min(60,refreshSeconds*Math.pow(2,liveFailures))*1000;if(Date.now()-lastRefresh<delay)return;
    lastRefresh=Date.now();loadLive();loadQuotes(true);
  }
  intervalSelect.addEventListener('change',function(){var value=Number(intervalSelect.value);refreshSeconds=[0,2,5,10,30].includes(value)?value:5;lastRefresh=0;writeSession();poll();});
  document.addEventListener('visibilitychange',function(){if(document.visibilityState==='visible'){lastRefresh=0;poll();}else updateAutoState();});
  window.addEventListener('pagehide',function(){closeMatches();clearLive();if(quoteController)quoteController.abort();if(noteController)noteController.abort();});

  root.querySelectorAll('[data-period]').forEach(function(button){button.addEventListener('click',function(){if(!symbol||period===button.dataset.period)return;period=button.dataset.period;updateControls();updateURL(paramsFromURL().page,false);loadQuotes();});});
  adjustSelect.addEventListener('change',function(){adjust=adjustSelect.value==='none'?'none':'qfq';if(symbol){updateURL(paramsFromURL().page,false);loadQuotes();}});
  retry.addEventListener('click',function(){loadQuotes(false);loadLive();});
  root.querySelector('.fn-market-refresh').addEventListener('click',function(){lastRefresh=Date.now();loadQuotes(true);loadLive();});
  function zoom(direction){if(!data||period==='minute')return;var visible=Math.min(count,data.bars.length-offset);count=Math.max(24,Math.min(data.bars.length,Math.round(visible*(direction>0?.8:1.25))));draw();}
  root.querySelector('.fn-market-zoom-out').addEventListener('click',function(){zoom(-1);});
  root.querySelector('.fn-market-zoom-in').addEventListener('click',function(){zoom(1);});
  root.querySelector('.fn-market-zoom-reset').addEventListener('click',function(){count=80;offset=0;hideBar();draw();});
  chart.addEventListener('pointermove',function(event){if(!plot)return;var rect=chart.getBoundingClientRect(),x=(event.clientX-rect.left)*plot.width/Math.max(1,rect.width);var index=Math.floor((x-plot.left)/plot.step);if(plot.minute){var distance=Infinity;plot.bars.forEach(function(p,i){var diff=Math.abs(plot.x(i)-x);if(diff<distance){distance=diff;index=i;}});}showBar(index);});
  chart.addEventListener('pointerleave',hideBar);chart.addEventListener('blur',hideBar);
  chart.addEventListener('keydown',function(event){if(!plot)return;if(event.key==='ArrowLeft'||event.key==='ArrowRight'){event.preventDefault();showBar(cursor<0?plot.bars.length-1:cursor+(event.key==='ArrowLeft'?-1:1));}else if(event.key==='Escape')hideBar();});
  chart.addEventListener('wheel',function(event){if(!data||period==='minute')return;if(event.ctrlKey||event.metaKey){event.preventDefault();zoom(event.deltaY<0?1:-1);}else if(event.shiftKey){event.preventDefault();offset=Math.max(0,Math.min(data.bars.length-24,offset+(event.deltaY>0?10:-10)));hideBar();draw();}},{passive:false});
  root.querySelector('.fn-market-note-pagination').addEventListener('click',function(event){var link=event.target.closest('[data-notes-page]');if(!link||event.metaKey||event.ctrlKey||event.shiftKey||event.altKey)return;event.preventDefault();var page=parseInt(link.dataset.notesPage,10);updateURL(page,false);loadNotes(page);});
  var writeLink=root.querySelector('a[href="#fn-market-write"]');if(writeLink)writeLink.addEventListener('click',function(){root.querySelector('.fn-market-write').open=true;});
  if(typeof ResizeObserver!=='undefined'){new ResizeObserver(draw).observe(chart);}else window.addEventListener('resize',draw);
  window.addEventListener('popstate',function(){var current=paramsFromURL();period=current.period;adjust=current.adjust;if(current.symbol)setStock(current.symbol,current.page,false);else{quoteSequence++;if(quoteController)quoteController.abort();closeMatches();clearLive();chart.removeAttribute('aria-busy');symbol='';input.value='';title.textContent='价格的变化，留给图表。';subtitle.textContent='判断的来由，留在文字里。';source.textContent='历史行情 · 可能存在延迟';resetChart();message('');loadNotes(1);}updateControls();});
  window.setInterval(poll,1000);
  readSession();
  var initial=paramsFromURL();period=initial.period;adjust=initial.adjust;updateControls();if(symbol||initial.symbol)setStock(initial.symbol||symbol,initial.page,false);else if(watchlist.length&&!new URL(window.location.href).searchParams.has('stock'))setStock(watchlist[0].symbol,1,false);updateAutoState();
}());
