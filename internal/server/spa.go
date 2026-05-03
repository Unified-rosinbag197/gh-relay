package server

import "github.ibm.com/soub4i/gh-relay/internal/version"

var spaHTML = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>gh-relay</title>
<link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/highlight.js/11.9.0/styles/github-dark.min.css">
<style>
  *, *::before, *::after { box-sizing: border-box; margin: 0; padding: 0; }
  :root {
    --bg:        #0d1117;
    --bg-panel:  #161b22;
    --bg-hover:  #21262d;
    --bg-active: #1f6feb33;
    --border:    #30363d;
    --text:      #e6edf3;
    --text-muted:#8b949e;
    --accent:    #58a6ff;
    --danger:    #f85149;
    --font-mono: 'SFMono-Regular', Consolas, 'Liberation Mono', Menlo, monospace;
    --font-ui:   -apple-system, BlinkMacSystemFont, 'Segoe UI', Helvetica, Arial, sans-serif;
    --radius:    6px;
    --tree-w:    280px;
  }
  html, body { height: 100%; overflow: hidden; background: var(--bg); color: var(--text); font-family: var(--font-ui); font-size: 14px; }

  #app { display: flex; flex-direction: column; height: 100%; }
  #header { display: flex; align-items: center; gap: 12px; padding: 10px 16px; background: var(--bg-panel); border-bottom: 1px solid var(--border); flex-shrink: 0; }
  #header .logo { font-weight: 700; font-size: 16px; color: var(--accent); letter-spacing: -0.5px; white-space: nowrap; }
  #header .meta { flex: 1; min-width: 0; }
  #header .repo-name { font-weight: 600; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
  #header .repo-desc { font-size: 12px; color: var(--text-muted); white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
  #header .badge { font-size: 11px; padding: 2px 8px; border-radius: 20px; border: 1px solid var(--border); color: var(--text-muted); flex-shrink: 0; }
  #header .badge.private { border-color: var(--danger); color: var(--danger); }
  #branch-select { background: var(--bg-hover); color: var(--text); border: 1px solid var(--border); border-radius: var(--radius); padding: 4px 8px; font-size: 13px; cursor: pointer; flex-shrink: 0; }
  #branch-select:focus { outline: 2px solid var(--accent); }
  .ro-badge { font-size: 11px; padding: 2px 8px; border-radius: 20px; background: var(--bg-active); color: var(--accent); border: 1px solid var(--accent); flex-shrink: 0; }

  #workspace { display: flex; flex: 1; overflow: hidden; }

  #sidebar { width: var(--tree-w); min-width: var(--tree-w); display: flex; flex-direction: column; border-right: 1px solid var(--border); overflow: hidden; }
  #tree-header { padding: 8px 12px; font-size: 12px; color: var(--text-muted); text-transform: uppercase; letter-spacing: 0.8px; font-weight: 600; border-bottom: 1px solid var(--border); flex-shrink: 0; display: flex; align-items: center; justify-content: space-between; }
  #tree-search { padding: 8px; flex-shrink: 0; border-bottom: 1px solid var(--border); }
  #tree-search input { width: 100%; background: var(--bg-hover); border: 1px solid var(--border); border-radius: var(--radius); color: var(--text); padding: 5px 8px; font-size: 13px; }
  #tree-search input::placeholder { color: var(--text-muted); }
  #tree-search input:focus { outline: 2px solid var(--accent); }
  #file-tree { overflow-y: auto; flex: 1; padding: 4px 0; }
  #file-tree::-webkit-scrollbar { width: 6px; }
  #file-tree::-webkit-scrollbar-thumb { background: var(--border); border-radius: 3px; }

  .tree-file { display: flex; align-items: center; gap: 6px; padding: 3px 0; cursor: pointer; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; user-select: none; border-radius: var(--radius); }
  .tree-file:hover { background: var(--bg-hover); }
  .tree-file.active { background: var(--bg-active); color: var(--accent); }
  .tree-dir { display: flex; flex-direction: column; user-select: none; }
  .tree-row { display: flex; align-items: center; gap: 6px; padding: 3px 0; cursor: pointer; white-space: nowrap; overflow: hidden; border-radius: var(--radius); }
  .tree-row:hover { background: var(--bg-hover); }
  .tree-icon { flex-shrink: 0; font-size: 13px; }
  .tree-name { overflow: hidden; text-overflow: ellipsis; font-size: 13px; }
  .tree-children { display: none; }
  .tree-dir.open > .tree-children { display: block; }

  #viewer { flex: 1; display: flex; flex-direction: column; overflow: hidden; }
  #viewer-header { padding: 8px 16px; border-bottom: 1px solid var(--border); display: flex; align-items: center; gap: 10px; flex-shrink: 0; font-size: 12px; color: var(--text-muted); }
  #viewer-header .file-path { font-family: var(--font-mono); color: var(--text); overflow: hidden; text-overflow: ellipsis; white-space: nowrap; flex: 1; }
  #viewer-header .file-size { flex-shrink: 0; }
  #viewer-body { flex: 1; overflow: auto; position: relative; }
  #viewer-body::-webkit-scrollbar { width: 8px; height: 8px; }
  #viewer-body::-webkit-scrollbar-thumb { background: var(--border); border-radius: 4px; }

  #code-wrap { display: flex; min-height: 100%; }
  #line-nums { padding: 16px 0 16px 16px; text-align: right; color: var(--text-muted); font-family: var(--font-mono); font-size: 13px; line-height: 1.6; user-select: none; flex-shrink: 0; min-width: 48px; border-right: 1px solid var(--border); margin-right: 0; white-space: pre; }
  #code-content { flex: 1; overflow: visible; }
  /* Override hljs defaults to fit our layout */
  #code-content pre { margin: 0; border-radius: 0; background: transparent !important; }
  #code-content pre code.hljs { padding: 16px; font-family: var(--font-mono); font-size: 13px; line-height: 1.6; background: transparent !important; white-space: pre; display: block; }

  #welcome { flex: 1; display: flex; flex-direction: column; align-items: center; justify-content: center; gap: 12px; color: var(--text-muted); }
  #welcome .big-icon { font-size: 48px; }
  #welcome h2 { color: var(--text); font-size: 18px; }
  .spinner { width: 32px; height: 32px; border: 3px solid var(--border); border-top-color: var(--accent); border-radius: 50%; animation: spin .8s linear infinite; }
  @keyframes spin { to { transform: rotate(360deg); } }
  #toast { position: fixed; bottom: 24px; right: 24px; background: var(--danger); color: #fff; padding: 10px 16px; border-radius: var(--radius); font-size: 13px; opacity: 0; pointer-events: none; transition: opacity .3s; z-index: 999; max-width: 360px; }
  #toast.show { opacity: 1; pointer-events: auto; }

  #commits-btn { background: var(--bg-hover); border: 1px solid var(--border); color: var(--text); border-radius: var(--radius); padding: 4px 10px; font-size: 12px; cursor: pointer; }
  #commits-btn:hover { background: var(--bg-active); }
  #download-btn { background: var(--bg-hover); border: 1px solid var(--border); color: var(--text); border-radius: var(--radius); padding: 4px 10px; font-size: 12px; cursor: pointer; text-decoration: none; }
  #download-btn:hover { background: var(--bg-active); }
  #commits-overlay { display: none; position: fixed; inset: 0; background: #000a; z-index: 100; align-items: center; justify-content: center; }
  #commits-overlay.show { display: flex; }
  #commits-modal { background: var(--bg-panel); border: 1px solid var(--border); border-radius: 10px; width: 640px; max-width: 95vw; max-height: 80vh; display: flex; flex-direction: column; }
  #commits-modal-header { padding: 14px 18px; border-bottom: 1px solid var(--border); display: flex; align-items: center; justify-content: space-between; font-weight: 600; }
  #commits-modal-body { overflow-y: auto; flex: 1; }
  .commit-row { padding: 12px 18px; border-bottom: 1px solid var(--border); display: flex; gap: 12px; }
  .commit-row:last-child { border-bottom: none; }
  .commit-sha { font-family: var(--font-mono); font-size: 12px; color: var(--accent); flex-shrink: 0; }
  .commit-msg { flex: 1; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
  .commit-meta { font-size: 12px; color: var(--text-muted); white-space: nowrap; }
  #commits-close { background: none; border: none; color: var(--text-muted); cursor: pointer; font-size: 18px; line-height: 1; }

  @media (max-width: 640px) { :root { --tree-w: 220px; } }
</style>
<script>/*__RELAY_TOKEN__*/</script>
<script>/*__ALLOW_DOWNLOAD__*/</script>
</head>
<body>
<div id="app">
  <header id="header">
    <span class="logo">gh-relay</span>&nbsp;version: ` + version.Version + `&nbsp;|&nbsp;<small>by <a href="https://github.com/soub4i" target="_blank" style="color:var(--accent)">soub4i</a></small>
    <div class="meta">
      <div class="repo-name" id="hdr-repo">Loading…</div>
      <div class="repo-desc" id="hdr-desc"></div>
    </div>
    <span class="ro-badge">read-only</span>
    <select id="branch-select" style="display:none"></select>
    <button id="commits-btn" style="display:none">⏱ History</button>
    <a id="download-btn" style="display:none">Download ZIP</a>
    <span class="badge" id="hdr-badge"></span>
  </header>
  <div id="workspace">
    <aside id="sidebar">
      <div id="tree-header">
        <span>Files</span>
        <span id="tree-count" style="font-weight:400"></span>
      </div>
      <div id="tree-search"><input id="search-input" type="search" placeholder="Filter files…" autocomplete="off"></div>
      <div id="file-tree"></div>
    </aside>
    <main id="viewer">
      <div id="welcome">
        <div class="big-icon">📂</div>
        <h2>Select a file to view</h2>
        <p>Browse the tree on the left to explore this repository.</p>
      </div>
    </main>
  </div>
</div>
<div id="toast"></div>
<div id="commits-overlay">
  <div id="commits-modal">
    <div id="commits-modal-header">
      <span>Recent Commits</span>
      <button id="commits-close">×</button>
    </div>
    <div id="commits-modal-body"></div>
  </div>
</div>

<script src="https://cdnjs.cloudflare.com/ajax/libs/highlight.js/11.9.0/highlight.min.js"></script>
<script src="https://cdnjs.cloudflare.com/ajax/libs/highlight.js/11.9.0/languages/go.min.js"></script>
<script src="https://cdnjs.cloudflare.com/ajax/libs/highlight.js/11.9.0/languages/rust.min.js"></script>
<script src="https://cdnjs.cloudflare.com/ajax/libs/highlight.js/11.9.0/languages/python.min.js"></script>
<script src="https://cdnjs.cloudflare.com/ajax/libs/highlight.js/11.9.0/languages/typescript.min.js"></script>
<script src="https://cdnjs.cloudflare.com/ajax/libs/highlight.js/11.9.0/languages/yaml.min.js"></script>
<script src="https://cdnjs.cloudflare.com/ajax/libs/highlight.js/11.9.0/languages/dockerfile.min.js"></script>
<script src="https://cdnjs.cloudflare.com/ajax/libs/highlight.js/11.9.0/languages/bash.min.js"></script>
<script src="https://cdnjs.cloudflare.com/ajax/libs/highlight.js/11.9.0/languages/sql.min.js"></script>
<script src="https://cdnjs.cloudflare.com/ajax/libs/highlight.js/11.9.0/languages/protobuf.min.js"></script>
<script>
hljs.configure({ ignoreUnescapedHTML: true });

var state = { info: null, tree: null, activeFile: null, filterText: '' };

function $(id) { return document.getElementById(id); }

function showToast(msg) {
  var el = $('toast');
  el.textContent = msg;
  el.classList.add('show');
  setTimeout(function() { el.classList.remove('show'); }, 4000);
}

function api(path) {
  return fetch(path, {
    headers: { 'X-Relay-Token': __RELAY_TOKEN__ }
  }).then(function(r) {
    if (r.status === 401) {
      document.body.innerHTML = '<div style="padding:40px;font-family:sans-serif;color:#f85149">Session expired, please reload the page.</div>';
      return new Promise(function() {});
    }
    if (!r.ok) throw new Error('HTTP ' + r.status + ' from ' + path);
    return r;
  });
}

async function boot() {
  try {
    var infoResp = await api('/api/info');
    state.info = await infoResp.json();
  } catch(e) {
    showToast('Failed to load repository info: ' + e.message);
    return;
  }

  $('hdr-repo').textContent = state.info.owner + ' / ' + state.info.repo;
  $('hdr-desc').textContent = state.info.description || '';
  var badge = $('hdr-badge');
  badge.textContent = state.info.private ? '🔒 Private' : '🌐 Public';
  badge.className = 'badge' + (state.info.private ? ' private' : '');

  var sel = $('branch-select');
  (state.info.branches || [state.info.branch]).forEach(function(b) {
    var opt = document.createElement('option');
    opt.value = b;
    opt.textContent = b;
    if (b === state.info.branch) opt.selected = true;
    sel.appendChild(opt);
  });
  sel.style.display = '';
  sel.addEventListener('change', function() { loadTree(sel.value); });

  $('commits-btn').style.display = '';
  $('commits-btn').addEventListener('click', openCommits);

  if (typeof __ALLOW_DOWNLOAD__ !== 'undefined' && __ALLOW_DOWNLOAD__) {
    var dlBtn = $('download-btn');
    dlBtn.style.display = '';
    dlBtn.addEventListener('click', async function(e) {
      e.preventDefault();
      dlBtn.textContent = 'Downloading…';
      try {
        var branch = sel.value || state.info.branch;
        var resp = await api('/api/download?branch=' + encodeURIComponent(branch));
        var blob = await resp.blob();
        var url = URL.createObjectURL(blob);
        var a = document.createElement('a');
        a.href = url;
        a.download = state.info.repo + '-' + branch + '.zip';
        a.click();
        URL.revokeObjectURL(url);
      } catch(err) {
        alert('Download failed: ' + err.message);
      }
      dlBtn.textContent = 'Download ZIP';
    });
  }

  loadTree(state.info.branch);
}

async function loadTree(branch) {
  $('file-tree').innerHTML = '<div style="padding:24px;text-align:center"><div class="spinner" style="margin:auto"></div></div>';
  try {
    var resp = await api('/api/tree?branch=' + encodeURIComponent(branch));
    var data = await resp.json();
    state.tree = data.tree || [];
    renderTree(state.tree);
  } catch(e) {
    $('file-tree').textContent = 'Error: ' + e.message;
    showToast('Could not load file tree: ' + e.message);
  }
}

function renderTree(entries) {
  var filter = state.filterText.toLowerCase();
  var filtered = filter ? entries.filter(function(e) { return e.path.toLowerCase().includes(filter); }) : entries;
  $('tree-count').textContent = filtered.length + ' files';
  var container = $('file-tree');
  container.innerHTML = '';

  if (filter) {
    filtered.forEach(function(e) {
      if (e.type !== 'blob') return;
      var row = document.createElement('div');
      row.className = 'tree-file' + (state.activeFile === e.path ? ' active' : '');
      row.style.paddingLeft = '12px';
      row.innerHTML = '<span class="tree-icon">' + fileIcon(e.path) + '</span><span class="tree-name" title="' + e.path + '">' + e.path + '</span>';
      row.addEventListener('click', function() { openFile(e); });
      container.appendChild(row);
    });
    return;
  }

  var root = {};
  entries.forEach(function(e) {
    var parts = e.path.split('/');
    var node = root;
    parts.forEach(function(p, i) {
      if (!node[p]) node[p] = { __entry: null, __children: {} };
      if (i === parts.length - 1) node[p].__entry = e;
      node = node[p].__children;
    });
  });

  function renderNode(nodeMap, depth) {
    var frag = document.createDocumentFragment();
    var keys = Object.keys(nodeMap).sort(function(a, b) {
      var aIsDir = Object.keys(nodeMap[a].__children).length > 0 && !nodeMap[a].__entry;
      var bIsDir = Object.keys(nodeMap[b].__children).length > 0 && !nodeMap[b].__entry;
      if (aIsDir && !bIsDir) return -1;
      if (!aIsDir && bIsDir) return 1;
      return a.toLowerCase() < b.toLowerCase() ? -1 : 1;
    });
    keys.forEach(function(key) {
      var item = nodeMap[key];
      var hasChildren = Object.keys(item.__children).length > 0;
      var entry = item.__entry;
      var pl = (depth * 12 + 12) + 'px';
      if (hasChildren) {
        var dir = document.createElement('div');
        dir.className = 'tree-dir';
        var row = document.createElement('div');
        row.className = 'tree-row';
        row.style.paddingLeft = pl;
        row.innerHTML = '<span class="tree-icon">📁</span><span class="tree-name">' + key + '</span>';
        var children = document.createElement('div');
        children.className = 'tree-children';
        children.appendChild(renderNode(item.__children, depth + 1));
        row.addEventListener('click', function(ev) {
          ev.stopPropagation();
          dir.classList.toggle('open');
          row.querySelector('.tree-icon').textContent = dir.classList.contains('open') ? '📂' : '📁';
        });
        dir.appendChild(row);
        dir.appendChild(children);
        frag.appendChild(dir);
      } else if (entry && entry.type === 'blob') {
        var file = document.createElement('div');
        file.className = 'tree-file' + (state.activeFile === entry.path ? ' active' : '');
        file.style.paddingLeft = pl;
        file.dataset.path = entry.path;
        file.innerHTML = '<span class="tree-icon">' + fileIcon(entry.path) + '</span><span class="tree-name" title="' + entry.path + '">' + key + '</span>';
        file.addEventListener('click', function() { openFile(entry); });
        frag.appendChild(file);
      }
    });
    return frag;
  }
  container.appendChild(renderNode(root, 0));
}

function fileIcon(path) {
  var name = path.split('/').pop().toLowerCase();
  var ext = name.includes('.') ? name.split('.').pop() : '';
  var icons = {
    'go':'🔵','js':'🟡','ts':'🔷','tsx':'🔷','jsx':'🟡','py':'🐍','rs':'🦀',
    'html':'🌐','css':'🎨','json':'📋','yaml':'📋','yml':'📋','md':'📝',
    'sh':'⚙️','bash':'⚙️','toml':'📋','sql':'🗄️','proto':'📡',
    'png':'🖼️','jpg':'🖼️','jpeg':'🖼️','gif':'🖼️','svg':'🖼️',
    'pdf':'📕','zip':'📦','tar':'📦','lock':'🔒','sum':'🔒',
  };
  var nameIcons = { 'dockerfile':'🐳', 'makefile':'⚙️', 'license':'📜', 'readme.md':'📝' };
  return nameIcons[name] || icons[ext] || '📄';
}

// Map file extensions to hljs language names
var EXT_LANG = {
  go:'go', rs:'rust', py:'python', js:'javascript', jsx:'javascript',
  ts:'typescript', tsx:'typescript', html:'html', htm:'html', css:'css',
  scss:'css', json:'json', yaml:'yaml', yml:'yaml', sh:'bash', bash:'bash',
  zsh:'bash', md:'markdown', sql:'sql', proto:'protobuf', xml:'xml',
  toml:'ini', dockerfile:'dockerfile', tf:'hcl',
};

var BINARY_EXTS = {
  png:1, jpg:1, jpeg:1, gif:1, bmp:1, webp:1, ico:1,
  pdf:1, zip:1, gz:1, tar:1, bz2:1, xz:1,
  exe:1, dll:1, so:1, dylib:1, bin:1, wasm:1, pyc:1, pyo:1, class:1,
};

async function openFile(entry) {
  state.activeFile = entry.path;
  document.querySelectorAll('.tree-file').forEach(function(el) {
    el.classList.toggle('active', el.dataset.path === entry.path);
  });

  var viewer = $('viewer');
  viewer.innerHTML =
    '<div id="viewer-header"><span class="file-path" id="vhdr-path"></span><span class="file-size" id="vhdr-size"></span></div>' +
    '<div id="viewer-body"><div style="display:flex;align-items:center;justify-content:center;height:100%"><div class="spinner"></div></div></div>';

  $('vhdr-path').textContent = entry.path;
  $('vhdr-size').textContent = formatBytes(entry.size || 0);

  try {
    var resp = await api('/api/blob?sha=' + encodeURIComponent(entry.sha) + '&path=' + encodeURIComponent(entry.path));
    var ct = resp.headers.get('Content-Type') || '';

    var nameParts = entry.path.split('/').pop().split('.');
    var fileExt = nameParts.length > 1 ? nameParts.pop().toLowerCase() : '';
    var isBinary = fileExt !== '' && BINARY_EXTS[fileExt] === 1;

    if (isBinary) {
      var blob = await resp.blob();
      var url = URL.createObjectURL(blob);
      var body = $('viewer-body');
      if (ct.startsWith('image/')) {
        body.innerHTML = '<div style="padding:24px;text-align:center"><img src="' + url + '" style="max-width:100%;max-height:70vh;border-radius:6px"></div>';
      } else {
        body.innerHTML = '<div style="padding:24px">Binary file, <a href="' + url + '" download="' + entry.path.split('/').pop() + '" style="color:var(--accent)">Download</a></div>';
      }
    } else {
      var text = await resp.text();
      renderCode(text, entry.path);
    }
  } catch(e) {
    $('viewer-body').innerHTML = '<div style="padding:24px;color:var(--danger)">Error: ' + e.message + '</div>';
    showToast('Failed to load file: ' + e.message);
  }
}

function renderCode(text, path) {
  var name = path.split('/').pop().toLowerCase();
  var ext = name.includes('.') ? name.split('.').pop() : '';

  // Special extensionless filenames
  var nameMap = { dockerfile:'dockerfile', makefile:'bash' };
  var lang = nameMap[name] || EXT_LANG[ext] || null;

  var lines = text.split('\n');
  if (lines[lines.length - 1] === '') lines.pop();
  var lineNums = lines.map(function(_, i) { return i + 1; }).join('\n');

  var highlighted;
  if (lang && hljs.getLanguage(lang)) {
    try {
      highlighted = hljs.highlight(text, { language: lang, ignoreIllegals: true }).value;
    } catch(e) {
      highlighted = escHtml(text);
    }
  } else {
    // Let hljs auto-detect, fall back to plain text
    try {
      highlighted = hljs.highlightAuto(text).value;
    } catch(e) {
      highlighted = escHtml(text);
    }
  }

  var body = $('viewer-body');
  body.innerHTML =
    '<div id="code-wrap">' +
      '<div id="line-nums">' + lineNums + '</div>' +
      '<div id="code-content"><pre><code class="hljs">' + highlighted + '</code></pre></div>' +
    '</div>';
}

function escHtml(s) { return s.replace(/&/g,'&amp;').replace(/</g,'&lt;').replace(/>/g,'&gt;'); }
function formatBytes(n) { if (n < 1024) return n + ' B'; if (n < 1048576) return (n/1024).toFixed(1) + ' KB'; return (n/1048576).toFixed(1) + ' MB'; }

async function openCommits() {
  $('commits-overlay').classList.add('show');
  var body = $('commits-modal-body');
  body.innerHTML = '<div style="padding:24px;text-align:center"><div class="spinner" style="margin:auto"></div></div>';
  var branch = $('branch-select').value || state.info.branch;
  try {
    var resp = await api('/api/commits?branch=' + encodeURIComponent(branch));
    var commits = await resp.json();
    if (!commits || commits.length === 0) { body.innerHTML = '<div style="padding:24px;color:var(--text-muted)">No commits found.</div>'; return; }
    body.innerHTML = commits.map(function(c) {
      var sha = c.sha ? c.sha.slice(0,7) : '?';
      var msg = c.commit && c.commit.message ? c.commit.message.split('\n')[0] : '(no message)';
      var author = c.commit && c.commit.author ? c.commit.author.name : '';
      var date = c.commit && c.commit.author && c.commit.author.date ? new Date(c.commit.author.date).toLocaleDateString() : '';
      return '<div class="commit-row"><span class="commit-sha">' + sha + '</span><span class="commit-msg" title="' + escHtml(msg) + '">' + escHtml(msg) + '</span><span class="commit-meta">' + escHtml(author) + ' · ' + date + '</span></div>';
    }).join('');
  } catch(e) {
    body.innerHTML = '<div style="padding:24px;color:var(--danger)">Error: ' + e.message + '</div>';
  }
}

$('commits-close').addEventListener('click', function() { $('commits-overlay').classList.remove('show'); });
$('commits-overlay').addEventListener('click', function(e) { if (e.target === this) this.classList.remove('show'); });
$('search-input').addEventListener('input', function() {
  state.filterText = this.value;
  if (state.tree) renderTree(state.tree);
});

boot().catch(function(e) { showToast('Boot error: ' + e.message); });
</script>
</body>
</html>`
