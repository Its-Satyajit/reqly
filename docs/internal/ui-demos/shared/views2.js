// oxlint-disable anti-slop/no-runtime-typeof
(function () {
  const A = window.ReqlyApp;
  const { D, S, esc, icon, methodChip, statusChip, ms, bytes, ago, fmtDate, fmtClock, highlight, diffLines, kvEditor, resolveVars } = A;
  const { undefinedVars } = A;
  const RVH = window.RVH;
  const ui = S.ui;
  const bind = (host) => document.dispatchEvent(new CustomEvent('noop'));
  const T = () => A.markDirty();
  const toastOk = (m) => A.toast(m, { type: 'success' });

  ui.ws = ui.ws || { url: 'wss://stream.paycore.dev/v1/events', proto: 'payments.v1', state: 'closed', msgs: [], timer: null };
  A.regView({
    id: 'websocket', title: 'WebSocket', icon: 'plug', group: 'Protocols',
    render() {
      const w = ui.ws;
      return `<div class="proto-page">
        <div class="urlbar">
          <span class="conn-badge conn-${w.state}">${icon(w.state === 'open' ? 'check' : w.state === 'connecting' ? 'clock' : 'x')} ${w.state}</span>
          <input class="url-input mono" value="${esc(w.url)}" data-vinput="ws-url" aria-label="WebSocket URL">
          <input class="url-input mono short" value="${esc(w.proto)}" data-vinput="ws-proto" placeholder="subprotocol" aria-label="Subprotocol">
          ${w.state === 'closed' ? `<button class="btn btn-primary" data-vact="ws-connect">${icon('plug')} Connect</button>` : `<button class="btn btn-danger" data-vact="ws-disconnect">${icon('stop')} Disconnect</button>`}
        </div>
        <div class="ws-grid">
          <div class="card pad">
            <h4>Message composer</h4>
            ${RVH.codeEditor('ws-msg', '{\n  "type": "payment.succeeded",\n  "echo": true\n}', { rows: 6 })}
            <div class="editor-tools">
              <button class="btn btn-primary btn-sm" data-vact="ws-send" ${w.state !== 'open' ? 'disabled' : ''}>${icon('export')} Send</button>
              <button class="btn btn-sm" data-vact="ws-ping" ${w.state !== 'open' ? 'disabled' : ''}>Ping</button>
              <button class="btn btn-sm" data-vact="ws-invalid">Send malformed JSON</button>
            </div>
            ${w.state !== 'open' && w.state !== 'connecting' ? RVH.banner('warn', 'Socket is closed — connect to enable sending.') : ''}
          </div>
          <div class="card pad ws-log-card">
            <div class="row-between"><h4>Message log <i class="btab-count">${w.msgs.length}</i></h4><button class="btn btn-sm" data-vact="ws-clear">${icon('trash')} Clear</button></div>
            <div class="ws-log">${w.msgs.length === 0 ? RVH.emptyState('plug', 'No traffic yet', 'Frames you send and events pushed by the server appear here.') : w.msgs.map(m => `<div class="frame ${m.dir === 'in' ? 'in' : m.dir === 'sys' ? 'sys' : 'out'}"><em>${esc(m.t)}</em><b>${m.dir === 'in' ? '← server' : m.dir === 'out' ? '→ client' : '· event'}</b><pre class="mono small">${highlight(m.text, 'json')}</pre></div>`).join('')}</div>
          </div>
        </div>
        <div class="card pad"><h4>Connection info</h4>
          <div class="insp-kv"><span>readyState</span><b>${w.state.toUpperCase()}</b><span>subprotocol</span><b class="mono">${esc(w.proto)}</b><span>extensions</span><b>permessage-deflate</b><span>ping interval</span><b>30s</b></div>
        </div>
      </div>`;
    },
    after(h) { bindViewActs2(h); }
  });
  function bindViewActs2(host) {
    host.querySelectorAll('[data-vact]').forEach(el => { if (el.dataset.bound) return; el.dataset.bound = '1'; el.addEventListener('click', () => V2[el.dataset.vact] && V2[el.dataset.vact](el)); });
    host.querySelectorAll('[data-vinput]').forEach(el => { if (el.dataset.bound) return; el.dataset.bound = '1'; el.addEventListener(el.tagName === 'SELECT' ? 'change' : 'input', () => VI2[el.dataset.vinput] && VI2[el.dataset.vinput](el)); });
  }
  window.bindViewActs2 = bindViewActs2;
  const V2 = {}, VI2 = {};
  VI2['ws-url'] = el => ui.ws.url = el.value;
  VI2['ws-proto'] = el => ui.ws.proto = el.value;
  V2['ws-connect'] = () => {
    const w = ui.ws; if (/error/.test(w.url)) { w.state = 'error'; A.renderMain(); A.toast('Handshake failed — 401 during upgrade', { type: 'error' }); return; }
    w.state = 'connecting'; A.renderMain();
    setTimeout(() => {
      w.state = 'open';
      w.msgs.push({ dir: 'sys', t: fmtClock(new Date().toISOString()), text: JSON.stringify({ event: 'connected', subprotocol: w.proto }) });
      clearInterval(w.timer);
      w.timer = setInterval(() => {
        const evs = [{ type: 'payment.succeeded', id: 'pay_' + Math.random().toString(36).slice(2, 9), amount: Math.round(Math.random() * 9000) }, { type: 'heartbeat', ts: Date.now() }];
        const e = evs[Math.floor(Math.random() * 2)];
        w.msgs.push({ dir: 'in', t: fmtClock(new Date().toISOString()), text: JSON.stringify(e) });
        A.renderMain();
      }, 2600);
      A.renderMain(); A.toast('WebSocket connected', { type: 'success' });
    }, 800);
  };
  V2['ws-disconnect'] = () => { clearInterval(ui.ws.timer); ui.ws.state = 'closed'; ui.ws.msgs.push({ dir: 'sys', t: fmtClock(new Date().toISOString()), text: '{"event":"closed","code":1000}' }); A.renderMain(); };
  V2['ws-send'] = () => { const txt = document.getElementById('ws-msg').value || '{}'; ui.ws.msgs.push({ dir: 'out', t: fmtClock(new Date().toISOString()), text: txt }); setTimeout(() => { ui.ws.msgs.push({ dir: 'in', t: fmtClock(new Date().toISOString()), text: JSON.stringify({ ack: true, echoed: JSON.parse(txt) }) }); A.renderMain(); }, 420); A.renderMain(); };
  V2['ws-ping'] = () => { ui.ws.msgs.push({ dir: 'out', t: fmtClock(new Date().toISOString()), text: '"PING"' }); setTimeout(() => { ui.ws.msgs.push({ dir: 'in', t: fmtClock(new Date().toISOString()), text: '"PONG" rtt=23ms' }); A.renderMain(); }, 200); A.renderMain(); };
  V2['ws-invalid'] = () => { ui.ws.msgs.push({ dir: 'out', t: fmtClock(new Date().toISOString()), text: '{broken json!!' }); ui.ws.msgs.push({ dir: 'in', t: fmtClock(new Date().toISOString()), text: JSON.stringify({ error: 'parse error', close_code: 1003 }) }); A.renderMain(); A.toast('Server rejected malformed frame (1003 unsupported data)', { type: 'error' }); };
  V2['ws-clear'] = () => { ui.ws.msgs = []; A.renderMain(); };

  ui.sse = ui.sse || { url: 'https://api.paycore.dev/v1/events/stream', state: 'idle', events: [], paused: false, filter: '', autoscroll: true };
  A.regView({
    id: 'sse', title: 'SSE', icon: 'signal', group: 'Protocols',
    render() {
      const s = ui.sse;
      const types = [...new Set(s.events.map(e => e.type))];
      const shown = s.events.filter(e => !s.filter || e.type === s.filter);
      return `<div class="proto-page">
        <div class="urlbar">
          <span class="conn-badge conn-${s.state === 'open' ? 'open' : s.state === 'reconnecting' ? 'connecting' : s.state === 'error' ? 'error' : 'closed'}">${icon(s.state === 'open' ? 'signal' : 'x')} ${s.state}</span>
          <input class="url-input mono" value="${esc(s.url)}" data-vinput="sse-url" aria-label="SSE endpoint">
          ${s.state === 'open' || s.state === 'reconnecting' ? `<button class="btn btn-danger" data-vact="sse-stop">${icon('stop')} Stop</button>` : `<button class="btn btn-primary" data-vact="sse-start">${icon('play')} Connect</button>`}
          ${s.state === 'reconnecting' ? '<span class="spinner sm"></span>' : ''}
        </div>
        <div class="sse-toolbar card pad">
          <label class="fld-inline"><span>Filter event type</span><select data-vinput="sse-filter"><option value="">all (${s.events.length})</option>${types.map(t => `<option ${s.filter === t ? 'selected' : ''}>${esc(t)}</option>`).join('')}</select></label>
          <button class="btn btn-sm ${s.paused ? 'btn-primary' : ''}" data-vact="sse-pause">${s.paused ? 'Resume stream' : 'Pause stream'}</button>
          <button class="btn btn-sm" data-vact="sse-clear">${icon('trash')} Clear</button>
          <label class="check"><input type="checkbox" ${s.autoscroll ? 'checked' : ''} data-vact-check="sse-scroll"> auto-scroll</label>
          <span class="st-flex"></span><span class="dim mono small">retry: 3000ms · last-event-id: ${s.events.filter(e => e.id).length || '—'}</span>
        </div>
        <div class="card pad sse-stream">
          ${shown.length === 0 ? (s.state === 'open' ? '<div class="load-row"><span class="spinner"></span> waiting for first event…</div>' : RVH.emptyState('signal', 'Stream is idle', 'Connect to an endpoint to watch server-sent events arrive in real time.')) :
          shown.map(e => `<div class="sse-event"><span class="chip chip-method m-info">${esc(e.type)}</span><code class="mono dim">${e.id || 'no-id'}</code><em>${e.t}</em><pre class="mono small">${esc(e.data)}</pre></div>`).join('')}
        </div>
      </div>`;
    },
    after(h) { bindViewActs2(h); }
  });
  VI2['sse-url'] = el => ui.sse.url = el.value;
  VI2['sse-filter'] = el => { ui.sse.filter = el.value; A.renderMain(); };
  V2['sse-start'] = () => {
    const s = ui.sse;
    if (/error/.test(s.url)) { s.state = 'error'; A.renderMain(); A.toast('Connection dropped — retrying in 3s', { type: 'error' }); return; }
    s.state = 'open'; s.paused = false; A.renderMain();
    clearInterval(s.timer);
    let n = 0;
    s.timer = setInterval(() => {
      if (s.paused) return;
      n++;
      const kinds = [['payment.updated', JSON.stringify({ id: 'pay_1Q8' + (400 + n), status: n % 3 ? 'processing' : 'succeeded' })], ['heartbeat', JSON.stringify({ ts: Date.now() })], ['invoice.finalized', JSON.stringify({ invoice: 'in_00' + n, total: n * 111 })]];
      const k = kinds[n % 3];
      s.events.push({ id: 'evt_' + String(n).padStart(3, '0'), type: k[0], data: k[1], t: fmtClock(new Date().toISOString()) });
      if (s.events.length > 60) s.events.shift();
      A.renderMain();
    }, 1100);
  };
  V2['sse-stop'] = () => { clearInterval(ui.sse.timer); ui.sse.state = 'idle'; A.renderMain(); A.toast('Event stream closed', { type: 'info' }); };
  V2['sse-pause'] = () => { ui.sse.paused = !ui.sse.paused; A.toast(ui.sse.paused ? 'Stream paused (buffering locally)' : 'Stream resumed', { type: 'info' }); A.renderMain(); };
  V2['sse-clear'] = () => { ui.sse.events = []; A.renderMain(); };

  ui.soap = ui.soap || { url: 'https://api.paycore.dev/soap/v1/Payments?wsdl', action: 'CreatePayment', envelope: `<?xml version="1.0"?>\n<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope"\n               xmlns:pay="urn:paycore:v1">\n  <soap:Header>\n    <wsse:Security xmlns:wsse="…/oasis-200401-wss-wssecurity-secext-1.0.xsd">\n      <wsse:UsernameToken>\n        <wsse:Username>svc-integrations</wsse:Username>\n      </wsse:UsernameToken>\n    </wsse:Security>\n  </soap:Header>\n  <soap:Body>\n    <pay:CreatePaymentRequest>\n      <amount>4250</amount>\n      <currency>USD</currency>\n      <customerId>cus_NffrFe</customerId>\n    </pay:CreatePaymentRequest>\n  </soap:Body>\n</soap:Envelope>`, resp: null, loading: false, fault: false };
  A.regView({
    id: 'soap', title: 'SOAP', icon: 'doc', group: 'Protocols',
    render() {
      const p = ui.soap;
      return `<div class="proto-page soap-page">
        <div class="urlbar">
          <span class="chip chip-method m-warn">SOAP 1.2</span>
          <input class="url-input mono" value="${esc(p.url)}" data-vinput="soap-url" aria-label="Endpoint WSDL URL">
          <button class="btn" data-vact="soap-import">${icon('import')} Import WSDL</button>
        </div>
        <div class="urlbar">
          <label class="fld-inline wide"><span>SOAPAction header</span><input class="mono" value="${esc(p.action)}" data-vinput="soap-action"></label>
          ${p.loading ? '<button class="btn btn-danger" data-vact="soap-cancel">Cancel</button>' : `<button class="btn btn-primary" data-vact="soap-send">${icon('play')} Send</button>`}
          <button class="btn" data-vact="soap-format">${icon('braces')} Pretty-print XML</button>
        </div>
        <div class="soap-cols">
          <div class="card pad"><h4>Request envelope</h4>${RVH.codeEditor('soap-env', p.envelope, { rows: 16 })}<div class="editor-tools"><span class="dim">WS-Security UsernameToken preset attached</span></div></div>
          <div class="card pad"><h4>Response ${p.fault ? RVH.badge('SOAP Fault', 'error') : p.resp ? RVH.badge('200 OK', 'ok') : ''}</h4>
            ${p.loading ? '<div class="load-row"><span class="spinner"></span> POSTing envelope…</div>' : ''}
            ${!p.loading && !p.resp ? RVH.emptyState('doc', 'No response yet', 'Send the envelope to see the XML answer or a fault.') : ''}
            ${!p.loading && p.resp ? (p.fault
              ? `<div class="banner banner-error">${icon('warn')}<div><b>soap:Fault</b><pre class="codeview mono small">${highlight(p.resp, 'xml')}</pre></div></div>`
              : `<pre class="codeview mono">${highlight(p.resp, 'xml')}</pre>`) : ''}
          </div>
        </div>
      </div>`;
    },
    after(h) { bindViewActs2(h); }
  });
  VI2['soap-url'] = el => ui.soap.url = el.value;
  VI2['soap-action'] = el => ui.soap.action = el.value;
  V2['soap-import'] = () => A.formModal({ title: 'Import WSDL', submitLabel: 'Parse', fields: [{ name: 'u', label: 'WSDL URL or file path', value: ui.soap.url }], onSubmit: v => { A.toast('Parsed PayCore v1 — 12 operations bound', { type: 'success' }); } });
  V2['soap-cancel'] = () => { ui.soap.loading = false; A.renderMain(); };
  V2['soap-format'] = () => { try { const x = new DOMParser().parseFromString(ui.soap.envelope, 'text/xml'); const ser = new XMLSerializer().serializeToString(x); ui.soap.envelope = ser.replace(/></g, '>\n<'); A.renderMain(); } catch (e) { A.toast('XML parse failed', { type: 'error' }); } };
  V2['soap-send'] = () => {
    ui.soap.loading = true; A.renderMain();
    setTimeout(() => {
      ui.soap.loading = false;
      if (/boom/.test(ui.soap.envelope)) {
        ui.soap.fault = true;
        ui.soap.resp = `<soap:Fault>\n  <faultcode>soap:Sender</faultcode>\n  <faultstring>Invalid amount — minor units required</faultstring>\n  <detail><pay:error code="E1002"/></detail>\n</soap:Fault>`;
      } else {
        ui.soap.fault = false;
        ui.soap.resp = `<soap:Envelope xmlns:soap="…">\n  <soap:Body>\n    <pay:CreatePaymentResponse>\n      <paymentId>pay_1Q85ne</paymentId>\n      <status>PROCESSING</status>\n    </pay:CreatePaymentResponse>\n  </soap:Body>\n</soap:Envelope>`;
      }
      A.renderMain();
    }, 800);
  };

  A.regView({
    id: 'environments', title: 'Environments', icon: 'layers', group: 'Data & Quality',
    render() {
      const selId = ui.envSel || D.environments[0].id;
      const env = D.environments.find(e => e.id === selId);
      const tab = ui.envTab || 'vars';
      const scopeCards = [
        ['Global', 'global', D.globalScope.vars.length + ' vars', 'Every workspace, every request'],
        ['Workspace', 'workspace', D.workspaceScope.vars.length + ' vars', 'This repository only'],
        ['Collection', 'collection', D.collectionScope.vars.length + ' vars', 'Inherited by folders & requests']
      ];
      return `<div class="env-layout">
        <aside class="env-list">
          <div class="side-head">Environments<button class="btn btn-icon btn-xs" data-vact="env-new" data-tip="New environment" aria-label="New environment">${icon('plus')}</button></div>
          ${D.environments.map(e => `<button class="env-card env-border-${e.color} ${e.id === S.envId ? 'is-active' : ''}" data-vact="env-select" data-id="${e.id}">
            <span class="env-dot d-${e.color}"></span><div><b>${esc(e.name)}</b><small>${e.vars.length} vars · ${e.secrets.length} secrets</small></div>
            ${e.id === S.envId ? RVH.badge('active', 'primary') : ''}
          </button>`).join('')}
          <div class="side-head" style="margin-top:14px">Shared scopes</div>
          ${scopeCards.map(c => `<button class="env-card ${ui.scopeSel === c[1] ? 'is-active' : ''}" data-vact="scope-select" data-id="${c[1]}"><span class="env-dot d-info"></span><div><b>${c[0]}</b><small>${c[2]} · ${c[3]}</small></div></button>`).join('')}
          <div class="side-head" style="margin-top:14px">Process / .env <em>read-only</em></div>
          ${D.processEnv.map(v => `<div class="proc-row mono">${esc(v.key)}<span>${esc(v.value)}</span><em>${v.source}</em></div>`).join('')}
        </aside>
        <section class="env-editor">
          ${env ? `<div class="row-between wrap">
            <div class="row-gap"><h3>${icon('layers')} ${esc(env.name)}</h3>${RVH.badge(env.color, env.color)}<button class="btn btn-sm ${S.envId === env.id ? 'btn-primary' : ''}" data-vact="env-activate" data-id="${env.id}">${S.envId === env.id ? 'Active ✓' : 'Set active'}</button></div>
            <div class="row-gap">
              <button class="btn btn-sm" data-vact="env-dup">${icon('copy')} Duplicate</button>
              <button class="btn btn-sm" data-vact="env-rename">${icon('edit')} Rename</button>
              <button class="btn btn-sm btn-danger-ghost" data-vact="env-del">${icon('trash')} Delete</button>
            </div></div>
            <div class="btabs">
              ${[['vars', `Variables (${env.vars.length})`], ['secret', `Secrets (${env.secrets.length})`], ['diff', 'Diff vs…'], ['preview', 'Resolution preview']].map(x => `<button class="btab ${tab === x[0] ? 'active' : ''}" data-vact="env-tab" data-tab="${x[0]}">${x[1]}</button>`).join('')}
            </div>` : `<div class="row-between"><h3>${icon('layers')} ${(ui.scopeSel || 'global')[0].toUpperCase() + (ui.scopeSel || 'global').slice(1)} scope</h3></div>`}
          <div class="env-pane">
            ${(!env || tab === 'vars') ? `${env ? '' : ''}${kvEditor(env ? 'env-vars-' + env.id : ui.scopeSel || 'global', env ? env.vars : D[(ui.scopeSel || 'global') + 'Scope'].vars, false)}
              <div class="tag-strip"><h6>Dynamic tags</h6>${D.dynamicTags.map(t => `<button class="snip" data-vact="copy-tag" data-code="${esc(t)}"><code>${esc(t)}</code></button>`).join('')}<span class="dim small">click to copy</span></div>` : ''}
            ${env && tab === 'secret' ? `
              ${RVH.banner('warn', 'Secrets are masked everywhere — CLI output, logs and test reports print <code>[SECRET]</code>. Values encrypt at rest via the OS keychain.')}
              ${kvEditor('env-secret-' + env.id, env.secrets, true)}` : ''}
            ${env && tab === 'diff' ? (() => {
              const otherId = ui.envDiffWith || D.environments.find(e => e.id !== env.id).id;
              const other = D.environments.find(e => e.id === otherId);
              const keys = [...new Set([...env.vars.map(v => v.key), ...other.vars.map(v => v.key)])];
              return `<div class="row-gap"><label class="fld-inline"><span>Compare with</span><select data-vinput="env-diff-with">${D.environments.filter(e => e.id !== env.id).map(e => `<option value="${e.id}" ${e.id === otherId ? 'selected' : ''}>${esc(e.name)}</option>`).join('')}</select></label></div>
              ${RVH.table(['Key', env.name, other.name, 'Δ'], keys.map(k => {
                const a = env.vars.find(v => v.key === k)?.value, b = other.vars.find(v => v.key === k)?.value;
                const st = a === undefined ? ['added', 'ok'] : b === undefined ? ['removed', 'error'] : a === b ? ['same', 'muted'] : ['changed', 'warn'];
                return [`<code class="mono">${esc(k)}</code>`, `<span class="mono ellip">${esc(a ?? '—')}</span>`, `<span class="mono ellip">${esc(b ?? '—')}</span>`, RVH.badge(st[0], st[1])];
              }))}`;
            })() : ''}
            ${env && tab === 'preview' ? (() => {
              const tpl = '{{baseUrl}}/payments/{{paymentId}}?limit={{limit}}&trace={{missingVar}}';
              const undef = undefinedVars(tpl);
              return `<p class="sub">Template resolution walks Global → Workspace → Collection → Environment → Local → Secret.</p>
              <div class="resolve-box"><h6>Template</h6><pre class="mono">${esc(tpl)}</pre><h6>Resolved against “${esc(env.name)}”</h6><pre class="mono resolved">${esc(resolveVars(tpl))}</pre></div>
              ${undef.length ? RVH.banner('error', `Undefined: ${undef.map(u => '<code>{{' + u + '}}</code>').join(' ')}`) : RVH.banner('success', 'All variables resolve cleanly.')}
              <div class="chain-vis">${['Global', 'Workspace', 'Collection', env.name, 'Local', 'Secret'].map((c, i) => `<span class="chain-node">${c}</span>${i < 5 ? icon('chev-r') : ''}`).join('')}</div>`;
            })() : ''}
            ${!env && tab === 'vars' ? '' : ''}
          </div>
          ${!env ? `<p class="dim">Editing shared scope variables. Process environment (.env) values are read-only — OS wins over dotenv.</p>` : ''}
        </section>
      </div>`;
    },
    after(h) { bindViewActs2(h); }
  });
  V2['env-select'] = el => { ui.envSel = el.dataset.id; ui.scopeSel = null; A.renderMain(); };
  V2['scope-select'] = el => { ui.scopeSel = el.dataset.id; ui.envSel = null; ui.envTab = 'vars'; A.renderMain(); };
  V2['env-tab'] = el => { ui.envTab = el.dataset.tab; A.renderMain(); };
  V2['env-activate'] = el => { S.envId = S.envId === el.dataset.id ? null : el.dataset.id; A.refresh(); A.toast(S.envId ? 'Environment activated — {{vars}} now resolve' : 'Environment deactivated', { type: 'success' }); };
  V2['env-new'] = () => A.formModal({ title: 'New environment', submitLabel: 'Create', fields: [{ name: 'name', label: 'Name', value: '' }], onSubmit: v => { D.environments.push({ id: A.uid(), name: v.name || 'new-env', color: 'info', vars: [{ key: 'baseUrl', value: 'https://', enabled: true }], secrets: [] }); ui.envSel = D.environments[D.environments.length - 1].id; A.renderMain(); toastOk('Environment created — file written to environments/' + (v.name || 'new-env') + '.yaml'); } });
  V2['env-dup'] = () => { const e = D.environments.find(x => x.id === ui.envSel); const c = JSON.parse(JSON.stringify(e)); c.id = A.uid(); c.name += '-copy'; D.environments.push(c); ui.envSel = c.id; A.renderMain(); toastOk('Duplicated'); };
  V2['env-rename'] = () => { const e = D.environments.find(x => x.id === ui.envSel); A.formModal({ title: 'Rename environment', submitLabel: 'Rename', fields: [{ name: 'n', label: 'Name', value: e.name }], onSubmit: v => { e.name = v.n; A.refresh(); } }); };
  V2['env-del'] = async () => { const e = D.environments.find(x => x.id === ui.envSel); if (!await A.confirmModal('Delete environment', `Delete <b>${esc(e.name)}</b>? The YAML file will be staged for deletion in Git.`, 'Delete', true)) return; D.environments = D.environments.filter(x => x.id !== ui.envSel); if (S.envId === e.id) S.envId = D.environments[0]?.id || null; ui.envSel = D.environments[0]?.id; A.refresh(); toastOk('Environment deleted'); };
  VI2['env-diff-with'] = el => { ui.envDiffWith = el.value; A.renderMain(); };
  V2['copy-tag'] = el => { navigator.clipboard?.writeText(el.dataset.code); toastOk(el.dataset.code + ' copied'); };

  ui.hist = ui.hist || { q: '', method: '', status: '', range: 'all', sel: null };
  A.regView({
    id: 'history', title: 'History', icon: 'clock', group: 'Data & Quality',
    render() {
      const f = ui.hist;
      let rows = D.history.slice();
      if (f.q) { const q = f.q.toLowerCase(); rows = rows.filter(h => [h.method, h.url, h.resolved, String(h.status), h.env, h.requestId || ''].join(' ').toLowerCase().includes(q)); }
      if (f.method) rows = rows.filter(h => h.method === f.method);
      if (f.status) rows = rows.filter(h => f.status === 'timeout' ? !!h.timeout : Math.floor((h.status || 0) / 100) === +f.status);
      if (f.range !== 'all') { const lim = { '1h': 36e5, '24h': 864e5, '7d': 6048e5 }[f.range]; rows = rows.filter(h => Date.now() - new Date(h.ts) < lim); }
      const sel = rows.find(h => h.id === f.sel);
      return `<div class="hist-layout">
        <div class="hist-main card">
          <div class="hist-toolbar">
            <div class="side-search grow">${icon('search')}<input placeholder='Full-text search history ("payments 500 staging")…' value="${esc(f.q)}" data-vinput="hist-q" aria-label="Search history"></div>
            <select data-vinput="hist-method" aria-label="Method filter"><option value="">method</option>${['GET', 'POST', 'PUT', 'PATCH', 'DELETE'].map(m => `<option ${f.method === m ? 'selected' : ''}>${m}</option>`).join('')}</select>
            <select data-vinput="hist-status" aria-label="Status filter"><option value="">status</option><option value="2" ${f.status === '2' ? 'selected' : ''}>2xx</option><option value="3" ${f.status === '3' ? 'selected' : ''}>3xx</option><option value="4" ${f.status === '4' ? 'selected' : ''}>4xx</option><option value="5" ${f.status === '5' ? 'selected' : ''}>5xx</option><option value="timeout" ${f.status === 'timeout' ? 'selected' : ''}>timeouts</option></select>
            <select data-vinput="hist-range" aria-label="Date range"><option value="all" ${f.range === 'all' ? 'selected' : ''}>any time</option><option value="1h" ${f.range === '1h' ? 'selected' : ''}>last hour</option><option value="24h" ${f.range === '24h' ? 'selected' : ''}>last 24h</option><option value="7d" ${f.range === '7d' ? 'selected' : ''}>last 7 days</option></select>
            <button class="btn btn-sm" data-vact="hist-export">${icon('download')} HAR</button>
            <button class="btn btn-sm btn-danger-ghost" data-vact="hist-clear">${icon('trash')} Clear all</button>
          </div>
          ${rows.length === 0 ? (D.history.length === 0 ? RVH.emptyState('clock', 'No history yet', 'Requests you send are stored locally in SQLite (FTS5) — searchable forever, never uploaded.', `<button class="btn btn-primary" data-action="nav" data-view="rest">Send your first request</button>`) : RVH.emptyState('search', 'No matching entries', `Nothing matches the current filters. <button class="linklike" data-vact="hist-reset">Reset filters</button>`)) : ''}
          ${rows.length > 0 ? `<div class="hist-rows">${rows.map(h => `<div class="hist-row ${f.sel === h.id ? 'active' : ''}" data-vact="hist-open" data-id="${h.id}" role="button" tabindex="0">
            <span class="mono dim small">${fmtClock(h.ts)}</span>
            ${methodChip(h.method)}
            <span class="mono ellip hist-url">${esc(h.resolved)}</span>
            ${h.timeout ? RVH.badge('TIMEOUT', 'error') : statusChip(h.status)}
            <span class="dim">${ms(h.time)}</span>
            <span class="dim">${bytes(h.size)}</span>
            ${RVH.badge(h.env || '—', h.env === 'production' ? 'error' : 'muted')}
            <span class="dim small">${ago(h.ts)}</span>
          </div>`).join('')}</div>` : ''}
          <div class="hist-foot dim small">${rows.length} of ${D.history.length} entries · stored locally · FTS5 indexed</div>
        </div>
        ${sel ? `<aside class="hist-detail card pad">
          <div class="row-between"><h4>Entry detail</h4><button class="btn btn-icon" data-vact="hist-close" aria-label="Close detail">${icon('x')}</button></div>
          <div class="insp-kv big">
            <span>when</span><b>${fmtDate(sel.ts)}</b>
            <span>method</span><b>${sel.method}</b>
            <span>resolved URL</span><b class="mono ellip">${esc(sel.resolved)}</b>
            <span>template</span><b class="mono ellip">${esc(sel.url)}</b>
            <span>status</span><b>${sel.timeout ? 'TIMEOUT (aborted)' : sel.status}</b>
            <span>latency</span><b>${ms(sel.time)}</b>
            <span>size</span><b>${bytes(sel.size)}</b>
            <span>environment</span><b>${esc(sel.env || '—')}</b>
            <span>request</span><b>${sel.requestId ? esc(A.walkTree(S.tree).find(n => n.id === sel.requestId)?.name || sel.requestId) : '(ad-hoc)'}</b>
          </div>
          <div class="row-gap wrap">
            <button class="btn btn-primary" data-vact="hist-replay" data-id="${sel.id}">${icon('refresh')} Replay request</button>
            <button class="btn" data-vact="hist-edit" data-id="${sel.id}">${icon('edit')} Open in editor</button>
            <button class="btn" data-vact="hist-curl" data-id="${sel.id}">${icon('terminal')} Copy cURL</button>
            <button class="btn" data-vact="hist-diffpair" data-id="${sel.id}">${icon('compare')} Diff vs…</button>
            <button class="btn btn-danger-ghost" data-vact="hist-del" data-id="${sel.id}">${icon('trash')} Delete entry</button>
          </div>
          <details class="hist-curl"><summary>cURL preview</summary><pre class="mono small">${esc(`curl -X ${sel.method} '${sel.resolved}' -H 'Authorization: Bearer [SECRET]'`)}</pre></details>
        </aside>` : ''}
      </div>`;
    },
    after(h) { bindViewActs2(h); }
  });
  A.regSidebar('history', () => `<div class="side-block"><div class="side-head">Saved searches</div>
    ${[['failed payments', 'q=payments status=5'], ['slow calls', 'range=24h sort=time'], ['staging traffic', 'q=staging']].map(s => `<button class="side-link" data-vact="hist-saved" data-q="${s[1]}">${icon('search')} ${s[0]}</button>`).join('')}
  </div>`);
  VI2['hist-q'] = el => { ui.hist.q = el.value; refreshHist(); };
  VI2['hist-method'] = el => { ui.hist.method = el.value; refreshHist(); };
  VI2['hist-status'] = el => { ui.hist.status = el.value; refreshHist(); };
  VI2['hist-range'] = el => { ui.hist.range = el.value; refreshHist(); };
  function refreshHist() {
    const view = A.VIEWS.history; const host = A.root.querySelector('.view-host');
    if (host && A.S.view === 'history') { host.innerHTML = view.render(); view.after(host); }
  }
  V2['hist-open'] = el => { ui.hist.sel = el.dataset.id; refreshHist(); };
  V2['hist-close'] = () => { ui.hist.sel = null; refreshHist(); };
  V2['hist-reset'] = () => { ui.hist = { q: '', method: '', status: '', range: 'all', sel: null }; refreshHist(); };
  V2['hist-replay'] = el => { const h = D.history.find(x => x.id === el.dataset.id); const node = A.walkTree(S.tree).find(n => n.id === h.requestId); if (node) { A.findOrMakeTab(node.id); A.navigate('rest'); setTimeout(() => { const t = A.activeTab(); t && A.sendRequest(t); }, 150); } else { A.toast('Replayed ad-hoc request from history', { type: 'info' }); A.log('sys', `replay ${h.method} ${h.resolved}`); } };
  V2['hist-edit'] = el => { const h = D.history.find(x => x.id === el.dataset.id); const tab = A.newTabFromNode({ id: A.uid(), name: 'History replay — ' + h.method + ' ' + h.url.split('?')[0].split('/').pop(), method: h.method, url: h.url, saved: false }); A.navigate('rest'); };
  V2['hist-curl'] = el => { navigator.clipboard?.writeText(`curl -X ${A.D.history.find(x => x.id === el.dataset.id).method} '…'`); toastOk('cURL copied'); };
  V2['hist-del'] = async el => { if (await A.confirmModal('Delete history entry', 'Remove this single entry from local storage?', 'Delete', true)) { D.history = D.history.filter(x => x.id !== el.dataset.id); ui.hist.sel = null; refreshHist(); toastOk('Entry deleted'); } };
  V2['hist-clear'] = async () => { if (await A.confirmModal('Clear entire history', 'All local request history will be permanently removed from <code>history.db</code>. This cannot be undone.', 'Clear history', true)) { D.history.length = 0; ui.hist.sel = null; refreshHist(); A.toast('History cleared', { type: 'success' }); } };
  V2['hist-export'] = () => { A.dl('paycore-history.har', JSON.stringify({ log: { version: '1.2', creator: { name: 'Reqly' }, entries: D.history.map(h => ({ startedDateTime: h.ts, request: { method: h.method, url: h.resolved }, response: { status: h.status }, time: h.time })) } }, null, 2), 'application/json'); toastOk('history.har exported'); };
  V2['hist-diffpair'] = el => { ui.diffMode = 'responses'; ui.diffRespA = el.dataset.id; ui.diffRespB = D.history.find(h => h.id !== el.dataset.id && h.status === 200)?.id || D.history[1]?.id; A.navigate('diff'); };
  V2['hist-saved'] = el => { const spec = Object.fromEntries(el.dataset.q.split(' ').map(p => p.split('='))); ui.hist = { ...{ q: '', method: '', status: '', range: 'all', sel: null }, q: spec.q === 'payments' ? 'payments' : spec.q || '', status: spec.status || '', range: spec.range || 'all' }; refreshHist(); };

  A.regView({
    id: 'scripting', title: 'Scripting', icon: 'terminal', group: 'Data & Quality',
    render() {
      const phase = ui.scriptPhase || 'pre';
      const reqOpts = A.walkTree(S.tree).filter(n => n.type === 'request');
      const target = ui.scriptTarget || (reqOpts[0] && reqOpts[0].id);
      const t = target ? A.findOrMakeTab(target) : null;
      const log = S.consoleEntries.slice(-14);
      return `<div class="script-layout">
        <div class="card pad script-left">
          <div class="row-between wrap">
            <div class="row-gap">
              ${RVH.segBtns('script-phase', ['pre', 'post'], phase).replace(/data-name="script-phase"/g, 'data-name="g-scriptphase"')}
              <select data-vinput="script-target" aria-label="Target request">${reqOpts.map(r => `<option value="${r.id}" ${target === r.id ? 'selected' : ''}>${r.method} ${esc(r.name)}</option>`).join('')}</select>
            </div>
            <button class="btn btn-primary" data-vact="script-run">${icon('play')} Execute</button>
          </div>
          ${t ? RVH.codeEditor('script-main', t.scripts[phase] || '', { lang: 'js', rows: 16, placeholder: '// reqly sandbox — Goja runtime, 500ms budget' }) : ''}
          <div class="editor-tools"><span class="dim">API surface: <code>reqly.env</code>, <code>reqly.vars</code>, <code>reqly.request</code>, <code>reqly.response</code>, <code>reqly.expect</code>, <code>reqly.test()</code>, <code>reqly.console</code></span>
          <button class="btn btn-sm" data-vact="insert-loop">${icon('plus')} Insert capture loop</button></div>
        </div>
        <aside class="script-right">
          <div class="card pad"><h4>Sandbox console</h4><div class="mini-console">${log.map(e => `<div class="cline cl-${e.level}"><span class="cl-t">${fmtClock(new Date(e.t).toISOString())}</span><span class="cl-badge">${e.level}</span><span class="cl-msg">${esc(e.msg)}</span></div>`).join('') || '<p class="empty-line">Execute a script to see output.</p>'}</div></div>
          <div class="card pad"><h4>Captured variables</h4>
            <div class="insp-kv"><span>tokenEpoch</span><b class="mono">1756012345678</b><span>accessToken</span><b>[SECRET]</b><span>page</span><b class="mono">1</b></div>
            <p class="dim small">Written via <code>reqly.env.set()</code>/<code>setSecret()</code> — persisted into environment files as plain-text (secrets encrypted).</p></div>
        </aside>
      </div>`;
    },
    after(h) {
      bindViewActs2(h);
      const ta = document.getElementById('script-main');
      ta && ta.addEventListener('input', () => { const t = A.activeTab(); });
    }
  });
  V2['seg'] = function (el) {
    if (el.dataset.name === 'g-scriptphase') { ui.scriptPhase = el.dataset.val; A.renderMain(); return; }
    if (el.dataset.name === 'set-density') { document.documentElement.dataset.density = el.dataset.val; A.renderMain(); return; }
  };
  VI2['script-target'] = el => { ui.scriptTarget = el.value; A.renderMain(); };
  V2['script-run'] = () => {
    const code = document.getElementById('script-main').value;
    A.log('sys', 'sandbox: executing script (' + code.split('\n').length + ' lines)');
    const lines = code.split('\n').filter(l => l.trim());
    lines.forEach((ln, i) => setTimeout(() => {
      if (/expect|test\(/.test(ln)) A.log('pass', 'assertion ✓');
      else if (/setSecret/.test(ln)) A.log('sec', 'stored secret [SECRET]');
      else if (/console/.test(ln)) A.log('log', ln.replace(/.*console\.\w+\(/, '').replace(/[');]/g, '') || '(value)');
      else if (/throw|boom/.test(ln)) A.log('err', 'GojaError: boom is not defined — line ' + (i + 1));
      else A.log('log', '→ ' + ln.trim().slice(0, 70));
    }, i * 140));
  };
  V2['insert-loop'] = () => { const ta = document.getElementById('script-main'); if (ta) { ta.setRangeText("\nfor (const p of reqly.response.json().data) {\n  reqly.console.log('payment', p.id);\n}", ta.selectionStart, ta.selectionStart, 'end'); ta.dispatchEvent(new Event('input')); A.renderMain(); } };

  ui.tests = ui.tests || { suite: D.testSuites[0].id, running: false, results: {}, runs: [{ at: ago(42), pass: 7, fail: 1, skip: 0, dur: '1.9s' }] };
  A.regView({
    id: 'tests', title: 'Test Runner', icon: 'flask', group: 'Data & Quality',
    render() {
      const cfg = ui.tests;
      const suite = D.testSuites.find(s => s.id === cfg.suite) || D.testSuites[0];
      const done = Object.keys(cfg.results).length;
      const counts = { pass: 0, fail: 0, skip: 0 };
      suite.tests.forEach(t => { const r = cfg.results[t.id]; if (r === 'pass') counts.pass++; else if (r === 'fail') counts.fail++; else if (r === 'skip') counts.skip++; else if (r == null && cfg.ranOnce) counts.skip++; });
      return `<div class="tests-layout">
        <aside class="suite-list card">
          <div class="side-head">Test suites<button class="btn btn-icon btn-xs" data-vact="suite-new" data-tip="New suite" aria-label="New suite">${icon('plus')}</button></div>
          ${D.testSuites.map(s => `<button class="suite-item ${s.id === cfg.suite ? 'active' : ''}" data-vact="suite-select" data-id="${s.id}">
            ${icon('flask')}<div><b>${esc(s.name)}</b><small class="mono">${esc(s.file)}</small></div>
            <span class="suite-dot ${Object.keys(cfg.results).length ? (s.tests.some(tst => cfg.results[tst.id] === 'fail') ? 'dot-error' : 'dot-ok') : ''}"></span>
          </button>`).join('')}
          <div class="side-foot-actions"><button class="btn btn-sm" data-vact="run-all">${icon('play')} Run all suites</button></div>
        </aside>
        <section class="suite-detail">
          <div class="row-between wrap">
            <div class="row-gap"><h3>${icon('flask')} ${esc(suite.name)}</h3>${RVH.badge(suite.tests.length + ' assertions', 'info')}</div>
            <div class="row-gap">
              ${cfg.running ? `<span class="spinner sm"></span><span class="dim">running…</span><button class="btn btn-danger btn-sm" data-vact="run-stop">Stop</button>` : `<button class="btn btn-primary" data-vact="run-suite">${icon('play')} Run suite</button>`}
              <button class="btn" data-vact="report-junit">${icon('download')} JUnit</button>
              <button class="btn" data-vact="report-json">${icon('download')} JSON</button>
              <button class="btn" data-vact="report-m42">${icon('doc')} M42 report</button>
            </div>
          </div>
          ${done > 0 || cfg.running ? `<div class="test-summary">
            ${RVH.badge(counts.pass + ' passed', counts.pass ? 'ok' : 'muted')}
            ${RVH.badge(counts.fail + ' failed', counts.fail ? 'error' : 'muted')}
            ${RVH.badge(counts.skip + ' skipped', counts.skip ? 'warn' : 'muted')}
            <div class="progress"><i style="width:${Math.round(done / Math.max(1, suite.tests.length) * 100)}%"></i></div>
          </div>` : ''}
          <div class="assert-list">
            ${suite.tests.map((t, i) => {
              const r = cfg.running && !cfg.results[t.id] && i <= cfg.cursor ? 'running' : cfg.results[t.id];
              return `<div class="assert-row ${r ? 'r-' + r : ''}">
                <span class="assert-state ${r || ''}">${r === 'pass' ? icon('check') : r === 'fail' ? icon('x') : r === 'running' ? '<span class="spinner sm"></span>' : r === 'skip' ? '–' : icon('clock')}</span>
                <div class="assert-body">
                  <b>${esc(t.name)}</b>
                  <div class="assert-spec mono small">${esc(assertSpec(t))}</div>
                  ${r === 'fail' ? `<div class="fail-detail"><span class="tok-key">expected</span> <code>${esc(t.expected || 'value present')}</code> · <span class="tok-key">actual</span> <code>null</code><button class="btn btn-tiny" data-vact="fail-diff">Open diff</button></div>` : ''}
                </div>
                ${RVH.badge(t.type, 'muted')}
              </div>`;
            }).join('')}
            <div class="add-assert">
              <h6>Add assertion</h6>
              <div class="assert-builder">
                <select id="ab-type">${D.assertionTypes.map(a => `<option value="${a.id}">${a.label}</option>`).join('')}</select>
                <input id="ab-source" class="mono" placeholder="source e.g. $.data[0].id">
                <select id="ab-op">${D.assertionOps.map(o => `<option>${o}</option>`).join('')}</select>
                <input id="ab-expected" class="mono" placeholder="expected">
                <button class="btn btn-primary" data-vact="assert-add">Add</button>
              </div>
            </div>
          </div>
          <div class="runs-history card pad">
            <h4>Run history</h4>
            ${RVH.table(['When', 'Passed', 'Failed', 'Skipped', 'Duration'], cfg.runs.slice(0, 6).map(r => [r.at, RVH.badge(String(r.pass), 'ok'), RVH.badge(String(r.fail), 'error'), String(r.skip), r.dur]))}
          </div>
        </section>
      </div>`;
    },
    after(h) { bindViewActs2(h); }
  });
  function assertSpec(t) {
    const src = t.source ? `reqly.expect(${t.type === 'jsonpath' ? "$." + t.source.replace(/^\$\.?/, '') : t.source})` : "reqly.expect(response)";
    return src + '.' + t.op.replace(' ', '(') + (t.expected ? `('${t.expected}')` : '()');
  }
  V2['suite-select'] = el => { ui.tests.suite = el.dataset.id; ui.tests.results = {}; ui.tests.ranOnce = false; A.renderMain(); };
  V2['suite-new'] = () => A.formModal({ title: 'New test suite', submitLabel: 'Create', fields: [{ name: 'n', label: 'File name', value: 'new.spec.yaml' }], onSubmit: v => { D.testSuites.push({ id: A.uid(), name: v.n, file: 'collections/tests/' + v.n, tests: [] }); ui.tests.suite = D.testSuites[D.testSuites.length - 1].id; A.renderMain(); toastOk('Suite scaffolded'); } });
  V2['run-stop'] = () => { clearTimeout(ui.tests.timer); ui.tests.running = false; A.renderMain(); A.toast('Run stopped early', { type: 'warn' }); };
  function runSuite(all) {
    const cfg = ui.tests;
    const suites = all ? D.testSuites : [D.testSuites.find(s => s.id === cfg.suite)];
    cfg.results = {}; cfg.running = true; cfg.cursor = -1; cfg.ranOnce = true;
    const flat = suites.flatMap(s => s.tests.map(t => ({ ...t, _suite: s.id })));
    A.log('sys', `test run: ${flat.length} assertions queued`);
    const step = () => {
      cfg.cursor++;
      if (cfg.cursor >= flat.length) {
        cfg.running = false;
        const p = flat.filter(t => cfg.results[t.id] === 'pass').length, f = flat.filter(t => cfg.results[t.id] === 'fail').length, sk = flat.filter(t => cfg.results[t.id] === 'skip' || !cfg.results[t.id]).length;
        cfg.runs.unshift({ at: fmtDate(new Date().toISOString()), pass: p, fail: f, skip: sk, dur: (flat.length * 0.27).toFixed(1) + 's' });
        A.renderMain();
        A.toast(f ? `Run finished — ${p} passed, ${f} failed` : `All green — ${p} assertions passed`, { type: f ? 'error' : 'success', action: f ? { label: 'Open failures', fn: () => A.log('err', 'failure: next_cursor missing') } : null });
        return;
      }
      const t = flat[cfg.cursor];
      const outcome = /cursor|legacy/i.test(t.name) ? (t.lastRun === 'skip' ? 'skip' : 'fail') : 'pass';
      cfg.results[t.id] = outcome;
      A.log(outcome === 'pass' ? 'pass' : outcome === 'fail' ? 'err' : 'warn', `${outcome}: ${t.name}`);
      A.renderMain();
      cfg.timer = setTimeout(step, 240);
    };
    step();
  }
  V2['run-suite'] = () => runSuite(false);
  V2['run-all'] = () => runSuite(true);
  V2['assert-add'] = () => {
    const suite = D.testSuites.find(s => s.id === ui.tests.suite);
    suite.tests.push({ id: A.uid(), name: document.getElementById('ab-source').value || 'unnamed assertion', type: document.getElementById('ab-type').value, source: document.getElementById('ab-source').value, op: document.getElementById('ab-op').value, expected: document.getElementById('ab-expected').value, lastRun: 'skip' });
    A.renderMain(); toastOk('Assertion added to suite');
  };
  V2['fail-diff'] = () => { A.modal({ title: 'Failure diff — $.next_cursor', wide: true, body: `<pre class="codeview mono">{
  "<span class="del-line">"next_cursor": "cur_1Q7ybn"</span>",
  "<span class="add-line">"next_cursor": null</span>"
}</pre>`, footer: `<button class="btn" onclick="ReqlyApp.closeTopModal()">Close</button>` }); };
  V2['report-junit'] = () => { A.dl('payments-report.xml', `<?xml version="1.0"?><testsuite name="payments.spec" tests="6" failures="1">…</testsuite>`, 'text/xml'); toastOk('JUnit report downloaded'); };
  V2['report-json'] = () => { A.dl('payments-report.json', JSON.stringify({ suite: 'payments.spec', results: ui.tests.results }, null, 2), 'application/json'); toastOk('JSON report downloaded'); };
  V2['report-m42'] = () => { A.modal({ title: 'M42 structured report', body: reportHtml(D.importReportSample), footer: `<button class="btn" onclick="ReqlyApp.closeTopModal()">Close</button><button class="btn btn-primary" onclick="ReqlyApp.toast('Report saved to reports/m42-latest.md',{type:'success'});ReqlyApp.closeTopModal()">Save markdown</button>` }); };

  ui.runner = ui.runner || { bulkLog: [], pageLog: [], conc: 4 };
  A.regView({
    id: 'runner', title: 'Bulk & Pagination', icon: 'zap', group: 'Data & Quality',
    render() {
      const r = ui.runner;
      return `<div class="runner-layout">
        <div class="card pad">
          <h4>${icon('zap')} Bulk runner</h4>
          <p class="sub">Run one request per CSV/JSON row. Variables map by column name.</p>
          ${RVH.codeEditor('bulk-csv', D.bulkInputSample, { rows: 5 })}
          <div class="editor-tools wrap">
            <label class="fld-inline"><span>Concurrency</span><input type="range" min="1" max="16" value="${r.conc}" data-vinput="bulk-conc"><b class="mono">${r.conc}</b></label>
            <label class="fld-inline"><span>Mode</span><select><option>parallel</option><option>sequential</option></select></label>
            <label class="check"><input type="checkbox" checked> stop on first failure</label>
            ${r.bulkRunning ? '<span class="spinner sm"></span>' : `<button class="btn btn-primary" data-vact="bulk-run">${icon('play')} Run bulk</button>`}
          </div>
          <div class="bulk-progress">${r.bulkLog.map(b => `<div class="bulk-row"><span class="mono">${esc(b.row)}</span>${statusChip(b.status)}<span class="dim">${ms(b.ms)}</span></div>`).join('')}${r.bulkSummary ? `<div class="banner banner-success">${icon('check')}<div>${r.bulkSummary}</div></div>` : ''}</div>
        </div>
        <div class="card pad">
          <h4>${icon('layers')} Pagination runner</h4>
          <div class="editor-tools wrap">
            <label class="fld-inline"><span>Strategy</span><select id="pg-strategy"><option>cursor</option><option>page number</option><option>offset</option><option>Link header</option></select></label>
            <label class="fld-inline"><span>Max pages</span><input class="mono" style="width:64px" value="10" id="pg-max"></label>
            <label class="fld-inline"><span>Stop when</span><select id="pg-stop"><option>has_more = false</option><option>empty page</option><option>same cursor twice</option></select></label>
            ${r.pgRunning ? '<span class="spinner sm"></span><button class="btn btn-sm btn-danger" data-vact="pg-stop">Stop</button>' : `<button class="btn btn-primary" data-vact="pg-run">${icon('play')} Crawl pages</button>`}
          </div>
          <div class="pg-log">${r.pageLog.map(p => `<div class="pg-row"><span class="mono">page ${p.n}</span><span class="mono dim">${esc(p.cursor)}</span>${statusChip(200)}<span class="dim">${p.rows} rows · ${ms(p.ms)}</span></div>`).join('')}${r.pgSummary ? `<div class="banner banner-success">${icon('check')}<div>${r.pgSummary}</div></div>` : ''}</div>
        </div>
      </div>`;
    },
    after(h) {
      bindViewActs2(h);
      const c = h.querySelector('[data-vinput="bulk-conc"]');
      c && c.addEventListener('input', () => { ui.runner.conc = +c.value; c.nextElementSibling.textContent = c.value; });
    }
  });
  V2['bulk-run'] = () => {
    const r = ui.runner; r.bulkLog = []; r.bulkRunning = true; r.bulkSummary = null;
    const rows = D.bulkInputSample.split('\n').slice(1);
    let i = 0;
    const step = () => {
      if (i >= rows.length) { r.bulkRunning = false; r.bulkSummary = `${rows.length} rows executed · concurrency ${r.conc} · 0 failures`; A.renderMain(); return; }
      const row = rows[i];
      r.bulkLog.push({ row, status: 201, ms: 120 + Math.round(Math.random() * 300) });
      i++; A.renderMain();
      setTimeout(step, 260);
    };
    step();
  };
  VI2['bulk-conc'] = () => {};
  V2['pg-run'] = () => {
    const r = ui.runner; r.pageLog = []; r.pgRunning = true; r.pgSummary = null;
    let n = 0, cur = 'start';
    const step = () => {
      n++;
      r.pageLog.push({ n, cursor: 'cur_' + Math.random().toString(36).slice(2, 8), rows: n === 4 ? 0 : 25, ms: 130 + Math.round(Math.random() * 120) });
      if (n >= 4) { r.pgRunning = false; r.pgSummary = 'Crawled 4 pages · 75 rows aggregated · stopped on empty page'; A.renderMain(); return; }
      A.renderMain(); setTimeout(step, 380);
    };
    step();
  };
  V2['pg-stop'] = () => { ui.runner.pgRunning = false; A.renderMain(); };

  ui.openapi = ui.openapi || { spec: D.openapiSpecs[0].id, tag: '', method: '', q: '', ep: null, tryResp: null, schemaOpen: {} };
  A.regView({
    id: 'openapi', title: 'OpenAPI Explorer', icon: 'book', group: 'Specs & Tooling',
    render() {
      const o = ui.openapi;
      const spec = D.openapiSpecs.find(s => s.id === o.spec);
      let eps = spec.endpoints;
      if (o.tag) eps = eps.filter(e => e.tag === o.tag);
      if (o.method) eps = eps.filter(e => e.method === o.method);
      if (o.q) { const q = o.q.toLowerCase(); eps = eps.filter(e => (e.path + e.summary).toLowerCase().includes(q)); }
      const ep = spec.endpoints.find(e => e.id === o.ep);
      const tags = [...new Set(spec.endpoints.map(e => e.tag))];
      return `<div class="oa-layout">
        <aside class="oa-side">
          <div class="row-between"><select data-vinput="oa-spec" aria-label="Specification selector">${D.openapiSpecs.map(s => `<option value="${s.id}" ${s.id === o.spec ? 'selected' : ''}>${esc(s.name)}</option>`).join('')}</select><span class="mono dim small">v${spec.version}</span></div>
          <p class="dim small">${spec.format} · ${spec.endpoints.length} operations · ${spec.schemas.length} schemas</p>
          <div class="side-search">${icon('search')}<input placeholder="Search endpoints…" value="${esc(o.q)}" data-vinput="oa-q"></div>
          <div class="chip-row">${['', ...tags].map(t => `<button class="chip chip-nav ${o.tag === t ? 'active' : ''}" data-vact="oa-tag" data-t="${t}">${t || 'all tags'}</button>`).join('')}</div>
          <div class="chip-row">${['', ...['GET', 'POST']].map(m => `<button class="chip chip-nav ${o.method === m ? 'active' : ''}" data-vact="oa-method" data-m="${m}">${m || 'all methods'}</button>`).join('')}</div>
          <div class="oa-tree">
            ${eps.length === 0 ? '<p class="empty-line">No endpoints match filters</p>' : tags.filter(t => eps.some(e => e.tag === t)).map(tag => `
              <div class="oa-tag">${esc(tag)}</div>
              ${eps.filter(e => e.tag === tag).map(e => `<button class="oa-ep ${o.ep === e.id ? 'active' : ''}" data-vact="oa-ep" data-id="${e.id}">${methodChip(e.method)}<span class="${e.deprecated ? 'strike' : ''}">${esc(e.path)}</span><small class="ellip">${esc(e.summary)}</small></button>`).join('')}`).join('')}
          </div>
          <div class="side-head" style="margin-top:10px">Security schemes</div>
          ${spec.security.map(sec => `<div class="sec-scheme">${icon('lock')}<div><b>${sec.scheme}</b><small>${sec.type}${sec.bearerFormat ? ' · ' + sec.bearerFormat : ''}${sec.in ? ' · in:' + sec.in + ':' + sec.name : ''}</small><p class="dim small">${esc(sec.desc)}</p></div></div>`).join('')}
        </aside>
        <section class="oa-detail">
          ${!ep ? RVH.emptyState('book', 'Select an endpoint', 'Browse the specification tree — every operation carries params, schemas and a live Try-it panel.') : `
          <div class="row-between wrap">
            <div class="row-gap">${methodChip(ep.method)}<h3 class="mono">${esc(ep.path)}</h3>${ep.deprecated ? RVH.badge('deprecated', 'warn') : ''}</div>
            <div class="row-gap">
              <button class="btn" data-vact="oa-docs">${icon('doc')} Generate docs</button>
              <button class="btn" data-vact="oa-import">${icon('import')} Create request</button>
            </div>
          </div>
          <p class="sub">${esc(ep.summary)}</p>
          ${ep.params && ep.params.length ? `<h5>Parameters</h5>${RVH.table(['Name', 'In', 'Schema'], ep.params.map(p => [`<code class="mono">${esc(p.name)}</code>`, RVH.badge(p.in, p.in === 'path' ? 'warn' : 'muted'), `<span class="mono small">${esc(p.schema)}</span>`]))}` : ''}
          ${ep.requestBody ? `<h5>Request body <span class="mono dim small">${ep.requestBody.contentType}${ep.requestBody.required ? ' · required' : ''}</span></h5>
            <div class="schema-view">${ep.requestBody.props.map(p => `<div class="schema-prop"><code>${esc(p.name)}</code><b class="mono">${esc(p.type)}</b>${p.required ? RVH.badge('required', 'error') : ''}<p class="dim small">${esc(p.desc || '')}</p></div>`).join('')}</div>` : ''}
          <h5>Responses</h5>
          ${ep.responses.map(r => { let ex = null; try { ex = JSON.parse(r.example); } catch (e) {} return `<div class="resp-example"><div class="row-gap">${statusChip(+r.code.split(',')[0] >= 100 ? +r.code : 200)}<b>${esc(r.desc)}</b></div>${ex ? `<pre class="codeview mono small">${highlight(JSON.stringify(ex, null, 2), 'json')}</pre>` : '<p class="dim small">No example payload in spec.</p>'}</div>`; }).join('')}
          <div class="try-it card pad">
            <div class="row-between"><h4>Try it</h4><select data-vinput="oa-server">${spec.servers.map(s => `<option>${esc(s)}</option>`).join('')}</select></div>
            ${o.tryLoading ? '<div class="load-row"><span class="spinner"></span> executing…</div>' : `<button class="btn btn-primary" data-vact="oa-try">${icon('play')} Send example request</button>`}
            ${o.tryResp ? `${statusChip(o.tryResp.status)}<span class="dim">${ms(o.tryResp.time)}</span><pre class="codeview mono small">${highlight(o.tryResp.body, 'json')}</pre>` : ''}
          </div>
          <h5>Related schemas</h5>
          <div class="chip-row">${spec.schemas.map(sc => `<button class="chip chip-nav" data-vact="oa-schema" data-n="${sc.name}">${sc.name}</button>`).join('')}</div>`}
        </section>
      </div>`;
    },
    after(h) { bindViewActs2(h); }
  });
  VI2['oa-spec'] = el => { ui.openapi.spec = el.value; ui.openapi.ep = null; A.renderMain(); };
  VI2['oa-q'] = el => { ui.openapi.q = el.value; refreshOA(); };
  function refreshOA() { const host = A.root.querySelector('.view-host'); if (host && S.view === 'openapi') { const v = A.VIEWS.openapi; host.innerHTML = v.render(); v.after(host); } }
  V2['oa-tag'] = el => { ui.openapi.tag = el.dataset.t; refreshOA(); };
  V2['oa-method'] = el => { ui.openapi.method = el.dataset.m; refreshOA(); };
  V2['oa-ep'] = el => { ui.openapi.ep = el.dataset.id; ui.openapi.tryResp = null; A.renderMain(); };
  V2['oa-try'] = () => { ui.openapi.tryLoading = true; A.renderMain(); setTimeout(() => { ui.openapi.tryLoading = false; ui.openapi.tryResp = { status: 200, time: 151, body: D.responses.ok.body }; A.renderMain(); }, 700); };
  V2['oa-import'] = () => { const spec = D.openapiSpecs.find(s => s.id === ui.openapi.spec); const ep = spec.endpoints.find(e => e.id === ui.openapi.ep); S.tree.push({ type: 'request', id: A.uid(), name: ep.summary, method: ep.method, url: '{{baseUrl}}' + ep.path.replace(/\{(\w+)\}/g, '{{$1}}'), saved: false }); A.renderSidebar(); toastOk(`Request created under Collections root — edit & save it into the right folder`); };
  V2['oa-schema'] = el => {
    const spec = D.openapiSpecs.find(s => s.id === ui.openapi.spec);
    const sc = spec.schemas.find(x => x.name === el.dataset.n);
    A.modal({ title: 'Schema — ' + sc.name, wide: true, body: `<div class="schema-view">${sc.props.map(p => `<div class="schema-prop"><code>${esc(p.name)}</code><b class="mono">${esc(p.type)}</b>${p.required ? RVH.badge('required', 'error') : ''}${p.enum ? RVH.badge('enum', 'info') : ''}<p class="dim small mono">${esc(p.enum || p.const || '')}</p></div>`).join('')}</div>`, footer: `<button class="btn btn-primary" onclick="ReqlyApp.closeTopModal()">Done</button>` });
  };
  V2['oa-docs'] = () => {
    const md = '# PayCore API\n\n## GET /payments\nList payments.\n\n**Parameters**\n| Name | In | Schema |\n|---|---|---|\n| limit | query | integer |\n\n**Responses**\n- `200` page of payments\n';
    A.modal({ title: 'Generated documentation (markdown)', wide: true, body: `<pre class="codeview mono small">${esc(md)}</pre>`, footer: `<button class="btn" onclick="ReqlyApp.closeTopModal()">Close</button><button class="btn btn-primary" onclick="ReqlyApp.dl('paycore-docs.md','Docs','text/markdown');ReqlyApp.toast('docs downloaded',{type:'success'})">Download .md</button>` });
  };

  ui.impexp = ui.impexp || { step: 1, format: 'auto', payload: '', invalid: false, progress: 0, done: false };
  A.regView({
    id: 'importexport', title: 'Import / Export', icon: 'import', group: 'Specs & Tooling',
    render() {
      const w = ui.impexp;
      const detected = w.payload.includes('curl ') ? 'curl' : w.payload.includes('"log"') ? 'har' : w.payload.includes('openapi:') ? 'openapi' : w.payload.includes('v2.1.0') ? 'postman' : w.payload.includes('__export_date') ? 'insomnia' : w.payload.includes('.bru') ? 'bruno' : '';
      return `<div class="ie-layout">
        <section class="ie-import card pad">
          <h3>${icon('import')} Import</h3>
          <div class="steps"><span class="step ${w.step >= 1 ? 'on' : ''}">1 Payload</span><span class="step ${w.step >= 2 ? 'on' : ''}">2 Preview</span><span class="step ${w.step >= 3 ? 'on' : ''}">3 Report</span></div>
          ${w.step === 1 ? `
            <div class="dropzone big" data-vact="ie-pick">${icon('download')}<p><b>Drop a file</b> or click to browse</p><p class="dim">cURL command, OpenAPI 3.x, HAR 1.2, Postman v2.1, Insomnia v4/v5, Bruno collection</p></div>
            <textarea class="mono ie-paste" placeholder="…or paste a cURL command / JSON / YAML here" data-vinput="ie-payload">${esc(w.payload)}</textarea>
            <div class="row-gap wrap">
              <span class="detected">${detected ? RVH.badge('auto-detected: ' + detected, 'ok') : RVH.badge('unknown format — pick manually', 'warn')}</span>
              <select data-vinput="ie-format">${['auto', ...D.importFormats.map(f => f.id)].map(f => `<option ${w.format === f ? 'selected' : ''}>${f}</option>`).join('')}</select>
              <label class="check"><input type="checkbox" ${w.invalid ? 'checked' : ''} data-vact-check="ie-invalid"> simulate validation errors</label>
              <button class="btn btn-primary" data-vact="ie-next" ${!w.payload ? 'disabled' : ''}>Continue →</button>
            </div>
            <div class="fmt-grid">${D.importFormats.map(f => `<div class="fmt-cell ${detected === f.id ? 'hit' : ''}"><b>${f.name}</b><span class="mono dim small">${f.ext}</span><code class="mono tiny">${esc(f.detect.slice(0, 34))}</code></div>`).join('')}</div>` : ''}
          ${w.step === 2 ? `
            ${w.invalid ? `<div class="banner banner-error">${icon('warn')}<div><b>Validation failed</b><ul class="err-list"><li>paths./payments.get.responses — missing 200 definition</li><li>components.schemas.Payment — unresolved $ref “#/Pet”</li><li>HAR entry #4 — no request URL</li></ul></div></div>` : `
            <p class="sub">Preview — 18 operations found, grouped by tag (showing first 5):</p>
            ${['Payments', 'Customers', 'Webhooks'].map(tag => `<div class="prev-group"><h6>${tag}</h6>${['GET /payments', 'POST /payments', 'GET /payments/{id}'].map(p => `<div class="prev-row">${methodChip(p.split(' ')[0])}<span class="mono">${p.split(' ')[1]}</span></div>`).join('')}</div>`).join('')}
            <button class="btn btn-sm" data-vact="ie-showall">Show all 18</button>`}
            <div class="row-gap wrap" style="margin-top:12px">
              <button class="btn" data-vact="ie-back">← Back</button>
              ${w.invalid ? '' : `<button class="btn btn-primary" data-vact="ie-run">${icon('play')} Import into workspace</button>`}
            </div>` : ''}
          ${w.step === 3 ? `
            ${w.progress < 100 ? `<div class="load-row"><span class="spinner"></span> importing… <b>${w.progress}%</b></div><div class="progress big"><i style="width:${w.progress}%"></i></div>` : `
            <div class="banner banner-success">${icon('check')}<div><b>Import complete.</b> 22 files written to the worktree — review them in Git before committing.</div></div>
            ${reportHtml(D.importReportSample)}`}
            <div class="row-gap wrap" style="margin-top:12px"><button class="btn" data-vact="ie-again">Import another</button>${w.progress >= 100 ? `<button class="btn btn-primary" data-action="nav" data-view="git">Review changes in Git →</button>` : ''}</div>` : ''}
        </section>
        <section class="ie-export card pad">
          <h3>${icon('export')} Export</h3>
          <div class="field"><label>What</label><select id="ex-what"><option>Whole workspace</option><option>Current collection</option><option>Active request</option><option>OpenAPI 3.0 spec (generated)</option><option>Documentation site</option></select></div>
          <div class="field"><label>Format</label><select id="ex-fmt"><option>Postman Collection v2.1</option><option>OpenAPI 3.0 (YAML)</option><option>HAR 1.2</option><option>Workspace directory (plain-text)</option><option>Code bundle (clients)</option></select></div>
          <div class="field"><label>Options</label>
            <label class="check"><input type="checkbox" checked id="ex-secrets"> mask secrets ([SECRET])</label>
            <label class="check"><input type="checkbox" checked> include environments</label>
            <label class="check"><input type="checkbox" checked> include test suites</label>
          </div>
          <button class="btn btn-primary" data-vact="ex-run">${icon('export')} Generate export</button>
          <div class="ex-result">
            ${w.exBusy ? '<div class="load-row"><span class="spinner"></span> generating…</div>' : ''}
            ${w.exDone ? `<div class="file-chip">${icon('doc')} paycore-workspace-export.zip <span class="dim">482 KB</span><button class="btn btn-sm" data-vact="ex-dl">Download</button></div>` : ''}
          </div>
        </section>
      </div>`;
    },
    after(h) { bindViewActs2(h); }
  });
  V2['ie-pick'] = () => { ui.impexp.payload = `curl -X POST https://api.paycore.dev/v1/payments -H 'Authorization: Bearer …' -d '{"amount":4250}'`; A.renderMain(); toastOk('Sample cURL loaded — detection ran automatically'); };
  VI2['ie-payload'] = el => { ui.impexp.payload = el.value; refreshIE(); };
  VI2['ie-format'] = el => { ui.impexp.format = el.value; };
  function refreshIE() { const host = A.root.querySelector('.view-host'); if (host && S.view === 'importexport') { const v = A.VIEWS.importexport; host.innerHTML = v.render(); v.after(host); } }
  V2['ie-next'] = () => { ui.impexp.step = 2; A.renderMain(); };
  V2['ie-back'] = () => { ui.impexp.step = 1; A.renderMain(); };
  V2['ie-showall'] = () => toastOk('Expanded full operation list (demo cap lifted)');
  V2['ie-run'] = () => {
    const w = ui.impexp; w.step = 3; w.progress = 0; A.renderMain();
    const iv = setInterval(() => { w.progress = Math.min(100, w.progress + 17); if (w.progress >= 100) { clearInterval(iv); A.toast('Import finished — M42 report ready', { type: 'success' }); } A.renderMain(); }, 240);
  };
  V2['ie-again'] = () => { ui.impexp = { step: 1, format: 'auto', payload: '', invalid: false, progress: 0 }; A.renderMain(); };
  V2['ex-run'] = () => { ui.impexp.exBusy = true; A.renderMain(); setTimeout(() => { ui.impexp.exBusy = false; ui.impexp.exDone = true; A.renderMain(); toastOk('Export bundle generated'); }, 900); };
  V2['ex-dl'] = () => { A.dl('paycore-workspace-export.zip', 'PK demo archive', 'application/zip'); toastOk('Downloaded'); };
  function reportHtml(rep) {
    return `<div class="report">
      <div class="rep-counts">${Object.entries(rep.counts).map(([k, v]) => `<span class="rep-count"><b>${v}</b><em>${k}</em></span>`).join('')}</div>
      ${rep.entries.map(en => `<div class="rep-cat rep-${en.severity}"><h6>${en.category} · ${RVH.badge(en.severity, en.severity === 'translated' ? 'ok' : en.severity === 'warned' ? 'warn' : 'error')}</h6><ul>${en.items.map(i => `<li class="mono small">${esc(i)}</li>`).join('')}</ul></div>`).join('')}
    </div>`;
  }

  ui.codegen = ui.codegen || { lang: 'JavaScript', client: 'fetch', opts: { auth: true, headers: true, asyncWrap: false }, err: false };
  A.regView({
    id: 'codegen', title: 'Code Generation', icon: 'braces', group: 'Specs & Tooling',
    render() {
      const c = ui.codegen;
      const reqs = A.walkTree(S.tree).filter(n => n.type === 'request');
      const src = reqs.find(r => r.id === c.srcId) || A.activeTab() && reqs.find(r => r.id === A.activeTab().nodeId) || reqs[0];
      if (!src) return RVH.emptyState('braces', 'Nothing to generate from', 'Select a request in Collections first.');
      const key = c.lang + '|' + c.client;
      let code = D.generatedCode[key];
      if (c.err) code = null;
      return `<div class="cg-layout">
        <aside class="cg-side card pad">
          <h4>Source</h4>
          <select data-vinput="cg-src">${reqs.map(r => `<option value="${r.id}" ${src.id === r.id ? 'selected' : ''}>${r.method} ${esc(r.name)}</option>`).join('')}</select>
          <h4>Language</h4>
          <div class="lang-list">${D.codegenTargets.map(t => `<button class="lang-item ${c.lang === t.lang ? 'active' : ''}" data-vact="cg-lang" data-l="${t.lang}">${icon('braces')} ${t.lang}</button>`).join('')}</div>
          <h4>Client</h4>
          <select data-vinput="cg-client">${(D.codegenTargets.find(t => t.lang === c.lang) || {}).clients?.map(cl => `<option ${c.client === cl ? 'selected' : ''}>${cl}</option>`).join('') || ''}</select>
          <h4>Options</h4>
          <label class="check"><input type="checkbox" ${c.opts.auth ? 'checked' : ''} data-opt="auth"> include authentication</label>
          <label class="check"><input type="checkbox" ${c.opts.headers ? 'checked' : ''} data-opt="headers"> generated headers (idempotency)</label>
          <label class="check"><input type="checkbox" ${c.opts.asyncWrap ? 'checked' : ''} data-opt="asyncWrap"> wrap in async main()</label>
          <p class="dim small">${icon('lock')} Secrets are emitted as environment lookups (<code>[SECRET]</code> masked) — never literal.</p>
        </aside>
        <section class="cg-main card">
          <div class="row-between wrap">
            <div class="row-gap">${methodChip(src.method)}<b class="mono">${esc(resolveVarsLite(src.url))}</b>${RVH.badge(c.lang + ' · ' + c.client, 'primary')}</div>
            <div class="row-gap">
              <button class="btn" data-vact="cg-copy">${icon('copy')} Copy</button>
              <button class="btn" data-vact="cg-dl">${icon('download')} Download</button>
            </div>
          </div>
          ${code ? `<pre class="codeview mono cg-code">${highlight(code, c.lang === 'JavaScript' ? 'js' : c.lang === 'Python' || c.lang === 'PHP' || c.lang === 'Rust' || c.lang === 'Go' ? 'js' : 'http')}</pre>` :
            RVH.emptyState('warn', 'Generation failed', 'The selected client has no template for this request kind. Pick another language/client combination.', `<button class="btn" data-vact="cg-fix">Use JavaScript · fetch</button>`)}
        </section>
      </div>`;
    },
    after(h) {
      bindViewActs2(h);
      h.querySelectorAll('[data-opt]').forEach(cb => cb.addEventListener('change', () => { ui.codegen.opts[cb.dataset.opt] = cb.checked; }));
    }
  });
  function resolveVarsLite(u) { return u.replace(/\{\{\w+\}\}/g, m => m); }
  VI2['cg-src'] = el => { ui.codegen.srcId = el.value; A.renderMain(); };
  VI2['cg-client'] = el => { ui.codegen.client = el.value; A.renderMain(); };
  V2['cg-lang'] = el => { ui.codegen.lang = el.dataset.l; ui.codegen.client = (D.codegenTargets.find(t => t.lang === el.dataset.l) || {}).clients?.[0]; ui.codegen.err = false; A.renderMain(); };
  V2['cg-copy'] = () => { navigator.clipboard?.writeText(D.generatedCode[ui.codegen.lang + '|' + ui.codegen.client] || ''); toastOk('Code copied to clipboard'); };
  V2['cg-dl'] = () => { A.dl('request.' + (ui.codegen.lang === 'Python' ? 'py' : ui.codegen.lang === 'Go' ? 'go' : ui.codegen.lang === 'Rust' ? 'rs' : ui.codegen.lang === 'PHP' ? 'php' : ui.codegen.lang === 'cURL' ? 'sh' : 'js'), D.generatedCode[ui.codegen.lang + '|' + ui.codegen.client] || '', 'text/plain'); toastOk('Snippet downloaded'); };
  V2['cg-fix'] = () => { ui.codegen.err = false; ui.codegen.lang = 'JavaScript'; ui.codegen.client = 'fetch'; A.renderMain(); };

  ui.mocks = ui.mocks || { sel: D.mockServers[0]?.id, editing: null, logTimer: null };
  A.regView({
    id: 'mocks', title: 'Mock Server', icon: 'server', group: 'Specs & Tooling',
    render() {
      const m = ui.mocks;
      const srv = D.mockServers.find(s => s.id === m.sel);
      return `<div class="mock-layout">
        <aside class="srv-list card">
          <div class="side-head">Servers<button class="btn btn-icon btn-xs" data-vact="srv-new" data-tip="New mock server" aria-label="New server">${icon('plus')}</button></div>
          ${D.mockServers.length === 0 ? RVH.emptyState('server', 'No mock servers', 'Spin one up from an OpenAPI spec or blank.', `<button class="btn btn-primary btn-sm" data-vact="srv-new">Create server</button>`) : D.mockServers.map(s => `<button class="srv-item ${s.id === m.sel ? 'active' : ''}" data-vact="srv-select" data-id="${s.id}">
            <i class="led led-${s.running ? 'ok' : 'muted'}"></i><div><b>${esc(s.name)}</b><small class="mono">localhost:${s.port} · ${s.hitCount} hits</small></div>
          </button>`).join('')}
        </aside>
        ${!srv ? '' : `<section class="srv-detail">
          <div class="row-between wrap">
            <div class="row-gap"><i class="led led-lg led-${srv.running ? 'ok pulse' : 'muted'}"></i><h3>${esc(srv.name)}</h3>${RVH.badge(srv.running ? 'running' : 'stopped', srv.running ? 'ok' : 'muted')}</div>
            <div class="row-gap">
              <button class="btn ${srv.running ? 'btn-danger' : 'btn-primary'}" data-vact="srv-toggle" data-id="${srv.id}">${srv.running ? icon('stop') + ' Stop' : icon('play') + ' Start'}</button>
              <button class="btn" data-vact="srv-url">${icon('copy')} Copy base URL</button>
              <button class="btn btn-danger-ghost" data-vact="srv-del">${icon('trash')} Delete</button>
            </div>
          </div>
          <p class="sub mono">${srv.running ? `http://localhost.${srv.port}.reqly.local` : `bind http://localhost:${srv.port} — press Start`} · uptime ${srv.running ? Math.floor((Date.now() - srv.uptimeStart) / 60000) + ' min' : '—'} · delay profile: realistic</p>
          <div class="row-between wrap" style="margin-bottom:8px"><h5>Endpoints</h5><button class="btn btn-sm" data-vact="ep-add">${icon('plus')} Add route</button></div>
          ${RVH.table(['Method', 'Path', 'Status', 'Delay', ''], srv.endpoints.map(e => [methodChip(e.method), `<code class="mono">${esc(e.path)}</code>`, statusChip(e.status), ms(e.delay),
            `<button class="btn btn-sm" data-vact="ep-edit" data-id="${e.id}">${icon('edit')} Edit</button>`]))}
          <div class="srv-logs card">
            <div class="dock-tabs"><span class="dock-tab active">Live logs</span><span class="dock-count">${srv.logs.length} requests served</span></div>
            <div class="dock-body mini-console">${srv.logs.slice(-10).reverse().map(l => `<div class="net-row"><span class="cl-t">${fmtClock(l.t)}</span>${methodChip(l.m)}<span class="mono">${esc(l.p)}</span>${statusChip(l.s)}<span class="dim">${ms(l.ms)}</span></div>`).join('') || '<p class="empty-line">No traffic since start.</p>'}</div>
          </div>
        </section>`}
      </div>`;
    },
    after(h) { bindViewActs2(h); }
  });
  A.regSidebar('mocks', () => `<div class="side-block"><div class="side-head">Quick templates</div>
    ${['Blank server', 'From OpenAPI spec', 'Record & replay'].map(t => `<button class="side-link" data-vact="srv-new">${icon('server')} ${t}</button>`).join('')}</div>`);
  V2['srv-select'] = el => { ui.mocks.sel = el.dataset.id; A.renderMain(); };
  V2['srv-toggle'] = el => {
    const srv = D.mockServers.find(s => s.id === el.dataset.id);
    if (!srv.running && S.simulatePortConflict) { S.simulatePortConflict = false; A.toast('Bind failed: port ' + srv.port + ' already in use — kill stale process or change port', { type: 'error', timeout: 6000 }); return; }
    srv.running = !srv.running;
    if (srv.running) {
      srv.uptimeStart = Date.now();
      clearInterval(ui.mocks.logTimer);
      ui.mocks.logTimer = setInterval(() => {
        if (!srv.running) return clearInterval(ui.mocks.logTimer);
        const ep = srv.endpoints[Math.floor(Math.random() * srv.endpoints.length)];
        srv.hitCount++;
        srv.logs.push({ t: new Date(now2()).toISOString(), m: ep.method, p: ep.path, s: ep.status, ms: ep.delay + Math.round(Math.random() * 20) });
        if (S.view === 'mocks') A.renderMain();
      }, 2200);
      A.toast(`Mock listening on localhost:${srv.port}`, { type: 'success' });
    } else { clearInterval(ui.mocks.logTimer); A.toast('Mock server stopped', { type: 'info' }); }
    A.renderMain(); A.updateInspector();
  };
  const now2 = () => Date.now();
  V2['srv-new'] = () => A.formModal({ title: 'New mock server', submitLabel: 'Create', fields: [ { name: 'n', label: 'Name', value: 'Untitled mock' }, { name: 'port', label: 'Port', value: String(9091 + D.mockServers.length) }, { name: 'tpl', label: 'Seed routes from', type: 'select', options: ['blank', 'PayCore API v1 (OpenAPI)', 'recorded history'] } ], onSubmit: v => { const s = { id: A.uid(), name: v.n, port: +v.port, running: false, uptimeStart: 0, hitCount: 0, endpoints: v.tpl !== 'blank' ? [{ id: A.uid(), method: 'GET', path: '/v1/payments', status: 200, delay: 120, headers: [], bodyType: 'json', body: '{ "data": [], "mocked": true }' }] : [], logs: [] }; D.mockServers.push(s); ui.mocks.sel = s.id; A.renderMain(); toastOk('Server created — start it when ready'); } });
  V2['srv-del'] = async () => { const srv = D.mockServers.find(s => s.id === ui.mocks.sel); if (!await A.confirmModal('Delete mock server', `Remove <b>${esc(srv.name)}</b> and its routes?`, 'Delete', true)) return; D.mockServers = D.mockServers.filter(s => s.id !== ui.mocks.sel); ui.mocks.sel = D.mockServers[0]?.id; A.renderMain(); toastOk('Server removed'); };
  V2['srv-url'] = () => { navigator.clipboard?.writeText('http://localhost:' + (D.mockServers.find(s => s.id === ui.mocks.sel) || {}).port); toastOk('Base URL copied'); };
  V2['ep-add'] = () => epModal(null);
  V2['ep-edit'] = el => epModal(el.dataset.id);
  function epModal(epId) {
    const srv = D.mockServers.find(s => s.id === ui.mocks.sel);
    const ep = srv.endpoints.find(e => e.id === epId) || { method: 'GET', path: '/', status: 200, delay: 100, body: '{\n  "mocked": true\n}' };
    A.formModal({
      title: epId ? 'Edit mock response' : 'Add mock route', submitLabel: epId ? 'Save route' : 'Add route', wide: true,
      fields: [
        { name: 'method', label: 'Method', type: 'select', options: ['GET', 'POST', 'PUT', 'PATCH', 'DELETE'], value: ep.method },
        { name: 'path', label: 'Path', value: ep.path },
        { name: 'status', label: 'Response status', type: 'select', options: ['200', '201', '204', '301', '400', '404', '429', '500', '503'], value: String(ep.status) },
        { name: 'delay', label: 'Delay (ms)', value: String(ep.delay), hint: 'Delay/error simulation for resilience testing.' },
        { name: 'body', label: 'Response body (JSON)', type: 'textarea', rows: 8, value: ep.body, hint: 'Supports template tags: {{$uuid}}, {{$timestamp}}' }
      ],
      onSubmit: v => {
        const rec = { id: epId || A.uid(), method: v.method, path: v.path, status: +v.status, delay: +v.delay, headers: [], bodyType: 'json', body: v.body };
        if (epId) srv.endpoints[srv.endpoints.findIndex(e => e.id === epId)] = rec; else srv.endpoints.push(rec);
        A.renderMain(); toastOk(epId ? 'Route updated' : 'Route added');
      }
    });
  }

  ui.diff = ui.diff || { base: 'v1.3.0', target: 'v1.4.2', ran: true, sev: '', q: '', sel: 'c2', mode: 'specs', respA: 'h2', respB: 'h9' };
  A.regView({
    id: 'diff', title: 'API Diff', icon: 'compare', group: 'Specs & Tooling',
    render() {
      const d = ui.diff;
      if (d.mode === 'responses') {
        const a = D.history.find(h => h.id === d.respA), b = D.history.find(h => h.id === d.respB);
        const rows = diffLines(sampleBodyFor(a), sampleBodyFor(b));
        return `<div class="diff-layout">
          <div class="row-gap wrap">
            <select data-vinput="diff-ra">${D.history.map(h => `<option value="${h.id}" ${d.respA === h.id ? 'selected' : ''}>${h.method} ${h.url.split('?')[0].split('/').pop()} · ${h.status}</option>`).join('')}</select>
            <select data-vinput="diff-rb">${D.history.map(h => `<option value="${h.id}" ${d.respB === h.id ? 'selected' : ''}>${h.method} ${h.url.split('?')[0].split('/').pop()} · ${h.status}</option>`).join('')}</select>
            <button class="btn" data-vact="diff-mode-specs">${icon('compare')} Switch to spec diff</button>
          </div>
          ${renderDiffRows(rows)}
        </div>`;
      }
      const dr = D.diffResult;
      let changes = dr.changes;
      if (d.sev) changes = changes.filter(c => d.sev === 'breaking' ? c.severity === 'breaking' : c.kind === d.sev);
      if (d.q) changes = changes.filter(c => (c.path + c.detail).toLowerCase().includes(d.q.toLowerCase()));
      const selChange = dr.changes.find(c => c.id === d.sel);
      return `<div class="diff-layout">
        <div class="card pad diff-controls">
          <div class="row-gap wrap">
            <label class="fld-inline"><span>Base</span><select data-vinput="diff-base">${['v1.2.0', 'v1.3.0', 'v1.4.2'].map(v => `<option ${d.base === v ? 'selected' : ''}>PayCore API ${v}</option>`).join('')}</select></label>
            <button class="btn btn-icon" data-vact="diff-swap" data-tip="Swap sides" aria-label="Swap">${icon('refresh')}</button>
            <label class="fld-inline"><span>Updated</span><select data-vinput="diff-target">${['v1.3.0', 'v1.4.2', 'v2.0.0-rc1'].map(v => `<option ${d.target === v ? 'selected' : ''}>PayCore API ${v}</option>`).join('')}</select></label>
            <button class="btn btn-primary" data-vact="diff-run">${icon('compare')} Compare</button>
            <span class="st-flex"></span>
            <button class="btn ${d.mode === 'responses' ? 'btn-primary' : ''}" data-vact="diff-mode-resp">${icon('clock')} Response diff (from history)</button>
          </div>
          ${d.ran ? `<div class="diff-stats">
            ${[['added', dr.stats.added, 'ok'], ['removed', dr.stats.removed, 'error'], ['changed', dr.stats.changed, 'warn'], ['BREAKING', dr.stats.breaking, 'breaking']].map(s => `<button class="stat-pill sp-${s[2]} ${d.sev === s[0].toLowerCase() ? 'sel' : ''}" data-vact="diff-sev" data-s="${s[0].toLowerCase()}"><b>${s[1]}</b><span>${s[0]}</span></button>`).join('')}
            <div class="side-search">${icon('search')}<input placeholder="Filter changes…" value="${esc(d.q)}" data-vinput="diff-q"></div>
          </div>` : ''}
        </div>
        ${d.ran ? `<div class="diff-cols">
          <div class="diff-list card">
            ${changes.map(c => `<button class="diff-row sev-${c.severity} ${d.sel === c.id ? 'active' : ''}" data-vact="diff-sel" data-id="${c.id}">
              <span class="sev-dot"></span>
              <div><b class="mono">${esc(c.path)}</b><p class="dim small">${esc(c.detail)}</p></div>
              ${RVH.badge(c.kind, c.severity === 'breaking' ? 'error' : c.severity === 'warn' ? 'warn' : 'info')}
            </button>`).join('') || '<p class="empty-line">No changes match the current filter.</p>'}
          </div>
          <div class="diff-detail card pad">
            <h4>Detail ${selChange ? '· ' + esc(selChange.path) : ''}</h4>
            ${!selChange ? '<p class="dim">Pick a change to inspect the exact delta.</p>' : `
              ${RVH.badge('severity: ' + selChange.severity, selChange.severity === 'breaking' ? 'error' : 'warn')}
              <p>${esc(selChange.detail)}</p>
              ${selChange.diff ? renderDiffRows(selChange.diff.map(l => ({ t: l[0], s: l[1] }))) : `<pre class="codeview mono small">${esc(unifiedStub(selChange))}</pre>`}
              ${selChange.severity === 'breaking' ? RVH.banner('error', '<b>Breaking.</b> Consumers relying on this contract will fail after deploy. Gate this release behind a major version bump.') : ''}`}
          </div>
        </div>` : RVH.emptyState('compare', 'Ready to compare', 'Choose two versions and hit Compare — breaking-change analysis runs locally.')}
      </div>`;
    },
    after(h) { bindViewActs2(h); }
  });
  function sampleBodyFor(h) { return h && h.status === 200 && h.size > 500 ? D.responses.ok.body : h ? D.responses.notFound.body : ''; }
  function renderDiffRows(rows) {
    return `<div class="diff-view mono">${rows.filter(r => r.t !== 'ctx' || true).map((r, i) => `<div class="dl dl-${r.t}"><span class="dl-sign">${r.t === 'add' ? '+' : r.t === 'del' ? '−' : ' '}</span><span class="dl-ln">${i + 1}</span><span class="dl-txt">${esc(r.s)}</span></div>`).join('')}</div>`;
  }
  function unifiedStub(c) {
    if (c.kind === 'added') return '+ ' + c.path + '\n+   (new)';
    if (c.kind === 'removed') return '− ' + c.path + '\n−   (gone in target)';
    return '~ ' + c.path + '\n~   modified in target spec';
  }
  VI2['diff-base'] = el => { ui.diff.base = el.value.split(' ')[1] || el.value; };
  VI2['diff-target'] = el => { ui.diff.target = el.value.split(' ')[1] || el.value; };
  VI2['diff-q'] = el => { ui.diff.q = el.value; refreshDiff(); };
  VI2['diff-ra'] = el => { ui.diff.respA = el.value; A.renderMain(); };
  VI2['diff-rb'] = el => { ui.diff.respB = el.value; A.renderMain(); };
  function refreshDiff() { const host = A.root.querySelector('.view-host'); if (host && S.view === 'diff') { const v = A.VIEWS.diff; host.innerHTML = v.render(); v.after(host); } }
  V2['diff-run'] = () => { ui.diff.ran = true; A.renderMain(); A.toast('Compared specs — 2 breaking changes detected', { type: 'warn' }); };
  V2['diff-swap'] = () => { const d = ui.diff; [d.base, d.target] = [d.target, d.base]; A.renderMain(); };
  V2['diff-sev'] = el => { const s = el.dataset.s; ui.diff.sev = ui.diff.sev === s ? '' : s; refreshDiff(); };
  V2['diff-sel'] = el => { ui.diff.sel = el.dataset.id; refreshDiff(); };
  V2['diff-mode-resp'] = () => { ui.diff.mode = 'responses'; A.renderMain(); };
  V2['diff-mode-specs'] = () => { ui.diff.mode = 'specs'; A.renderMain(); };

  ui.jwt = ui.jwt || { token: '', decoded: null, err: null };
  A.regView({
    id: 'jwttool', title: 'JWT Inspector', icon: 'key', group: 'Specs & Tooling',
    render() {
      const j = ui.jwt;
      return `<div class="jwt-layout">
        <div class="card pad">
          <h4>${icon('key')} Paste a token</h4>
          <textarea class="mono jwt-in" rows="4" placeholder="eyJhbGciOi…" data-vinput="jwt-token">${esc(j.token)}</textarea>
          <div class="editor-tools"><button class="btn btn-primary" data-vact="jwt-decode">${icon('search')} Decode</button><button class="btn" data-vact="jwt-sample">Load sample token</button><button class="btn" data-vact="jwt-clear">Clear</button></div>
          ${j.err ? RVH.banner('error', '<b>Not a valid JWT.</b> Expected three base64url segments joined by dots.') : ''}
        </div>
        ${j.decoded ? `<div class="jwt-cols">
          <div class="card pad"><h4>Header <span class="mono dim small">alg: ${j.decoded.h.alg}</span></h4><div class="insp-kv">${Object.entries(j.decoded.h).map(([k, v]) => `<span>${k}</span><b class="mono">${esc(String(v))}</b>`).join('')}</div></div>
          <div class="card pad"><h4>Payload claims</h4>
            <div class="insp-kv">${Object.entries(j.decoded.p).map(([k, v]) => `<span>${esc(k)}</span><b class="mono">${esc(typeof v === 'object' ? JSON.stringify(v) : String(v)).slice(0, 44)}</b>`).join('')}</div>
            <div class="validity-bar"><i style="--pct:${j.decoded.pct}%"></i><span>issued ${new Date(j.decoded.p.iat * 1000).toLocaleDateString()} → expires ${new Date(j.decoded.p.exp * 1000).toLocaleString()} · ${j.decoded.pct > 100 ? '<b class="bad">EXPIRED</b>' : j.decoded.pct + '% of lifetime elapsed'}</span></div>
          </div>
          <div class="card pad"><h4>Signature</h4>
            <pre class="mono small sig">${esc(j.decoded.sig)}</pre>
            ${RVH.banner('info', 'Local verification (RS256 via JWKS) ships later — decode-only today. Nothing leaves this machine.')}
          </div>
        </div>` : ''}
      </div>`;
    },
    after(h) { bindViewActs2(h); }
  });
  VI2['jwt-token'] = el => ui.jwt.token = el.value;
  V2['jwt-sample'] = () => { ui.jwt.token = D.jwtSample; A.renderMain(); };
  V2['jwt-clear'] = () => { ui.jwt = { token: '', decoded: null, err: null }; A.renderMain(); };
  V2['jwt-decode'] = () => {
    const parts = ui.jwt.token.trim().split('.');
    if (parts.length !== 3) { ui.jwt.decoded = null; ui.jwt.err = true; A.renderMain(); return; }
    const dec = (s) => { try { return JSON.parse(atob(s.replace(/-/g, '+').replace(/_/g, '/') + '==='.slice(0, (4 - s.length % 4) % 4))); } catch (e) { return {}; } };
    const h = dec(parts[0]), p = dec(parts[1]);
    const pct = p.iat && p.exp ? Math.min(999, Math.round((Date.now() / 1000 - p.iat) / Math.max(1, p.exp - p.iat) * 100)) : 0;
    ui.jwt.err = false; ui.jwt.decoded = { h, p, pct, sig: parts[2].slice(0, 46) + '…' };
    A.renderMain();
  };

  ui.git = ui.git || { selFile: null, msg: '', resolving: false };
  A.regView({
    id: 'git', title: 'Git Panel', icon: 'git', group: 'Ops',
    render() {
      const g = D.gitState;
      const staged = g.changes.filter(c => c.staged && !c.conflict);
      const unstaged = g.changes.filter(c => !c.staged && !c.conflict);
      const conflicts = g.changes.filter(c => c.conflict);
      const sel = g.changes.find(c => c.file === ui.git.selFile) || g.changes[0];
      return `<div class="git-layout">
        <aside class="git-files card">
          <div class="side-head">Working tree <span class="btab-count">${g.changes.length}</span></div>
          ${conflicts.length ? `<div class="git-group conflict">${icon('warn')} Unmerged — resolve before commit</div>` : ''}
          ${conflicts.map(c => fileRow(c, true)).join('')}
          <div class="git-group">Staged (${staged.length})</div>
          ${staged.map(c => fileRow(c)).join('')}
          <div class="git-group">Unstaged (${unstaged.length})</div>
          ${unstaged.map(c => fileRow(c)).join('')}
          <div class="commit-box">
            <textarea id="git-msg" rows="3" placeholder="Commit message (feat(scope): summary)" aria-label="Commit message">${esc(ui.git.msg)}</textarea>
            <div class="editor-tools"><label class="check"><input type="checkbox"> amend last</label><label class="check"><input type="checkbox"> sign-off</label></div>
            <button class="btn btn-primary wide" data-vact="git-commit" ${staged.length === 0 ? 'disabled' : ''}>${icon('commit')} Commit ${staged.length} file${staged.length === 1 ? '' : 's'}</button>
          </div>
        </aside>
        <section class="git-center">
          <div class="card pad diff-card">
            <div class="row-between wrap">
              <h4>${icon('compare')} Diff — <span class="mono">${esc(sel?.file || '')}</span> ${sel ? RVH.badge(sel.status, sel.conflict ? 'error' : sel.status === 'A' ? 'ok' : 'warn') : ''}</h4>
              <div class="row-gap"><button class="btn btn-sm" data-vact="git-stage-all">Stage all</button><button class="btn btn-sm" data-vact="git-discard">Discard…</button></div>
            </div>
            ${sel && sel.conflict ? conflictView(g.conflictFile) : renderDiffRows(diffLines(oldContentFor(sel), newContentFor(sel)))}
          </div>
          <div class="card pad">
            <h4>Recent commits — ${esc(g.branch)}</h4>
            <div class="commits">${g.commits.map(c => `<div class="commit-row"><span class="graph-dot"></span><code class="mono">${c.hash}</code><b>${esc(c.msg)}</b><span class="dim">${esc(c.who)} · ${c.when}</span></div>`).join('')}</div>
          </div>
        </section>
        <aside class="git-inspector">
          <div class="card pad"><h4>Branch</h4>
            <div class="branch-chip big">${icon('branch')} <strong>${esc(g.branch)}</strong></div>
            <div class="insp-kv"><span>upstream</span><b class="mono">${esc(g.upstream)}</b><span>ahead</span><b>${g.ahead}</b><span>behind</span><b>${g.behind}</b></div>
            <div class="row-gap wrap">
              <button class="btn btn-sm" data-action="git-sync" data-kind="fetch">${icon('refresh')} Fetch</button>
              <button class="btn btn-sm" data-action="git-sync" data-kind="pull">${icon('download')} Pull</button>
              <button class="btn btn-sm btn-primary" data-action="git-sync" data-kind="push">${icon('export')} Push</button>
            </div>
          </div>
          <div class="card pad"><h4>All branches</h4>
            ${g.branches.map(b => `<div class="health-row"><i class="led ${b === g.branch ? 'led-ok' : 'led-muted'}"></i><span class="mono">${esc(b)}</span>${b === g.branch ? RVH.badge('HEAD', 'primary') : ''}</div>`).join('')}
          </div>
          <div class="card pad privacy-note">${icon('shield')} <p>Everything here reads the plain-text worktree. No network calls unless you press Fetch/Pull/Push.</p></div>
        </aside>
      </div>`;
    },
    after(h) {
      bindViewActs2(h);
      const msgEl = document.getElementById('git-msg');
      msgEl && msgEl.addEventListener('input', () => ui.git.msg = msgEl.value);
    }
  });
  function fileRow(c, conflict) {
    return `<div class="git-file ${conflict ? 'is-conflict' : ''} ${ui.git.selFile === c.file ? 'sel' : ''}" data-vact="git-file" data-f="${esc(c.file)}" role="button" tabindex="0">
      <span class="gs gs-${c.status}">${c.status}</span><span class="mono ellip">${esc(c.file)}</span>
      <span class="mono dim tiny">${c.lines}</span>
      ${c.staged && !conflict ? `<button class="btn btn-icon btn-xs" data-vact="git-unstage" data-f="${esc(c.file)}" data-tip="Unstage" aria-label="Unstage">${icon('x')}</button>` : conflict ? `<button class="btn btn-tiny" data-vact="git-resolve" data-f="${esc(c.file)}">Resolve</button>` : `<button class="btn btn-icon btn-xs" data-vact="git-stage" data-f="${esc(c.file)}" data-tip="Stage" aria-label="Stage">${icon('plus')}</button>`}
    </div>`;
  }
  function oldContentFor(c) {
    if (!c) return '';
    const base = { M: ['variables:', '  baseUrl: https://api.paycore.dev/v1', '  limit: 25', 'auth:', '  bearer: {{accessToken}}'], D: ['# payments.spec.yaml', '- name: list returns 200', '  op: equals', '  expected: "200"'], A: ['', '', ''] }[c.status] || [''];
    return c.status === 'A' ? '' : base.join('\n');
  }
  function newContentFor(c) {
    if (!c) return '';
    if (c.conflict) return ['variables:', '<<<<<<< HEAD', '  baseUrl: https://api.paycore.dev/v1', '=======', '  baseUrl: https://staging.paycore.io/v1', '>>>>>>> origin/main', '  limit: 25'].join('\n');
    if (c.status === 'D') return '';
    if (c.status === 'A') return ['name: send-test-event', 'method: POST', 'request:', '  url: "{{baseUrl}}/v1/webhooks/{{webhookId}}/test"', 'auth: inherit', 'scripts:', '  post: |', '    reqly.test("queued", () => reqly.expect(reqly.response.status).toBe(202))'].join('\n');
    return ['variables:', '  baseUrl: https://api.paycore.dev/v1', '  limit: 50', 'auth:', '  bearer: {{accessToken}}', '  refreshSkewMs: 45000'].join('\n');
  }
  function conflictView(cf) {
    return `<div class="merge-view">
      ${RVH.banner('error', '<b>Merge conflict</b> — incoming change from <code>origin/main</code> touches the same lines. Choose a side or merge manually.')}
      <div class="merge-cols">
        <div><h6>Ours — ${esc(D.gitState.branch)}</h6><pre class="codeview mono small ours">${esc(cf.ours.join('\n'))}</pre><button class="btn btn-sm" data-vact="merge-take" data-side="ours">Take ours</button></div>
        <div><h6>Theirs — origin/main</h6><pre class="codeview mono small theirs">${esc(cf.theirs.join('\n'))}</pre><button class="btn btn-sm" data-vact="merge-take" data-side="theirs">Take theirs</button></div>
      </div>
      <h6 style="margin-top:10px">Manual result (editable)</h6>
      <textarea class="mono merge-editor" rows="6">${esc(cf.ours.concat(cf.theirs.slice(1)).join('\n'))}</textarea>
      <div class="row-gap" style="margin-top:8px">
        <button class="btn btn-primary" data-vact="merge-done">Mark resolved & stage</button>
        <button class="btn" data-vact="merge-abort">Abort merge</button>
      </div>
    </div>`;
  }
  V2['git-file'] = el => { ui.git.selFile = el.dataset.f; A.renderMain(); };
  V2['git-stage'] = el => { const c = D.gitState.changes.find(x => x.file === el.dataset.f); if (c) c.staged = true; A.renderMain(); toastOk('Staged ' + el.dataset.f.split('/').pop()); };
  V2['git-unstage'] = el => { const c = D.gitState.changes.find(x => x.file === el.dataset.f); if (c) c.staged = false; A.renderMain(); };
  V2['git-stage-all'] = () => { D.gitState.changes.forEach(c => { if (!c.conflict) c.staged = true; }); A.renderMain(); toastOk('All conflict-free files staged'); };
  V2['git-discard'] = async () => { if (await A.confirmModal('Discard changes', 'Revert unstaged modifications in <code>collections/payments/get-payment-by-id.json</code>? This loses local edits (Git-clean files only).', 'Discard', true)) { toastOk('Changes discarded'); } };
  V2['git-commit'] = () => {
    const msg = ui.git.msg || 'chore: update workspace';
    D.gitState.commits.unshift({ hash: Math.random().toString(16).slice(2, 9), msg, who: 'you', when: 'just now' });
    D.gitState.ahead++;
    D.gitState.changes = D.gitState.changes.filter(c => !c.staged);
    ui.git.msg = ''; A.renderMain();
    A.toast(`Committed "${msg}" — branch ahead by ${D.gitState.ahead}`, { type: 'success', action: { label: 'Push now', fn: () => A.toast('Pushed to origin/main', { type: 'success' }) } });
  };
  V2['git-resolve'] = () => { ui.git.selFile = 'environments/dev.yaml'; A.renderMain(); };
  V2['merge-take'] = el => {
    const cf = D.gitState.conflictFile;
    const ta = document.querySelector('.merge-editor');
    ta.value = (el.dataset.side === 'ours' ? cf.ours : cf.theirs).join('\n');
    A.toast('Applied ' + el.dataset.side + ' version into manual editor', { type: 'info' });
  };
  V2['merge-done'] = () => {
    const cs = D.gitState.changes.find(c => c.conflict);
    if (cs) { cs.conflict = false; cs.status = 'M'; cs.lines = '+3 −2'; }
    A.renderMain(); A.toast('Conflict resolved — file staged for commit', { type: 'success' });
  };
  V2['merge-abort'] = async () => { if (await A.confirmModal('Abort merge', 'Return to pre-merge state? All conflict markers disappear and pulled commits are rolled back.', 'Abort merge', true)) A.toast('Merge aborted — tree restored', { type: 'info' }); };

  ui.settings = ui.settings || { section: 'appearance' };
  const SET_SECTIONS = [['appearance', 'Appearance', 'eye'], ['editor', 'Editor', 'edit'], ['network', 'Network', 'rest'], ['storage', 'Storage', 'server'], ['privacy', 'Privacy', 'shield'], ['shortcuts', 'Shortcuts', 'grid'], ['about', 'About', 'info']];
  A.regSidebar('settings', () => `<div class="side-block"><div class="side-head">Settings</div>
    ${SET_SECTIONS.map(s => `<button class="side-link ${ui.settings.section === s[0] ? 'active' : ''}" data-vact="set-nav" data-id="${s[0]}">${icon(s[2])} ${s[1]}</button>`).join('')}</div>`);
  A.regView({
    id: 'settings', title: 'Settings', icon: 'gear', group: '',
    render() {
      const sec = ui.settings.section;
      const bodies = {
        appearance: `<div class="card pad"><h4>Theme</h4>
          <div class="theme-cards">
            ${['light', 'dark'].map(th => `<button class="theme-card theme-preview-${th} ${document.documentElement.dataset.theme === th ? 'sel' : ''}" data-vact="set-theme" data-th="${th}"><div class="tp-bar"></div><div class="tp-side"></div><div class="tp-main"></div><span>${th}</span></button>`).join('')}
          </div>
          <h4 style="margin-top:14px">Density</h4>
          ${RVH.segBtns('set-density', ['compact', 'cozy'], document.documentElement.dataset.density || 'compact')}
          <h4 style="margin-top:14px">Accent</h4><div class="accent-row">#ff6f52 Reqly coral — brand accent locked across themes</div></div>`,
        editor: `<div class="card pad"><h4>Editor</h4>
          <label class="fld"><span>Font size</span><input type="range" min="11" max="18" value="${getComputedStyle(document.documentElement).getPropertyValue('--fs-mono-pt') ? '13' : '13'}" data-vact-none></label>
          <label class="fld"><span>Tab size</span><select><option>2</option><option selected>4</option></select></label>
          <label class="check"><input type="checkbox" checked> word wrap</label>
          <label class="check"><input type="checkbox"> minimap</label>
          <label class="check"><input type="checkbox" checked> format on save (json/yaml canonical)</label></div>`,
        network: `<div class="card pad"><h4>Network defaults</h4>
          ${RVH.table(['Setting', 'Value'], [['Timeout', '30 000 ms'], ['Follow redirects', 'up to 10'], ['SSL verification', 'strict'], ['HTTP version', 'negotiated (h2/h1.1)'], ['Retry policy', '2× fixed 400ms on 429/502/503/504 + network errors'], ['Proxy', 'system (disabled in demo)']])}
          <p class="dim small">${icon('info')} Per-request overrides live in Request settings (⋯ menu in REST builder).</p></div>`,
        storage: `<div class="card pad"><h4>Local storage</h4>
          ${[['Workspaces (plain-text)', 82, '12.4 MB'], ['History SQLite + FTS5', 34, '38 MB'], ['Cookie jars', 8, '1.1 MB'], ['Secret cache (keychain)', 3, '4 KB']].map(r => `<div class="usage-row"><span>${r[0]}</span><div class="progress"><i style="width:${r[1]}%"></i></div><em class="mono">${r[2]}</em></div>`).join('')}
          <div class="row-gap wrap" style="margin-top:12px">
            <button class="btn btn-sm btn-danger-ghost" data-vact="clear-hist-db">Clear history DB</button>
            <button class="btn btn-sm btn-danger-ghost" data-vact="vacuum">Vacuum database</button>
            <button class="btn btn-sm btn-danger-ghost" data-vact="factory-reset">Factory reset app data…</button>
          </div></div>`,
        privacy: `<div class="card pad privacy-hero">${icon('shield', 'big-shield')}<h3>Zero telemetry. Always.</h3>
          <p>Reqly has no analytics SDK, no crash reporter, no account system. Requests, responses, secrets and cookies never leave your machine. The networking stack makes zero outbound connections except the requests you explicitly send.</p>
          <div class="insp-kv"><span>telemetry opt-out</span><b>hardcoded — not configurable</b><span>crash reporting</span><b>none</b><span>update check</span><b>manual only</b><span>accounts</span><b>none — local-first forever</b></div></div>`,
        shortcuts: `<div class="card pad"><h4>Keyboard</h4>
          ${RVH.table(['Keys', 'Action'], [['⌘K', 'Command palette'], ['⌘↵', 'Send request'], ['⌘S', 'Save request'], ['⌘B', 'Toggle sidebar'], ['Esc', 'Close overlay / cancel in-flight request'], ['?', 'Shortcut cheat-sheet']])}
          <p class="dim small">Custom rebindings arrive with the plugins milestone.</p></div>`,
        about: `<div class="card pad about-card">
          <div class="row-gap">${icon('logo', 'welcome-logo')}<div><h3>Reqly</h3><p class="sub">v1.4.2-demo · Go core · Wails v3 shell</p></div></div>
          <div class="row-gap wrap" style="margin:12px 0"><button class="btn" data-vact="update-check">${icon('refresh')} Check for updates</button><span data-role="upd-state" class="dim">last checked: today</span></div>
          <p class="dim small">MIT licensed. Built local-first & Git-native: collections, environments, requests and tests are plain-text files you can read, diff and commit.</p>
          <button class="btn btn-danger-ghost" data-vact="factory-reset">Reset all application data…</button></div>`
      };
      return `<div class="settings-layout"><div class="settings-body">${bodies[sec] || ''}</div></div>`;
    },
    after(h) { bindViewActs2(h); }
  });
  V2['set-nav'] = el => { ui.settings.section = el.dataset.id; A.renderMain(); A.renderSidebar(); };
  V2['set-theme'] = el => { document.documentElement.dataset.theme = el.dataset.th; A.renderMain(); A.toast('Theme applied — ' + el.dataset.th, { type: 'success' }); };
  V2['clear-hist-db'] = async () => { if (await A.confirmModal('Clear history DB', 'Delete all rows from <code>history.db</code> (WAL mode)? FTS index rebuilds automatically.', 'Clear', true)) toastOk('history.db cleared'); };
  V2['vacuum'] = () => { A.toast('VACUUM complete — reclaimed 4.2 MB', { type: 'success' }); };
  V2['update-check'] = () => { const st = document.querySelector('[data-role="upd-state"]'); if (st) { st.innerHTML = '<span class="spinner sm"></span> checking…'; setTimeout(() => st.textContent = 'You are on the latest version (manual check only — no phoning home)', 900); } };
  V2['factory-reset'] = async () => { if (await A.confirmModal('Factory reset', 'This wipes local settings, cached tokens and demo state from this browser. Workspaces on disk stay untouched.', 'Wipe & reload', true)) { try { localStorage.removeItem('reqly_demo_first_run'); } catch (e) {} location.reload(); } };
})();
