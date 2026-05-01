package server

import (
	"github.ibm.com/soub4i/gh-relay/internal/version"
)

// spaHTML is the single-file SPA served to all guests.
// It bundles: Prism.js (syntax highlighting), a file tree, a code viewer,
var spaHTML = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>gh-relay</title>
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
    --accent2:   #3fb950;
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

  /* tree-file: a single clickable row */
  .tree-file { display: flex; align-items: center; gap: 6px; padding: 3px 0; cursor: pointer; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; user-select: none; border-radius: var(--radius); }
  .tree-file:hover { background: var(--bg-hover); }
  .tree-file.active { background: var(--bg-active); color: var(--accent); }
  /* tree-dir: column container — header row on top, children below */
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
  #line-nums { padding: 16px 0 16px 16px; text-align: right; color: var(--text-muted); font-family: var(--font-mono); font-size: 13px; line-height: 1.6; user-select: none; flex-shrink: 0; min-width: 48px; border-right: 1px solid var(--border); margin-right: 16px; white-space: pre; }
  #code-content { padding: 16px 16px 16px 0; font-family: var(--font-mono); font-size: 13px; line-height: 1.6; white-space: pre; flex: 1; tab-size: 2; }

  /* ── Prism theme (GitHub Dark) ───────────────────────────── */
  .token.comment, .token.prolog, .token.doctype, .token.cdata { color: #8b949e; font-style: italic; }
  .token.punctuation { color: #e6edf3; }
  .token.namespace { opacity: .7; }
  .token.property, .token.tag, .token.boolean, .token.number, .token.constant, .token.symbol, .token.deleted { color: #79c0ff; }
  .token.selector, .token.attr-name, .token.string, .token.char, .token.builtin, .token.inserted { color: #a5d6ff; }
  .token.operator, .token.entity, .token.url, .language-css .token.string, .style .token.string { color: #d2a8ff; }
  .token.atrule, .token.attr-value, .token.keyword { color: #ff7b72; }
  .token.function, .token.class-name { color: #d2a8ff; }
  .token.regex, .token.important, .token.variable { color: #ffa657; }
  .token.important, .token.bold { font-weight: bold; }
  .token.italic { font-style: italic; }

  #welcome { flex: 1; display: flex; flex-direction: column; align-items: center; justify-content: center; gap: 12px; color: var(--text-muted); }
  #welcome .big-icon { font-size: 48px; }
  #welcome h2 { color: var(--text); font-size: 18px; }
  .spinner { width: 32px; height: 32px; border: 3px solid var(--border); border-top-color: var(--accent); border-radius: 50%; animation: spin .8s linear infinite; }
  @keyframes spin { to { transform: rotate(360deg); } }
  #toast { position: fixed; bottom: 24px; right: 24px; background: var(--danger); color: #fff; padding: 10px 16px; border-radius: var(--radius); font-size: 13px; opacity: 0; pointer-events: none; transition: opacity .3s; z-index: 999; max-width: 360px; }
  #toast.show { opacity: 1; pointer-events: auto; }

  #commits-btn { background: var(--bg-hover); border: 1px solid var(--border); color: var(--text); border-radius: var(--radius); padding: 4px 10px; font-size: 12px; cursor: pointer; }
  #commits-btn:hover { background: var(--bg-active); }
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

  @media (max-width: 640px) {
    :root { --tree-w: 220px; }
  }
</style>
</head>
<body>
<div id="app">
  <header id="header">
    <span class="logo">gh-relay</span> version: ` + version.Version + `  | <small>by <a href="https://github.com/soub4i" target="_blank" style="color: var(--accent);">soub4i</a></small>
    <div class="meta">
      <div class="repo-name" id="hdr-repo">Loading…</div>
      <div class="repo-desc" id="hdr-desc"></div>
    </div>
    <span class="ro-badge">read-only</span>
    <select id="branch-select" style="display:none"></select>
    <button id="commits-btn" style="display:none">⏱ History</button>
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
        <div class="big-icon">🗂️</div>
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

<script>
// ── Minimal Prism.js core + language support (inlined, no CDN) ──────────────
// We include a stripped-down Prism that covers the most common languages.
var Prism = (function() {
  var _self = {};
  _self.languages = {};
  _self.languages.extend = function(id, redef) {
    var lang = _self.util.clone(_self.languages[id]);
    for (var key in redef) { lang[key] = redef[key]; }
    return lang;
  };
  _self.util = {
    clone: function deepClone(o, visited) {
      var clone, id;
      visited = visited || [];
      switch (_self.util.type(o)) {
        case 'Object':
          if (visited.indexOf(o) !== -1) { return o; }
          clone = {};
          visited.push(o);
          for (var key in o) { if (o.hasOwnProperty(key)) { clone[key] = deepClone(o[key], visited); } }
          return clone;
        case 'Array':
          if (visited.indexOf(o) !== -1) { return o; }
          clone = [];
          visited.push(o);
          o.forEach(function(v, i) { clone[i] = deepClone(v, visited); });
          return clone;
        default: return o;
      }
    },
    type: function(o) {
      return Object.prototype.toString.call(o).match(/\[object (\w+)\]/)[1];
    },
    objId: function(obj) {
      if (!obj['__id']) { Object.defineProperty(obj, '__id', { value: ++Prism.util.objId.uid }); }
      return obj['__id'];
    }
  };
  _self.util.objId.uid = 0;

  _self.languages.markup = {
    'comment': /<!--[\s\S]*?-->/,
    'prolog': /<\?[\s\S]+?\?>/,
    'doctype': /<!DOCTYPE[\s\S]+?>/i,
    'cdata': /<!\[CDATA\[[\s\S]*?]]>/i,
    'tag': { pattern: /<\/?(?!\d)[^\s>\/=$<%]+(?:\s+[^\s>\/=]+(?:=(?:("|')(?:\\[\s\S]|(?!\1)[^\\])*\1|[^\s'">=]+))?)*\s*\/?>/i, greedy: true, inside: { 'tag': { pattern: /^<\/?[^\s>\/]+/, inside: { 'punctuation': /^<\/?/, 'namespace': /^[^\s>\/:]+:/ } }, 'attr-value': { pattern: /=(?:("|')(?:\\[\s\S]|(?!\1)[^\\])*\1|[^\s'">=]+)/i, inside: { 'punctuation': [/^=/, /("|')/, /^(?!["'])./, /(?!["']).$/ ] } }, 'punctuation': /\/?>/,'attr-name': { pattern: /[^\s>\/]+/, inside: { 'namespace': /^[^\s>\/:]+:/ } } } },
    'entity': /&[\da-z]{1,8};/i,
  };
  _self.languages.html = _self.languages.markup;
  _self.languages.mathml = _self.languages.markup;
  _self.languages.svg = _self.languages.markup;

  _self.languages.css = {
    'comment': /\/\*[\s\S]*?\*\//,
    'atrule': { pattern: /@[\w-]+[\s\S]*?(?:;|(?=\s*\{))/, inside: { 'rule': /@[\w-]+/ } },
    'url': /url\((?:(["'])(?:\\(?:\r\n|[\s\S])|(?!\1)[^\\\r\n])*\1|.*?)\)/i,
    'selector': /[^{}\s][^{};]*?(?=\s*\{)/,
    'string': { pattern: /("|')(?:\\(?:\r\n|[\s\S])|(?!\1)[^\\\r\n])*\1/, greedy: true },
    'property': /[-_a-z\xA0-\uFFFF][-\w\xA0-\uFFFF]*(?=\s*:)/i,
    'important': /\B!important\b/i,
    'function': /[-a-z0-9]+(?=\()/i,
    'punctuation': /[(){};:,]/
  };

  _self.languages.clike = {
    'comment': [
      { pattern: /(^|[^\\])\/\*[\s\S]*?(?:\*\/|$)/, lookbehind: true },
      { pattern: /(^|[^\\:])\/\/.*/, lookbehind: true, greedy: true }
    ],
    'string': { pattern: /(["'])(?:\\(?:\r\n|[\s\S])|(?!\1)[^\\\r\n])*\1/, greedy: true },
    'class-name': { pattern: /((?:\b(?:class|interface|extends|implements|trait|instanceof|new)\s+)|(?:catch\s+\())[\w.\\]+/i, lookbehind: true, inside: { 'punctuation': /[.\\]/ } },
    'keyword': /\b(?:if|else|while|do|for|return|in|instanceof|function|new|try|throw|catch|finally|null|break|continue)\b/,
    'boolean': /\b(?:true|false)\b/,
    'function': /\w+(?=\()/,
    'number': /\b0x[\da-f]+\b|(?:\b\d+\.?\d*|\B\.\d+)(?:e[+-]?\d+)?/i,
    'operator': /[<>]=?|[!=]=?=?|--?|\+\+?|&&?|\|\|?|[?*/~^%]/,
    'punctuation': /[{}[\];(),.:]/
  };

  _self.languages.javascript = _self.languages.extend('clike', {
    'class-name': [ _self.languages.clike['class-name'], { pattern: /(^|[^$\w\xA0-\uFFFF])[_$A-Z\xA0-\uFFFF][$\w\xA0-\uFFFF]*(?=\.(?:prototype|constructor))/, lookbehind: true } ],
    'keyword': [{ pattern: /((?:^|})\s*)(?:catch|finally)\b/, lookbehind: true }, /\b(?:as|async|await|break|case|class|const|continue|debugger|default|delete|do|else|enum|export|extends|for|from|function|get|if|implements|import|in|instanceof|interface|let|new|null|of|package|private|protected|public|return|set|static|super|switch|this|throw|try|typeof|undefined|var|void|while|with|yield)\b/],
    'number': /\b(?:(?:0[xX][\dA-Fa-f]+|0[bB][01]+|0[oO][0-7]+)n?|\d+n|NaN|Infinity)\b|(?:\B|\b\d+)\.?\d*(?:[Ee][+-]?\d+)?/,
    'function': /[_$a-zA-Z\xA0-\uFFFF][$\w\xA0-\uFFFF]*(?=\s*(?:\.\s*(?:apply|bind|call)\s*)?\()/,
    'operator': /--|\+\+|\*\*=?|=>|&&|\|\||[!<>]=?|>>>?=?|[-+*/%&|^~]=?|[?:]/,
  });
  _self.languages.js = _self.languages.javascript;

  _self.languages.go = {
    'comment': [
      { pattern: /(^|[^\\])\/\*[\s\S]*?(?:\*\/|$)/, lookbehind: true },
      { pattern: /(^|[^\\:])\/\/.*/, lookbehind: true, greedy: true }
    ],
    'string': { pattern: /(["` + "`" + `])(?:\\[\s\S]|(?!\1)[^\\\r\n])*\1/, greedy: true },
    'keyword': /\b(?:break|case|chan|const|continue|default|defer|else|fallthrough|for|func|go|goto|if|import|interface|map|package|range|return|select|struct|switch|type|var)\b/,
    'builtin': /\b(?:append|cap|close|complex|copy|delete|error|false|imag|iota|len|make|new|nil|panic|print|println|real|recover|true)\b/,
    'boolean': /\b(?:true|false)\b/,
    'number': /\b0x[\da-f]+\b|\b\d+\.?\d*(?:e[+-]?\d+)?(?:i)?\b/i,
    'operator': /:=|[<>]=?|[!]=?|&&|\|\||[+\-*\/%&^|~]=?|<<|>>/,
    'punctuation': /[{}[\]();,.]/
  };

  _self.languages.python = _self.languages.extend('clike', {
    'comment': { pattern: /(^|[^\\])#.*/, lookbehind: true },
    'string': { pattern: /(?:[rub]|rb|br)?(?:("""|''')[\s\S]*?\1|("|')(?:\\.|(?!\2)[^\\\r\n])*\2)/i, greedy: true },
    'keyword': /\b(?:and|as|assert|async|await|break|class|continue|def|del|elif|else|except|exec|finally|for|from|global|if|import|in|is|lambda|nonlocal|not|or|pass|print|raise|return|try|while|with|yield)\b/,
    'builtin': /\b(?:__import__|abs|all|any|ascii|bin|bool|bytearray|bytes|callable|chr|classmethod|compile|complex|delattr|dict|dir|divmod|enumerate|eval|exec|filter|float|format|frozenset|getattr|globals|hasattr|hash|help|hex|id|input|int|isinstance|issubclass|iter|len|list|locals|map|max|memoryview|min|next|object|oct|open|ord|pow|property|range|repr|reversed|round|set|setattr|slice|sorted|staticmethod|str|sum|super|tuple|type|vars|zip)\b/,
    'boolean': /\b(?:True|False|None)\b/,
    'number': /(?:\b(?=\d)|\B(?=\.))(?:0[bo])?(?:(?:\d|0x[\da-f])[\da-f]*\.?\d*|\.\d+)(?:e[+-]?\d+)?j?\b/i,
    'operator': /[-+%=]=?|!=|\*\*?=?|\/\/?=?|<[<=>]?|>[=>]?|[&|^~]|\b(?:or|and|not|in|is)\b/,
    'punctuation': /[{}[\];(),.:]/
  });

  _self.languages.rust = _self.languages.extend('clike', {
    'comment': [
      { pattern: /\/\/!.+|\/\/(?!\/).*|\/\*[\s\S]*?\*\//, greedy: true },
    ],
    'keyword': /\b(?:abstract|as|async|await|become|box|break|const|continue|crate|do|dyn|else|enum|extern|final|fn|for|if|impl|in|let|loop|macro|match|mod|move|mut|override|priv|pub|ref|return|self|Self|static|struct|super|trait|try|type|typeof|unsafe|unsized|use|virtual|where|while|yield)\b/,
    'builtin': /\b(?:bool|char|f32|f64|i8|i16|i32|i64|i128|isize|str|u8|u16|u32|u64|u128|usize|String|Vec|Option|Result|Box|Rc|Arc|Cell|RefCell|HashMap|HashSet)\b/,
  });

  _self.languages.typescript = _self.languages.extend('javascript', {
    'class-name': { pattern: /(\b(?:class|extends|implements|instanceof|interface|new|type)\s+)(?!keyof\b)[_$A-Za-z\xA0-\uFFFF][$\w\xA0-\uFFFF]*/, lookbehind: true },
    'keyword': /\b(?:abstract|as|async|await|break|case|catch|class|const|constructor|continue|debugger|declare|default|delete|do|else|enum|export|extends|finally|for|from|function|get|if|implements|import|in|instanceof|interface|is|keyof|let|module|namespace|new|null|of|package|private|protected|public|readonly|return|require|set|static|super|switch|this|throw|try|type|typeof|undefined|var|void|while|with|yield)\b/,
  });
  _self.languages.ts = _self.languages.typescript;

  _self.languages.json = {
    'property': { pattern: /(^|[^\\])"(?:\\.|[^\\"\r\n])*"(?=\s*:)/, lookbehind: true, greedy: true },
    'string': { pattern: /(^|[^\\])"(?:\\.|[^\\"\r\n])*"(?!\s*:)/, lookbehind: true, greedy: true },
    'null': { pattern: /\bnull\b/, alias: 'keyword' },
    'boolean': /\b(?:true|false)\b/,
    'number': /-?\d+\.?\d*(?:e[+-]?\d+)?/i,
    'punctuation': /[{}[\],]/,
    'operator': /:/
  };

  _self.languages.yaml = {
    'scalar': { pattern: /([|>][-+\d]*(?:\n[ \t]+.+)+)/, lookbehind: true, alias: 'string' },
    'comment': /#.*/,
    'key': { pattern: /(\s*(?:^|[:\-,[{\r\n?])[ \t]*(?:![^\s]+)?[ \t]*)[^\r\n{[\]},#\!\|>]+?(?=\s*:\s)/, lookbehind: true, alias: 'atrule' },
    'directive': { pattern: /(^[ \t]*)%.+/m, lookbehind: true, alias: 'important' },
    'datetime': { pattern: /([:\-,[{\r\n](?:[ \t]+|(?!.)))\d{4}-\d\d?-\d\d?(?:[tT]|[ \t]+)\d\d?:\d{2}:\d{2}(?:\.\d*)?(?:[ \t]*Z|[-+]\d\d?(?::\d{2})?)?(?=[ \t]*$)/m, lookbehind: true, alias: 'number' },
    'boolean': { pattern: /([:\-,[{\r\n](?:[ \t]+|(?!.)))(?:true|false)[ \t]*(?=$|,|\]|\})/im, lookbehind: true, alias: 'important' },
    'null': { pattern: /([:\-,[{\r\n](?:[ \t]+|(?!.)))(?:null|~)[ \t]*(?=$|,|\]|\})/im, lookbehind: true, alias: 'important' },
    'string': { pattern: /([:\-,[{\r\n](?:[ \t]+|(?!.)))("|')(?:(?!\2)[^\\\r\n]|\\.)*\2(?=[ \t]*(?:$|,|\]|\}))/m, lookbehind: true, greedy: true },
    'number': { pattern: /([:\-,[{\r\n](?:[ \t]+|(?!.)))[+-]?(?:0x[\da-f]+|0o[0-7]+|(?:\d+\.?\d*|\.?\d+)(?:e[+-]?\d+)?|\.inf|\.nan)[ \t]*(?=$|,|\]|\})/im, lookbehind: true },
    'tag': /![^\s]+/,
    'punctuation': /----|\.\.\.|[:\[\]{},]/
  };

  // ── Tokenizer ──────────────────────────────────────────────
  function matchPattern(pattern, pos, text, lookbehind) {
    pattern.lastIndex = pos;
    var match = pattern.exec(text);
    if (match && lookbehind && match[1]) {
      var lookbehindLength = match[1].length;
      match.index += lookbehindLength;
      match[0] = match[0].slice(lookbehindLength);
    }
    return match;
  }

  function matchGrammar(text, tokenList, grammar, startNode, startPos, rematch) {
    for (var token in grammar) {
      if (!grammar.hasOwnProperty(token) || !grammar[token]) { continue; }
      var patterns = grammar[token];
      patterns = Array.isArray(patterns) ? patterns : [patterns];
      for (var j = 0; j < patterns.length; ++j) {
        if (rematch && rematch.cause == token + ',' + j) { return; }
        var patternObj = patterns[j];
        var inside = patternObj.inside;
        var lookbehind = !!patternObj.lookbehind;
        var greedy = !!patternObj.greedy;
        var alias = patternObj.alias;
        var pattern = patternObj.pattern || patternObj;
        if (greedy && !pattern.sticky) {
          var flags = pattern.toString().match(/[imsuy]*$/)[0];
          pattern = RegExp(pattern.source, flags + (flags.indexOf('g') !== -1 ? '' : 'g'));
        }
        var currentNode = startNode.next;
        for (var pos = startPos; currentNode !== tokenList.tail; pos += currentNode.value.length, currentNode = currentNode.next) {
          if (rematch && pos >= rematch.reach) { break; }
          var str = currentNode.value;
          if (typeof str !== 'string') { continue; }
          var removeCount = 1;
          var match;
          if (greedy) {
            match = matchPattern(pattern, pos, text, lookbehind);
            if (!match) { break; }
            var from = match.index;
            var to = from + match[0].length;
            var p = pos;
            p += currentNode.value.length;
            while (from >= p) { currentNode = currentNode.next; p += currentNode.value.length; }
            p -= currentNode.value.length;
            pos = p;
            if (typeof currentNode.value !== 'string') { continue; }
            for (var k = currentNode; k !== tokenList.tail && (p < to || typeof k.value === 'string'); k = k.next) {
              removeCount++;
              p += k.value.length;
            }
            removeCount--;
            str = text.slice(pos, p);
            match.index -= pos;
          } else {
            match = matchPattern(pattern, 0, str, lookbehind);
            if (!match) { continue; }
          }
          var from = match.index;
          var matchStr = match[0];
          var before = str.slice(0, from);
          var after = str.slice(from + matchStr.length);
          var reach = pos + str.length;
          if (rematch && reach > rematch.reach) { rematch.reach = reach; }
          var removeFrom = currentNode.prev;
          if (before) { removeFrom = addAfter(tokenList, removeFrom, before); pos += before.length; }
          removeRange(tokenList, removeFrom, removeCount);
          var wrapped = new Token(token, inside ? tokenize(matchStr, inside) : matchStr, alias, matchStr);
          currentNode = addAfter(tokenList, removeFrom, wrapped);
          if (after) { addAfter(tokenList, currentNode, after); }
          if (removeCount > 1) {
            var nestedRematch = { cause: token + ',' + j, reach: reach };
            matchGrammar(text, tokenList, grammar, currentNode.prev, pos, nestedRematch);
            if (rematch && nestedRematch.reach > rematch.reach) { rematch.reach = nestedRematch.reach; }
          }
          break;
        }
      }
    }
  }

  function LinkedList() { var head = { value: null, prev: null, next: null }; var tail = { value: null, prev: head, next: null }; head.next = tail; this.head = head; this.tail = tail; this.length = 0; }
  function addAfter(list, node, value) { var next = node.next; var newNode = { value: value, prev: node, next: next }; node.next = newNode; next.prev = newNode; list.length++; return newNode; }
  function removeRange(list, node, count) { var next = node.next; for (var i = 0; i < count && next !== list.tail; i++) { next = next.next; } node.next = next; next.prev = node; list.length -= i; }
  function toArray(list) { var array = []; var node = list.head.next; while (node !== list.tail) { array.push(node.value); node = node.next; } return array; }

  function Token(type, content, alias, matchedStr) { this.type = type; this.content = content; this.alias = alias; this.length = (matchedStr || '').length | 0; }

  function tokenize(text, grammar) {
    var tokenList = new LinkedList();
    addAfter(tokenList, tokenList.head, text);
    matchGrammar(text, tokenList, grammar, tokenList.head, 0, null);
    return toArray(tokenList);
  }

  function stringify(o, language) {
    if (typeof o == 'string') { return escapeHtml(o); }
    if (Array.isArray(o)) { return o.map(function(e) { return stringify(e, language); }).join(''); }
    var env = { type: o.type, content: stringify(o.content, language), tag: 'span', classes: ['token', o.type], attributes: {}, language: language };
    var aliases = o.alias;
    if (aliases) { env.classes = env.classes.concat(aliases); }
    return '<span class="' + env.classes.join(' ') + '">' + env.content + '</span>';
  }

  function escapeHtml(s) { return s.replace(/&/g,'&amp;').replace(/</g,'&lt;').replace(/>/g,'&gt;'); }

  _self.highlight = function(text, grammar, language) {
    return stringify(tokenize(text, grammar), language);
  };

  _self.highlightElement = function(el) {
    var language = el.className.match(/language-(\w+)/);
    if (!language) { return; }
    language = language[1];
    var grammar = _self.languages[language];
    if (!grammar) { el.textContent = el.textContent; return; }
    el.innerHTML = _self.highlight(el.textContent, grammar, language);
  };

  return _self;
})();


var state = {
  info: null,
  tree: null,          // flat list of TreeEntry
  activeFile: null,    // currently open file path
  filterText: '',
};

function $(id) { return document.getElementById(id); }

function showToast(msg) {
  var el = $('toast');
  el.textContent = msg;
  el.classList.add('show');
  setTimeout(function() { el.classList.remove('show'); }, 4000);
}

function api(path) {
  return fetch(path, { credentials: 'same-origin' }).then(function(r) {
   if (r.status === 401) {
      window.location.href = '/auth/login';
      return new Promise(function() {});
    }
    if (!r.ok) throw new Error('HTTP ' + r.status + ' from ' + path);
    return r;
  });
}

// ── Boot ─────────────────────────────────────────────────────
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

  // Branch selector.
  var sel = $('branch-select');
  (state.info.branches || [state.info.branch]).forEach(function(b) {
    var opt = document.createElement('option');
    opt.value = b;
    opt.textContent = b;
    if (b === state.info.branch) { opt.selected = true; }
    sel.appendChild(opt);
  });
  sel.style.display = '';
  sel.addEventListener('change', function() { loadTree(sel.value); });

  $('commits-btn').style.display = '';
  $('commits-btn').addEventListener('click', openCommits);

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
    // Flat list when filtering.
    filtered.forEach(function(e) {
      if (e.type !== 'blob') return;
      var row = document.createElement('div');
      row.className = 'tree-file' + (state.activeFile === e.path ? ' active' : '');
      row.style.paddingLeft = '12px';
      row.innerHTML = '<span class="tree-icon">' + fileIcon(e.path) + '</span><span class="tree-name" title="' + e.path + '">' + e.path + '</span>';
      row.addEventListener('click', function() { openFile(e); row.classList.add('active'); });
      container.appendChild(row);
    });
    return;
  }

  // Build nested tree structure.
  var root = {};
  entries.forEach(function(e) {
    var parts = e.path.split('/');
    var node = root;
    parts.forEach(function(p, i) {
      if (!node[p]) { node[p] = { __entry: null, __children: {} }; }
      if (i === parts.length - 1) { node[p].__entry = e; }
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
        row.innerHTML = '<span class="tree-icon">📂</span><span class="tree-name">' + key + '</span>';
        var children = document.createElement('div');
        children.className = 'tree-children';
        children.appendChild(renderNode(item.__children, depth + 1));
        row.addEventListener('click', function(ev) { ev.stopPropagation(); dir.classList.toggle('open'); row.querySelector('.tree-icon').textContent = dir.classList.contains('open') ? '📂' : '📁'; });
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
  var ext = (path.split('.').pop() || '').toLowerCase();
  var icons = {
    'go':'🔵','js':'🟡','ts':'🔷','tsx':'🔷','jsx':'🟡','py':'🐍','rs':'🦀',
    'html':'🌐','css':'🎨','json':'📋','yaml':'📋','yml':'📋','md':'📝',
    'sh':'⚙️','bash':'⚙️','toml':'📋','sql':'🗄️','proto':'📡',
    'dockerfile':'🐳','png':'🖼️','jpg':'🖼️','jpeg':'🖼️','gif':'🖼️','svg':'🖼️',
    'pdf':'📕','zip':'📦','tar':'📦','lock':'🔒','sum':'🔒', 
  };
  return icons[ext] || '📄';
}

async function openFile(entry) {
  state.activeFile = entry.path;

  // Update active state in tree.
  document.querySelectorAll('.tree-file').forEach(function(el) {
    el.classList.toggle('active', el.dataset.path === entry.path);
  });

  var viewer = $('viewer');
  viewer.innerHTML =
    '<div id="viewer-header"><span class="file-path" id="vhdr-path"></span><span class="file-size" id="vhdr-size"></span></div>' +
    '<div id="viewer-body"><div style="display:flex;align-items:center;justify-content:center;height:100%;"><div class="spinner"></div></div></div>';

  $('vhdr-path').textContent = entry.path;
  $('vhdr-size').textContent = formatBytes(entry.size || 0);

  try {
    var resp = await api('/api/blob?sha=' + encodeURIComponent(entry.sha) + '&path=' + encodeURIComponent(entry.path));
    var ct = resp.headers.get('Content-Type') || '';
    // Determine binary by file extension — never trust Content-Type alone
    // because the server may return application/octet-stream for extensionless
    // files like Dockerfile, Makefile, LICENSE which are plain text.
    var BINARY_EXTS = {
      png:1, jpg:1, jpeg:1, gif:1, bmp:1, webp:1, ico:1,
      pdf:1,
      zip:1, gz:1, tar:1, bz2:1, xz:1,
      exe:1, dll:1, so:1, dylib:1, bin:1,
      wasm:1, pyc:1, pyo:1, class:1,
    };
    // For files with no extension (Dockerfile, Makefile, LICENSE, etc.)
    // the ext equals the full filename — treat those as text.
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
        body.innerHTML = '<div style="padding:24px">Binary file — <a href="' + url + '" download="' + entry.path.split('/').pop() + '" style="color:var(--accent)">Download</a></div>';
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
  var ext = (path.split('.').pop() || '').toLowerCase();
  var langMap = { js:'javascript', jsx:'javascript', ts:'typescript', tsx:'typescript', py:'python', rs:'rust', go:'go', html:'html', css:'css', json:'json', yaml:'yaml', yml:'yaml', md:'markup', sh:'bash', bash:'bash' };
  var lang = langMap[ext];
  var grammar = lang && Prism.languages[lang];

  var lines = text.split('\n');
  // If last line is empty (trailing newline), remove to avoid phantom line.
  if (lines[lines.length - 1] === '') { lines.pop(); }

  var lineNums = lines.map(function(_, i) { return i + 1; }).join('\n');
  var highlighted = grammar ? Prism.highlight(text, grammar, lang) : escHtml(text);

  var body = $('viewer-body');
  body.innerHTML =
    '<div id="code-wrap">' +
      '<div id="line-nums">' + lineNums + '</div>' +
      '<div id="code-content">' + highlighted + '</div>' +
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
$('commits-overlay').addEventListener('click', function(e) { if (e.target === this) { this.classList.remove('show'); } });

$('search-input').addEventListener('input', function() {
  state.filterText = this.value;
  if (state.tree) { renderTree(state.tree); }
});

boot().catch(function(e) { showToast('Boot error: ' + e.message); });
</script>
</body>
</html>`
