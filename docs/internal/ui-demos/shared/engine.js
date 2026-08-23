// oxlint-disable anti-slop/no-runtime-typeof
(function () {
  const D = window.REQLY_DATA;
  const store = (() => {
    try { const s = window.localStorage; s.getItem('__t'); return s; }
    catch (e) { const m = {}; return { getItem: k => (k in m ? m[k] : null), setItem: (k, v) => { m[k] = String(v); }, removeItem: k => { delete m[k]; } }; }
  })();

  const esc = (s) => String(s == null ? '' : s).replace(/[&<>"']/g, (c) => ({ '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;' }[c]));
  const uid = () => 'id' + Math.random().toString(36).slice(2, 9);
  const pad = (n) => String(n).padStart(2, '0');
  const fmtClock = (iso) => { const d = new Date(iso); return `${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`; };
  const fmtDate = (iso) => new Date(iso).toLocaleString(undefined, { month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' });
  const ago = (iso) => {
    const s = Math.max(1, Math.floor((Date.now() - new Date(iso).getTime()) / 1000));
    if (s < 60) return s + 's ago';
    if (s < 3600) return Math.floor(s / 60) + 'm ago';
    if (s < 86400) return Math.floor(s / 3600) + 'h ago';
    return Math.floor(s / 86400) + 'd ago';
  };
  const bytes = (n) => n >= 1048576 ? (n / 1048576).toFixed(2) + ' MB' : n >= 1024 ? (n / 1024).toFixed(1) + ' KB' : n + ' B';
  const ms = (n) => n >= 1000 ? (n / 1000).toFixed(2) + ' s' : Math.round(n) + ' ms';

  const ICONS = {
    logo: '<path d="M12 2l9 5v10l-9 5-9-5V7z"/><path d="M12 22V12M12 12L3 7M12 12l9-5"/>',
    home: '<path d="M3 10l9-7 9 7v10a1 1 0 01-1 1h-5v-6h-6v6H4a1 1 0 01-1-1z"/>',
    folder: '<path d="M3 7a2 2 0 012-2h4l2 2h8a2 2 0 012 2v9a2 2 0 01-2 2H5a2 2 0 01-2-2z"/>',
    'folder-open': '<path d="M3 7a2 2 0 012-2h4l2 2h8a2 2 0 012 2v1H5.5L3 18z"/><path d="M3 18l2.5-8H21l-2.5 8a1 1 0 01-1 .9H4a1 1 0 01-1-.9z"/>',
    rest: '<circle cx="12" cy="12" r="9"/><path d="M3 12h18M12 3a15 15 0 010 18M12 3a15 15 0 000 18"/>',
    graphql: '<path d="M12 3l8 4.5v9L12 21l-8-4.5v-9z"/><circle cx="12" cy="3" r="1.6"/><circle cx="20" cy="7.5" r="1.6"/><circle cx="20" cy="16.5" r="1.6"/><circle cx="12" cy="21" r="1.6"/><circle cx="4" cy="16.5" r="1.6"/><circle cx="4" cy="7.5" r="1.6"/>',
    bolt: '<path d="M13 2L4 14h6l-1 8 9-12h-6z"/>',
    plug: '<path d="M9 3v6m6-6v6M5 9h14v3a7 7 0 01-7 7 7 7 0 01-7-7zM12 19v3"/>',
    signal: '<path d="M2 20h.01M7 20v-4m5 4V10m5 10V4"/>',
    doc: '<path d="M6 2h9l5 5v13a2 2 0 01-2 2H6a2 2 0 01-2-2V4a2 2 0 012-2z"/><path d="M14 2v6h6"/>',
    layers: '<path d="M12 2l10 6-10 6L2 8z"/><path d="M2 14l10 6 10-6"/>',
    clock: '<circle cx="12" cy="12" r="9"/><path d="M12 7v5l3 3"/>',
    flask: '<path d="M9 3h6M10 3v6L4.5 19a2 2 0 001.8 3h11.4a2 2 0 001.8-3L14 9V3"/><path d="M7.5 15h9"/>',
    terminal: '<rect x="2" y="4" width="20" height="16" rx="2"/><path d="M6 9l3 3-3 3M12 15h6"/>',
    book: '<path d="M4 4a2 2 0 012-2h5v20H6a2 2 0 00-2-2z"/><path d="M20 4a2 2 0 00-2-2h-5v20h5a2 2 0 002-2z"/>',
    'import': '<path d="M12 3v12m0 0l-4-4m4 4l4-4"/><path d="M4 17v2a2 2 0 002 2h12a2 2 0 002-2v-2"/>',
    export: '<path d="M12 15V3m0 0L8 7m4-4l4 4"/><path d="M4 17v2a2 2 0 002 2h12a2 2 0 002-2v-2"/>',
    braces: '<path d="M8 3H7a2 2 0 00-2 2v4a2 2 0 01-2 2 2 2 0 012 2v4a2 2 0 002 2h1M16 3h1a2 2 0 012 2v4a2 2 0 002 2 2 2 0 00-2 2v4a2 2 0 01-2 2h-1"/>',
    server: '<rect x="3" y="3" width="18" height="7" rx="1.5"/><rect x="3" y="14" width="18" height="7" rx="1.5"/><path d="M7 6.5h.01M7 17.5h.01"/>',
    compare: '<circle cx="6" cy="6" r="3"/><circle cx="18" cy="18" r="3"/><path d="M6 9v3a3 3 0 003 3h3M18 15v-3a3 3 0 00-3-3h-3"/>',
    branch: '<circle cx="6" cy="5" r="2.5"/><circle cx="6" cy="19" r="2.5"/><circle cx="18" cy="9" r="2.5"/><path d="M6 7.5v9M18 11.5a6 6 0 01-6 6"/>',
    gear: '<circle cx="12" cy="12" r="3.2"/><path d="M19 12a7 7 0 00-.1-1.2l2-1.5-2-3.4-2.3 1a7 7 0 00-2-1.2L14.2 3h-4l-.4 2.7a7 7 0 00-2 1.2l-2.3-1-2 3.4 2 1.5A7 7 0 005 12a7 7 0 00.1 1.2l-2 1.5 2 3.4 2.3-1a7 7 0 002 1.2l.4 2.7h4l.4-2.7a7 7 0 002-1.2l2.3 1 2-3.4-2-1.5A7 7 0 0019 12z"/>',
    key: '<circle cx="7" cy="14" r="4"/><path d="M10 11L20 3m-4 4l2.5 2.5M13 8l2.5 2.5"/>',
    zap: '<path d="M13 2L4 14h6l-1 8 9-12h-6z"/>',
    search: '<circle cx="11" cy="11" r="7"/><path d="M21 21l-4.3-4.3"/>',
    plus: '<path d="M12 5v14M5 12h14"/>',
    x: '<path d="M6 6l12 12M18 6L6 18"/>',
    'chev-r': '<path d="M9 5l7 7-7 7"/>',
    'chev-d': '<path d="M5 9l7 7 7-7"/>',
    dots: '<circle cx="5" cy="12" r="1.6"/><circle cx="12" cy="12" r="1.6"/><circle cx="19" cy="12" r="1.6"/>',
    play: '<path d="M7 4l13 8-13 8z"/>',
    stop: '<rect x="6" y="6" width="12" height="12" rx="1.5"/>',
    copy: '<rect x="9" y="9" width="12" height="12" rx="2"/><path d="M5 15H4a2 2 0 01-2-2V4a2 2 0 012-2h9a2 2 0 012 2v1"/>',
    download: '<path d="M12 3v12m0 0l-4-4m4 4l4-4M4 21h16"/>',
    trash: '<path d="M4 7h16M9 7V4h6v3M6 7l1 13a2 2 0 002 2h6a2 2 0 002-2l1-13"/>',
    edit: '<path d="M4 20h4L20.5 7.5a2.1 2.1 0 00-3-3L5 17z"/><path d="M13.5 6.5l3 3"/>',
    check: '<path d="M4 12l6 6L20 6"/>',
    warn: '<path d="M12 3l10 18H2z"/><path d="M12 10v5m0 3h.01"/>',
    info: '<circle cx="12" cy="12" r="9"/><path d="M12 11v5m0-8h.01"/>',
    lock: '<rect x="5" y="11" width="14" height="9" rx="2"/><path d="M8 11V7a4 4 0 018 0v4"/>',
    eye: '<path d="M2 12s3.5-7 10-7 10 7 10 7-3.5 7-10 7-10-7-10-7z"/><circle cx="12" cy="12" r="3"/>',
    'eye-off': '<path d="M3 3l18 18M10 6a10 10 0 012 -.2c6.5 0 10 6.2 10 6.2a17 17 0 01-3 3.6M6.6 6.9C3.8 8.8 2 12 2 12s3.5 7 10 7a10 10 0 004 -.8M9.9 9.9a3 3 0 004.2 4.2"/>',
    refresh: '<path d="M21 12a9 9 0 11-2.6-6.4"/><path d="M21 3v6h-6"/>',
    git: '<circle cx="6" cy="5" r="2.5"/><circle cx="6" cy="19" r="2.5"/><circle cx="18" cy="9" r="2.5"/><path d="M6 7.5v9M18 11.5a6 6 0 01-6 6"/>',
    commit: '<circle cx="12" cy="12" r="4"/><path d="M2 12h6m8 0h6"/>',
    grid: '<rect x="3" y="3" width="7" height="7" rx="1.5"/><rect x="14" y="3" width="7" height="7" rx="1.5"/><rect x="3" y="14" width="7" height="7" rx="1.5"/><rect x="14" y="14" width="7" height="7" rx="1.5"/>',
    sliders: '<path d="M4 8h10M18 8h2M4 16h2M10 16h10"/><circle cx="16" cy="8" r="2"/><circle cx="8" cy="16" r="2"/>',
    shield: '<path d="M12 3l8 3v6c0 5-3.5 8-8 9-4.5-1-8-4-8-9V6z"/><path d="M9 12l2 2 4-4"/>'
  };
  const icon = (n, cls) => `<svg class="ic ${cls || ''}" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round">${ICONS[n] || ICONS.info}</svg>`;

  const METHOD_COLORS = { GET: 'ok', POST: 'info', PUT: 'warn', PATCH: 'warn', DELETE: 'error', HEAD: 'muted', OPTIONS: 'muted' };
  const methodChip = (m) => `<span class="chip chip-method m-${METHOD_COLORS[m] || 'muted'}">${m}</span>`;
  const statusChip = (s) => {
    if (!s) return `<span class="chip chip-status st-none">—</span>`;
    const cls = s === 0 ? 'st-timeout' : s < 300 ? 'st-ok' : s < 400 ? 'st-redir' : s < 500 ? 'st-warn' : 'st-error';
    return `<span class="chip chip-status ${cls}">${s === 0 ? 'TIMEOUT' : s}${s >= 200 && s < 300 ? '' : ''}</span>`;
  };

  const S = {
    view: 'rest', workspaceId: D.workspaces[0].id, envId: 'env-dev',
    tabs: [], activeTab: null,
    tree: JSON.parse(JSON.stringify(D.collectionTree)),
    emptyWorkspace: false, firstRunDone: !!store.getItem('reqly_demo_first_run'),
    consoleEntries: [{ t: Date.now(), level: 'sys', msg: 'reqly sandbox ready — reqly.* API attached' }],
    dockTab: 'console', sidebarVisible: true,
    simulateConflict: false, simulatePortConflict: false,
    ui: {}, lastLatency: null, sending: false,
    secretsRevealed: {}, expandedSchema: {}
  };
  let OPTS = { shell: 'ide', brand: 'Reqly', responseMode: 'split', theme: 'dark' };
  let ROOT = null;

  const envById = () => D.environments.find(e => e.id === S.envId);
  const wsById = () => D.workspaces.find(w => w.id === S.workspaceId);

  function resolveVars(text, depth = 0) {
    if (!text || depth > 4) return text || '';
    const env = envById();
    const pool = {};
    D.globalScope.vars.forEach(v => pool[v.key] = v.value);
    if (env) env.vars.forEach(v => pool[v.key] = v.value);
    env && env.secrets.forEach(v => pool[v.key] = '[SECRET]');
    pool.limit = pool.limit || '25';
    return String(text).replace(/\{\{(\w+)\}\}/g, (m, k) => k in pool ? resolveVars(pool[k], depth + 1) : m)
      .replace(/\{\{\$timestamp\}\}/g, Math.floor(Date.now() / 1000))
      .replace(/\{\{\$uuid\}\}/g, (window.crypto && crypto.randomUUID) ? crypto.randomUUID() : 'u-' + Math.random().toString(36).slice(2))
      .replace(/\{\{\$isoTimestamp\}\}/g, new Date().toISOString())
      .replace(/\{\{\$randomInt\}\}/g, String(Math.floor(Math.random() * 9000) + 1000))
      .replace(/\{\{\$(\w+)\}\}/g, '<dynamic:$1>');
  }
  const undefinedVars = (text) => {
    const out = []; const seen = {};
    String(text || '').replace(/\{\{(\w+)\}\}/g, (m, k) => { if (!(k in seen)) { seen[k] = 1; out.push(k); } return m; });
    const known = new Set(D.globalScope.vars.map(v => v.key).concat((envById() || { vars: [], secrets: [] }).vars.map(v => v.key)));
    return out.filter(k => !known.has(k));
  };

  function highlight(code, lang) {
    let h = esc(code);
    if (lang === 'json') {
      h = h.replace(/("(\\u[\da-fA-F]{4}|\\[^u]|[^\\"])*")(\s*:)?|\b(true|false|null)\b|-?\d+(\.\d+)?([eE][+-]?\d+)?/g, (m) => {
        let cls = 'tok-num';
        if (/^"/.test(m)) cls = /:$/.test(m) ? 'tok-key' : 'tok-str';
        if (/^(true|false|null)/.test(m)) cls = 'tok-lit';
        return `<span class="${cls}">${m}</span>`;
      });
    } else if (lang === 'js') {
      h = h.replace(/(\/\/[^\n]*)|(&quot;(?:[^&]|&(?!quot;))*?&quot;|'(?:[^'])*?'|`(?:[^`])*?`)|\b(const|let|var|function|await|async|return|if|else|new|throw|import|from|export)\b|\b(reqly|console|fetch|JSON|Math|Date|process|require)\b/g,
        (m, cm, str, kw, obj) => cm ? `<span class="tok-com">${cm}</span>` : str ? `<span class="tok-str">${str}</span>` : kw ? `<span class="tok-kw">${kw}</span>` : `<span class="tok-obj">${obj}</span>`);
    } else if (lang === 'xml') {
      h = h.replace(/(&lt;!--[\s\S]*?--&gt;)|(\/?[a-zA-Z:][\w:-]*)&gt;|([\w:-]+)=(&quot;[^&]*?&quot;)/g,
        (m, com, tag, attr, val) => com ? `<span class="tok-com">${com}</span>` : tag ? `<span class="tok-tag">${tag}&gt;</span>` : `<span class="tok-attr">${attr}=</span><span class="tok-str">${val}</span>`);
    } else if (lang === 'http') {
      h = h.replace(/^([A-Z]+)([^\n]*)$/gm, '<span class="tok-kw">$1</span><span class="tok-str">$2</span>')
        .replace(/^([A-Za-z-]+):(.*)$/gm, '<span class="tok-key">$1</span>:<span class="tok-str">$2</span>');
    }
    return h;
  }

  function jsonPathGet(obj, path) {
    try {
      const parts = path.replace(/^\$\.?/, '').split(/[.[\]]+/).filter(Boolean);
      let cur = obj;
      for (const p of parts) {
        if (Array.isArray(cur)) { const i = parseInt(p); if (isNaN(i)) return undefined; cur = cur[i]; }
        else if (cur && typeof cur === 'object') cur = cur[p];
        else return undefined;
      }
      return cur;
    } catch (e) { return undefined; }
  }

  function diffLines(aText, bText) {
    const a = String(aText || '').split('\n'), b = String(bText || '').split('\n');
    const n = a.length, m = b.length;
    const dp = Array.from({ length: n + 1 }, () => new Uint16Array(m + 1));
    for (let i = n - 1; i >= 0; i--) for (let j = m - 1; j >= 0; j--)
      dp[i][j] = a[i] === b[j] ? dp[i + 1][j + 1] + 1 : Math.max(dp[i + 1][j], dp[i][j + 1]);
    const out = []; let i = 0, j = 0;
    while (i < n && j < m) {
      if (a[i] === b[j]) { out.push({ t: 'ctx', s: a[i] }); i++; j++; }
      else if (dp[i + 1][j] >= dp[i][j + 1]) { out.push({ t: 'del', s: a[i++] }); }
      else { out.push({ t: 'add', s: b[j++] }); }
    }
    while (i < n) out.push({ t: 'del', s: a[i++] });
    while (j < m) out.push({ t: 'add', s: b[j++] });
    return out;
  }

  let MODALS = [];
  function modal({ title, body, footer, wide, onClose }) {
    const wrap = document.createElement('div');
    wrap.className = 'modal-backdrop';
    wrap.innerHTML = `<div class="modal ${wide ? 'modal-wide' : ''}" role="dialog" aria-modal="true" aria-label="${esc(title)}">
      <div class="modal-head"><h3>${esc(title)}</h3><button class="btn btn-icon" data-close aria-label="Close">${icon('x')}</button></div>
      <div class="modal-body">${body}</div>${footer ? `<div class="modal-foot">${footer}</div>` : ''}</div>`;
    ROOT.querySelector('.overlays').appendChild(wrap);
    const api = { el: wrap, close: () => { wrap.remove(); MODALS = MODALS.filter(m => m !== api); onClose && onClose(); } };
    wrap.addEventListener('mousedown', (e) => { if (e.target === wrap) api.close(); });
    wrap.querySelector('[data-close]').addEventListener('click', api.close);
    MODALS.push(api);
    const f = wrap.querySelector('input,textarea,select'); f && f.focus();
    return api;
  }
  function closeTopModal() { MODALS.length && MODALS[MODALS.length - 1].close(); }

  function confirmModal(title, msg, label, danger, extraBody) {
    return new Promise((res) => {
      const m = modal({
        title, body: `<p class="modal-msg">${msg}</p>${extraBody || ''}`,
        footer: `<button class="btn" data-no>Cancel</button><button class="btn ${danger ? 'btn-danger' : 'btn-primary'}" data-yes>${esc(label)}</button>`
      });
      m.el.querySelector('[data-no]').onclick = () => { m.close(); res(false); };
      m.el.querySelector('[data-yes]').onclick = () => { m.close(); res(true); };
    });
  }
  function formModal({ title, fields, submitLabel, wide, onSubmit }) {
    const bodyHtml = fields.map(f => {
      const id = 'ff_' + f.name;
      const input = f.type === 'select'
        ? `<select id="${id}" name="${f.name}" ${f.disabled ? 'disabled' : ''}>${(f.options || []).map(o => `<option value="${esc(o)}" ${o === f.value ? 'selected' : ''}>${esc(o)}</option>`).join('')}</select>`
        : f.type === 'textarea'
          ? `<textarea id="${id}" name="${f.name}" rows="${f.rows || 5}" placeholder="${esc(f.placeholder || '')}">${esc(f.value || '')}</textarea>`
          : `<input id="${id}" name="${f.name}" type="${f.type || 'text'}" value="${esc(f.value == null ? '' : f.value)}" placeholder="${esc(f.placeholder || '')}" ${f.disabled ? 'disabled' : ''}>`;
      return `<div class="field"><label for="${id}">${esc(f.label)}</label>${input}${f.hint ? `<p class="field-hint">${f.hint}</p>` : ''}</div>`;
    }).join('');
    const m = modal({
      title, wide, body: `<form class="modal-form" data-form>${bodyHtml}<button class="btn btn-primary" type="submit">${esc(submitLabel)}</button></form>`
    });
    m.el.querySelector('[data-form]').addEventListener('submit', (e) => {
      e.preventDefault();
      const vals = {}; new FormData(e.target).forEach((v, k) => vals[k] = v);
      fields.filter(f => f.type === 'checkbox').forEach(f => vals[f.name] = m.el.querySelector(`[name=${f.name}]`).checked);
      m.close(); onSubmit && onSubmit(vals);
    });
    return m;
  }

  const TOASTS = [];
  function toast(msg, opt = {}) {
    const t = document.createElement('div');
    t.className = `toast toast-${opt.type || 'info'}`;
    t.innerHTML = `<span class="toast-dot"></span><div class="toast-body"><span>${msg}</span>${opt.action ? `<button class="btn btn-tiny" data-tact>${esc(opt.action.label)}</button>` : ''}</div><button class="btn btn-icon btn-tiny" data-tx>${icon('x')}</button>`;
    ROOT.querySelector('.toasts').appendChild(t);
    requestAnimationFrame(() => t.classList.add('in'));
    const kill = () => { t.classList.remove('in'); setTimeout(() => t.remove(), 220); };
    t.querySelector('[data-tx]').onclick = kill;
    if (opt.action) t.querySelector('[data-tact]').onclick = () => { opt.action.fn(); kill(); };
    setTimeout(kill, opt.timeout || 4200);
    TOASTS.push(t);
  }

  let MENU = null;
  function menu(x, y, items) {
    closeMenu();
    const el = document.createElement('div');
    el.className = 'ctx-menu';
    el.innerHTML = items.map((it, i) => it === '-' ? '<hr>' :
      `<button data-i="${i}" class="${it.danger ? 'danger' : ''}">${it.icon ? icon(it.icon) : ''}<span>${esc(it.label)}</span>${it.hint ? `<kbd>${esc(it.hint)}</kbd>` : ''}</button>`).join('');
    document.body.appendChild(el);
    const r = el.getBoundingClientRect();
    el.style.left = Math.min(x, innerWidth - r.width - 8) + 'px';
    el.style.top = Math.min(y, innerHeight - r.height - 8) + 'px';
    el.addEventListener('click', (e) => { const b = e.target.closest('[data-i]'); if (b) { closeMenu(); items[+b.dataset.i].fn(); } });
    MENU = el;
  }
  function closeMenu() { MENU && MENU.remove(); MENU = null; }

  function paletteCommands() {
    const cmds = [];
    Object.values(VIEWS).forEach(v => cmds.push({ label: 'Go to: ' + v.title, hint: v.group, icon: v.icon, fn: () => navigate(v.id) }));
    cmds.push(
      { label: 'Simulate: first launch experience', icon: 'zap', fn: () => { S.firstRunDone = false; welcomeFlow(); } },
      { label: 'Simulate: empty workspace', icon: 'folder', fn: () => { S.emptyWorkspace = true; refresh(); toast('Workspace emptied — collections cleared', { type: 'warn' }); } },
      { label: 'Simulate: restore populated workspace', icon: 'refresh', fn: () => { S.emptyWorkspace = false; refresh(); toast('Sample collections restored', { type: 'success' }); } },
      { label: 'Simulate: save conflict on next save', icon: 'warn', fn: () => { S.simulateConflict = true; toast('Next ⌘S will hit a save conflict', { type: 'warn' }); } },
      { label: 'Simulate: mock-server port already in use', icon: 'server', fn: () => toast('Bind failed: port 9090 already in use by reqly-mock (stale)', { type: 'error', timeout: 6000 }) },
      { label: 'Simulate: request timeout (open slow endpoint)', icon: 'clock', fn: () => { const t = findOrMakeTab('r-slow'); setActive(t.id); navigate('rest'); setTimeout(() => sendRequest(t), 250); } },
      { label: 'Toggle secret visibility (all)', icon: 'eye', fn: () => { const anyShown = Object.values(S.secretsRevealed).some(Boolean); S.secretsRevealed = {}; if (!anyShown) S.secretsRevealed._all = true; refresh(); } },
      { label: 'Clear console output', icon: 'trash', fn: () => { S.consoleEntries = []; renderDock(); } },
      { label: 'Show keyboard shortcuts', icon: 'grid', fn: shortcutsModal },
      { label: 'Reset demo state (reload)', icon: 'refresh', fn: () => location.reload() }
    );
    return cmds;
  }

  let PAL = null;
  function palette() {
    if (PAL) { PAL.close(); PAL = null; return; }
    const m = modal({
      title: '', wide: true,
      body: `<div class="palette"><div class="palette-input">${icon('search')}<input placeholder="Type a command… views, simulations, tools" aria-label="Command search"></div><div class="palette-list" role="listbox"></div><div class="palette-foot"><kbd>↑↓</kbd> navigate <kbd>↵</kbd> run <kbd>esc</kbd> close</div></div>`
    });
    m.el.classList.add('palette-wrap');
    m.el.querySelector('.modal').classList.add('modal-palette');
    const input = m.el.querySelector('input');
    const list = m.el.querySelector('.palette-list');
    let sel = 0, filtered = [];
    const draw = () => {
      filtered = paletteCommands().filter(c => c.label.toLowerCase().includes(input.value.toLowerCase()));
      sel = Math.min(sel, Math.max(0, filtered.length - 1));
      list.innerHTML = filtered.map((c, i) => `<button class="pal-item ${i === sel ? 'sel' : ''}" data-i="${i}">${icon(c.icon || 'chev-r')}<span>${esc(c.label)}</span><em>${esc(c.hint || '')}</em></button>`).join('') || '<p class="empty-line">No matching commands</p>';
      const s = list.querySelector('.sel'); try { s && s.scrollIntoView && s.scrollIntoView({ block: 'nearest' }); } catch (e) {}
    };
    input.addEventListener('input', () => { sel = 0; draw(); });
    input.addEventListener('keydown', (e) => {
      if (e.key === 'ArrowDown') { sel = Math.min(sel + 1, filtered.length - 1); draw(); e.preventDefault(); }
      if (e.key === 'ArrowUp') { sel = Math.max(sel - 1, 0); draw(); e.preventDefault(); }
      if (e.key === 'Enter' && filtered[sel]) { const c = filtered[sel]; PAL = null; m.close(); c.fn(); }
    });
    list.addEventListener('click', (e) => { const b = e.target.closest('[data-i]'); if (b) { const c = filtered[+b.dataset.i]; PAL = null; m.close(); c.fn(); } });
    draw();
    input.focus();
    PAL = m;
    const origClose = m.close.bind(m); m.close = () => { origClose(); PAL = null; };
  }

  function shortcutsModal() {
    modal({
      title: 'Keyboard shortcuts', wide: true,
      body: `<table class="kv-table shortcuts-table"><tbody>
        ${[['⌘/Ctrl K', 'Command palette'], ['⌘/Ctrl ↵', 'Send active request'], ['⌘/Ctrl S', 'Save active request'], ['Esc', 'Close overlay / cancel'], ['⌘/Ctrl B', 'Toggle sidebar'], ['?', 'This dialog (when not editing)']].map(r => `<tr><td><kbd>${r[0]}</kbd></td><td>${r[1]}</td></tr>`).join('')}
      </tbody></table>`
    });
  }

  function welcomeFlow() {
    if (S.firstRunDone) return;
    modal({
      title: 'Welcome to Reqly', wide: true,
      body: `<div class="welcome-grid">
        <div class="welcome-hero">${icon('logo', 'welcome-logo')}<h4>Local-first. Git-native. Zero telemetry.</h4>
        <p>This workspace mirrors plain-text files in <code>~/dev/paycore-ws</code>. Everything you see lives in your repository.</p></div>
        <ol class="welcome-steps">
          <li><strong>Pick an environment</strong> — variables like <code>{{baseUrl}}</code> resolve through 6 scopes.</li>
          <li><strong>Send your first request</strong> — “List payments” is preloaded.</li>
          <li><strong>Explore the surface</strong> — GraphQL, gRPC, WebSocket, SSE, SOAP, mocks, diffs, tests…</li>
          <li><strong>Press ⌘K</strong> — jump anywhere, or trigger realistic states (timeouts, conflicts).</li>
        </ol></div>`,
      footer: `<label class="check"><input type="checkbox" checked data-noshow> Don't show again</label><button class="btn btn-primary" data-go>Start exploring</button>`
    }).el.addEventListener('click', (e) => {
      if (e.target.closest('[data-go]')) {
        if (e.target.closest('.modal-foot').querySelector('[data-noshow]').checked) { S.firstRunDone = true; store.setItem('reqly_demo_first_run', '1'); }
        closeTopModal();
      }
    });
    S.firstRunDone = true;
  }

  function newTabFromNode(node) {
    const det = D.requestDetail[node.id] || JSON.parse(JSON.stringify(D.defaultDetail));
    const dup = S.tabs.find(t => t.nodeId === node.id);
    if (dup) { S.activeTab = dup.id; return dup; }
    const tab = {
      id: uid(), nodeId: node.id, name: node.name, method: node.method, url: node.url,
      protocol: node.protocol || 'rest',
      params: JSON.parse(JSON.stringify(det.params)), headers: JSON.parse(JSON.stringify(det.headers)),
      auth: JSON.parse(JSON.stringify(det.auth)), body: JSON.parse(JSON.stringify(det.body)),
      scripts: JSON.parse(JSON.stringify(det.scripts)), savedName: node.name, dirty: !node.saved,
      response: { state: 'idle' }, builderTab: 'params', resTab: 'body', resMode: 'pretty', resSearch: '', jsonPath: '', pathHit: ''
    };
    S.tabs.push(tab);
    S.activeTab = tab.id;
    return tab;
  }
  const findOrMakeTab = (nodeId) => {
    const found = walkTree(S.tree).find(n => n.id === nodeId);
    const ex = S.tabs.find(t => t.nodeId === nodeId);
    if (ex) { S.activeTab = ex.id; return ex; }
    return newTabFromNode(found);
  };
  const activeTab = () => S.tabs.find(t => t.id === S.activeTab);
  function walkTree(nodes, parent = null, out = []) {
    nodes.forEach(n => { out.push({ ...n, _parent: parent }); if (n.children) walkTree(n.children, n, out); });
    return out;
  }

  function sendRequest(tab) {
    if (!tab || S.sending) return;
    tab.response = { state: 'loading', startedAt: Date.now(), mode: tab.resMode };
    S.sending = true; renderStatus(); callHook('renderResponse');
    const url = resolveVars(tab.url);
    let kind = 'ok';
    if (/\/slow/.test(url)) kind = 'slow';
    else if (/fail/.test(url)) kind = 'fail';
    else if (/teapot/.test(url)) kind = 'teapot';
    else if (/oauth\/token/.test(url)) kind = 'token';
    else if (/pay_missing|payments\/[a-z_]*missing/.test(url)) kind = 'notFound';
    else if (/graphql/.test(url)) kind = 'graphql';
    else if (tab.method === 'POST' && /payments$/.test(url.split('?')[0])) kind = 'created';

    const finish = () => {
      S.sending = false;
      if (kind === 'graphql') {
        tab.response = {
          state: 'done', status: 200, statusText: 'OK', time: 173, size: 1240, url,
          headers: { 'content-type': 'application/json' }, cookies: [],
          body: JSON.stringify({ data: { customer: { id: 'cus_NffrFe', email: 'nina@ferrite.io', orders: [ { id: 'ord_8841', total: 4250, payments: [{ id: 'pay_1Q84xa', status: 'SUCCEEDED' }] } ] } } }, null, 2)
        };
      } else {
        const r = D.responses[kind];
        const jitter = Math.round(r.time * (0.85 + Math.random() * 0.3));
        tab.response = { ...r, state: 'done', time: jitter, size: kind === 'large' ? r.size : r.size + Math.floor(Math.random() * 40), url };
      }
      S.lastLatency = tab.response.time;
      D.history.unshift({ id: uid(), ts: new Date().toISOString(), method: tab.method, url: tab.url, resolved: url, status: tab.response.status, time: tab.response.time, size: tab.response.size, env: envById()?.name || '—', requestId: tab.nodeId });
      log('sys', `${tab.method} ${url} → ${tab.response.status} in ${ms(tab.response.time)}`);
      runScriptSim(tab, 'post');
      callHook('historyChanged');
      callHook('renderResponse'); renderStatus(); refreshTabsBar();
    };

    if (kind === 'slow') {
      tab.response.timeoutTimer = setTimeout(() => {
        S.sending = false;
        tab.response = { state: 'timeout', url, message: 'No response within 30 000 ms — connection aborted locally.' };
        D.history.unshift({ id: uid(), ts: new Date().toISOString(), method: tab.method, url: tab.url, resolved: url, status: 0, time: 30000, size: 0, env: envById()?.name, requestId: tab.nodeId, timeout: true });
        log('err', 'ETIMEDOUT after 30s — ' + url);
        callHook('renderResponse'); renderStatus();
      }, 6000);
    } else {
      const latency = kind === 'fail' ? 380 : 140 + Math.random() * 260;
      tab.response.cancelTimer = setTimeout(() => { tab.response.cancelTimer = null; finish(); }, latency * (kind === 'large' ? 6 : 1));
    }
    runScriptSim(tab, 'pre');
    callHook('renderResponse');
  }
  function cancelRequest(tab) {
    if (tab?.response?.cancelTimer) clearTimeout(tab.response.cancelTimer);
    if (tab?.response?.timeoutTimer) clearTimeout(tab.response.timeoutTimer);
    if (tab?.response?.state === 'loading') {
      tab.response = { state: 'cancelled', url: tab.url };
      S.sending = false;
      log('warn', 'request cancelled by user');
      callHook('renderResponse'); renderStatus();
    }
  }
  function runScriptSim(tab, phase) {
    const script = (tab.scripts || {})[phase];
    if (!script || !script.trim()) return;
    const lines = script.split('\n').filter(l => l.trim());
    lines.forEach((ln, i) => {
      setTimeout(() => {
        if (/reqly\.console\.log/.test(ln)) log('log', ln.replace(/.*reqly\.console\.log\(|['");]/g, '').trim() || '(value)');
        else if (/reqly\.expect|reqly\.test/.test(ln)) log('pass', 'assertion ✓ ' + ln.trim().slice(0, 72));
        else if (/setSecret/i.test(ln)) log('sec', 'secret stored: accessToken = [SECRET]');
        else log('log', '→ ' + ln.trim().slice(0, 80));
      }, 120 + i * 160);
    });
  }
  function log(level, msg) {
    S.consoleEntries.push({ t: Date.now(), level, msg });
    if (S.consoleEntries.length > 200) S.consoleEntries.shift();
    renderDock();
  }

  function saveActive() {
    const t = activeTab(); if (!t) return;
    if (S.simulateConflict) {
      S.simulateConflict = false;
      modal({
        title: 'Save conflict', wide: true,
        body: `<div class="conflict-box">${icon('warn', 'conflict-ic')}<p><code>${esc(fileOf(t))}</code> changed on disk while you were editing (commit <code>7b21de0</code> pulled “fix(auth): refresh race”).</p>
        <div class="conflict-cols"><div><h5>Your version</h5><pre>${esc(t.url)}</pre></div><div><h5>Disk version</h5><pre>{{baseUrl}}/v1/payments?limit={{limit}}&amp;safe=true</pre></div></div></div>`,
        footer: `<button class="btn" data-x>Compare in diff viewer</button><button class="btn" data-r>Reload from disk (discard mine)</button><button class="btn btn-primary" data-o>Overwrite disk version</button>`
      }).el.addEventListener('click', (e) => {
        if (e.target.closest('[data-o]')) { closeTopModal(); doSave(); }
        if (e.target.closest('[data-r]')) { closeTopModal(); t.dirty = false; refresh(); toast('Reloaded request from disk', { type: 'info' }); }
        if (e.target.closest('[data-x]')) { closeTopModal(); navigate('diff'); toast('Opened workspace diff viewer', { type: 'info' }); }
      });
      return;
    }
    doSave();
  }
  function doSave() {
    const t = activeTab(); if (!t) return;
    t.dirty = false; t.savedName = t.name;
    const n = walkTree(S.tree).find(x => x.id === t.nodeId);
    if (n) { n.name = t.name; n.saved = true; }
    renderSidebar(); refreshTabsBar();
    toast(`Saved ${fileOf(t)} — formatted, atomic write (0644)`, { type: 'success', action: { label: 'View diff', fn: () => navigate('git') } });
  }
  const fileOf = (t) => `collections/${(walkTree(S.tree).find(x => x.id === t.nodeId)?._parent?.name || 'root').toLowerCase().replace(/\W+/g, '-')}/${t.name.toLowerCase().replace(/\W+/g, '-')}.json`;

  function navigate(view) {
    if (!VIEWS[view]) return;
    S.view = view;
    refresh();
  }

  function refresh() {
    renderNavRail();
    updateInspector();
    renderTitleblock();
    renderTopbar();
    renderSidebar(); renderMain(); refreshTabsBar(); renderStatus(); renderDock();
  }

  const VIEWS = {}; const SIDEBARS = {}; const HOOKS = {};
  function regView(v) { VIEWS[v.id] = v; }
  function regSidebar(id, fn) { SIDEBARS[id] = fn; }
  function callHook(n, ...a) { HOOKS[n] && HOOKS[n](...a); }

  function renderSidebar() {
    const html = (() => { const fn = SIDEBARS[S.view]; return fn ? fn() : sidebarCollections(); })();
    const host = ROOT.querySelector('[data-mount="sidebar"]');
    if (host) { host.style.display = S.sidebarVisible ? '' : 'none'; host.innerHTML = html; if (typeof window.bindViewActs2 === 'function') window.bindViewActs2(host); }
    const dr = ROOT.querySelector('[data-mount="drawer"]');
    if (dr) { dr.innerHTML = html; if (typeof window.bindViewActs2 === 'function') window.bindViewActs2(dr); }
  }
  function sidebarCollections() {
    if (S.emptyWorkspace) {
      return `<div class="side-block"><div class="side-head">Collections</div>
      <div class="empty empty-sm">${icon('folder')}<p>No collections yet</p><button class="btn btn-primary btn-sm" data-action="new-request">Create first request</button><p class="empty-sub">or drop a Postman / Insomnia / Bruno export anywhere</p></div></div>`;
    }
    const q = (S.ui.treeQuery || '').toLowerCase();
    const filterTree = (nodes) => nodes.map(n => {
      if (n.type === 'folder') { const kids = filterTree(n.children || []); return kids.length ? { ...n, children: kids, open: q ? true : n.open } : (n.name.toLowerCase().includes(q) ? { ...n, children: [] } : null); }
      return !q || n.name.toLowerCase().includes(q) || n.url.toLowerCase().includes(q) ? n : null;
    }).filter(Boolean);
    const shown = q ? filterTree(S.tree) : S.tree;
    return `<div class="side-block side-tree-block">
      <div class="side-head">Collections<button class="btn btn-icon btn-xs" data-action="tree-menu" data-tip="Collection actions" aria-label="Collection actions">${icon('dots')}</button></div>
      <div class="side-search">${icon('search')}<input placeholder="Filter requests…" value="${esc(S.ui.treeQuery || '')}" data-action-input="tree-query" aria-label="Filter requests"></div>
      <div class="tree">${shown.map(n => treeNode(n, 0)).join('')}</div>
      <div class="side-foot-actions">
        <button class="btn btn-sm" data-action="new-folder">${icon('folder')} Folder</button>
        <button class="btn btn-sm" data-action="new-request">${icon('plus')} Request</button>
      </div>
      ${OPTS.shell === 'atlas' && D.gitState.changes.length ? `<div class="scm-mini">
        <div class="side-head scm-head">${icon('branch')} Source control <span class="scm-count">${D.gitState.changes.length}</span></div>
        ${D.gitState.changes.slice(0, 5).map(c => `<button class="scm-row" data-action="nav" data-view="git" data-tip="${esc(c.file)}"><span class="gs gs-${c.status}">${c.status}</span><span class="ellip">${esc(c.file.split('/').pop())}</span><em class="mono">${esc(c.lines.split(' ')[0])}</em></button>`).join('')}
        <button class="scm-all" data-action="nav" data-view="git">Open Git panel ${icon('chev-r')}</button>
      </div>` : ''}</div>`;
  }
  function treeNode(n, depth) {
    const hasKids = n.children && n.children.length;
    if (n.type === 'folder') {
      return `<div class="tree-node" style="--d:${depth}">
        <div class="tree-row row-folder" role="button" tabindex="0" data-action="toggle-folder" data-id="${n.id}" aria-expanded="${!!n.open}">
          <span class="tw ${n.open ? 'open' : ''}">${icon(hasKids ? 'chev-d' : 'chev-r')}</span>${icon(n.auth ? 'lock' : 'folder', 'tfolder')}<span class="tname">${esc(n.name)}</span>
          <span class="row-more" role="button" tabindex="0" data-action="node-menu" data-id="${n.id}" data-tip="Folder actions">${icon('dots')}</span></div>
        ${n.open !== false && hasKids ? `<div class="tree-kids">${n.children.map(c => treeNode(c, depth + 1)).join('')}</div>` : ''}
      </div>`;
    }
    const active = S.activeTab && S.tabs.find(t => t.id === S.activeTab)?.nodeId === n.id;
    return `<div class="tree-node" style="--d:${depth}">
      <div class="tree-row row-request ${active ? 'active' : ''}" role="button" tabindex="0" data-action="open-request" data-id="${n.id}">
        ${methodChip(n.method)}<span class="tname">${esc(n.name)}</span>
        ${!n.saved && !active ? '<span class="unsaved-dot" data-tip="Unsaved changes"></span>' : ''}
        <span class="row-more" role="button" tabindex="0" data-action="node-menu" data-id="${n.id}" data-tip="Request actions">${icon('dots')}</span></div>
    </div>`;
  }

  function renderMain() {
    const host = ROOT.querySelector('[data-mount="main"]');
    const v = VIEWS[S.view];
    const crumbs = breadcrumbHtml();
    host.innerHTML = `${crumbs ? `<div class="crumbs">${crumbs}</div>` : ''}<div class="view-host view-${S.view}">${v ? v.render() : '<p class="empty-line">Unknown view</p>'}</div>`;
    v.after && v.after(host);
  }
  function breadcrumbHtml() {
    const t = activeTab();
    if ((S.view === 'rest' || S.view === 'codegen') && t) {
      const node = walkTree(S.tree).find(n => n.id === t.nodeId);
      const parts = [wsById().name, node?._parent?.name, t.savedName || t.name].filter(Boolean);
      return parts.map((p, i) => `${i ? icon('chev-r', 'crumb-sep') : ''}<span class="${i === parts.length - 1 ? 'crumb-last' : ''}">${esc(p)}</span>`).join('');
    }
    return '';
  }

  function refreshTabsBar() {
    const host = ROOT.querySelector('[data-mount="tabs"]');
    if (!host) return;
    if (!S.tabs.length) { host.innerHTML = ''; host.style.display = 'none'; return; }
    host.style.display = '';
    host.innerHTML = S.tabs.map(t => `
      <button class="tab ${t.id === S.activeTab ? 'active' : ''}" data-action="focus-tab" data-id="${t.id}" role="tab" aria-selected="${t.id === S.activeTab}">
        ${methodChip(t.method)}<span class="tab-name">${esc(t.name)}</span>${t.dirty ? '<span class="dirty-dot" data-tip="Unsaved changes"></span>' : ''}
        <span class="tab-x" data-action="close-tab" data-id="${t.id}" role="button" aria-label="Close tab">${icon('x')}</span>
      </button>`).join('') + `<button class="btn btn-icon btn-xs tab-new" data-action="new-request" data-tip="New request (⌘N)" aria-label="New request">${icon('plus')}</button>`;
  }

  function renderStatus() {
    const host = ROOT.querySelector('[data-mount="status"]');
    if (!host) return;
    const env = envById();
    const ws = wsById();
    host.innerHTML = `
      <span class="st-item">${icon('branch')} ${esc(ws.branch)}</span>
      <span class="st-item ${env ? 'st-env-' + env.color : 'st-env-none'}">${icon('layers')} ${env ? esc(env.name) : 'no environment selected'}</span>
      <span class="st-item">${icon('doc')} ${walkTree(S.tree).length} requests</span>
      ${S.lastLatency != null ? `<span class="st-item">last ${ms(S.lastLatency)}</span>` : ''}
      <span class="st-flex"></span>
      <span class="st-item st-telemetry">${icon('shield')} Zero telemetry · local-first</span>
      <span class="st-item">v1.4.2-demo</span>`;
  }

  function renderDock() {
    const host = ROOT.querySelector('[data-mount="dock"]');
    if (!host) return;
    const tabs = [['console', 'Console'], ['tests', 'Test output'], ['network', 'Network']];
    host.innerHTML = `<div class="dock-tabs">${tabs.map(t => `<button class="dock-tab ${S.dockTab === t[0] ? 'active' : ''}" data-action="dock-tab" data-id="${t[0]}">${t[1]}</button>`).join('')}
      <span class="dock-count">${S.consoleEntries.length} lines</span>
      <button class="btn btn-icon btn-xs" data-action="clear-console" data-tip="Clear" aria-label="Clear console">${icon('trash')}</button></div>
      <div class="dock-body">${
        S.dockTab === 'console' ? (S.consoleEntries.map(e => `<div class="cline cl-${e.level}"><span class="cl-t">${fmtClock(new Date(e.t).toISOString())}</span><span class="cl-badge">${e.level}</span><span class="cl-msg">${esc(e.msg)}</span></div>`).join('') || '<p class="empty-line">Console is empty — run a request with pre/post scripts.</p>')
        : S.dockTab === 'tests' ? `<p class="empty-line">Run a suite in Test Runner — assertion stream lands here.</p>`
        : `<div class="netlog">${D.history.slice(0, 6).map(h => `<div class="net-row">${methodChip(h.method)}<span class="mono">${esc(h.resolved.slice(0, 52))}</span>${statusChip(h.status)}<span>${ms(h.time)}</span></div>`).join('')}</div>`
      }</div>`;
  }

  function renderNavRail() {
    const rail = ROOT.querySelector('[data-mount="rail"]');
    if (rail) {
      const labeled = OPTS.shell === 'atlas';
      rail.innerHTML = RAIL_GROUPS.map(g => `<div class="rail-group">${g.items.map(id => {
        const v = VIEWS[id]; if (!v) return '';
        return `<button class="rail-btn ${S.view === id ? 'active' : ''}" data-action="nav" data-view="${id}" data-tip="${esc(v.title)}" aria-label="${esc(v.title)}">${icon(v.icon)}${labeled ? `<span class="lbl">${esc(v.title)}</span>` : ''}</button>`;
      }).join('')}</div>`).join('') + `<div class="rail-group rail-bottom">
        <button class="rail-btn ${S.view === 'settings' ? 'active' : ''}" data-action="nav" data-view="settings" data-tip="Settings" aria-label="Settings">${icon('gear')}${labeled ? '<span class="lbl">Settings</span>' : ''}</button></div>`;
    }
    const sess = ROOT.querySelector('[data-mount="sessions"]');
    if (sess) sess.innerHTML = sessionsHtml();
  }
  const RAIL_GROUPS = [
    { items: ['overview', 'rest', 'graphql', 'grpc', 'websocket', 'sse', 'soap'] },
    { items: ['environments', 'history', 'scripting', 'tests', 'runner'] },
    { items: ['openapi', 'importexport', 'codegen', 'mocks', 'diff'] },
    { items: ['jwttool', 'git'] }
  ];

  function renderTopbar() {
    const tb = ROOT.querySelector('[data-mount="topbar"]');
    if (!tb) return;
    tb.innerHTML = topbarHtml();
  }
  function topbarHtml() {
    const ws = wsById(); const env = envById();
    return `
      <div class="tb-left">
        <button class="btn btn-icon ${S.sidebarVisible ? '' : 'is-off'}" data-action="toggle-sidebar" data-tip="Toggle sidebar (⌘B)" aria-label="Toggle sidebar">${icon('sliders')}</button>
        <div class="ws-pill" tabindex="0" data-action="ws-menu" data-tip="Switch workspace">
          ${icon('logo', 'ws-logo')}<span class="ws-name">${esc(ws.name)}</span>${icon('chev-d', 'chev')}
        </div>
      </div>
      <div class="tb-mid">
        <button class="cmdk-btn" data-action="palette" aria-label="Open command palette">${icon('search')}<span>Jump to anything…</span><kbd>⌘K</kbd></button>
      </div>
      <div class="tb-right">
        <div class="env-select" data-role="env-select">
          <button class="btn env-btn ${env ? 'env-' + env.color : 'env-none'}" data-action="env-menu" data-tip="Environment (6 scopes)">${icon('layers')}<span>${env ? esc(env.name) : 'No environment'}</span>${icon('chev-d', 'chev')}</button>
        </div>
        <button class="btn btn-icon" data-action="nav" data-view="settings" data-tip="Settings" aria-label="Settings">${icon('gear')}</button>
      </div>`;
  }
  function menuAnchor(e) {
    const el = (e.target instanceof Element ? e.target : null);
    const btn = el && el.closest('[data-action]');
    return btn ? btn.getBoundingClientRect() : (el ? el.getBoundingClientRect() : { left: e.clientX, bottom: e.clientY, right: e.clientX });
  }
  function envMenu(e) {
    const r = menuAnchor(e);
    const items = [
      ...D.environments.map(en => ({ label: (en.id === S.envId ? '● ' : '○ ') + en.name, icon: 'layers', fn: () => { S.envId = en.id; refresh(); toast(`Environment switched to ${en.name}`, { type: 'success' }); } })),
      '-',
      { label: 'Manage environments…', icon: 'sliders', fn: () => navigate('environments') },
      { label: 'No environment', icon: 'x', fn: () => { S.envId = null; refresh(); toast('No environment active — {{vars}} stay unresolved', { type: 'warn' }); } }
    ];
    menu(r.left, r.bottom + 4, items);
  }

  function openRequestNode(id) {
    const node = walkTree(S.tree).find(n => n.id === id);
    if (!node) return;
    if (node.protocol && node.protocol !== 'rest') {
      const destMap = { graphql: 'graphql', grpc: 'grpc', websocket: 'websocket', sse: 'sse', soap: 'soap' };
      newTabFromNode(node);
      navigate(destMap[node.protocol]);
      return;
    }
    newTabFromNode(node);
    navigate('rest');
    renderMain();
  }

  function nodeContextMenu(e, id) {
    const node = walkTree(S.tree).find(n => n.id === id);
    if (!node) return;
    const items = node.type === 'folder' ? [
      { label: 'New request here', icon: 'plus', fn: () => createRequestInFolder(node.id) },
      { label: 'New folder', icon: 'folder', fn: () => renameModal('New folder', '', (name) => { node.children.push({ type: 'folder', id: uid(), name, open: true, children: [] }); renderSidebar(); }) },
      { label: 'Rename…', icon: 'edit', fn: () => renameModal('Rename folder', node.name, (name) => { node.name = name; renderSidebar(); }) },
      '-', { label: 'Delete folder', icon: 'trash', danger: true, fn: async () => { if (await confirmModal('Delete folder', `Delete <b>${esc(node.name)}</b> and its contents? Files will be removed from the Git worktree.`, 'Delete', true)) { delNode(node.id); toast('Folder deleted (staged as deleted in Git)', { type: 'success' }); } } }
    ] : [
      { label: 'Open', icon: 'play', fn: () => openRequestNode(id) },
      { label: 'Duplicate request', icon: 'copy', fn: () => duplicateNode(id) },
      { label: 'Rename…', icon: 'edit', fn: () => renameModal('Rename request', node.name, (name) => { node.name = name; const t = S.tabs.find(t => t.nodeId === id); if (t) { t.name = name; t.dirty = true; refreshTabsBar(); } renderSidebar(); }) },
      { label: 'Copy resolved URL', icon: 'copy', fn: () => { navigator.clipboard?.writeText(resolveVars(node.url)); toast('Resolved URL copied', { type: 'success' }); } },
      { label: 'Generate code…', icon: 'braces', fn: () => { findOrMakeTab(id); navigate('codegen'); } },
      '-', { label: 'Delete request', icon: 'trash', danger: true, fn: async () => { if (await confirmModal('Delete request', `Delete <b>${esc(node.name)}</b>? The request file stays recoverable from Git history.`, 'Delete', true)) { delNode(id); toast('Request deleted (Git tracks deletion)', { type: 'success' }); } } }
    ];
    menu(e.clientX, e.clientY, items);
  }
  function delNode(id) {
    const rm = (ns) => ns.forEach((n, i) => { if (n.id === id) ns.splice(i, 1); else if (n.children) rm(n.children); });
    rm(S.tree);
    S.tabs = S.tabs.filter(t => t.nodeId !== id);
    if (!S.tabs.find(t => t.id === S.activeTab)) S.activeTab = S.tabs[0]?.id || null;
    renderSidebar(); refreshTabsBar(); renderMain();
  }
  function duplicateNode(id) {
    const node = walkTree(S.tree).find(n => n.id === id);
    const copy = JSON.parse(JSON.stringify(node)); copy.id = uid(); copy.name += ' copy'; copy.saved = false;
    const parent = node._parent ? node._parent.children : S.tree;
    const idx = parent.findIndex(n => n.id === id);
    parent.splice(idx + 1, 0, copy);
    renderSidebar(); toast('Duplicated — new request file staged as added', { type: 'success' });
  }
  function renameModal(title, initial, cb) {
    formModal({ title, submitLabel: title.startsWith('Rename') ? 'Rename' : 'Create', fields: [{ name: 'val', label: 'Name', value: initial }], onSubmit: (v) => v.val && cb(v.val) });
  }
  function createRequestInFolder(folderId) {
    formModal({
      title: 'New request', submitLabel: 'Create',
      fields: [
        { name: 'name', label: 'Name', value: 'New request' },
        { name: 'method', label: 'Method', type: 'select', options: ['GET', 'POST', 'PUT', 'PATCH', 'DELETE'] },
        { name: 'url', label: 'URL', value: '{{baseUrl}}/' }
      ],
      onSubmit: (v) => {
        const node = walkTree(S.tree).find(n => n.id === folderId) || { children: S.tree };
        node.children = node.children || [];
        node.children.push({ type: 'request', id: uid(), name: v.name, method: v.method, url: v.url, saved: false });
        if (node.open !== undefined) node.open = true;
        renderSidebar();
        toast('Request created (draft — unsaved)', { type: 'success' });
      }
    });
  }

  const kvRow = (scopeId, row, i, secret) => `<div class="kv-row ${row.enabled === false ? 'kv-disabled' : ''}">
    <button class="kv-check ${row.enabled === false ? 'off' : ''}" data-action="kv-toggle" data-scope="${scopeId}" data-i="${i}" data-tip="${row.enabled === false ? 'Enable' : 'Disable'} row" aria-label="Toggle row">${icon('check')}</button>
    <input class="kv-key mono" value="${esc(row.key)}" data-action-input="kv-key" data-scope="${scopeId}" data-i="${i}" placeholder="key" aria-label="Key">
    <input class="kv-val mono" type="${secret && !(S.secretsRevealed._all || S.secretsRevealed[scopeId + i]) ? 'password' : 'text'}" value="${esc(row.value)}" data-action-input="kv-val" data-scope="${scopeId}" data-i="${i}" placeholder="value" aria-label="Value">
    ${secret ? `<button class="btn btn-icon btn-xs" data-action="reveal-secret" data-scope="${scopeId}" data-i="${i}" data-tip="${secret && !(S.secretsRevealed._all || S.secretsRevealed[scopeId + i]) ? 'Reveal once' : 'Hide'}" aria-label="Toggle visibility">${icon(secret && !(S.secretsRevealed._all || S.secretsRevealed[scopeId + i]) ? 'eye' : 'eye-off')}</button>` : ''}
    <span class="drag-handle" data-tip="Drag to reorder">${icon('dots')}</span>
    <button class="btn btn-icon btn-xs kv-del" data-action="kv-del" data-scope="${scopeId}" data-i="${i}" data-tip="Remove row" aria-label="Remove row">${icon('trash')}</button>
  </div>`;
  const kvEditor = (scopeId, rows, secret) => `<div class="kv-editor" data-scope="${scopeId}">
    <div class="kv-head"><span></span><span>KEY</span><span>VALUE</span><span></span><span></span><span></span></div>
    ${rows.map((r, i) => kvRow(scopeId, r, i, secret)).join('')}
    <button class="kv-add" data-action="kv-add" data-scope="${scopeId}">${icon('plus')} Add row</button>
    ${rows.length === 0 ? '<p class="empty-line">This table is empty.</p>' : ''}
  </div>`;
  function getKvScope(scopeId) {
    const t = activeTab();
    if (scopeId === 'global') return D.globalScope.vars;
    if (scopeId === 'workspace') return D.workspaceScope.vars;
    if (scopeId === 'collection') return D.collectionScope.vars;
    if (scopeId.startsWith('env-vars-')) return D.environments.find(e => e.id === scopeId.slice(9)).vars;
    if (scopeId.startsWith('env-secret-')) return D.environments.find(e => e.id === scopeId.slice(11)).secrets;
    if (scopeId === 'form') return t.body.form || [];
    return t[scopeId] || [];
  }

  function largeBodyHtml() {
    const rows = Array.from({ length: 240 }, (_, i) => ({ i: i, id: 'tx_' + (10000 + i), amount: Math.round(Math.random() * 99999), currency: ['usd', 'eur', 'gbp'][i % 3], status: ['succeeded', 'pending', 'refunded'][i % 3], created: new Date(now - i * 36e5).toISOString() }));
    return { rows, note: `Previewing first 240 of 5 214 entries (${bytes(1048576)}) — virtualized.` };
  }
  const now = Date.now();

  function workspaceMenu(e) {
    const items = [
      ...D.workspaces.map(w => ({ label: (w.id === S.workspaceId ? '● ' : '○ ') + w.name + '  ·  ' + w.requests + ' reqs', icon: 'logo', fn: () => { S.workspaceId = w.id; refresh(); toast(`Switched to ${w.name} (${w.path})`, { type: 'success' }); } })),
      '-',
      { label: 'New workspace…', icon: 'plus', fn: () => formModal({ title: 'New workspace', submitLabel: 'Create', fields: [ { name: 'name', label: 'Name', value: '' }, { name: 'path', label: 'Local path (Git repository)', value: '~/dev/new-ws' } ], onSubmit: (v) => { D.workspaces.push({ id: uid(), name: v.name || 'Untitled', path: v.path, branch: 'main', requests: 0, envs: 0 }); toast('Workspace created — run `git init` or clone into it', { type: 'success' }); } }) },
      { label: 'Workspace settings…', icon: 'gear', fn: () => navigate('settings') },
      { label: 'Reveal in file manager', icon: 'folder', fn: () => toast(wsById().path + ' — local directory (demo)', { type: 'info' }) }
    ];
    const r = menuAnchor(e);
    menu(r.left, r.bottom + 4, items);
  }

  function sessionsHtml() {
    let n = 0;
    return RAIL_GROUPS.map(g => `<div class="ss-group">${g.items.map(id => {
      const v = VIEWS[id]; if (!v) return '';
      const key = (++n % 10);
      return `<button class="ss-item ${S.view === id ? 'active' : ''}" data-action="nav" data-view="${id}"><span class="ss-key mono">${key}</span><span class="ss-name">${esc(v.title)}</span></button>`;
    }).join('')}</div>`).join('');
  }

  function commitStripHtml() {
    const g = D.gitState;
    return `<div class="commit-strip" role="navigation" aria-label="Recent commits">
      <span class="cs-label">${icon('branch')} HEAD →</span>
      <div class="cs-track">${g.commits.slice(0, 6).map((c, i) => `<button class="cs-node ${i === 0 ? 'is-head' : ''}" data-action="commit-info" data-hash="${esc(c.hash)}" data-msg="${esc(c.msg)}" data-who="${esc(c.who)}" data-tip="${esc(c.hash)} · ${esc(c.msg)}"><i></i><span class="mono">${esc(c.hash)}</span></button>`).join('<span class="cs-line" aria-hidden="true"></span>')}</div>
      <button class="btn btn-xs cs-sync" data-action="git-sync" data-kind="fetch">sync ↑${g.ahead} ↓${g.behind}</button>
    </div>`;
  }

  function titleblockHtml() {
    const ids = Object.keys(VIEWS);
    const idx = Math.max(1, ids.indexOf(S.view) + 1);
    const env = envById();
    return `
      <table class="tb-grid" aria-label="Sheet title block"><tbody>
        <tr>
          <th scope="row">PROJECT</th><td>${esc(wsById().name)}</td>
          <th scope="row">REVISION</th><td class="mono">${esc(D.gitState.branch)}</td>
          <th scope="row">SHEET</th><td class="mono">${String(idx).padStart(2, '0')} / ${String(ids.length).padStart(2, '0')}</td>
        </tr>
        <tr>
          <th scope="row">ENVIRONMENT</th><td class="mono ${env ? 'env-t-' + env.color : ''}">${env ? esc(env.name) : '— none —'}</td>
          <th scope="row">DATE</th><td class="mono">${new Date().toISOString().slice(0, 10)}</td>
          <th scope="row">WORKTREE</th><td class="mono">${D.gitState.changes.length} changed · ↑${D.gitState.ahead}↓${D.gitState.behind}</td>
        </tr>
      </tbody></table>`;
  }
  function renderTitleblock() {
    const host = ROOT.querySelector('[data-mount="titleblock"]');
    if (host) host.innerHTML = titleblockHtml();
  }


  function branchbarHtml() {
    const g = D.gitState;
    return `<div class="branchbar">
      <button class="branch-chip" data-action="branch-menu">${icon('branch')}<strong>${esc(g.branch)}</strong>${icon('chev-d', 'chev')}</button>
      <span class="sync-pill" data-action="git-sync" data-tip="Fetch / pull / push">↑${g.ahead} ↓${g.behind} <em>with ${esc(g.upstream)}</em></span>
      <button class="btn btn-sm" data-action="git-sync" data-kind="pull">${icon('download')} Pull</button>
      <button class="btn btn-sm" data-action="git-sync" data-kind="push">${icon('export')} Push</button>
      <span class="st-flex"></span>
      <span class="mono repo-path">${esc(g.repo)}</span>
    </div>`;
  }

  function A_wsSlug() { return wsById().path.replace(/^~\//, '').replace(/[^a-z0-9-]/gi, ''); }

  const SHELLS = {
    workbench: () => `<div class="app wb theme-${OPTS.theme}">
      <header class="wb-title">
        <button class="btn btn-icon" data-action="toggle-wb-side" data-tip="Toggle explorer" aria-label="Toggle explorer">${icon('sliders')}</button>
        <div class="wb-id">${icon('logo')}<b>Reqly</b></div>
        <div class="ws-pill" tabindex="0" data-action="ws-menu" data-tip="Switch workspace"><span class="ws-name">${esc(wsById().name)}</span><small class="mono">${esc(wsById().path)}</small>${icon('chev-d', 'chev')}</div>
        <button class="cmdk-btn" data-action="palette" aria-label="Command palette">${icon('search')}<span>Search or jump…</span><kbd>⌘K</kbd></button>
        <span class="wb-spring"></span>
        <button class="branch-chip" data-action="branch-menu">${icon('branch')}<strong>${esc(D.gitState.branch)}</strong>${icon('chev-d', 'chev')}</button>
        <button class="btn env-btn" data-action="env-menu">${icon('layers')}<span>${envById() ? esc(envById().name) : 'No environment'}</span>${icon('chev-d', 'chev')}</button>
        <button class="btn btn-icon" data-action="nav" data-view="settings" data-tip="Settings" aria-label="Settings">${icon('gear')}</button>
      </header>
      <div class="wb-body">
        <aside class="wb-side" data-mount="sidebar"></aside>
        <div class="gutter gutter-side" data-split="side" role="separator" aria-orientation="vertical"></div>
        <section class="maincol">
          <div class="tabstrip wb-tabs" data-mount="tabs" role="tablist"></div>
          <div class="viewhost wb-viewhost" data-mount="main"></div>
        </section>
        <div class="gutter gutter-insp" data-split="inspector" role="separator" aria-orientation="vertical"></div>
        <aside class="wb-insp" data-mount="inspector"></aside>
      </div>
      ${commitStripHtml()}
      <footer class="statusbar wb-status" data-mount="status"></footer>
    </div>`,
    terminal: () => `<div class="app tt theme-${OPTS.theme}">
      <header class="tt-bar">
        <span class="tt-brand">${icon('logo')} REQ<b>LY</b><em>v1.4.2</em></span>
        <span class="tt-ws mono">${esc(wsById().name)}</span>
        <span class="st-flex"></span>
        <button class="btn env-btn tt-env" data-action="env-menu">${icon('layers')}<span>${envById() ? esc(envById().name) : 'no env'}</span></button>
        <button class="btn btn-icon" data-action="nav" data-view="settings" data-tip="Settings" aria-label="Settings">${icon('gear')}</button>
      </header>
      <div class="tt-cmdrow">
        <button class="tt-prompt" data-action="palette" aria-label="Open command line">
          <span class="ps1">reqly</span><span class="ps-path mono">~/${A_wsSlug()}</span><span class="ps-caret" aria-hidden="true"></span>
          <em>type to jump — views, simulations, states…</em><kbd>⌘K</kbd>
        </button>
        <button class="branch-chip tt-branch" data-action="branch-menu">${icon('branch')}${esc(D.gitState.branch)}<i class="led led-ok"></i></button>
      </div>
      <div class="tt-body">
        <nav class="tt-sessions" data-mount="sessions" aria-label="Views"></nav>
        <div class="gutter gutter-side" data-split="side" role="separator"></div>
        <section class="maincol">
          <div class="tabstrip tt-tabs" data-mount="tabs" role="tablist"></div>
          <div class="viewhost tt-view" data-mount="main"></div>
        </section>
        <div class="gutter gutter-insp" data-split="inspector" role="separator"></div>
        <aside class="tt-tail" data-mount="dock"></aside>
      </div>
      <footer class="tt-statusline"><span class="mode-chip">NORMAL</span><span data-mount="status"></span></footer>
    </div>`,
    blueprint: () => `<div class="bp-frame theme-${OPTS.theme}">
      <div class="app bp">
        <header class="bp-band">
          <div class="bp-mark">${icon('logo')}<div><b>Reqly</b><small>API engineering sheet</small></div></div>
          <div class="bp-tb" data-mount="titleblock"></div>
          <div class="bp-tools">
            <button class="cmdk-btn bp-cmdk" data-action="palette">${icon('search')}<span>Find</span><kbd>⌘K</kbd></button>
            <button class="btn env-btn bp-env" data-action="env-menu">${icon('layers')}<span>${envById() ? esc(envById().name) : 'no env'}</span></button>
            <button class="btn btn-icon" data-action="ws-menu" data-tip="Workspaces" aria-label="Workspaces">${icon('layers')}</button>
            <button class="btn btn-icon" data-action="nav" data-view="settings" data-tip="Settings" aria-label="Settings">${icon('gear')}</button>
          </div>
        </header>
        <div class="bp-body">
          <aside class="bp-index" data-mount="sidebar"></aside>
          <div class="gutter gutter-side" data-split="side" role="separator"></div>
          <section class="maincol">
            <div class="tabstrip bp-tabs" data-mount="tabs" role="tablist"></div>
            <div class="viewhost bp-sheet-area" data-mount="main"></div>
            <footer class="statusbar bp-status" data-mount="status"></footer>
          </section>
          <div class="gutter gutter-insp" data-split="inspector" role="separator"></div>
          <aside class="bp-notes" data-mount="inspector"></aside>
        </div>
      </div>
    </div>`,
    atlas: () => `<div class="app atl theme-${OPTS.theme}">
      <header class="atl-bar">
        <div class="atl-pill atl-brandpill">
          <button class="atl-ws" data-action="ws-menu" data-tip="Switch workspace">${icon('logo')}<span>${esc(wsById().name)}</span>${icon('chev-d', 'chev')}</button>
          <span class="atl-div"></span>
          <button class="atl-env" data-action="env-menu" data-tip="Environment">${icon('layers')}<span>${envById() ? esc(envById().name) : 'none'}</span></button>
          <span class="atl-div"></span>
          <button class="atl-branch" data-action="branch-menu" data-tip="Branch">${icon('branch')}${esc(D.gitState.branch)}</button>
        </div>
        <button class="cmdk-btn atl-cmdk" data-action="palette" aria-label="Command palette">${icon('search')}<span>Search anything…</span><kbd>⌘K</kbd></button>
        <div class="atl-pill atl-toolpill">
          <button class="atl-icon" data-action="toggle-sidebar" data-tip="Toggle panel" aria-label="Toggle sidebar">${icon('sliders')}</button>
          <button class="atl-icon" data-action="toggle-inspector" data-tip="Context" aria-label="Toggle context panel">${icon('book')}</button>
          <span class="atl-div"></span>
          <button class="atl-icon" data-action="nav" data-view="settings" data-tip="Settings" aria-label="Settings">${icon('gear')}</button>
        </div>
      </header>
      <div class="atl-body">
        <div class="rail-spacer" aria-hidden="true"></div>
        <nav class="rail atl-rail" data-mount="rail" aria-label="Views"></nav>
        <div class="gutter gutter-side" data-split="side" role="separator"></div>
        <aside class="atl-side" data-mount="sidebar"></aside>
        <section class="maincol">
          <div class="tabstrip atl-tabs" data-mount="tabs" role="tablist"></div>
          <div class="viewhost atl-main" data-mount="main"></div>
        </section>
      </div>
      <footer class="atl-footline"><div class="atl-status" data-mount="status"></div><div class="atl-dockpill" data-mount="dock"></div></footer>
      <aside class="atl-ctx" data-mount="inspector"></aside>
    </div>`,
    monograph: () => `<div class="mg theme-${OPTS.theme}">
      <header class="mg-mast">
        <div class="mast-top">
          <div class="mg-word">${icon('logo', 'brand-logo')}<b>Reqly</b><span>a working specification for ${esc(wsById().name)}</span></div>
          <div class="mast-meta">
            <button class="cmdk-btn mg-cmdk" data-action="palette" aria-label="Search">${icon('search')}<span>Search</span><kbd>⌘K</kbd></button>
            <button class="btn env-btn mg-env" data-action="env-menu">${icon('layers')}<span>${envById() ? esc(envById().name) : 'none'}</span></button>
            <button class="branch-chip mg-branch" data-action="branch-menu">${icon('branch')}${esc(D.gitState.branch)}</button>
            <button class="btn btn-icon" data-action="nav" data-view="settings" data-tip="Settings" aria-label="Settings">${icon('gear')}</button>
          </div>
        </div>
        <div class="mast-rule" aria-hidden="true"></div>
      </header>
      <div class="mg-body">
        <aside class="mg-outline" data-mount="sidebar"></aside>
        <div class="gutter gutter-side" data-split="side" role="separator"></div>
        <section class="maincol mg-maincol">
          <div class="folios" data-mount="tabs" role="tablist"></div>
          <div class="viewhost mg-doc" data-mount="main"></div>
        </section>
        <div class="gutter gutter-insp" data-split="inspector" role="separator"></div>
        <aside class="mg-margin" data-mount="inspector"></aside>
      </div>
      <footer class="mg-foot" data-mount="status"></footer>
    </div>`,
    git: () => `<div class="app shell-git theme-${OPTS.theme}">
      ${branchbarHtml()}
      <header class="topbar git-topbar" data-mount="topbar"></header>
      <div class="git-body">
        <aside class="sidebar sidebar-git" data-mount="sidebar"></aside>
        <div class="gutter gutter-side" data-split="side" role="separator"></div>
        <section class="maincol">
          <div class="tabstrip tabstrip-git" data-mount="tabs" role="tablist"></div>
          <div class="viewhost" data-mount="main"></div>
          <footer class="statusbar" data-mount="status"></footer>
        </section>
      </div>
    </div>`
  };

  function mount(rootEl, opts = {}) {
    Object.assign(OPTS, opts);
    ROOT = rootEl;
    ROOT.classList.add('reqly-root');
    ROOT.innerHTML = SHELLS[OPTS.shell]();
    ROOT.insertAdjacentHTML('beforeend', '<div class="overlays"></div><div class="toasts" aria-live="polite"></div>');
    if (!S.tabs.length && !S.emptyWorkspace) findOrMakeTab('r-list-payments');
    bindDelegation();
    initSplitters();
    startClock();
    refresh();
    setTimeout(() => { if (!S.firstRunDone) welcomeFlow(); }, 500);
    return window.ReqlyApp;
  }

  function inspectorHtml() {
    const t = activeTab();
    const h = D.history[0];
    return `<div class="insp-block"><h5>LIVE REQUEST</h5>
      ${t ? `<div class="insp-kv"><span>method</span><b>${t.method}</b><span>host</span><b class="mono">${esc((resolveVars(t.url).split('/').slice(0, 3).join('/')))}</b><span>scheme</span><b>https</b><span>auth</span><b>${esc(t.auth?.type || 'none')}</b></div>`
      : '<p class="empty-line">No active request</p>'}</div>
      <div class="insp-block"><h5>GIT STATE</h5><div class="insp-kv"><span>branch</span><b>${esc(D.gitState.branch)}</b><span>dirty files</span><b class="warn-t">${D.gitState.changes.length}</b><span>ahead/behind</span><b>${D.gitState.ahead}/${D.gitState.behind}</b></div></div>
      <div class="insp-block"><h5>LATEST RESPONSE</h5>${h ? `<div class="insp-kv"><span>status</span><b>${h.status}</b><span>time</span><b>${ms(h.time)}</b><span>size</span><b>${bytes(h.size)}</b><span>env</span><b>${esc(h.env || '—')}</b></div>` : '<p class="empty-line">—</p>'}</div>
      <div class="insp-block insp-health"><h5>SERVICE HEALTH</h5>
        ${[['api.paycore.dev', 'ok'], ['staging.paycore.io', 'warn'], ['ledger.svc', 'error']].map(s => `<div class="health-row"><i class="led led-${s[1]}"></i><span>${s[0]}</span><em>${s[1]}</em></div>`).join('')}</div>`;
  }
  function updateInspector() {
    const ins = ROOT.querySelector('[data-mount="inspector"]');
    if (ins) ins.innerHTML = inspectorHtml();
  }

  let CLOCK = null;
  function startClock() {
    clearInterval(CLOCK);
    CLOCK = setInterval(() => {
      const el = ROOT.querySelector('[data-role="clock"]');
      if (el) el.textContent = fmtClock(new Date().toISOString());
    }, 1000);
  }

  function initSplitters() {
    ROOT.querySelectorAll('[data-split]').forEach(g => {
      g.addEventListener('pointerdown', (e) => {
        e.preventDefault();
        g.setPointerCapture(e.pointerId);
        const kind = g.dataset.split;
        const move = (ev) => {
          const r = ROOT.querySelector('.app').getBoundingClientRect();
          if (kind === 'side') document.documentElement.style.setProperty('--side-w', Math.max(180, ev.clientX - r.left - 56) + 'px');
          if (kind === 'dock') document.documentElement.style.setProperty('--dock-h', Math.max(120, r.bottom - ev.clientY) + 'px');
          if (kind === 'inspector') document.documentElement.style.setProperty('--insp-w', Math.max(200, r.right - ev.clientX) + 'px');
        };
        const up = () => { g.removeEventListener('pointermove', move); g.removeEventListener('pointerup', up); };
        g.addEventListener('pointermove', move);
        g.addEventListener('pointerup', up);
      });
      g.addEventListener('dblclick', () => { ['--side-w', '--dock-h', '--insp-w'].forEach(v => document.documentElement.style.removeProperty(v)); });
    });
  }

  function bindDelegation() {
    ROOT.addEventListener('click', (e) => {
      const moreBtn = e.target.closest('.row-more');
      if (moreBtn && moreBtn.dataset.action === 'node-menu') { e.stopPropagation(); nodeContextMenu(e, moreBtn.dataset.id); return; }
      const act = e.target.closest('[data-action]');
      if (!act) return;
      const a = act.dataset.action;
      const actions = {
        nav: () => navigate(act.dataset.view),
        palette: () => palette(),
        'toggle-sidebar': () => {
          const dr = ROOT.querySelector('[data-mount="drawer"]');
          if (dr) { dr.classList.toggle('open'); return; }
          const atl = ROOT.querySelector('.atl');
          if (atl) { atl.toggleAttribute('data-side-hidden'); return; }
          S.sidebarVisible = !S.sidebarVisible;
          const side = ROOT.querySelector('[data-mount="sidebar"]');
          if (side) side.style.display = S.sidebarVisible ? '' : 'none';
        },
        'toggle-inspector': () => {
          const ins = ROOT.querySelector('[data-mount="inspector"]');
          if (ins) ins.classList.toggle('open');
        },
        'toggle-wb-side': () => {
          const s = ROOT.querySelector('.wb-side');
          if (s) s.classList.toggle('closed');
        },
        'commit-info': () => {
          const b = e.target.closest('[data-action="commit-info"]');
          if (b) toast(`commit ${b.dataset.hash} — ${b.dataset.msg} · ${b.dataset.who}`, { type: 'info' });
        },
        'ws-menu': () => workspaceMenu(e),
        'env-menu': () => envMenu(e),
        'open-request': () => openRequestNode(act.dataset.id),
        'node-menu': () => nodeContextMenu(e, act.dataset.id),
        'toggle-folder': () => { const n = walkTree(S.tree).find(x => x.id === act.dataset.id); n.open = n.open === false ? true : !n.open; renderSidebar(); },
        'tree-menu': () => menu(e.clientX, e.clientY, [
          { label: 'Run collection…', icon: 'play', fn: () => { navigate('runner'); toast('Collection runner opened', { type: 'info' }); } },
          { label: 'Export collection…', icon: 'export', fn: () => navigate('importexport') },
          { label: 'New folder', icon: 'folder', fn: () => renameModal('New folder', '', (name) => { S.tree.push({ type: 'folder', id: uid(), name, open: true, children: [] }); renderSidebar(); }) },
          '-', { label: 'Open containing folder', icon: 'folder', fn: () => toast('collections/ — plain-text JSON on disk', { type: 'info' }) }
        ]),
        'focus-tab': () => { S.activeTab = act.dataset.id; refreshTabsBar(); callHook('renderResponse'); renderMain(); },
        'close-tab': async () => {
          const t = S.tabs.find(t => t.id === act.dataset.id);
          if (t?.dirty && !(await confirmModal('Unsaved changes', `<b>${esc(t.name)}</b> has unsaved edits. Close anyway?`, 'Discard & close', true))) return;
          S.tabs = S.tabs.filter(x => x.id !== act.dataset.id);
          if (S.activeTab === act.dataset.id) S.activeTab = S.tabs[0]?.id || null;
          refreshTabsBar(); renderMain(); renderSidebar();
        },
        'new-request': () => createRequestInFolder(null),
        'new-folder': () => renameModal('New folder', '', (name) => { S.tree.push({ type: 'folder', id: uid(), name, open: true, children: [] }); renderSidebar(); }),
        'dock-tab': () => { S.dockTab = act.dataset.id; renderDock(); },
        'clear-console': () => { S.consoleEntries = []; renderDock(); },
        'kv-toggle': () => { const rows = getKvScope(act.dataset.scope); const r = rows[+act.dataset.i]; if (r) { r.enabled = r.enabled === false ? true : false; markDirty(); callHook('renderResponse'); renderMain(); } },
        'kv-del': () => { getKvScope(act.dataset.scope).splice(+act.dataset.i, 1); markDirty(); renderMain(); },
        'kv-add': () => { getKvScope(act.dataset.scope).push({ key: '', value: '', enabled: true }); markDirty(); renderMain(); const inp = App.root.querySelector(`.kv-editor[data-scope="${act.dataset.scope}"] .kv-row:last-of-type .kv-key`); inp && inp.focus(); },
        'reveal-secret': () => { const k = act.dataset.scope + act.dataset.i; S.secretsRevealed[k] = !S.secretsRevealed[k]; renderMain(); },
        'branch-menu': () => menu(e.clientX, e.clientY, [
          ...D.gitState.branches.map(b => ({ label: (b === D.gitState.branch ? '● ' : '○ ') + b, icon: 'branch', fn: () => { D.gitState.branch = b; toast(`Checked out ${b}`, { type: 'success' }); renderStatus(); refresh(); } })),
          '-', { label: 'New branch from HEAD…', icon: 'plus', fn: () => formModal({ title: 'Create branch', submitLabel: 'Create', fields: [{ name: 'name', label: 'Branch name', value: 'feat/' }] , onSubmit: v => { D.gitState.branches.push(v.name); D.gitState.branch = v.name; toast(`Created & switched to ${v.name}`, { type: 'success' }); renderStatus(); } }) },
          { label: 'Manage branches → Git view', icon: 'git', fn: () => navigate('git') }
        ]),
        'git-sync': () => {
          const kind = act.dataset.kind || 'fetch';
          toast(`${kind === 'push' ? 'Pushing' : kind === 'pull' ? 'Pulling' : 'Fetching'} origin/main…`, { type: 'info', timeout: 1400 });
          setTimeout(() => {
            if (kind === 'pull' && D.gitState.conflictFile) {
              navigate('git');
              modal({ title: 'Pull resulted in conflicts', wide: true, body: `<div class="conflict-box">${icon('warn', 'conflict-ic')}<p><code>environments/dev.yaml</code> conflicts with incoming commit <code>7b21de0</code>. Resolve it in the merge editor.</p></div>`, footer: `<button class="btn btn-primary" onclick="document.querySelector('.modal-backdrop [data-close]')?.click()">Open merge editor</button>` });
            } else toast(`${kind} complete — working tree ${kind === 'push' ? 'published' : 'up to date'}`, { type: 'success' });
          }, 900);
        }
      };
      actions[a] && actions[a]();
    });

    ROOT.addEventListener('input', (e) => {
      const el = e.target.closest('[data-action-input]');
      if (!el) return;
      const a = el.dataset.actionInput;
      if (a === 'tree-query') { S.ui.treeQuery = el.value; renderSidebar(); const inp = App.root.querySelector('[data-action-input="tree-query"]'); inp && (inp.focus(), inp.setSelectionRange(inp.value.length, inp.value.length)); }
      if (a === 'kv-key' || a === 'kv-val') {
        const rows = getKvScope(el.dataset.scope); const row = rows[+el.dataset.i];
        if (row) { row[a === 'kv-key' ? 'key' : 'value'] = el.value; markDirty(); }
      }
      callHook('live-input', a, el);
    });

    ROOT.addEventListener('keydown', (e) => {
      const typing = /^(input|textarea|select)$/i.test(e.target.tagName) || e.target.isContentEditable;
      if ((e.metaKey || e.ctrlKey) && e.key.toLowerCase() === 'k') { e.preventDefault(); palette(); return; }
      if ((e.metaKey || e.ctrlKey) && e.key === 'Enter') { e.preventDefault(); const t = activeTab(); t && sendRequest(t); return; }
      if ((e.metaKey || e.ctrlKey) && e.key.toLowerCase() === 's') { e.preventDefault(); saveActive(); return; }
      if ((e.metaKey || e.ctrlKey) && e.key.toLowerCase() === 'b') { e.preventDefault(); S.sidebarVisible = !S.sidebarVisible; renderSidebar(); return; }
      if (e.key === 'Escape') { if (MODALS.length) closeTopModal(); else if (MENU) closeMenu(); else cancelRequest(activeTab()); return; }
      if (e.key === '?' && !typing) { shortcutsModal(); }
    });
    ROOT.addEventListener('contextmenu', (e) => {
      const row = e.target.closest('.row-request');
      if (row) { e.preventDefault(); nodeContextMenu(e, row.dataset.id); }
    });
  }
  function markDirty() {
    const t = activeTab();
    if (t && !t.dirty) { t.dirty = true; refreshTabsBar(); }
  }

  window.ReqlyApp = {
    D, get S() { return S; }, get opts() { return OPTS; }, VIEWS, RAIL_GROUPS,
    esc, uid, icon, methodChip, statusChip, ago, fmtClock, fmtDate, bytes, ms, highlight, jsonPathGet, diffLines, resolveVars, undefinedVars,
    modal, closeTopModal, confirmModal, formModal, toast, menu, closeMenu, palette, shortcutsModal, welcomeFlow,
    regView, regSidebar, on: (n, f) => HOOKS[n] = f, refresh, navigate, renderMain, renderSidebar, refreshTabsBar, renderStatus, renderDock,
    mount, updateInspector, inspectorHtml, branchbarHtml, commitStripHtml, titleblockHtml, sessionsHtml,
    openRequestNode, nodeContextMenu, newTabFromNode, findOrMakeTab, activeTab, walkTree,
    sendRequest, cancelRequest, saveActive, log, kvEditor, kvRow, getKvScope, largeBodyHtml, envById, wsById, fileOf,
    get root() { return ROOT; },
    markDirty
  };
})();
