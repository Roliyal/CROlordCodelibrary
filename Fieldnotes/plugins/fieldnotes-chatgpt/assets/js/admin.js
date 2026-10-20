(function () {
  'use strict';
  var button = document.getElementById('fnchat-diagnose');
  var output = document.getElementById('fnchat-diagnostic-result');
  var config = window.fieldnotesChatGPTAdmin;
  if (!button || !output || !config) return;
  var original = button.textContent;
  function paragraph(text, className) { var p = document.createElement('p'); p.className = className || ''; p.textContent = text || ''; output.appendChild(p); }
  button.addEventListener('click', function () {
    if (button.disabled) return;
    button.disabled = true; button.textContent = config.working; output.replaceChildren(); output.setAttribute('aria-busy', 'true');
    var body = new URLSearchParams(); body.set('action', 'fieldnotes_chatgpt_diagnose'); body.set('nonce', config.nonce);
    fetch(config.ajaxUrl, {method: 'POST', credentials: 'same-origin', cache: 'no-store', headers: {'Content-Type': 'application/x-www-form-urlencoded; charset=UTF-8'}, body: body.toString()}).then(function (response) {
      return response.json().catch(function () { throw new Error(config.error); });
    }).then(function (result) {
      if (!result || !result.success || !result.data) throw new Error((result && result.data && result.data.message) || config.error);
      var report = result.data;
      output.dataset.state = report.ok ? 'ok' : 'issue';
      paragraph(report.message, 'fnchat-diagnostic-title'); paragraph(report.hint);
      var rows = [['结果代码', report.code], ['API HTTP 状态', report.http_status || '未获得 HTTP 响应'], ['检测耗时', report.duration_ms + ' ms'], ['当前模型', report.model], ['Key 来源', report.key_source], ['对话请求时限', report.request_timeout + ' 秒'], ['代理配置', report.proxy_configured ? 'WordPress 已配置代理' : '未设置 WordPress 代理']];
      if (report.request_id) rows.push(['请求标识', report.request_id]);
      if (report.diagnostic_timeout) rows.push(['本次检测总时限', report.diagnostic_timeout + ' 秒（连接阶段可能更早超时）']);
      if (report.technical) rows.push(['技术详情（已脱敏）', report.technical]);
      var list = document.createElement('dl');
      rows.forEach(function (row) {var term = document.createElement('dt'), description = document.createElement('dd');term.textContent = row[0];description.textContent = String(row[1] == null ? '' : row[1]);list.append(term, description);});
      output.appendChild(list);
    }).catch(function (error) { output.dataset.state = 'issue';paragraph(error.message || config.error); }).finally(function () {button.disabled = false;button.textContent = original;output.removeAttribute('aria-busy');});
  });
}());
