// oxlint-disable anti-slop/no-runtime-typeof
(function () {
  const A = window.ReqlyApp;
  const { D, S, esc, icon, methodChip, statusChip, ms, bytes, ago, fmtDate, fmtClock, highlight, jsonPathGet, resolveVars, undefinedVars, kvEditor } = A;
  const ui = S.ui;

  const RVH = {};
  RVH.segBtns = (name, opts, cur) => `<div class="seg" role="tablist">${opts.map(o => `<button class="seg-btn ${o === cur ? 'active' : ''}" data-vact="seg" data-name="${name}" data-val="${esc(o)}" role="tab" aria-selected="${o === cur}">${esc(o)}</button>`).join('')}</div>`;
  RVH.codeEditor = (id, val, opt = {}) => {
    const lines = Math.max((val || '').split('\n').length, opt.rows || 8);
    return `<div class="code-editor ${opt.cls || ''}"><div class="ce-gutter mono">${Array.from({ length: lines }, (_, i) => `<span>${i + 1}</span>`).join('')}</div><textarea id="${id}" class="mono ce-area" spellcheck="false" wrap="${opt.wrap === false ? 'off' : 'soft'}" placeholder="${esc(opt.placeholder || '')}">${esc(val)}</textarea></div>`;
  };
  RVH.emptyState = (ic, title, sub, cta) => `<div class="empty">${icon(ic)}<h4>${esc(title)}</h4><p>${sub || ''}</p>${cta || ''}</div>`;
  RVH.banner = (kind, html) => `<div class="banner banner-${kind}">${icon(kind === 'error' ? 'warn' : kind === 'success' ? 'check' : 'info')}<div>${html}</div></div>`;
  RVH.table = (head, rows, cls) => `<div class="tbl-wrap ${cls || ''}"><table class="tbl"><thead><tr>${head.map(h => `<th>${h}</th>`).join('')}</tr></thead><tbody>${rows.map(r => `<tr>${r.map(c => `<td>${c}</td>`).join('')}</tr>`).join('') || '<tr><td colspan="9" class="empty-line">No rows</td></tr>'}</tbody></table></div>`;
  RVH.badge = (txt, tone) => `<span class="chip chip-${tone || 'muted'}">${esc(txt)}</span>`;
  window.RVH = RVH;

  window.RVH = RVH;

  function bindViewActs(host) {
    host.querySelectorAll('[data-vact]').forEach(el => {
      const k = el.dataset.vact;
      if (el.dataset.bound) return;
      el.dataset.bound = '1';
      if (k === 'seg') el.addEventListener('click', () => VACT.seg(el));
      else if (VACT[k]) el.addEventListener('click', () => VACT[k](el));
    });
    host.querySelectorAll('[data-vinput]').forEach(el => {
      const k = el.dataset.vinput;
      if (el.dataset.bound) return;
      el.dataset.bound = '1';
      if (k === 'script-area') el.addEventListener('input', () => { const t = A.activeTab(); t.scripts[el.dataset.phase] = el.value; t.dirty = true; A.refreshTabsBar(); });
      else if (VINPUT[k]) {
        el.addEventListener(el.tagName === 'SELECT' ? 'change' : 'input', () => VINPUT[k](el));
      }
    });
  }
  const VACT = {}, VINPUT = {};

  A.regView({
    id: 'overview', title: 'Overview', icon: 'home', group: '',
    render() {
      const runsToday = D.history.filter(h => Date.now() - new Date(h.ts) < 864e5).length;
      const okRuns = D.history.filter(h => h.status >= 200 && h.status < 300).length;
      const rate = Math.round(okRuns / Math.max(1, D.history.length) * 100);
      return `
      <div class="ov-hero card">
        <div>
          <h2>${esc(A.wsById().name)}</h2>
          <p class="sub">Git-native workspace · <span class="mono">${esc(A.wsById().path)}</span> · branch <b>${esc(D.gitState.branch)}</b></p>
          <div class="ov-chips">
            ${RVH.badge(A.wsById().requests + ' requests', 'primary')}
            ${RVH.badge(D.environments.length + ' environments', 'info')}
            ${RVH.badge(D.testSuites.reduce((a, s) => a + s.tests.length, 0) + ' assertions', 'ok')}
            ${RVH.badge('zero telemetry', 'muted')}
          </div>
        </div>
        <div class="ov-stats">
          <div class="stat"><b>${runsToday}</b><span>runs today</span></div>
          <div class="stat"><b>${rate}%</b><span>passing</span></div>
          <div class="stat"><b>${ms(Math.round(D.history.reduce((a, h) => a + h.time, 0) / D.history.length))}</b><span>avg latency</span></div>
        </div>
      </div>
      <div class="ov-grid">
        <div class="card pad">
          <h4>Quick actions</h4>
          <div class="qa-grid">
            ${[['rest', 'rest', 'New request'], ['importexport', 'import', 'Import API'], ['mocks', 'server', 'Spin up a mock'], ['diff', 'compare', 'Diff two specs'], ['codegen', 'braces', 'Generate client'], ['openapi', 'book', 'Browse OpenAPI']].map(q => `<button class="qa-tile" data-action="nav" data-view="${q[0]}">${icon(q[1])}<span>${q[2]}</span></button>`).join('')}
          </div>
        </div>
        <div class="card pad">
          <h4>Recent activity</h4>
          <div class="recent-list">${D.history.slice(0, 6).map(h => `<button class="recent-row" data-vact="open-hist" data-id="${h.id}">${methodChip(h.method)}<span class="mono ellip">${esc(h.url)}</span>${statusChip(h.status)}<em>${ago(h.ts)}</em></button>`).join('')}</div>
        </div>
        <div class="card pad">
          <h4>Environment</h4>
          <p class="sub">Active: <b>${A.envById() ? esc(A.envById().name) : 'none selected'}</b></p>
          <div class="mini-kv">${A.envById() ? A.envById().vars.slice(0, 4).map(v => `<div><code>{{${esc(v.key)}}}</code><span class="mono ellip">${esc(v.value.slice(0, 34))}</span></div>`).join('') : '<p class="empty-line">No variables resolve right now.</p>'}</div>
          <button class="btn btn-sm" data-action="nav" data-view="environments">Manage environments</button>
        </div>
        <div class="card pad">
          <h4>Repository snapshot</h4>
          <div class="git-mini">
            <div class="health-row"><i class="led led-ok"></i><span>working tree</span><em>${D.gitState.changes.length} changed</em></div>
            <div class="health-row"><i class="led led-warn"></i><span>ahead / behind</span><em>${D.gitState.ahead}↑ ${D.gitState.behind}↓</em></div>
            <div class="health-row"><i class="led led-error"></i><span>conflicts</span><em>1 unresolved</em></div>
          </div>
          <button class="btn btn-sm" data-action="nav" data-view="git">Open Git panel</button>
        </div>
        <div class="card pad protocols-card">
          <h4>Protocol clients</h4>
          <div class="proto-tiles">
            ${[['graphql', 'graphql', 'GraphQL'], ['grpc', 'bolt', 'gRPC'], ['websocket', 'plug', 'WebSocket'], ['sse', 'signal', 'SSE'], ['soap', 'doc', 'SOAP']].map(p => `<button class="proto-tile" data-action="nav" data-view="${p[0]}">${icon(p[1])}<span>${p[2]}</span></button>`).join('')}
          </div>
        </div>
      </div>`;
    },
    after(host) { bindViewActs(host); }
  });
  VACT['open-hist'] = (el) => { ui.histSel = el.dataset.id; A.navigate('history'); };
  VACT['idle-open'] = (el) => {
    const h = D.history.find(x => x.id === el.dataset.id);
    if (!h) return;
    const node = h.requestId && A.walkTree(S.tree).find(n => n.id === h.requestId);
    if (node) A.findOrMakeTab(node.id);
    else A.newTabFromNode({ id: A.uid(), name: 'Replay — ' + (h.url.split('?')[0].split('/').pop() || h.method), method: h.method, url: h.url, saved: false });
    A.navigate('rest');
  };
  RVH.VACT = VACT; RVH.VINPUT = VINPUT;

  const BUILDER_TABS = [['params', 'Params'], ['headers', 'Headers'], ['auth', 'Auth'], ['body', 'Body'], ['scripts', 'Scripts']];
  function restLayout(builderHtml, respHtml) {
    if (A.opts.responseMode === 'stacked') return `<div class="rest-stacked"><div class="pane builder-pane">${builderHtml}</div><div class="pane resp-pane">${respHtml}</div></div>`;
    if (A.opts.responseMode === 'third') return `<div class="rest-third"><div class="pane builder-pane">${builderHtml}</div><div class="vsep"></div><div class="pane resp-pane resp-pane-cmd">${respHtml}</div></div>`;
    return `<div class="rest-split"><div class="pane builder-pane">${builderHtml}</div><div class="vsep vsep-drag" data-tip="Drag to resize"></div><div class="pane resp-pane">${respHtml}</div></div>`;
  }

  function authEditor(t) {
    const a = t.auth;
    const types = ['none', 'bearer', 'basic', 'apikey', 'oauth2', 'aws sigv4', 'edgegrid'];
    let fields = '';
    if (a.type === 'bearer') fields = `<label class="fld"><span>Token</span><input class="mono auth-fld" data-fld="token" type="${/SECRET/.test(a.token) ? 'password' : 'text'}" value="${esc(a.token || '')}" placeholder="access token"></label>`;
    if (a.type === 'basic') fields = `<label class="fld"><span>Username</span><input class="mono auth-fld" data-fld="username" value="${esc(a.username || '')}"></label><label class="fld"><span>Password</span><input class="mono auth-fld" data-fld="password" type="password" value="${esc(a.password || '')}"></label>`;
    if (a.type === 'apikey') fields = `<label class="fld"><span>Key</span><input class="mono auth-fld" data-fld="key" value="${esc(a.key || 'X-Api-Key')}"></label><label class="fld"><span>Value</span><input class="mono auth-fld" data-fld="value" type="password" value="${esc(a.value || '')}"></label><label class="fld"><span>Add to</span><select class="auth-fld" data-fld="in"><option>header</option><option>query</option></select></label>`;
    if (a.type === 'oauth2') fields = `
      <div class="oauth-grid">
        <label class="fld"><span>Grant type</span><select class="auth-fld" data-fld="grant"><option ${a.grant === 'client_credentials' ? 'selected' : ''}>client_credentials</option><option ${a.grant === 'authorization_code + PKCE' ? 'selected' : ''}>authorization_code + PKCE</option><option ${a.grant === 'device flow' ? 'selected' : ''}>device flow</option><option ${a.grant === 'refresh_token' ? 'selected' : ''}>refresh_token</option></select></label>
        <label class="fld"><span>Auth URL</span><input class="mono auth-fld" data-fld="authUrl" value="${esc(a.authUrl || 'https://api.paycore.dev/oauth/authorize')}"></label>
        <label class="fld"><span>Token URL</span><input class="mono auth-fld" data-fld="tokenUrl" value="${esc(a.tokenUrl || '{{baseUrl}}/../oauth/token')}"></label>
        <label class="fld"><span>Client ID</span><input class="mono auth-fld" data-fld="clientId" value="{{clientId}}"></label>
        <label class="fld"><span>Client secret</span><input class="mono auth-fld" data-fld="clientSecret" type="password" value="{{clientSecret}}"></label>
        <label class="fld"><span>Scopes</span><input class="mono auth-fld" data-fld="scopes" value="payments:read payments:write"></label>
      </div>
      <div class="oauth-actions"><button class="btn btn-sm" data-vact="oauth-fetch">${icon('key')} Get token</button><span class="tok-state" data-role="tok-state">Token cached · expires in 58 min</span></div>`;
    if (a.type === 'aws sigv4') fields = `<label class="fld"><span>Access key</span><input class="mono auth-fld" data-fld="ak" value="${esc(a.ak || 'AKIA…')}"></label><label class="fld"><span>Secret key</span><input class="mono auth-fld" data-fld="sk" type="password" value="${esc(a.sk || '')}"></label><label class="fld"><span>Region</span><input class="mono auth-fld" data-fld="region" value="us-east-1"></label><label class="fld"><span>Service</span><input class="mono auth-fld" data-fld="svc" value="execute-api"></label>`;
    if (a.type === 'edgegrid') fields = `<label class="fld"><span>Client token</span><input class="mono auth-fld" data-fld="ct" value="eg-client-…"></label><label class="fld"><span>Access token</span><input class="mono auth-fld" data-fld="at" value="eg-access-…"></label><label class="fld"><span>Client secret</span><input class="mono auth-fld" data-fld="cs" type="password" value=""></label><p class="field-hint">EG1-HMAC-SHA256 signing applied automatically.</p>`;
    return `<div class="auth-editor">
      <label class="fld"><span>Type</span><select data-vinput="auth-type">${types.map(x => `<option ${a.type === x ? 'selected' : ''}>${x}</option>`).join('')}</select></label>
      ${fields}
      ${A.walkTree(S.tree).find(n => n.id === t.nodeId)?._parent?.auth ? `<p class="inherit-note">${icon('lock')} Folder “Authentication” sets OAuth2 — this request inherits unless overridden.</p>` : ''}
    </div>`;
  }

  function bodyEditor(t) {
    const b = t.body;
    const kinds = ['none', 'json', 'raw', 'form', 'multipart', 'binary'];
    let inner = '';
    if (b.type === 'json') {
      const invalid = b.text && !validJson(b.text);
      inner = `${invalid ? RVH.banner('error', '<b>Invalid JSON.</b> Fix syntax before sending — parser will reject this payload.') : ''}
        <div class="editor-tools"><button class="btn btn-sm" data-vact="fmt-json">${icon('braces')} Format</button><span class="mono dim">${(b.text || '').length} chars</span></div>
        ${RVH.codeEditor('body-json', b.text, { lang: 'json', rows: 12, placeholder: '{ "amount": 1000 }' })}`;
    } else if (b.type === 'raw') inner = `<div class="editor-tools"><select data-vinput="raw-lang">${['text', 'javascript', 'xml', 'yaml'].map(l => `<option ${(b.lang || 'text') === l ? 'selected' : ''}>${l}</option>`).join('')}</select></div>${RVH.codeEditor('body-raw', b.text || '', { rows: 10 })}`;
    else if (b.type === 'form') inner = kvEditor('form', b.form || [], false);
    else if (b.type === 'multipart') inner = `${kvEditor('form', b.form || [], false)}<p class="field-hint">${icon('doc')} Rows with a <b>file</b> value open a file picker on send. Boundary is generated per request.</p><button class="btn btn-sm" data-vact="pick-file">${icon('plus')} Attach file…</button>`;
    else if (b.type === 'binary') inner = `<button class="btn" data-vact="pick-file">${icon('download')} Choose binary file…</button> <span class="dim">${b.file || 'no file selected'}</span>`;
    return `<div class="body-kind">${RVH.segBtns('body-kind', kinds, b.type)}</div>${inner}`;
  }
  const validJson = (s) => { try { JSON.parse(s); return true; } catch (e) { return false; } };

  function scriptEditor(t, phase) {
    const snips = D.scriptSnippets;
    return `<div class="scripts-wrap">
      <div class="snippets"><h6>reqly.* helpers</h6>${snips.map(s => `<button class="snip" data-vact="insert-snip" data-phase="${phase}" data-code="${esc(s.code)}"><code>${esc(s.label)}</code></button>`).join('')}</div>
      <div class="script-main">
        ${RVH.codeEditor(`script-${phase}`, t.scripts[phase] || '', { lang: 'js', rows: 10, placeholder: phase === 'pre' ? '// runs before request — mutate reqly.request / env' : '// runs after response — assert, capture secrets' })}
        <div class="editor-tools"><span class="dim">${phase === 'pre' ? 'Pre-request' : 'Post-response'} · Goja sandbox · timeout 500ms</span><button class="btn btn-sm" data-vact="run-script" data-phase="${phase}">${icon('play')} Run once</button></div>
      </div></div>`;
  }

  function builderHtml(t) {
    const undef = undefinedVars([t.url, t.body?.text, ...(t.params || []).map(p => p.value)].filter(Boolean).join(' '));
    const counts = { params: t.params.length, headers: t.headers.filter(h => h.enabled !== false).length, body: t.body.type === 'none' ? '' : t.body.type, scripts: (t.scripts.pre ? 1 : 0) + (t.scripts.post ? 1 : 0) };
    return `
    <div class="urlbar">
      <select class="method-select m-${{ GET: 'ok', POST: 'info', PUT: 'warn', PATCH: 'warn', DELETE: 'error' }[t.method] || 'muted'}" data-vinput="method">${['GET', 'POST', 'PUT', 'PATCH', 'DELETE', 'HEAD', 'OPTIONS'].map(m => `<option ${t.method === m ? 'selected' : ''}>${m}</option>`).join('')}</select>
      <input class="url-input mono" data-vinput="url" value="${esc(t.url)}" placeholder="https:// or {{baseUrl}}/path" aria-label="Request URL">
      ${t.response.state === 'loading'
        ? `<button class="btn btn-danger send-btn" data-vact="cancel">${icon('stop')} Cancel</button>`
        : `<button class="btn btn-primary send-btn" data-vact="send">${icon('play')} Send</button>`}
      <button class="btn" data-vact="save" data-tip="Save (⌘S)" aria-label="Save request">${icon('check')} Save</button>
      <button class="btn btn-icon" data-vact="req-menu" data-tip="Request actions" aria-label="More actions">${icon('dots')}</button>
    </div>
    ${t.dirty ? `<div class="dirty-strip">${icon('edit')} Unsaved changes — <button class="linklike" data-vact="save">save now</button> or <button class="linklike" data-vact="revert">revert</button></div>` : ''}
    ${undef.length ? RVH.banner('warn', `Unresolved variables: ${undef.map(v => `<code>{{${esc(v)}}}</code>`).join(' ')} — will be sent literally.`) : ''}
    <div class="btabs" role="tablist">
      ${BUILDER_TABS.map(b => `<button class="btab ${t.builderTab === b[0] ? 'active' : ''}" role="tab" aria-selected="${t.builderTab === b[0]}" data-vact="btab" data-tab="${b[0]}">${b[1]}${counts[b[0]] !== '' && counts[b[0]] > 0 ? `<i class="btab-count">${counts[b[0]]}</i>` : ''}${b[0] === 'params' && t.params.some(p => p.enabled === false) ? '<i class="btab-off">·</i>' : ''}</button>`).join('')}
    </div>
    <div class="bpane">
      ${t.builderTab === 'params' ? kvEditor('params', t.params, false) : ''}
      ${t.builderTab === 'headers' ? kvEditor('headers', t.headers, false) : ''}
      ${t.builderTab === 'auth' ? authEditor(t) : ''}
      ${t.builderTab === 'body' ? bodyEditor(t) : ''}
      ${t.builderTab === 'scripts' ? `<div class="script-phases">${scriptEditor(t, 'pre')}${scriptEditor(t, 'post')}</div>` : ''}
    </div>`;
  }

  function skeletonLines() { return Array.from({ length: 7 }, (_, i) => `<div class="sk-line" style="width:${88 - i * 9}%"></div>`).join(''); }

  function respDoneHtml(t) {
    const r = t.response;
    const bodyObj = (() => { try { return JSON.parse(r.body); } catch (e) { return null; } })();
    const modes = ['pretty', 'raw', 'table'];
    let bodyMain = '';
    const searched = t.resSearch ? (r.body.toLowerCase().split(t.resSearch.toLowerCase()).length - 1) : 0;
    if (r.largeRows && !ui.largeLoaded) {
      const lg = A.largeBodyHtml();
      bodyMain = `<p class="large-note">${icon('info')} ${lg.note}</p>
        <div class="tbl-scroll">${RVH.table(['#', 'id', 'amount', 'currency', 'status', 'created'], lg.rows.slice(0, 60).map(x => [x.i, x.id, x.amount, x.currency, RVH.badge(x.status, x.status === 'succeeded' ? 'ok' : 'warn'), x.created]))}</div>
        <button class="btn" data-vact="load-large">${icon('download')} Render full payload (virtualized)</button>`;
    } else if (r.largeRows && ui.largeLoaded) {
      const lg = A.largeBodyHtml();
      bodyMain = `<p class="dim small">${lg.note}</p><div class="tbl-scroll">${RVH.table(['#', 'id', 'amount', 'currency', 'status'], lg.rows.map(x => [x.i, x.id, x.amount, x.currency, x.status]))}</div><button class="btn btn-sm" data-vact="download-body">${icon('download')} Download full JSON (${bytes(r.size)})</button>`;
    } else if (!r.body && !r.size) {
      bodyMain = RVH.emptyState('doc', 'Empty body', 'This response carried no content (typical for 204 No Content).');
    } else if (t.resMode === 'pretty' && bodyObj) {
      bodyMain = `<pre class="codeview mono" tabindex="0">${highlight(JSON.stringify(bodyObj, null, 2), 'json')}</pre>`;
    } else if (t.resMode === 'table' && bodyObj) {
      const arr = Array.isArray(bodyObj.data || bodyObj) ? (bodyObj.data || bodyObj) : Object.entries(bodyObj).map(([k, v]) => ({ key: k, value: typeof v === 'object' ? JSON.stringify(v) : String(v) }));
      const cols = [...new Set(arr.flatMap(o => Object.keys(typeof o === 'object' ? o : { value: o })))].slice(0, 7);
      bodyMain = `<div class="tbl-scroll">${RVH.table(cols, arr.map(row => cols.map(c => { const v = typeof row === 'object' ? row[c] : row; return typeof v === 'object' ? '<code class="mono">' + esc(JSON.stringify(v)?.slice(0, 40)) + '</code>' : esc(String(v ?? '—')); })))}</div>`;
    } else {
      bodyMain = `<pre class="codeview mono">${highlight(r.body || '', r.body?.trim()?.startsWith('{') ? 'json' : 'http')}</pre>`;
    }
    const pathRes = t.pathHit !== '' && bodyObj ? jsonPathGet(bodyObj, t.jsonPath) : undefined;
    return `
    <div class="resp-meta">
      ${statusChip(r.status)}<span class="st-text">${esc(r.statusText || '')}</span>
      <span class="meta-sep">·</span><span>${ms(r.time)}</span>
      <span class="meta-sep">·</span><span>${bytes(r.size)}</span>
      ${r.status >= 400 ? RVH.badge('error response', 'error') : ''}
      ${RVH.badge('1 attempt', 'muted')}
      <span class="st-flex"></span>
      <span class="mono dim req-id">req ${esc(r.headers['x-request-id'] || '—')}</span>
    </div>
    <div class="resp-toolbar">
      <div class="rtabs">
        ${[['body', 'Body'], ['headers', `Headers (${Object.keys(r.headers).length})`], ['cookies', `Cookies (${(r.cookies || []).length})`]].map(x => `<button class="rtab ${t.resTab === x[0] ? 'active' : ''}" data-vact="rtab" data-tab="${x[0]}">${x[1]}</button>`).join('')}
      </div>
      ${t.resTab === 'body' ? `<div class="resp-tools">
        ${modes.map(m => `<button class="seg-mini ${t.resMode === m ? 'active' : ''}" data-vact="resmode" data-m="${m}">${m}</button>`).join('')}
        <span class="tool-sep"></span>
        <div class="jp-box">${icon('search')}<input placeholder="JSONPath e.g. $.data[0].id" value="${esc(t.jsonPath)}" data-vinput="json-path" aria-label="JSONPath query"><button class="btn btn-tiny" data-vact="jp-apply">Go</button></div>
        <div class="jp-box">${icon('search')}<input placeholder="Find in body" value="${esc(t.resSearch || '')}" data-vinput="res-search" aria-label="Search body"><em class="dim">${t.resSearch ? searched + ' hits' : ''}</em></div>
        <button class="btn btn-icon" data-vact="copy-body" data-tip="Copy response" aria-label="Copy response">${icon('copy')}</button>
        <button class="btn btn-icon" data-vact="download-body" data-tip="Download" aria-label="Download response">${icon('download')}</button>
      </div>` : ''}
    </div>
    ${t.pathHit !== '' ? `<div class="jp-result">${pathRes === undefined ? RVH.badge('JSONPath: no match', 'error') : `<span class="dim">JSONPath result</span> <code class="mono jp-val">${esc(typeof pathRes === 'object' ? JSON.stringify(pathRes, null, 1) : String(pathRes)).slice(0, 300)}</code> <button class="btn btn-tiny" data-vact="jp-copy">Copy</button> <button class="linklike" data-vact="jp-clear">clear</button>`}</div>` : ''}
    ${t.resSearch && searched ? `<div class="search-note">${searched} occurrence${searched > 1 ? 's' : ''} of “${esc(t.resSearch)}” — matches highlighted in Raw mode.</div>` : ''}
    <div class="resp-body-host ${t.resTab === 'body' ? '' : 'hidden'}">
      ${t.resTab === 'body' ? bodyMain : ''}
      ${t.resTab === 'headers' ? `<div class="tbl-wrap"><table class="tbl kv-read"><tbody>${Object.entries(r.headers).map(([k, v]) => `<tr><td class="mono tok-key">${esc(k)}</td><td class="mono">${esc(v)}</td></tr>`).join('')}</tbody></table></div>` : ''}
      ${t.resTab === 'cookies' ? ((r.cookies || []).length ? `<div class="tbl-wrap">${RVH.table(['Name', 'Value', 'Domain', 'Expires', 'Flags'], r.cookies.map(c => [`<code class="mono">${esc(c.name)}</code>`, `<code class="mono">${esc(c.value)}</code>`, esc(c.domain), esc(c.expires), `${c.httpOnly ? RVH.badge('httpOnly', 'info') : ''} ${c.secure ? RVH.badge('secure', 'ok') : ''}`]))}</div><div class="jar-bar">${icon('shield')} Cookie jar persisted locally (<code>history.db</code>, FTS5-indexed, 0600) <button class="btn btn-sm" data-vact="clear-jar">Clear jar</button></div>` : RVH.emptyState('shield', 'No cookies set', 'The server did not return Set-Cookie headers.')) : ''}
    </div>`;
  }

  function respHtml(t) {
    const st = t.response.state;
    const head = `<div class="resp-head"><h3>Response</h3><span class="resp-sub">${esc(resolveVars(t.url).slice(0, 46))}</span></div>`;
    if (st === 'idle') {
      const seen = new Set();
      const recent = D.history.filter(h => h.status > 0 && h.status < 400 && !seen.has(h.url) && seen.add(h.url)).slice(0, 4);
      return `${head}<div class="resp-stage"><div class="idle-card">
        <div class="idle-glow">${icon('zap')}</div>
        <h4>Ready to send</h4>
        <p>Press <kbd>⌘↵</kbd> or hit Send — pre-request scripts run first, then the response lands here.</p>
        <button class="btn btn-primary" data-vact="send">${icon('play')} Send request</button>
        ${recent.length ? `<div class="idle-recent"><span>Recent</span>${recent.map(h => `<button class="idle-chip" data-vact="idle-open" data-id="${h.id}" data-tip="Open in editor">${methodChip(h.method)}<span class="ellip">${esc((h.url.split('?')[0].split('/').pop() || h.url).slice(0, 22))}</span></button>`).join('')}</div>` : ''}
      </div></div>`;
    }
    if (st === 'loading') return `${head}<div class="resp-stage resp-loading">
      <div class="load-row"><span class="spinner"></span><b>Sending ${esc(t.method)}…</b><span class="dim" data-role="elapsed">0 ms</span><button class="btn btn-sm btn-danger" data-vact="cancel">${icon('stop')} Cancel</button></div>
      ${skeletonLines()}</div>`;
    if (st === 'cancelled') return `${head}<div class="resp-stage">${RVH.emptyState('stop', 'Request cancelled', 'No data was received.', `<button class="btn btn-primary" data-vact="send">${icon('refresh')} Send again</button>`)}</div>`;
    if (st === 'timeout') return `${head}<div class="resp-stage">${RVH.emptyState('clock', 'ETIMEDOUT — request timed out', esc(t.response.message) + '<br>The socket was aborted locally after 30 000 ms.', `<button class="btn" data-vact="settings-timeout">${icon('sliders')} Adjust timeout</button> <button class="btn btn-primary" data-vact="send">${icon('refresh')} Retry</button>`)}</div>`;
    return `${head}<div class="resp-done ${t.response.status >= 400 ? 'is-error' : ''}">${respDoneHtml(t)}</div>`;
  }

  A.regView({
    id: 'rest', title: 'REST Client', icon: 'rest', group: 'Protocols',
    render() {
      const t = A.activeTab();
      if (!t) return RVH.emptyState('folder', 'No request open', 'Pick one from Collections, or create a fresh draft.', `<button class="btn btn-primary" data-action="new-request">${icon('plus')} New request</button>`);
      return restLayout(`<div class="pane-inner">${builderHtml(t)}</div>`, `<div class="pane-inner" id="resp-host">${respHtml(t)}</div>`);
    },
    after(host) {
      bindViewActs(host);
      startElapsedTicker();
    }
  });

  let ELAPSED = null;
  function startElapsedTicker() {
    clearInterval(ELAPSED);
    ELAPSED = setInterval(() => {
      const t = A.activeTab();
      if (!t || t.response.state !== 'loading') { clearInterval(ELAPSED); return; }
      const el = document.querySelector('[data-role="elapsed"]');
      if (el) el.textContent = (Date.now() - t.response.startedAt) + ' ms';
    }, 90);
  }

  A.regSidebar('rest', null);
  function bindRespHost() {
    const t = A.activeTab(); const host = document.getElementById('resp-host');
    if (!host || !t) return;
    host.innerHTML = respHtml(t);
    bindViewActs(host);
    if (window.bindViewActs2) window.bindViewActs2(host);
  }
  A.on('renderResponse', bindRespHost);
  A.on('live-input', (name) => { if (name === 'tree-query') { /* handled */ } });

  const T = A.markDirty;
  VACT.send = () => { const t = A.activeTab(); t && A.sendRequest(t); };
  VACT.cancel = () => { const t = A.activeTab(); t && A.cancelRequest(t); };
  VACT.save = () => A.saveActive();
  VACT.revert = () => { const t = A.activeTab(); const n = A.walkTree(S.tree).find(x => x.id === t.nodeId); if (n) { t.name = n.name; t.dirty = false; A.refresh(); toastOnce('Reverted to last saved version'); } };
  VACT.btab = (el) => { const t = A.activeTab(); t.builderTab = el.dataset.tab; A.renderMain(); };
  VACT.rtab = (el) => { const t = A.activeTab(); t.resTab = el.dataset.tab; bindRespHost(); };
  VACT.resmode = (el) => { const t = A.activeTab(); t.resMode = el.dataset.m; bindRespHost(); };
  VACT.jpApply = null;
  VACT['jp-apply'] = () => { const t = A.activeTab(); t.pathHit = '1'; bindRespHost(); };
  VACT['jp-clear'] = () => { const t = A.activeTab(); t.pathHit = ''; t.jsonPath = ''; bindRespHost(); };
  VACT['jp-copy'] = () => { navigator.clipboard?.writeText(document.querySelector('.jp-val')?.textContent || ''); toastOnce('Copied JSONPath result'); };
  VACT['copy-body'] = () => { const t = A.activeTab(); navigator.clipboard?.writeText(t.response.body || ''); toastOnce('Response body copied'); };
  VACT['download-body'] = () => { const t = A.activeTab(); downloadFile(fileSafeName(t) + '.json', t.response.largeRows ? JSON.stringify(A.largeBodyHtml().rows) : (t.response.body || ''), 'application/json'); toastOnce('Response downloaded'); };
  VACT['load-large'] = () => { ui.largeLoaded = true; const h = document.getElementById('resp-host'); const t = A.activeTab(); if (h && t) { h.innerHTML = respHtml(t); bindViewActs(h); } toastOnce('Rendered virtualized preview'); };
  VACT['clear-jar'] = async () => { if (await A.confirmModal('Clear cookie jar', 'All stored cookies for this workspace will be removed from local storage.', 'Clear jar', true)) toastOnce('Cookie jar cleared'); };
  VACT['fmt-json'] = () => { const t = A.activeTab(); try { t.body.text = JSON.stringify(JSON.parse(t.body.text), null, 2); A.renderMain(); } catch (e) { A.toast('Cannot format — JSON is invalid', { type: 'error' }); } };
  VACT['pick-file'] = () => { const t = A.activeTab(); A.modal({ title: 'Choose file', body: `<div class="dropzone">${icon('download')}<p>Drop a file here or</p><button class="btn btn-primary">Browse…</button><p class="dim small">receipt_2026-08.pdf · 84 KB selected (demo)</p></div>`, footer: `<button class="btn" onclick="ReqlyApp.closeTopModal()">Cancel</button><button class="btn btn-primary" data-pick>Attach receipt_2026-08.pdf</button>` }).el.addEventListener('click', e => { if (e.target.closest('[data-pick]')) { if (t.body.type === 'binary') t.body.file = 'receipt_2026-08.pdf'; else (t.body.form = t.body.form || []).push({ key: 'file', value: 'receipt_2026-08.pdf', enabled: true }); A.closeTopModal(); A.renderMain(); T(); } }); };
  VACT['insert-snip'] = (el) => { const area = document.getElementById('script-' + el.dataset.phase); if (area) { const p = area.selectionStart; area.setRangeText('\n' + el.dataset.code, p, p, 'end'); area.dispatchEvent(new Event('input')); area.focus(); } };
  VACT['run-script'] = (el) => { const t = A.activeTab(); const phase = el.dataset.phase; A.log('sys', `sandbox: running ${phase} script for “${t.name}”`); setTimeout(() => A.log(phase === 'pre' ? 'log' : 'pass', phase === 'pre' ? 'env.set(traceHeader) ✓' : 'assertion stream finished — 0 failures'), 350); A.toast(`${phase} script executed cleanly`, { type: 'success' }); };
  VACT['settings-timeout'] = () => requestSettingsModal();
  VACT['oauth-fetch'] = () => { const st = document.querySelector('[data-role="tok-state"]'); if (st) { st.innerHTML = '<span class="spinner sm"></span> requesting…'; setTimeout(() => { st.innerHTML = '✓ access_token stored as secret <code>[SECRET]</code>'; A.log('sec', 'oauth2: acquired bearer token → env secret accessToken'); A.toast('OAuth2 token acquired and masked', { type: 'success' }); }, 900); } };
  VACT['req-menu'] = (el) => {
    const t = A.activeTab(); const r = el.getBoundingClientRect();
    A.menu(r.left - 120, r.bottom + 4, [
      { label: 'Duplicate request', icon: 'copy', fn: () => A.walkTree(S.tree).find(n => n.id === t.nodeId) && dupCurrent(t) },
      { label: 'Rename…', icon: 'edit', fn: () => A.formModal({ title: 'Rename request', submitLabel: 'Rename', fields: [{ name: 'n', label: 'Name', value: t.name }], onSubmit: v => { t.name = v.n; t.dirty = true; A.refreshTabsBar(); A.renderMain(); A.renderSidebar(); } }) },
      { label: 'Generate code…', icon: 'braces', fn: () => A.navigate('codegen') },
      { label: 'Request settings…', icon: 'sliders', fn: () => requestSettingsModal() },
      '-', { label: 'Copy as cURL', icon: 'terminal', fn: () => { navigator.clipboard?.writeText(curlOf(t)); toastOnce('cURL copied'); } },
      { label: 'Save as example response', icon: 'doc', fn: () => toastOnce('Example saved to request file') }
    ]);
  };
  function dupCurrent(t) { const n = A.walkTree(S.tree).find(x => x.id === t.nodeId); n && (function () { const copy = JSON.parse(JSON.stringify(n)); copy.id = A.uid(); copy.name += ' copy'; copy.saved = false; const parent = n._parent ? n._parent.children : S.tree; parent.splice(parent.indexOf(n) + 1, 0, copy); A.renderSidebar(); A.toast('Duplicated', { type: 'success' }); })(); }
  function curlOf(t) { const url = resolveVars(t.url); const hdrs = t.headers.filter(h => h.enabled !== false).map(h => `-H '${h.key}: ${resolveVars(h.value)}'`).join(' '); const body = t.body.text ? `--data '${t.body.text.replace(/\n\s*/g, '')}'` : ''; return `curl -X ${t.method} '${url}' ${hdrs} ${body}`.trim(); }
  function fileSafeName(t) { return t.name.toLowerCase().replace(/\W+/g, '-') || 'response'; }
  function downloadFile(name, content, mime) { const b = new Blob([content], { type: mime }); const u = URL.createObjectURL(b); const a = document.createElement('a'); a.href = u; a.download = name; a.click(); setTimeout(() => URL.revokeObjectURL(u), 2000); }
  A.dl = downloadFile;
  function toastOnce(msg) { A.toast(msg, { type: 'success' }); }
  function requestSettingsModal() {
    A.formModal({
      title: 'Request settings', submitLabel: 'Apply', wide: true,
      fields: [
        { name: 'timeout', label: 'Timeout (ms)', value: '30000' },
        { name: 'redirects', label: 'Follow redirects', type: 'select', options: ['follow (max 10)', 'off'] },
        { name: 'ssl', label: 'SSL verification', type: 'select', options: ['strict', 'relaxed (unsafe)'] },
        { name: 'retryCount', label: 'Retry count', value: '2', hint: 'Retries network errors + 429/502/503/504 honoring Retry-After.' },
        { name: 'retryDelay', label: 'Retry delay (ms)', value: '400' },
        { name: 'strategy', label: 'Backoff strategy', type: 'select', options: ['fixed', 'exponential'] }
      ],
      onSubmit: v => { A.toast(`Settings saved — retries: ${v.retryCount}× ${v.strategy}`, { type: 'success' }); }
    });
  }

  VINPUT.method = (el) => { const t = A.activeTab(); t.method = el.value; t.dirty = true; A.refresh(); };
  VINPUT.url = (el) => { const t = A.activeTab(); t.url = el.value; t.dirty = true; A.refreshTabsBar(); const undef = undefinedVars(el.value); const b = document.querySelector('.banner-warn'); if (!undef.length && b && b.textContent.includes('Unresolved')) b.remove(); };
  VINPUT['auth-type'] = (el) => { const t = A.activeTab(); t.auth = { type: el.value }; t.dirty = true; A.renderMain(); };
  VINPUT['raw-lang'] = (el) => { const t = A.activeTab(); t.body.lang = el.value; t.dirty = true; };
  VINPUT['json-path'] = (el) => { const t = A.activeTab(); t.jsonPath = el.value; };
  VINPUT['res-search'] = (el) => { const t = A.activeTab(); t.resSearch = el.value; bindRespHost(); };
  VACT.seg = (el) => {
    const t = A.activeTab(); const name = el.dataset.name, val = el.dataset.val;
    if (!t) return;
    if (name === 'body-kind') { t.body.type = val; if ((val === 'form' || val === 'multipart') && !t.body.form) t.body.form = []; if (val === 'json' && !t.body.text) t.body.text = '{\n  \n}'; }
    if (name.startsWith('gql-')) ui.gql[name] = val;
    t.dirty = true; A.renderMain();
  };

  const GQL_SAMPLE = `query CustomerOrders($id: ID!) {\n  customer(id: $id) {\n    id\n    email\n    orders(first: 5) {\n      id\n      total\n      payments { id status }\n    }\n  }\n}`;
  ui.gql = ui.gql || { endpoint: '{{graphqlUrl}}', query: GQL_SAMPLE, vars: '{\n  "id": "cus_NffrFe"\n}', resp: null, loading: false, tab: 'query', schemaType: 'Customer' };

  function gqlOps(q) { return [...q.matchAll(/(query|mutation|subscription)\s+(\w+)/g)].map(m => m[2]); }

  A.regView({
    id: 'graphql', title: 'GraphQL', icon: 'graphql', group: 'Protocols',
    render() {
      const g = ui.gql;
      const ops = gqlOps(g.query);
      const types = D.graphqlSchema.types;
      const selType = types.find(t => t.name === g.schemaType) || types[0];
      return `<div class="gql-layout">
        <div class="gql-main">
          <div class="urlbar">
            <span class="chip chip-method m-info">GQL</span>
            <input class="url-input mono" value="${esc(g.endpoint)}" data-vinput="gql-endpoint" aria-label="GraphQL endpoint">
            <select data-vinput="gql-op" aria-label="Operation">${['<auto>', ...ops].map(o => `<option ${g.op === o ? 'selected' : ''}>${esc(o)}</option>`).join('')}</select>
            <button class="btn btn-icon" data-vact="gql-introspect" data-tip="Refresh introspection" aria-label="Introspect">${icon('refresh')}</button>
            ${g.loading ? `<button class="btn btn-danger" data-vact="gql-cancel">${icon('stop')} Cancel</button>` : `<button class="btn btn-primary" data-vact="gql-run">${icon('play')} Run</button>`}
          </div>
          <div class="gtabs">
            ${['query', 'variables', 'headers'].map(x => `<button class="btab ${g.tab === x ? 'active' : ''}" data-vact="seg" data-name="gql-tab" data-val="${x}">${x[0].toUpperCase() + x.slice(1)}</button>`).join('')}
            <span class="st-flex"></span><button class="btn btn-sm" data-vact="gql-prettify">Prettify</button>
          </div>
          ${g.tab === 'query' ? RVH.codeEditor('gql-query', g.query, { rows: 14 }) : ''}
          ${g.tab === 'variables' ? RVH.codeEditor('gql-vars', g.vars, { rows: 10, placeholder: '{ "id": "cus_x" }' }) : ''}
          ${g.tab === 'headers' ? kvEditor('headers', [{ key: 'Authorization', value: 'Bearer {{accessToken}}', enabled: true }], false) : ''}
          <div class="gql-resp">
            ${g.loading ? `<div class="load-row"><span class="spinner"></span> Executing <b>${esc(g.op === '<auto>' ? ops[0] || 'anonymous' : g.op)}</b>…</div>` : ''}
            ${!g.loading && !g.resp ? RVH.emptyState('graphql', 'No result yet', 'Run the operation to see data here.') : ''}
            ${!g.loading && g.resp ? (g.resp.errors ? `<div class="gql-errors">${g.resp.errors.map(e => `<div class="banner banner-error">${icon('warn')}<div><b>${esc(e.message)}</b><pre class="mono small">${esc(JSON.stringify(e.path || []))}</pre></div></div>`).join('')}</div>` : `<pre class="codeview mono">${highlight(JSON.stringify(g.resp.data, null, 2), 'json')}</pre>`) : ''}
          </div>
        </div>
        <aside class="gql-schema">
          <div class="side-head">Schema browser<button class="btn btn-icon btn-xs" data-vact="gql-introspect" data-tip="Reload">${icon('refresh')}</button></div>
          <div class="side-search">${icon('search')}<input placeholder="Filter types…" data-vinput="gql-schema-filter" value="${esc(g.schemaFilter || '')}"></div>
          <div class="schema-types">${types.filter(t => !g.schemaFilter || t.name.toLowerCase().includes(g.schemaFilter.toLowerCase())).map(t => `<button class="schema-type ${selType.name === t.name ? 'active' : ''}" data-vact="gql-type" data-n="${t.name}">${icon(t.kind === 'enum' ? 'braces' : 'doc')} ${t.name} <em>${t.kind}</em></button>`).join('')}</div>
          <div class="schema-fields">
            <h6>${selType.name} <em>${selType.kind}</em></h6>
            ${selType.fields.map(f => `<div class="schema-field"><code>${f.name}${f.args ? '(' + f.args.map(a => a.name + ': ' + a.type).join(', ') + ')' : ''}</code><b>: ${f.type}</b>${f.desc ? `<p>${esc(f.desc)}</p>` : ''}</div>`).join('')}
          </div>
        </aside>
      </div>`;
    },
    after(host) {
      bindViewActs(host);
      const q = document.getElementById('gql-query');
      q && q.addEventListener('input', () => { ui.gql.query = q.value; });
      const v = document.getElementById('gql-vars');
      v && v.addEventListener('input', () => { ui.gql.vars = v.value; });
    }
  });
  VINPUT['gql-endpoint'] = (el) => ui.gql.endpoint = el.value;
  VINPUT['gql-op'] = (el) => ui.gql.op = el.value;
  VINPUT['gql-schema-filter'] = (el) => { ui.gql.schemaFilter = el.value; const list = document.querySelector('.schema-types'); if (list) list.innerHTML = D.graphqlSchema.types.filter(t => !el.value || t.name.toLowerCase().includes(el.value.toLowerCase())).map(t => `<button class="schema-type ${ui.gql.schemaType === t.name ? 'active' : ''}" data-vact="gql-type" data-n="${t.name}">${icon(t.kind === 'enum' ? 'braces' : 'doc')} ${t.name} <em>${t.kind}</em></button>`).join(''); };
  VACT['gql-type'] = (el) => { ui.gql.schemaType = el.dataset.n; A.renderMain(); };
  VACT['gql-prettify'] = () => { try { const parts = ui.gql.query.split(/\s(?=(?:query|mutation|subscription))/); ui.gql.query = parts.join('\n\n'); A.renderMain(); } catch (e) {} };
  VACT['gql-introspect'] = (el) => { const b = el.closest('.gql-layout').querySelector('[data-vact="gql-introspect"] svg'); b && (b.style.animation = 'spin 1s linear 2'); A.toast('Introspection __schema refreshed (214 types)', { type: 'success' }); };
  VACT['gql-cancel'] = () => { ui.gql.loading = false; A.renderMain(); A.toast('GraphQL execution cancelled', { type: 'warn' }); };
  VACT['gql-run'] = () => {
    ui.gql.loading = true; A.renderMain();
    setTimeout(() => {
      ui.gql.loading = false;
      const bad = /boomboom|__crash/.test(ui.gql.query);
      ui.gql.resp = bad
        ? { errors: [{ message: 'Cannot query field “nope” on type “Customer”.' }] }
        : { data: { customer: { id: 'cus_NffrFe', email: 'nina@ferrite.io', orders: [{ id: 'ord_8841', total: 4250, payments: [{ id: 'pay_1Q84xa', status: 'SUCCEEDED' }] }, { id: 'ord_8790', total: 1200, payments: [{ id: 'pay_1Q84td', status: 'PENDING' }] }] } } };
      A.renderMain();
      A.log('sys', `graphql ${gqlOps(ui.gql.query)[0] || 'anon'} → ${bad ? 'errors' : 'OK'}`);
    }, 750);
  };

  const GRPC_MSG_DEFAULTS = {
    CreatePayment: '{\n  "amount": { "value": 4250, "currency": "USD" },\n  "customer_id": "cus_NffrFe"\n}',
    GetPayment: '{\n  "payment_id": "pay_1Q84xa"\n}',
    StreamPaymentEvents: '{\n  "filter": { "statuses": ["SUCCEEDED"] }\n}',
    UploadReceipts: '{\n  "chunk_b64": "JVBERi0xLjQK…" ,\n  "seq": 1\n}',
    ChatWithAgent: '{\n  "text": "why was pay_1Q84xa refunded?",\n  "session_id": "sess_01"\n}',
    GetCustomer: '{\n  "customer_id": "cus_NffrFe"\n}',
    ListCustomers: '{\n  "page_size": 25\n}'
  };
  ui.grpc = ui.grpc || { server: 'api.paycore.dev:443', tls: true, svc: 0, m: 0, meta: [], frames: [], state: 'idle', trailers: null, timer: null };

  A.regView({
    id: 'grpc', title: 'gRPC', icon: 'bolt', group: 'Protocols',
    render() {
      const g = ui.grpc;
      const svc = D.grpcServices[g.svc];
      const mth = svc.methods[g.m];
      return `<div class="grpc-layout">
        <div class="grpc-main card pad">
          <div class="urlbar">
            <input class="url-input mono" value="${esc(g.server)}" data-vinput="grpc-server" aria-label="Server address">
            ${RVH.segBtns('grpc-tls', ['TLS', 'plaintext'], g.tls ? 'TLS' : 'plaintext')}
            <button class="btn" data-vact="grpc-reflect" data-tip="Server reflection / proto import">${icon('refresh')} Reflect</button>
            ${g.state === 'streaming' ? `<button class="btn btn-danger" data-vact="grpc-cancel">${icon('stop')} Stop stream</button>` : ''}
          </div>
          <div class="grpc-selects">
            <label class="fld"><span>Service</span><select data-vinput="grpc-svc">${D.grpcServices.map((s, i) => `<option value="${i}" ${g.svc === i ? 'selected' : ''}>${s.name}</option>`).join('')}</select></label>
            <label class="fld"><span>Method</span><select data-vinput="grpc-method">${svc.methods.map((m, i) => `<option value="${i}" ${g.m === i ? 'selected' : ''}>${m.name}</option>`).join('')}</select></label>
            <div class="stream-badges">
              ${RVH.badge(mth.streaming + ' stream', mth.streaming === 'unary' ? 'ok' : 'warn')}
              ${RVH.badge(mth.input + ' → ' + mth.output, 'muted')}
            </div>
          </div>
          <div class="grpc-cols">
            <div>
              <h6>Message <span class="mono dim">(${mth.input})</span></h6>
              ${RVH.codeEditor('grpc-msg', GRPC_MSG_DEFAULTS[mth.name], { rows: 8 })}
              <div class="editor-tools">
                ${mth.streaming === 'unary' || mth.streaming === 'server' ? `<button class="btn btn-primary" data-vact="grpc-invoke">${icon('play')} Invoke</button>` : `<button class="btn" data-vact="grpc-send-frame">${icon('export')} Send frame</button><button class="btn btn-primary" data-vact="grpc-half-close">${icon('check')} Half-close</button>`}
              </div>
            </div>
            <div>
              <h6>Metadata</h6>
              ${kvEditor('grpc-meta', g.meta, false)}
              <div class="editor-tools"><span class="dim">deadline: 15s · compression: identity</span></div>
            </div>
          </div>
          <div class="grpc-resp">
            <h6>Responses ${g.trailers ? RVH.badge('trailers: ' + g.trailers.code, g.trailers.code === 'OK' ? 'ok' : 'error') : ''}</h6>
            ${g.frames.length === 0 ? RVH.emptyState('bolt', 'No messages yet', 'Invoke a unary call or open a stream.') : `<div class="frames">${g.frames.map(f => `<div class="frame ${f.dir}"><span class="frame-dir">${f.dir === 'in' ? '←' : '→'} ${f.dir === 'in' ? mth.output : mth.input}</span><pre class="mono">${highlight(JSON.stringify(f.msg, null, 1), 'json')}</pre><em>${f.t}</em></div>`).join('')}</div>`}
          </div>
        </div>
        <aside class="insp-block-col">
          <div class="insp-block"><h5>PROTO</h5><pre class="mono small proto-snippet">syntax = "proto3";\npaycore.v1;\n\nservice PaymentService {\n${svc.methods.map(m => `  rpc ${m.name}(${m.input}) returns (${m.streaming === 'unary' ? m.output : 'stream ' + m.output});`).join('\n')}\n}</pre></div>
          <div class="insp-block"><h5>CHANNEL</h5><div class="insp-kv"><span>address</span><b class="mono">${esc(g.server)}</b><span>security</span><b>${g.tls ? 'TLS 1.3' : 'plaintext'}</b><span>user-agent</span><b>reqly-grpc/1.4</b></div></div>
        </aside>
      </div>`;
    },
    after(host) { bindViewActs(host); }
  });
  VINPUT['grpc-server'] = (el) => ui.grpc.server = el.value;
  VINPUT['grpc-svc'] = (el) => { ui.grpc.svc = +el.value; ui.grpc.m = 0; ui.grpc.frames = []; ui.grpc.trailers = null; A.renderMain(); };
  VINPUT['grpc-method'] = (el) => { ui.grpc.m = +el.value; ui.grpc.frames = []; ui.grpc.trailers = null; A.renderMain(); };
  VACT.seg = VACT.seg;
  const origSeg = VACT.seg.bind({});
  VACT.seg = (el) => { if (el.dataset.name === 'grpc-tls') { ui.grpc.tls = el.dataset.val === 'TLS'; A.renderMain(); return; } origSeg(el); };
  VACT['grpc-reflect'] = () => A.toast('Reflection OK — 2 services, 7 methods discovered', { type: 'success' });
  VACT['grpc-cancel'] = () => { clearTimeout(ui.grpc.timer); ui.grpc.state = 'idle'; ui.grpc.trailers = { code: 'CANCELLED' }; A.renderMain(); };
  VACT['grpc-invoke'] = () => {
    const g = ui.grpc; const mth = D.grpcServices[g.svc].methods[g.m];
    g.frames = []; g.trailers = null;
    if (mth.streaming === 'unary') {
      g.frames.push({ dir: 'out', msg: { _sent: true }, t: fmtClock(new Date().toISOString()) });
      A.renderMain();
      setTimeout(() => {
        g.frames.push({ dir: 'in', msg: mth.name === 'GetPayment' ? { payment: { id: 'pay_1Q84xa', amount: { value: 4250 }, status: 'SUCCEEDED' } } : { payment: { id: 'pay_new_g1', status: 'PROCESSING' } }, t: fmtClock(new Date().toISOString()) });
        g.trailers = { code: 'OK', detail: 'grpc-status: 0 · grpc-message:' };
        A.renderMain(); A.log('sys', `grpc ${mth.name} → OK in 41ms`);
      }, 650);
    } else {
      g.state = 'streaming';
      let n = 0;
      const push = () => {
        if (g.state !== 'streaming') return;
        n++;
        g.frames.push({ dir: 'in', msg: { event: 'PAYMENT_SUCCEEDED', seq: n, payment_id: 'pay_' + (7000 + n * 13) }, t: fmtClock(new Date().toISOString()) });
        if (n >= 5) { g.trailers = { code: 'OK' }; g.state = 'idle'; }
        else g.timer = setTimeout(push, 620);
        A.renderMain(); updateGrpcRespOnly();
      };
      push();
    }
    A.renderMain();
  };
  function updateGrpcRespOnly() { A.renderMain(); }
  VACT['grpc-send-frame'] = () => {
    const g = ui.grpc;
    g.frames.push({ dir: 'out', msg: { chunk_b64: 'JVBERi0xLjQK' + Math.random().toString(36).slice(2, 8), seq: g.frames.filter(f => f.dir === 'out').length + 1 }, t: fmtClock(new Date().toISOString()) });
    A.renderMain();
  };
  VACT['grpc-half-close'] = () => {
    const g = ui.grpc;
    g.frames.push({ dir: 'in', msg: { uploaded: g.frames.filter(f => f.dir === 'out').length * 128, checksum_ok: true }, t: fmtClock(new Date().toISOString()) });
    g.trailers = { code: 'OK' };
    A.renderMain();
  };
})();
