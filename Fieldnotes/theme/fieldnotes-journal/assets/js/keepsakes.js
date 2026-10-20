(function () {
  'use strict';
  var root=document.querySelector('.boughs-keepsakes');if(!root||root.dataset.jdReady)return;root.dataset.jdReady='1';
  var bar=root.querySelector('.fn-keepsakes-selection'),all=root.querySelector('.fn-keepsakes-select-page'),review=root.querySelector('.fn-keepsakes-review');
  var dialog=root.querySelector('.fn-keepsakes-dialog'),list=root.querySelector('.fn-keepsakes-list'),notice=root.querySelector('.fn-keepsakes-list-status');
  var key='fieldnotes:keepsakes:v1:'+window.location.pathname,selected=[],products=[],returnFocus=null;
  function valid(item){return item&&/^\d{1,12}$/.test(String(item.id))&&typeof item.title==='string'&&item.title.length<=300&&typeof item.keyword==='string'&&item.keyword.length<=300&&typeof item.sku==='string'&&(!item.sku||/^[1-9][0-9]{4,17}$/.test(item.sku));}
  function element(tag,cls,text){var el=document.createElement(tag);if(cls)el.className=cls;if(text!=null)el.textContent=text;return el;}
  function productURL(item){return item.sku?'https://item.jd.com/'+item.sku+'.html':'https://search.jd.com/Search?keyword='+encodeURIComponent(item.keyword||item.title)+'&enc=utf-8';}
  try{var stored=JSON.parse(window.sessionStorage.getItem(key)||'[]');if(Array.isArray(stored))selected=stored.filter(valid).slice(0,40);}catch(ignore){}
  root.querySelectorAll('[data-keepsake]').forEach(function(card){
    try{var product=JSON.parse(card.dataset.keepsake);if(!valid(product))return;products.push({card:card,item:product,input:card.querySelector('.fn-keepsake-choose input')});card.querySelector('.fn-keepsake-choose').hidden=false;}catch(ignore){}
  });
  function save(){try{window.sessionStorage.setItem(key,JSON.stringify(selected));}catch(ignore){notice.textContent='浏览器未允许临时保存；本页仍可使用清单。';}}
  function update(){
    var checked=0;products.forEach(function(product){product.input.checked=selected.some(function(item){return item.id===product.item.id;});product.card.classList.toggle('is-selected',product.input.checked);if(product.input.checked)checked++;});
    all.checked=products.length>0&&checked===products.length;all.indeterminate=checked>0&&checked<products.length;all.disabled=!products.length;
    root.querySelector('.fn-keepsakes-selection-count').textContent=selected.length?'已选择 '+selected.length+' 件':'尚未选择';review.disabled=!selected.length;root.querySelector('.fn-keepsakes-clear').disabled=!selected.length;
  }
  function add(item){if(!selected.some(function(current){return current.id===item.id;})){if(selected.length>=40)return false;selected.push(item);}return true;}
  products.forEach(function(product){product.input.addEventListener('change',function(){if(product.input.checked){if(!add(product.item)){notice.textContent='清单最多保留 40 件，整理后可以继续选择。';openList();}}else selected=selected.filter(function(item){return item.id!==product.item.id;});save();update();});});
  all.addEventListener('change',function(){var choose=all.checked;products.forEach(function(product){if(choose)add(product.item);else selected=selected.filter(function(item){return item.id!==product.item.id;});});save();update();});
  root.querySelector('.fn-keepsakes-clear').addEventListener('click',function(){selected=[];save();update();});
  function renderList(){
    list.replaceChildren();selected.forEach(function(item){
      var row=element('li'),copy=element('div'),link=element('a','',item.sku?'查看商品 ↗':'去选择款式 ↗'),remove=element('button','fn-keepsakes-remove','×');
      copy.append(element('h3','',item.title),element('p','',item.sku?'京东商品 '+item.sku:'先在京东找到合适的商品与规格'));
      link.href=productURL(item);link.target='_blank';link.rel='noopener noreferrer';remove.type='button';remove.dataset.remove=item.id;remove.setAttribute('aria-label','从选购清单移除 '+item.title);row.append(copy,link,remove);list.appendChild(row);
    });
    if(!selected.length)list.appendChild(element('li','fn-keepsakes-list-empty','清单已清空。再挑选几件喜欢的物品吧。'));
    root.querySelector('.fn-keepsakes-copy').disabled=!selected.length;
  }
  function openList(){
    returnFocus=review;notice.textContent='';renderList();
    if(typeof dialog.showModal==='function')dialog.showModal();else{dialog.setAttribute('open','');dialog.setAttribute('role','dialog');dialog.setAttribute('aria-modal','true');}
    document.body.classList.add('fn-keepsakes-open');root.querySelector('.fn-keepsakes-close').focus();
  }
  function closeList(){if(typeof dialog.close==='function')dialog.close();else dialog.removeAttribute('open');document.body.classList.remove('fn-keepsakes-open');if(returnFocus&&returnFocus.isConnected)returnFocus.focus();}
  review.addEventListener('click',openList);root.querySelector('.fn-keepsakes-close').addEventListener('click',closeList);
  dialog.addEventListener('cancel',function(event){event.preventDefault();closeList();});
  dialog.addEventListener('click',function(event){if(event.target===dialog){var rect=dialog.getBoundingClientRect();if(event.clientX<rect.left||event.clientX>rect.right||event.clientY<rect.top||event.clientY>rect.bottom)closeList();}});
  dialog.addEventListener('keydown',function(event){
    if(event.key==='Escape'){event.preventDefault();closeList();return;}
    if(event.key!=='Tab')return;var focusable=Array.from(dialog.querySelectorAll('a[href],button:not(:disabled),textarea:not(:disabled)')),first=focusable[0],last=focusable[focusable.length-1];
    if(event.shiftKey&&document.activeElement===first){event.preventDefault();last.focus();}else if(!event.shiftKey&&document.activeElement===last){event.preventDefault();first.focus();}
  });
  list.addEventListener('click',function(event){var button=event.target.closest('[data-remove]');if(!button)return;selected=selected.filter(function(item){return item.id!==button.dataset.remove;});save();update();renderList();root.querySelector('.fn-keepsakes-close').focus();});
  root.querySelector('.fn-keepsakes-copy').addEventListener('click',function(){
    var text=selected.map(function(item,i){return(i+1)+'. '+item.title+'\n'+productURL(item);}).join('\n\n');
    if(window.isSecureContext&&navigator.clipboard&&navigator.clipboard.writeText){navigator.clipboard.writeText(text).then(function(){notice.textContent='选购清单已复制。';}).catch(function(){manualCopy(text);});}else manualCopy(text);
  });
  function manualCopy(text){
    var area=dialog.querySelector('.fn-keepsakes-copy-text');if(!area){area=element('textarea','fn-keepsakes-copy-text');area.readOnly=true;area.setAttribute('aria-label','选购清单，可全选复制');dialog.insertBefore(area,notice);}area.value=text;area.focus();area.select();notice.textContent='清单已选中，请复制后留存。';
  }
  bar.hidden=false;update();
}());
