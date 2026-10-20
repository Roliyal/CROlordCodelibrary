(function () {
  'use strict';
  var roots = document.querySelectorAll('.fnchat');
  var config = window.fieldnotesChatGPT;
  if (!roots.length || !config) return;
  var root = document.querySelector('.chatgpt-page-content .fnchat') || document.querySelector('#main-content .fnchat') || roots[0];
  // Covers direct template/widget output as well as filtered page content.
  Array.prototype.forEach.call(roots, function (node) { if (node !== root) node.remove(); });
  if (root.dataset.fnchatInitialized === '1') return;
  root.dataset.fnchatInitialized = '1';
  var strings = config.strings || {};
  window.addEventListener('pageshow', function (event) { if (event.persisted) window.location.reload(); });

  function sendRequest(action, fields) {
    var body = new URLSearchParams();
    body.set('action', action);
    body.set('nonce', config.nonce);
    Object.keys(fields || {}).forEach(function (key) { body.set(key, fields[key]); });
    return fetch(config.ajaxUrl, {
      method: 'POST', credentials: 'same-origin', cache: 'no-store',
      headers: { 'Content-Type': 'application/x-www-form-urlencoded; charset=UTF-8' },
      body: body.toString()
    }).then(function (response) {
      return response.json().catch(function () { throw new Error(strings.networkError); });
    }).then(function (result) {
      if (!result || !result.success) {
        var error = new Error((result && result.data && result.data.message) || strings.networkError);
        error.code = result && result.data && result.data.code;
        throw error;
      }
      return result;
    });
  }

  var unlockForm = root.querySelector('.fnchat-unlock');
  if (unlockForm) {
    var input = unlockForm.querySelector('input[name="password"]');
    var button = unlockForm.querySelector('button[type="submit"]');
    var status = unlockForm.querySelector('.fnchat-form-status');
    var show = unlockForm.querySelector('.fnchat-show-password');
    var unlocking = false;
    if (show) show.addEventListener('click', function () {
      var visible = input.type === 'password';
      input.type = visible ? 'text' : 'password';
      show.setAttribute('aria-pressed', String(visible));
      show.setAttribute('aria-label', visible ? show.dataset.hide : show.dataset.show);
    });
    unlockForm.addEventListener('submit', function (event) {
      event.preventDefault();
      if (unlocking) return;
      if (!input.value.trim()) {
        status.textContent = strings.passwordRequired;
        input.setAttribute('aria-invalid', 'true');
        input.focus();
        return;
      }
      unlocking = true;
      button.disabled = true;
      button.querySelector('span').textContent = strings.working;
      input.removeAttribute('aria-invalid');
      status.textContent = '';
      sendRequest('fieldnotes_chatgpt_unlock', { password: input.value }).then(function () {
        input.value = '';
        window.location.reload();
      }).catch(function (error) {
        status.textContent = error.message || strings.networkError;
        input.setAttribute('aria-invalid', 'true');
        input.focus();
        unlocking = false;
        button.disabled = false;
        button.querySelector('span').textContent = strings.enter || '进入对话';
      });
    });
    return;
  }

  var log = root.querySelector('.fnchat-log');
  var form = root.querySelector('.fnchat-compose');
  var textarea = root.querySelector('#fnchat-message');
  if (!log || !form || !textarea) return;
  var sendButton = form.querySelector('button[type="submit"]');
  var clearButton = root.querySelector('.fnchat-clear');
  var lockButton = root.querySelector('.fnchat-lock');
  var reauthButton = root.querySelector('.fnchat-reauth');
  var composeStatus = root.querySelector('#fnchat-compose-status');
  var welcomeTemplate = log.querySelector('.fnchat-welcome');
  welcomeTemplate = welcomeTemplate ? welcomeTemplate.cloneNode(true) : null;
  var history = [];
  var busy = false;
  var locking = false;
  var conversation = '', revision = '', pendingApprovals = [], uncertain = false;
  var connectPanel = root.querySelector('.fnchat-connect-panel');
  var connectToggle = root.querySelector('.fnchat-connections-toggle');
  var toolCount = root.querySelector('.fnchat-tool-count');
  var approvalArea = root.querySelector('.fnchat-approval-area');
  var exportButton = root.querySelector('.fnchat-export');
  var media = null;
  if (window.FieldnotesChatMedia) media = new window.FieldnotesChatMedia(root, config, {request: sendRequest, say: say, change: updateControls, error: handleError});
  else say('附件组件未加载，请刷新页面后重试。');

  function say(text) { if (composeStatus) composeStatus.textContent = text || ''; }
  function scrollLog() { log.scrollTop = log.scrollHeight; }
  function resizeInput() { textarea.style.height = 'auto'; textarea.style.height = Math.min(textarea.scrollHeight, 160) + 'px'; }
  function restoreWelcome() {
    if (!log.children.length && welcomeTemplate) log.appendChild(welcomeTemplate.cloneNode(true));
  }
  function addMessage(role, text, pending) {
    var welcome = log.querySelector('.fnchat-welcome');
    if (welcome) welcome.remove();
    var row = document.createElement('article');
    row.className = 'fnchat-message fnchat-message-' + role + (pending ? ' is-pending' : '');
    var label = document.createElement('span');
    label.className = 'fnchat-message-label';
    label.textContent = role === 'assistant' ? 'Echoes' : '你';
    var body = document.createElement('div');
    body.className = 'fnchat-message-body';
    body.textContent = text;
    row.append(label, body);
    log.appendChild(row);
    scrollLog();
    return { row: row, body: body };
  }

  // Build a small Markdown subset with DOM nodes. Model text never becomes HTML.
  function inlineText(parent, text) {
    var pattern = /(`[^`\n]+`|\*\*[^*\n]+\*\*|\[[^\]\n]+\]\(https?:\/\/[^\s)]+\))/g;
    var start = 0, match;
    while ((match = pattern.exec(text))) {
      parent.appendChild(document.createTextNode(text.slice(start, match.index)));
      var value = match[0], node;
      if (value.charAt(0) === '`') { node = document.createElement('code'); node.textContent = value.slice(1, -1); }
      else if (value.slice(0, 2) === '**') { node = document.createElement('strong'); node.textContent = value.slice(2, -2); }
      else {
        var split = value.indexOf('](');
        node = document.createElement('a'); node.textContent = value.slice(1, split);
        node.href = value.slice(split + 2, -1); node.target = '_blank'; node.rel = 'noopener noreferrer';
      }
      parent.appendChild(node); start = pattern.lastIndex;
    }
    parent.appendChild(document.createTextNode(text.slice(start)));
  }
  function renderReply(parent, text) {
    parent.replaceChildren();
    var lines = text.replace(/\r\n?/g, '\n').split('\n'), i = 0;
    while (i < lines.length) {
      var line = lines[i], match, node;
      if (!line.trim()) { i++; continue; }
      if (/^\s*```/.test(line)) {
        var codeLines = []; i++;
        while (i < lines.length && !/^\s*```/.test(lines[i])) codeLines.push(lines[i++]);
        i++;
        node = document.createElement('pre');
        var code = document.createElement('code'); code.textContent = codeLines.join('\n');
        node.appendChild(code); parent.appendChild(node); continue;
      }
      if (line.indexOf('|') !== -1 && i + 1 < lines.length && /^\s*\|?\s*:?-{3,}/.test(lines[i+1]) && /^[\s|:\-]+$/.test(lines[i+1])) {
        var wrap = document.createElement('div'); wrap.className = 'fnchat-table-wrap';
        var table = document.createElement('table'), thead = document.createElement('thead'), tbody = document.createElement('tbody');
        function cells(value) { return value.trim().replace(/^\||\|$/g, '').split('|').map(function (cell) { return cell.trim(); }); }
        function tableRow(values, header) { var tr = document.createElement('tr'); values.forEach(function (value) { var td = document.createElement(header ? 'th' : 'td'); if (header) td.scope = 'col'; inlineText(td, value); tr.appendChild(td); }); return tr; }
        thead.appendChild(tableRow(cells(line), true)); i += 2;
        while (i < lines.length && lines[i].trim() && lines[i].indexOf('|') !== -1) tbody.appendChild(tableRow(cells(lines[i++]), false));
        table.append(thead, tbody); wrap.appendChild(table); parent.appendChild(wrap); continue;
      }
      if ((match = /^(#{1,6})\s+(.+)$/.exec(line))) {
        node = document.createElement(match[1].length <= 2 ? 'h3' : 'h4'); inlineText(node, match[2]); parent.appendChild(node); i++; continue;
      }
      if (/^\s*(?:[-*+] |\d+\. )/.test(line)) {
        var ordered = /^\s*\d+\. /.test(line);
        node = document.createElement(ordered ? 'ol' : 'ul');
        while (i < lines.length && (match = (ordered ? /^\s*\d+\. (.*)$/ : /^\s*[-*+] (.*)$/).exec(lines[i]))) {
          var item = document.createElement('li'); inlineText(item, match[1]); node.appendChild(item); i++;
        }
        parent.appendChild(node); continue;
      }
      if (/^>\s?/.test(line)) {
        node = document.createElement('blockquote'); var quoted = [];
        while (i < lines.length && /^>\s?/.test(lines[i])) quoted.push(lines[i++].replace(/^>\s?/, ''));
        inlineText(node, quoted.join('\n')); parent.appendChild(node); continue;
      }
      var paragraph = [line]; i++;
      while (i < lines.length && lines[i].trim() && !/^(?:\s*```|#{1,6}\s|\s*[-*+] |\s*\d+\. |>)/.test(lines[i])) paragraph.push(lines[i++]);
      node = document.createElement('p'); inlineText(node, paragraph.join('\n')); parent.appendChild(node);
    }
  }
  function legacyCopy(text) {
    return new Promise(function (resolve, reject) {
      var active = document.activeElement;
      var helper = document.createElement('textarea');
      helper.value = text; helper.setAttribute('readonly', '');
      helper.style.position = 'fixed'; helper.style.opacity = '0';
      document.body.appendChild(helper); helper.select();
      var copied = false;
      try { copied = document.execCommand('copy'); } catch (error) { copied = false; }
      helper.remove();
      if (active && active.focus) active.focus({ preventScroll: true });
      copied ? resolve() : reject(new Error(strings.copyFailed));
    });
  }
  function copyText(text) {
    if (window.isSecureContext && navigator.clipboard && navigator.clipboard.writeText) {
      return navigator.clipboard.writeText(text).catch(function () { return legacyCopy(text); });
    }
    return legacyCopy(text);
  }
  function addCopyButton(row, text) {
    var copy = document.createElement('button');
    copy.type = 'button'; copy.className = 'fnchat-copy'; copy.textContent = strings.copy || '复制回复';
    copy.addEventListener('click', function () {
      copyText(text).then(function () {
        copy.textContent = strings.copied || '已复制';
        window.setTimeout(function () { copy.textContent = strings.copy || '复制回复'; }, 1600);
      }).catch(function () { say(strings.copyFailed); });
    });
    row.appendChild(copy);
  }

  function selectedTools() {
    return Array.prototype.map.call(root.querySelectorAll('.fnchat-connect-list input:checked'), function (input) { return input.value; });
  }
  function updateControls() {
    var unavailable = root.dataset.ready !== '1';
    var files = media ? media.snapshot() : [];
    var uploading = media && media.loading();
    sendButton.disabled = busy || locking || unavailable || pendingApprovals.length > 0 || uncertain || uploading || media && media.hasErrors() || (!textarea.value.trim() && !files.length);
    textarea.disabled = busy || locking || unavailable;
    if (clearButton) clearButton.disabled = busy || locking || uploading;
    if (lockButton) lockButton.disabled = busy || locking || uploading;
    if (exportButton) exportButton.disabled = !history.length || busy;
    if (connectToggle) connectToggle.disabled = busy || locking || pendingApprovals.length > 0;
    root.querySelectorAll('.fnchat-connect-list input').forEach(function (input) { input.disabled = busy || locking || !!conversation; });
    root.querySelectorAll('[data-prompt]').forEach(function (button) { button.disabled = busy || locking || unavailable || pendingApprovals.length > 0 || uncertain; });
    if (media) media.block(busy || locking || pendingApprovals.length > 0 || uncertain);
    root.classList.toggle('is-busy', busy);
    log.setAttribute('aria-busy', String(busy));
  }
  function handleError(error) {
    if (reauthButton && error.code === 'session_expired') reauthButton.hidden = false;
  }
  function setConnections(open) { connectPanel.hidden = !open; connectToggle.setAttribute('aria-expanded', String(open)); }
  var list = root.querySelector('.fnchat-connect-list');
  (config.servers || []).forEach(function (server) {
    var label = document.createElement('label'), input = document.createElement('input'), text = document.createElement('span'), strong = document.createElement('strong'), small = document.createElement('small');
    input.type = 'checkbox'; input.value = server.label;
    strong.textContent = server.title; small.textContent = server.host;
    text.append(strong, small); label.append(input, text); list.appendChild(label);
    input.addEventListener('change', function () { var count = selectedTools().length; toolCount.textContent = count; toolCount.hidden = !count; });
  });
  if (!(config.servers || []).length) {
    var empty = document.createElement('p'); empty.className = 'fnchat-connect-empty'; empty.textContent = '还没有连接工具。管理员可在「设置 → ChatGPT 接入」中添加远程 MCP 服务。'; list.appendChild(empty);
  }
  connectToggle.addEventListener('click', function () { setConnections(connectPanel.hidden); });
  root.querySelector('.fnchat-close-connections').addEventListener('click', function () { setConnections(false); connectToggle.focus(); });
  document.addEventListener('click', function (event) { if (!connectPanel.hidden && !connectPanel.contains(event.target) && !connectToggle.contains(event.target)) setConnections(false); });
  root.addEventListener('keydown', function (event) { if (event.key === 'Escape' && !connectPanel.hidden) { setConnections(false); connectToggle.focus(); } });
  function activityLines(row, activity) {
    (activity || []).forEach(function (entry) {
      var detail = document.createElement('p'); detail.className = 'fnchat-activity';
      detail.textContent = entry.server + ' / ' + entry.name + (entry.ok ? ' · 已完成' : ' · 调用未完成');
      row.appendChild(detail);
    });
  }
  function acceptResult(data, reply) {
    if (!data || !/^[a-f0-9]{32}$/.test(data.conversation || '') || !Array.isArray(data.approvals)) throw new Error(strings.replyError);
    conversation = data.conversation; revision = data.revision || ''; pendingApprovals = data.approvals; uncertain = false;
    var text = typeof data.text === 'string' ? data.text : '';
    reply.row.classList.remove('is-pending');
    if (text) {
      renderReply(reply.body, text); addCopyButton(reply.row, text); history.push({role:'assistant', content:text});
    } else {
      reply.body.textContent = pendingApprovals.length ? '需要你的确认，才能继续使用工具。' : '工具处理已结束。';
    }
    activityLines(reply.row, data.activity);
    if (data.incomplete) { var note = document.createElement('p'); note.className = 'fnchat-activity'; note.textContent = '回答达到本次输出上限，可能尚未完整。'; reply.row.appendChild(note); }
    showApprovals(); updateControls(); scrollLog();
  }
  function showApprovals() {
    approvalArea.replaceChildren(); approvalArea.hidden = !pendingApprovals.length;
    if (!pendingApprovals.length) return;
    var heading = document.createElement('h3'); heading.textContent = '在继续之前，确认一下。';
    var intro = document.createElement('p'); intro.textContent = '以下参数会发送给对应服务。请逐项选择允许或拒绝。';
    var approvalForm = document.createElement('form'); var decisions = {};
    approvalArea.append(heading, intro, approvalForm);
    pendingApprovals.forEach(function (entry, index) {
      var item = document.createElement('div'); item.className = 'fnchat-approval-item';
      var name = document.createElement('strong'); name.textContent = entry.name;
      var host = document.createElement('small'); host.textContent = entry.server + ' · ' + entry.host;
      var details = document.createElement('details'); details.open = true;
      var summary = document.createElement('summary'); summary.textContent = '将要发送的参数';
      var args = document.createElement('pre');
      try { args.textContent = JSON.stringify(JSON.parse(entry.arguments), null, 2); } catch (error) { args.textContent = entry.arguments; }
      details.append(summary, args);
      var options = document.createElement('div'); options.className = 'fnchat-approval-options';
      [true, false].forEach(function (allowed) {
        var label = document.createElement('label'), radio = document.createElement('input'); radio.type = 'radio'; radio.name = 'approval-' + index; radio.value = allowed ? 'allow' : 'deny';
        label.append(radio, document.createTextNode(allowed ? '允许这次调用' : '拒绝'));
        radio.addEventListener('change', function () { decisions[entry.id] = allowed; confirm.disabled = Object.keys(decisions).length !== pendingApprovals.length; });
        options.appendChild(label);
      });
      item.append(name, host, details, options); approvalForm.appendChild(item);
    });
    var confirm = document.createElement('button'); confirm.type = 'submit'; confirm.className = 'fnchat-approval-submit'; confirm.textContent = '确认并继续'; confirm.disabled = true; approvalForm.appendChild(confirm);
    approvalForm.addEventListener('submit', function (event) {
      event.preventDefault(); if (busy || Object.keys(decisions).length !== pendingApprovals.length) return;
      busy = true; say(''); setConnections(false); updateControls();
      approvalForm.querySelectorAll('input,button').forEach(function (node) { node.disabled = true; });
      var reply = addMessage('assistant', '正在处理工具请求…', true);
      sendRequest('fieldnotes_chatgpt_approve', {conversation: conversation, revision: revision, decisions: JSON.stringify(decisions)}).then(function (result) {
        acceptResult(result.data, reply);
      }).catch(function (error) {
        reply.row.remove(); uncertain = true; say(error.message || strings.networkError); handleError(error);
        var notice = document.createElement('p'); notice.className = 'fnchat-activity'; notice.textContent = '本次确认已停止。请检查对应服务的结果，随后开启新对话。'; approvalArea.appendChild(notice);
      }).finally(function () { busy = false; updateControls(); });
    });
  }
  function submitMessage(value) {
    var message = (value || '').trim(), files = media ? media.snapshot() : [];
    if ((!message && !files.length) || sendButton.disabled || busy || locking || message.length > 5000) return;
    var user = addMessage('user', message, false);
    if (files.length && media) user.row.insertBefore(media.messageFiles(files), user.body);
    if (!message) user.body.hidden = true;
    var reply = addMessage('assistant', strings.thinking || '正在思考…', true);
    busy = true; say(''); if (reauthButton) reauthButton.hidden = true; setConnections(false);
    textarea.value = ''; resizeInput(); updateControls();
    sendRequest('fieldnotes_chatgpt_chat', {message:message, conversation:conversation, attachments:JSON.stringify(media ? media.refs() : []), tools:JSON.stringify(selectedTools())}).then(function (result) {
      // Validate before consuming the draft or moving attachments out of the composer.
      if (!result.data || !/^[a-f0-9]{32}$/.test(result.data.conversation || '') || !Array.isArray(result.data.approvals)) throw new Error(strings.replyError);
      history.push({role:'user', content:message + (files.length ? '\n\n附件：' + files.map(function (file) { return file.name; }).join('、') : '')});
      if (media) media.commit();
      acceptResult(result.data, reply);
    }).catch(function (error) {
      user.row.remove(); reply.row.remove(); restoreWelcome(); textarea.value = message; resizeInput();
      say((error.message || strings.networkError) + ' ' + (strings.retry || '草稿与附件已保留，可以再次发送。')); handleError(error);
    }).finally(function () { busy = false; updateControls(); if (!textarea.disabled) textarea.focus({preventScroll:true}); });
  }
  form.addEventListener('submit', function (event) { event.preventDefault(); submitMessage(textarea.value); });
  textarea.addEventListener('input', function () { resizeInput(); updateControls(); });
  textarea.addEventListener('keydown', function (event) {
    if (event.key === 'Enter' && !event.shiftKey && !event.isComposing && event.keyCode !== 229) { event.preventDefault(); submitMessage(textarea.value); }
  });
  root.addEventListener('click', function (event) {
    var prompt = event.target.closest('[data-prompt]');
    if (!prompt || prompt.disabled || textarea.disabled) return;
    textarea.value = prompt.dataset.prompt || ''; resizeInput(); updateControls(); textarea.focus();
  });
  async function cleanConversation() {
    var pendingCleanup = false;
    if (conversation) {
      var result = await sendRequest('fieldnotes_chatgpt_reset', {conversation:conversation});
      pendingCleanup = result.data && result.data.cleanup_pending;
    }
    if (media) await media.discard();
    history = []; conversation = ''; revision = ''; pendingApprovals = []; uncertain = false;
    log.replaceChildren(); restoreWelcome(); showApprovals(); resizeInput();
    return pendingCleanup;
  }
  if (clearButton) clearButton.addEventListener('click', function () {
    if (clearButton.disabled) return;
    busy = true; say(''); setConnections(false); updateControls();
    cleanConversation().then(function (pendingCleanup) {
      say(pendingCleanup ? '新对话已开启。旧附件未能立即删除，将按一小时过期策略清理。' : '');
    }).catch(function (error) { say(error.message || strings.networkError); handleError(error); }).finally(function () { busy = false; updateControls(); if (!textarea.disabled) textarea.focus(); });
  });
  if (lockButton) lockButton.addEventListener('click', function () {
    if (lockButton.disabled) return;
    locking = true; say(''); updateControls();
    // Lock must still succeed if best-effort attachment cleanup cannot reach OpenAI.
    cleanConversation().catch(function () {}).then(function () { return sendRequest('fieldnotes_chatgpt_lock', {}); }).then(function () {
      history = []; log.replaceChildren(); textarea.value = ''; window.location.reload();
    }).catch(function (error) { locking = false; updateControls(); say(error.message || strings.networkError); });
  });
  if (exportButton) exportButton.addEventListener('click', function () {
    if (!history.length) return;
    var text = '# Echoes\n\n' + history.map(function (turn) { return '## ' + (turn.role === 'user' ? '你' : 'Echoes') + '\n\n' + turn.content; }).join('\n\n');
    var blob = new Blob([text], {type:'text/markdown;charset=utf-8'}), url = URL.createObjectURL(blob), link = document.createElement('a');
    link.href = url; link.download = 'echoes-' + new Date().toISOString().slice(0,10) + '.md'; document.body.appendChild(link); link.click(); link.remove(); window.setTimeout(function () { URL.revokeObjectURL(url); }, 1000);
  });
  if (reauthButton) reauthButton.addEventListener('click', function () { window.location.reload(); });
  if (window.visualViewport) {
    var viewport = function () { root.style.setProperty('--fnchat-vh', Math.round(window.visualViewport.height) + 'px'); };
    window.visualViewport.addEventListener('resize', viewport); viewport();
  }
  resizeInput(); updateControls();
}());
