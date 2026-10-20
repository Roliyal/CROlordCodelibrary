(function () {
  'use strict';
  // Local previews use object URLs. Only opaque server references enter chat requests.
  window.FieldnotesChatMedia = function (root, config, callbacks) {
    var form = root.querySelector('.fnchat-compose');
    var tray = root.querySelector('.fnchat-attachments');
    var input = root.querySelector('.fnchat-file-input');
    var attach = root.querySelector('.fnchat-attach');
    var overlay = root.querySelector('.fnchat-drop-label');
    var preview = root.querySelector('.fnchat-preview');
    var content = root.querySelector('.fnchat-preview-content');
    var title = root.querySelector('.fnchat-preview-title');
    var close = root.querySelector('.fnchat-preview-close');
    var items = [], archived = [], pending = 0, blocked = false, dragDepth = 0, active = null;
    var api = {};
    function sizeLabel(bytes) { return bytes < 1048576 ? Math.max(1, Math.round(bytes / 1024)) + ' KB' : (bytes / 1048576).toFixed(1) + ' MB'; }
    function element(tag, className, text) { var node = document.createElement(tag); if (className) node.className = className; if (text != null) node.textContent = text; return node; }
    function changed() {
      pending = items.filter(function (item) { return item.status === 'uploading'; }).length;
      attach.disabled = blocked || !config.uploadsEnabled || root.dataset.ready !== '1';
      render(); if (callbacks.change) callbacks.change();
    }
    function upload(item) {
      if (item.status === 'uploading' || blocked) return Promise.resolve();
      item.status = 'uploading'; item.error = ''; changed();
      var prepare = item.kind === 'video' && !item.prepared ? videoFrames(item.file) : Promise.resolve(item.prepared || item.file);
      return prepare.then(function (file) {
        item.prepared = file;
        if (file.size > config.uploadLimit) throw new Error('文件超过本站 ' + sizeLabel(config.uploadLimit) + ' 的上传上限。');
        var body = new FormData(); body.set('action', 'fieldnotes_chatgpt_upload'); body.set('nonce', config.nonce); body.set('file', file, file.name);
        if (item.kind === 'video') { body.set('kind', 'video_frames'); body.set('source_name', item.file.name); }
        return fetch(config.ajaxUrl, {method: 'POST', credentials: 'same-origin', cache: 'no-store', body: body}).then(function (response) { return response.json().catch(function () { throw new Error('上传未完成，请核对服务器上传限制或网络。'); }); });
      }).then(function (result) {
        if (!result || !result.success || !result.data || !/^[a-f0-9]{32}$/.test(result.data.ref || '')) {
          var error = new Error(result && result.data && result.data.message || '附件没有上传成功，请重试。');
          error.code = result && result.data && result.data.code; throw error;
        }
        item.ref = result.data.ref; item.status = 'ready'; item.transcript = result.data.transcript || ''; item.expires = result.data.expires;
      }).catch(function (error) {
        item.status = 'error'; item.error = error.message || '上传失败，请重试。';
        callbacks.say(item.error); if (callbacks.error) callbacks.error(error);
      }).finally(changed);
    }
    function kind(file) {
      var ext = file.name.split('.').pop().toLowerCase();
      if (/^(jpg|jpeg|png|webp|gif)$/.test(ext)) return 'image';
      if (/^(mp4|mov)$/.test(ext) || ext === 'webm' && /^video\//.test(file.type)) return 'video';
      if (/^(mp3|mpga|mpeg|m4a|wav|webm)$/.test(ext)) return 'audio';
      if (/^(pdf|txt|md|csv|tsv|json|doc|docx|ppt|pptx|xls|xlsx)$/.test(ext)) return 'document';
      return '';
    }
    function add(files) {
      if (blocked || !config.uploadsEnabled || root.dataset.ready !== '1') return;
      var chain = Promise.resolve();
      Array.prototype.forEach.call(files || [], function (file) {
        if (items.length >= 4) { callbacks.say('每条消息最多添加 4 份附件。'); return; }
        var type = kind(file);
        if (!type) { callbacks.say('暂不支持这个文件格式。'); return; }
        if (type === 'audio' && !config.audioEnabled) { callbacks.say('本站尚未开启音频转写。'); return; }
        if (!file.size || file.size > (type === 'video' ? 100 * 1048576 : config.uploadLimit)) { callbacks.say(type === 'video' ? '视频请控制在 100 MB 内。' : '文件超过本站 ' + sizeLabel(config.uploadLimit) + ' 的上传上限。'); return; }
        var item = {file: file, name: file.name, kind: type, status: 'queued', ref: '', url: '', thumbnail: '', transcript: '', error: ''};
        if (/^(image|audio|video)$/.test(type)) item.url = URL.createObjectURL(file);
        if (type === 'image') item.thumbnail = item.url;
        items.push(item);
        chain = chain.then(function () { return upload(item); });
      });
      changed();
    }
    function hint(item) {
      if (item.status === 'error') return '读取失败 · 点击重试';
      if (item.status === 'queued') return '等待上传';
      if (item.status === 'uploading') return item.kind === 'audio' ? '正在上传并转写…' : item.kind === 'video' ? '正在准备四帧摘要…' : '正在上传并准备…';
      return item.kind === 'video' ? '四帧视觉摘要 · 不含声音' : item.kind === 'audio' ? '语音已转成文字' : sizeLabel(item.file.size) + ' · 已就绪';
    }
    function chip(item, removable) {
      var wrapper = element('div', 'fnchat-file-chip' + (item.status === 'error' ? ' is-error' : ''));
      var open = element('button', 'fnchat-file-open'); open.type = 'button';
      open.setAttribute('aria-label', (item.status === 'error' && removable ? '重试上传 ' : '预览 ') + item.name);
      if (item.thumbnail) { var img = element('img'); img.src = item.thumbnail; img.alt = ''; open.appendChild(img); }
      else open.appendChild(element('span', 'fnchat-file-mark', item.kind === 'audio' ? 'AUDIO' : item.kind === 'video' ? 'VIDEO' : item.file.name.split('.').pop().slice(0,4)));
      var info = element('span', 'fnchat-file-info'); info.append(element('strong', '', item.name), element('small', '', hint(item))); open.appendChild(info);
      open.addEventListener('click', function () { if (item.status === 'error' && removable) { upload(item); } else { showPreview(item, open); } });
      wrapper.appendChild(open);
      if (removable) {
        var remove = element('button', 'fnchat-file-remove', '×'); remove.type = 'button'; remove.setAttribute('aria-label', '移除 ' + item.name);
        remove.disabled = blocked || item.status === 'uploading' || item.status === 'queued';
        remove.addEventListener('click', function () {
          if (remove.disabled) return;
          remove.disabled = true;
          var work = item.ref ? callbacks.request('fieldnotes_chatgpt_remove_file', {ref: item.ref}) : Promise.resolve();
          work.then(function (result) {
            if (result && result.data && result.data.cleanup_pending) callbacks.say('已移出对话；远端附件未能立即删除，将按一小时过期策略清理。');
            items = items.filter(function (entry) { return entry !== item; }); release(item); changed();
          }).catch(function (error) {
            if (error.code === 'file_expired') { items = items.filter(function (entry) { return entry !== item; }); release(item); changed(); }
            else { callbacks.say(error.message); remove.disabled = false; }
          });
        });
        wrapper.appendChild(remove);
      }
      return wrapper;
    }
    function render() {
      tray.replaceChildren(); tray.hidden = items.length === 0;
      items.forEach(function (item) { tray.appendChild(chip(item, true)); });
    }
    function release(item) {
      if (item.url) URL.revokeObjectURL(item.url);
      if (item.thumbnail && item.thumbnail !== item.url && item.thumbnail.indexOf('blob:') === 0) URL.revokeObjectURL(item.thumbnail);
    }
    function showPreview(item, origin) {
      active = origin; title.textContent = item.name; content.replaceChildren();
      if (item.kind === 'image' || item.kind === 'video') {
        var media = element(item.kind === 'video' ? 'video' : 'img'); media.src = item.url;
        if (item.kind === 'video') { media.controls = true; media.playsInline = true; } else media.alt = item.name;
        content.appendChild(media);
        if (item.kind === 'video') content.appendChild(element('p', '', '对话只读取四张等间隔关键帧，不分析视频声音；这里可预览原视频。'));
      } else if (item.kind === 'audio') {
        var audio = element('audio'); audio.controls = true; audio.src = item.url; content.appendChild(audio);
        content.appendChild(element('p', '', '对话将使用语音转写，不分析音色、音乐及其他声音。'));
        if (item.transcript) content.appendChild(element('pre', '', item.transcript));
      } else {
        content.appendChild(element('p', '', sizeLabel(item.file.size) + ' · 文档将作为附件交给模型阅读。'));
        if (/\.(txt|md|csv|tsv|json)$/i.test(item.name)) {
          var pre = element('pre', '', '正在读取本地预览…'); content.appendChild(pre);
          item.file.slice(0, 24000).text().then(function (text) { pre.textContent = text + (item.file.size > 24000 ? '\n…预览已截断，模型接收完整文件。' : ''); }).catch(function () { pre.textContent = '本地预览不可用，附件仍可发送。'; });
        }
      }
      if (item.error) content.appendChild(element('p', '', item.error));
      preview.hidden = false; root.querySelector('.fnchat-shell').setAttribute('inert', ''); close.focus();
    }
    function closePreview() {
      preview.hidden = true; content.querySelectorAll('video,audio').forEach(function (media) { media.pause(); }); content.replaceChildren();
      root.querySelector('.fnchat-shell').removeAttribute('inert'); if (active && active.isConnected) active.focus();
    }
    close.addEventListener('click', closePreview);
    preview.addEventListener('click', function (event) { if (event.target === preview) closePreview(); });
    preview.addEventListener('keydown', function (event) {
      if (event.key === 'Escape') { event.preventDefault(); closePreview(); }
      if (event.key === 'Tab') {
        var nodes = Array.prototype.slice.call(preview.querySelectorAll('button,video[controls],audio[controls],a[href]'));
        var first = nodes[0], last = nodes[nodes.length - 1];
        if (event.shiftKey && document.activeElement === first) { event.preventDefault(); last.focus(); }
        else if (!event.shiftKey && document.activeElement === last) { event.preventDefault(); first.focus(); }
      }
    });
    function videoFrames(file) {
      // A single contact sheet keeps uploads bounded and makes the actual video input explicit.
      return new Promise(function (resolve, reject) {
        var video = document.createElement('video'), url = URL.createObjectURL(file), stopped = false;
        video.muted = true; video.playsInline = true; video.preload = 'auto';
        var timer = window.setTimeout(function () { finish(new Error('无法读取这个视频，请换成可播放的 MP4 或 WebM。')); }, 35000);
        function finish(error, value) { if (stopped) return; stopped = true; window.clearTimeout(timer); video.pause(); video.removeAttribute('src'); video.load(); URL.revokeObjectURL(url); error ? reject(error) : resolve(value); }
        video.onerror = function () { finish(new Error('浏览器无法解码这个视频，请转换为 MP4 后重试。')); };
        video.onloadedmetadata = async function () {
          if (!Number.isFinite(video.duration) || video.duration <= 0) { finish(new Error('无法取得视频时长。')); return; }
          var canvas = document.createElement('canvas'); canvas.width = 1280; canvas.height = 800;
          var ctx = canvas.getContext('2d'); if (!ctx) { finish(new Error('浏览器不支持视频关键帧读取。')); return; }
          ctx.fillStyle = '#f5f5f7'; ctx.fillRect(0,0,1280,800);
          try {
            for (var i = 0; i < 4; i++) {
              var time = Math.min(video.duration - .001, video.duration * (i + .5) / 4);
              await new Promise(function (res, rej) {
                var wait = window.setTimeout(function () { rej(new Error('读取视频画面超时。')); }, 6500);
                video.onseeked = function () { window.clearTimeout(wait); res(); }; video.currentTime = Math.max(.001, time);
              });
              if (stopped) return;
              var cellX = i % 2 * 640, cellY = Math.floor(i / 2) * 400;
              var ratio = Math.min(624 / video.videoWidth, 350 / video.videoHeight);
              var width = video.videoWidth * ratio, height = video.videoHeight * ratio;
              ctx.drawImage(video, cellX + (640-width)/2, cellY + 8 + (350-height)/2, width, height);
              ctx.fillStyle = '#515156'; ctx.font = '16px sans-serif'; ctx.fillText('FRAME ' + (i+1) + '  /  ' + Math.floor(time/60) + ':' + String(Math.floor(time%60)).padStart(2,'0'), cellX + 16, cellY + 385);
            }
            canvas.toBlob(function (blob) { if (!blob) { finish(new Error('关键帧导出失败。')); return; } finish(null, new File([blob], 'video-frames.jpg', {type:'image/jpeg'})); }, 'image/jpeg', .86);
          } catch (error) { finish(error); }
        };
        video.src = url;
      });
    }
    attach.addEventListener('click', function () { if (!attach.disabled) input.click(); });
    input.addEventListener('change', function () { add(input.files); input.value = ''; });
    form.addEventListener('paste', function (event) {
      if (blocked || !event.clipboardData || !event.clipboardData.files.length) return;
      event.preventDefault(); add(event.clipboardData.files);
    });
    form.addEventListener('dragenter', function (event) { if (event.dataTransfer && Array.prototype.indexOf.call(event.dataTransfer.types, 'Files') !== -1 && !blocked) { event.preventDefault(); dragDepth++; overlay.hidden = false; } });
    form.addEventListener('dragover', function (event) { if (!blocked && event.dataTransfer && Array.prototype.indexOf.call(event.dataTransfer.types, 'Files') !== -1) { event.preventDefault(); event.dataTransfer.dropEffect = 'copy'; } });
    form.addEventListener('dragleave', function (event) { event.preventDefault(); dragDepth = Math.max(0, dragDepth-1); if (!dragDepth) overlay.hidden = true; });
    form.addEventListener('drop', function (event) { event.preventDefault(); dragDepth = 0; overlay.hidden = true; if (event.dataTransfer) add(event.dataTransfer.files); });
    api.loading = function () { return pending > 0 || items.some(function (item) { return item.status === 'queued'; }); };
    api.hasErrors = function () { return items.some(function (item) { return item.status === 'error'; }); };
    api.snapshot = function () { return items.slice(); };
    api.refs = function () { return items.map(function (item) { return item.ref; }); };
    api.block = function (value) { blocked = value; attach.disabled = blocked || !config.uploadsEnabled || root.dataset.ready !== '1'; tray.querySelectorAll('.fnchat-file-remove').forEach(function (button, index) { button.disabled = blocked || items[index].status === 'uploading' || items[index].status === 'queued'; }); };
    api.messageFiles = function (files) { var group = element('div', 'fnchat-message-files'); files.forEach(function (item) { group.appendChild(chip(item, false)); }); return group; };
    api.commit = function () { archived = archived.concat(items); items = []; changed(); };
    api.discard = async function () {
      for (var i = 0; i < items.length; i++) if (items[i].ref) {
        try { await callbacks.request('fieldnotes_chatgpt_remove_file', {ref: items[i].ref}); }
        catch (error) { if (error.code !== 'file_expired') throw error; }
      }
      items.concat(archived).forEach(release); items = []; archived = []; changed();
    };
    api.add = add; api.closePreview = closePreview;
    window.addEventListener('pagehide', function () { items.concat(archived).forEach(release); });
    changed(); return api;
  };
}());
