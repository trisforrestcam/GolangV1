#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""Convert Markdown docs -> polished HTML (light theme v2).

Features:
- Blockquote -> styled callout cards (auto-typed by emoji); fences inside
  quotes are highlighted properly (python-markdown can't do this natively).
- Unified code renderer: pygments + lang chip + Copy button, with
  language auto-guessing when missing.
- Tables wrapped for horizontal scroll, task lists, wiki-links [[...]].
- Right-side TOC with scrollspy (wide screens), progress bar, back-to-top.
- Generates index.html hub page. Vietnamese UI, Windows-friendly (no macOS chrome).

Usage:
  python3 md2html.py [docs_dir]      # default: ./docs (relative to cwd)

Requires: pip install markdown pygments
"""
import html as html_mod
import re
import sys
from pathlib import Path

import markdown
from pygments import highlight as pyg_highlight
from pygments.formatters import HtmlFormatter
from pygments.lexers import get_lexer_by_name
from pygments.util import ClassNotFound

ROOT = Path(sys.argv[1]).resolve() if len(sys.argv) > 1 else Path("docs").resolve()

MD_EXTENSIONS = ["tables", "fenced_code", "codehilite", "toc", "sane_lists"]
MD_EXT_CONFIG = {
    "codehilite": {"guess_lang": False, "pygments_style": "friendly"},
    "toc": {"permalink": False},
}

WIKI = re.compile(r"\[\[([^\]|#]+)(#[^\]|]+)?(?:\|([^\]]+))?\]\]")
EMOJI = re.compile("[\U0001F000-\U0001FAFF\u2600-\u27BF\u2B00-\u2BFF\uFE0F]")
FENCE_OPEN = re.compile(r"^\s{0,3}(`{3,}|~{3,})\s*([\w#+-]*)")

# ---------------------------------------------------------------- wiki links

def wikilink_sub(m):
    target = m.group(1).strip()
    anchor = m.group(2) or ""
    label = m.group(3)
    if not label:
        label = target + (("#" + anchor.lstrip("#")) if anchor else "")
    if anchor:
        anchor = "#" + re.sub(r"[^\w\- ]", "", anchor.lstrip("#")).strip().lower().replace(" ", "-")
    return f'<a class="wikilink" href="{target}.html{anchor}">{label}</a>'

# ---------------------------------------------------------------- code blocks

PYG_FORMATTER = HtmlFormatter(cssclass="codehilite", style="friendly", nowrap=False)

def guess_lang(code: str):
    s = code.strip()
    if not s:
        return None
    if re.search(r"\bfunc |\bpackage \w|:=|\bimport \(|\bfmt\.|\bchan\b|\bstruct \{|\binterface \{|\bnil\b", s):
        return "go"
    if re.search(r"\bfunction\b|=>|\blet |\bconst |\bvar |\bconsole\.log|\brequire\(", s):
        return "javascript"
    if re.search(r"\bSELECT\b|\bINSERT INTO\b", s, re.I):
        return "sql"
    if s.startswith("$ ") or re.search(r"^\$ ", s, re.M) or s.startswith("#!"):
        return "bash"
    if s.startswith("{") and s.endswith("}"):
        return "json"
    return None

def render_code(lang: str, code: str) -> str:
    lang = (lang or "").strip().lower() or guess_lang(code)
    try:
        lexer = get_lexer_by_name(lang or "text", stripnl=False)
    except ClassNotFound:
        lang, lexer = None, get_lexer_by_name("text", stripnl=False)
    body = pyg_highlight(code, lexer, PYG_FORMATTER)
    chip = (f'<span class="lang-chip">{html_mod.escape(lang)}</span>' if lang
            else '<span class="lang-chip"></span>')
    head = (f'<div class="codehead">{chip}'
            f'<button class="copybtn" type="button" aria-label="Copy code">Copy</button></div>')
    return f'<div class="codeblock">{head}{body}</div>'

# ---------------------------------------------------------------- fence hoisting

def split_fences(text: str):
    """Split into [('md', str) | ('code', lang, code)] preserving order."""
    parts, buf = [], []
    lines = text.split("\n")
    i, n = 0, len(lines)
    while i < n:
        m = FENCE_OPEN.match(lines[i])
        if m:
            fence, lang = m.group(1), m.group(2)
            close = re.compile(r"^\s{0,3}" + re.escape(fence[0]) + "{" + str(len(fence)) + ",}\s*$")
            j, code_lines = i + 1, []
            while j < n and not close.match(lines[j]):
                code_lines.append(lines[j])
                j += 1
            parts.append(("code", lang, "\n".join(code_lines)))
            i = j + 1
        else:
            buf.append(lines[i])
            i += 1
    if buf:
        parts.append(("md", "\n".join(buf)))
    return parts

def hoist_fences(text: str, stash) -> str:
    out = []
    for part in split_fences(text):
        if part[0] == "md":
            out.append(part[1])
        else:
            out.extend(["", stash(render_code(part[1], part[2])), ""])
    return "\n".join(out)

# ---------------------------------------------------------------- blockquotes

CALLOUT_MAP = [
    ("📦", "compare"), ("🆚", "compare"), ("⚖", "compare"),
    ("⚠", "warning"), ("❗", "warning"), ("🚨", "warning"), ("🪤", "warning"),
    ("⛔", "warning"), ("🐛", "warning"), ("💢", "warning"),
    ("💡", "tip"), ("🤔", "tip"), ("⭐", "tip"), ("🌟", "tip"), ("🏆", "tip"),
    ("✅", "success"), ("✔", "success"), ("👍", "success"), ("🎉", "success"),
    ("❌", "danger"), ("🛑", "danger"), ("🔴", "danger"), ("🔥", "danger"),
    ("📌", "info"), ("🎯", "info"), ("🧠", "info"), ("⚡", "info"), ("🚀", "info"),
    ("🔭", "info"), ("🔍", "info"), ("🎓", "info"), ("⏱", "info"), ("❓", "info"),
]

def classify_callout(text: str) -> str:
    m = EMOJI.search(text)
    if m:
        for icon, cls in CALLOUT_MAP:
            if m.group(0).startswith(icon):
                return cls
    return "note"

def render_segments(text: str, stash) -> str:
    text = hoist_fences(text, stash)
    md = markdown.Markdown(extensions=MD_EXTENSIONS, extension_configs=MD_EXT_CONFIG)
    body = md.convert(text)
    body = re.sub(r"<li>\[ \]\s*", '<li class="task"><input type="checkbox" disabled> ', body)
    body = re.sub(r"<li>\[[xX]\]\s*", '<li class="task"><input type="checkbox" checked disabled> ', body)
    return body

def transform_blockquotes(text: str, stash, depth: int = 0) -> str:
    out, buf = [], []

    def flush():
        nonlocal buf
        if buf:
            out.append("")
            out.append(render_blockquote(buf, stash, depth))
            out.append("")
            buf = []

    for line in text.split("\n"):
        if re.match(r"^\s*>", line):
            buf.append(line)
        else:
            flush()
            out.append(line)
    flush()
    return "\n".join(out)

def strip_quote_prefix(line: str) -> str:
    return re.sub(r"^\s*>\s?", "", line)

def render_blockquote(qlines, stash, depth: int) -> str:
    inner = "\n".join(strip_quote_prefix(l) for l in qlines)
    if depth < 3:
        inner = transform_blockquotes(inner, stash, depth + 1)

    # Extract a title line: starts with emoji, e.g. "**📦 Node.js Comparison:**"
    title = icon = None
    lines = inner.split("\n")
    for i, l in enumerate(lines):
        s = l.strip()
        if not s:
            continue
        if FENCE_OPEN.match(s):
            break
        core = re.sub(r"^\*+|\*+$", "", s).strip()
        m = EMOJI.search(core[:5])
        if m and len(core) <= 80:
            icon = m.group(0)
            title = core.replace(icon, "", 1).strip().rstrip(":").rstrip("*").strip()
            lines = lines[:i] + lines[i + 1:]
        break
    inner = "\n".join(lines)

    cls = classify_callout((icon or "") + (title or ""))
    rendered = render_segments(inner, stash)
    head = ""
    if title:
        head = (f'<div class="callout-title"><span class="callout-icon">{icon}</span>'
                f'<span>{html_mod.escape(title)}</span></div>')
    return (f'<div class="callout callout-{cls}">{head}'
            f'<div class="callout-body">{rendered}</div></div>')

# ---------------------------------------------------------------- full render

def md_to_body(text: str) -> str:
    text = WIKI.sub(wikilink_sub, text)
    tokens = []

    def stash(block_html: str) -> str:
        tok = f"@@BLOCK{len(tokens)}@@"
        tokens.append((tok, block_html))
        return tok

    text = transform_blockquotes(text, stash)
    text = hoist_fences(text, stash)
    body = markdown.Markdown(
        extensions=MD_EXTENSIONS, extension_configs=MD_EXT_CONFIG).convert(text)
    body = re.sub(r"<li>\[ \]\s*", '<li class="task"><input type="checkbox" disabled> ', body)
    body = re.sub(r"<li>\[[xX]\]\s*", '<li class="task"><input type="checkbox" checked disabled> ', body)
    body = body.replace("<table>", '<div class="tablewrap"><table>')
    body = body.replace("</table>", "</table></div>")

    for tok, block in reversed(tokens):  # reverse: nested tokens live inside outer blocks
        body = body.replace(f"<p>{tok}</p>", block)
        body = body.replace(tok, block)
    return body

# ---------------------------------------------------------------- template

PYGMENTS_CSS = HtmlFormatter(style="friendly").get_style_defs(".codehilite")

CSS = """
:root {
  --bg1:#f7f9fd; --bg2:#eef3fb; --card:#ffffff;
  --ink:#1e293b; --ink-strong:#0f172a; --muted:#64748b;
  --accent:#2563eb; --accent-2:#4f46e5; --accent-soft:#eff6ff;
  --border:#e2e8f0; --code-bg:#f6f8fa;
}
* { box-sizing: border-box; }
html { scroll-behavior: smooth; }
body {
  margin: 0; color: var(--ink); font-size: 16px; line-height: 1.75;
  font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue",
               Arial, "Noto Sans", "Noto Sans Vietnamese", sans-serif;
  background: linear-gradient(160deg, var(--bg1) 0%, var(--bg2) 55%, #e9f0fa 100%) fixed;
}
#progress { position: fixed; top: 0; left: 0; height: 3px; width: 0; z-index: 70;
  background: linear-gradient(90deg, var(--accent), var(--accent-2), #7c3aed); }
#backtop { position: fixed; right: 26px; bottom: 26px; width: 44px; height: 44px;
  border-radius: 50%; border: 1px solid var(--border); background: #fff;
  color: var(--accent); font-size: 1.15rem; cursor: pointer; opacity: 0;
  pointer-events: none; transition: all .25s; box-shadow: 0 8px 20px rgba(15,23,42,.14); z-index: 60; }
#backtop.show { opacity: 1; pointer-events: auto; }
#backtop:hover { background: var(--accent); color: #fff; }
.wrap { max-width: 1240px; margin: 0 auto; padding: 28px 24px 72px; }
.crumb { max-width: 1240px; margin: 0 auto; padding: 18px 24px 0; font-size: .9rem; color: var(--muted); }
.crumb a { color: var(--accent); text-decoration: none; font-weight: 600; }
.crumb a:hover { text-decoration: underline; }
.card {
  background: var(--card); border: 1px solid var(--border); border-radius: 18px;
  padding: 44px 52px 52px; box-shadow: 0 14px 40px rgba(15, 23, 42, .07);
  margin-top: 14px;
}
@media (max-width: 640px) { .card { padding: 26px 18px 34px; } }
h1, h2, h3, h4, h5, h6 { color: var(--ink-strong); line-height: 1.3; font-weight: 700; }
h1 {
  font-size: 2.1rem; margin: 0 0 .7em; padding-bottom: .45em;
  border-bottom: 3px solid var(--accent-soft);
  background: linear-gradient(90deg, var(--accent), var(--accent-2));
  -webkit-background-clip: text; background-clip: text; -webkit-text-fill-color: transparent;
}
h2 { font-size: 1.5rem; margin: 2em 0 .6em; padding-bottom: .3em; border-bottom: 2px solid #edf2f9; }
h3 { font-size: 1.22rem; margin: 1.6em 0 .5em; }
h3::before { content: "▍"; color: var(--accent); margin-right: 6px; font-weight: 400; }
h4 { font-size: 1.05rem; margin: 1.4em 0 .4em; }
p { margin: .8em 0; }
a { color: var(--accent); text-decoration: none; }
a:hover { text-decoration: underline; }
a.wikilink { font-weight: 600; }
ul, ol { padding-left: 1.6em; margin: .8em 0; }
li { margin: .3em 0; }
li.task { list-style: none; margin-left: -1.4em; }
li.task input { margin-right: .5em; accent-color: var(--accent); }
code {
  font-family: "JetBrains Mono", "Fira Code", "SF Mono", Consolas, Menlo, monospace;
  font-size: .88em; background: #eef2f8; color: #be185d;
  padding: .14em .42em; border-radius: 6px; border: 1px solid #e2e8f0;
}
/* --- code blocks --- */
.codeblock { margin: 1.5em 0; border-radius: 12px; border: 1px solid var(--border);
  background: var(--code-bg); overflow: hidden; box-shadow: 0 2px 6px rgba(15,23,42,.05); }
.codehead { display: flex; align-items: center; justify-content: space-between;
  padding: 6px 14px 6px 16px; background: #edf1f7; border-bottom: 1px solid var(--border); }
.lang-chip { font-size: .7rem; letter-spacing: .1em; text-transform: uppercase;
  font-weight: 800; color: #94a3b8; }
.copybtn { display: inline-flex; align-items: center; gap: 5px; border: 1px solid var(--border);
  background: #fff; color: #64748b; font-size: .72rem; font-weight: 700; letter-spacing: .03em;
  padding: 3px 11px; border-radius: 6px; cursor: pointer; transition: all .15s; font-family: inherit; }
.copybtn:hover { color: var(--accent); border-color: var(--accent); }
.copybtn.copied { color: #15803d; border-color: #86efac; background: #f0fdf4; }
.codeblock .codehilite { border: none; border-radius: 0; margin: 0; box-shadow: none; }
.codeblock .codehilite pre { margin: 0; }
.codehilite {
  background: var(--code-bg); border: 1px solid var(--border); border-radius: 12px;
  padding: 16px 18px; margin: 1.2em 0; overflow-x: auto;
}
.codehilite pre { margin: 0; background: none; }
.codehilite code, pre code { background: none; border: none; color: inherit; padding: 0;
  font-size: .875rem; line-height: 1.65; }
.callout .codeblock { box-shadow: none; }
.callout .codeblock, .callout .codehilite, .callout code { background: #fff; }
.callout .codehead { background: rgba(237,241,247,.6); }
/* --- tables --- */
.tablewrap { overflow-x: auto; margin: 1.2em 0; }
table { border-collapse: collapse; width: 100%; font-size: .94em; margin: 0; }
th, td { border: 1px solid var(--border); padding: 9px 13px; text-align: left; vertical-align: top; }
th { background: #f1f5f9; color: var(--ink-strong); font-weight: 700; white-space: nowrap; }
tbody tr:nth-child(even) { background: #f8fafc; }
tbody tr:hover { background: var(--accent-soft); }
.callout table { background: #fff; }
/* --- callouts (blockquotes) --- */
.callout { border-radius: 14px; border: 1px solid var(--border); padding: 16px 20px;
  margin: 1.5em 0; background: #f8fafc; }
.callout .callout-title { display: flex; align-items: center; gap: 10px; font-weight: 800;
  margin: 0 0 .5em; font-size: .98rem; letter-spacing: .01em; }
.callout .callout-icon { font-size: 1.1rem; line-height: 1; }
.callout .callout-body > :first-child { margin-top: 0; }
.callout .callout-body > :last-child { margin-bottom: 0; }
.callout-info    { background: #eff6ff; border-color: #bfdbfe; }
.callout-info    .callout-title { color: #1d4ed8; }
.callout-tip     { background: #fefce8; border-color: #fde68a; }
.callout-tip     .callout-title { color: #b45309; }
.callout-warning { background: #fff7ed; border-color: #fed7aa; }
.callout-warning .callout-title { color: #c2410c; }
.callout-danger  { background: #fef2f2; border-color: #fecaca; }
.callout-danger  .callout-title { color: #b91c1c; }
.callout-success { background: #f0fdf4; border-color: #bbf7d0; }
.callout-success .callout-title { color: #15803d; }
.callout-compare { background: #eef2ff; border-color: #c7d2fe; }
.callout-compare .callout-title { color: #4338ca; }
.callout-note    { background: #f8fafc; border-color: var(--border); }
.callout-note    .callout-title { color: #334155; }
blockquote { margin: 1.2em 0; padding: .7em 1.3em; color: #334155;
  border-left: 4px solid var(--accent); background: var(--accent-soft);
  border-radius: 0 12px 12px 0; }
blockquote p { margin: .3em 0; }
hr { border: none; border-top: 1px solid var(--border); margin: 2.2em 0; }
kbd { background: #fff; border: 1px solid #cbd5e1; border-bottom-width: 2px;
  border-radius: 6px; padding: 1px 6px; font-size: .85em; font-family: inherit; }
/* --- TOC --- */
.toc { position: fixed; top: 84px; right: 26px; width: 250px; max-height: calc(100vh - 130px);
  overflow-y: auto; font-size: .82rem; background: rgba(255,255,255,.92);
  backdrop-filter: blur(8px); border: 1px solid var(--border); border-radius: 14px;
  padding: 14px 12px; box-shadow: 0 10px 28px rgba(15,23,42,.07); display: none; z-index: 50; }
.toc-title { font-weight: 800; font-size: .72rem; letter-spacing: .1em;
  text-transform: uppercase; color: var(--muted); margin: 0 0 8px 10px; }
.toc a { display: block; color: #475569; padding: 3px 6px 3px 10px;
  border-left: 2px solid transparent; line-height: 1.45; }
.toc a.l3 { padding-left: 24px; font-size: .94em; }
.toc a:hover { color: var(--accent); text-decoration: none; }
.toc a.active { color: var(--accent); border-left-color: var(--accent); font-weight: 600; }
@media (min-width: 1800px) { .toc { display: block; } }
/* --- pager & footer --- */
.pager { display: flex; justify-content: space-between; gap: 16px; margin-top: 26px; }
.pager a {
  flex: 1; background: var(--card); border: 1px solid var(--border); border-radius: 12px;
  padding: 12px 18px; color: var(--ink-strong); font-weight: 600;
  box-shadow: 0 4px 12px rgba(15,23,42,.05); transition: all .2s;
}
.pager a:hover { border-color: var(--accent); text-decoration: none;
  box-shadow: 0 6px 16px rgba(37,99,235,.14); transform: translateY(-1px); }
.pager .next { text-align: right; }
.pager .dir { display: block; font-size: .78rem; color: var(--muted); font-weight: 500; }
footer { text-align: center; color: var(--muted); font-size: .85rem; margin-top: 34px; }
@media print { #progress, #backtop, .toc, .pager { display: none !important; }
  body { background: #fff; } .card { box-shadow: none; border: none; } }
"""

INDEX_CSS = """
.wrap.wide { max-width: 1340px; }
.hero { text-align: center; padding: 64px 10px 26px; }
.hero-emoji { font-size: 3.4rem; margin-bottom: 8px; }
.hero h1 { font-size: 2.6rem; border: none; display: inline-block;
  background: linear-gradient(90deg, #2563eb, #4f46e5, #7c3aed);
  -webkit-background-clip: text; background-clip: text; -webkit-text-fill-color: transparent; }
.hero .sub { color: var(--muted); font-size: 1.05rem; }
.stats { display: flex; gap: 10px; justify-content: center; margin-top: 20px; flex-wrap: wrap; }
.stat { background: #fff; border: 1px solid var(--border); border-radius: 999px;
  padding: 6px 16px; font-size: .85rem; font-weight: 700; color: #475569;
  box-shadow: 0 4px 12px rgba(15,23,42,.05); }
.grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(400px, 1fr));
  gap: 22px; margin-top: 30px; }
.series-card { background: var(--card); border: 1px solid var(--border); border-radius: 18px;
  padding: 26px 28px; box-shadow: 0 12px 32px rgba(15,23,42,.06); }
.series-card h2 { border: none; margin: 0; font-size: 1.28rem; display: flex;
  align-items: center; gap: 10px; flex-wrap: wrap; }
.series-icon { font-size: 1.5rem; }
.badge { background: var(--accent-soft); color: var(--accent); font-size: .72rem;
  padding: 3px 11px; border-radius: 999px; font-weight: 800; }
.desc { color: var(--muted); font-size: .92rem; margin: 8px 0 12px; line-height: 1.6; }
.filelist { list-style: none; padding: 0; margin: 0; }
.filelist li { border-top: 1px dashed #eef2f7; margin: 0; }
.filelist a { display: flex; justify-content: space-between; align-items: center;
  gap: 12px; padding: 10px 4px; color: var(--ink); font-weight: 600; transition: all .15s; }
.filelist a::after { content: "›"; color: var(--accent); font-weight: 800; font-size: 1.1rem; }
.filelist a:hover { color: var(--accent); text-decoration: none; padding-left: 10px; }
@media (max-width: 500px) { .grid { grid-template-columns: 1fr; } }
"""

CSS = CSS + INDEX_CSS

SCRIPT = """
(function(){
  var bar=document.getElementById('progress');
  var top=document.getElementById('backtop');
  var links=[].slice.call(document.querySelectorAll('.toc a'));
  function onScroll(){
    var h=document.documentElement;
    var max=h.scrollHeight-h.clientHeight||1;
    bar.style.width=(h.scrollTop/max*100).toFixed(2)+'%';
    top.classList.toggle('show', h.scrollTop>600);
    if(links.length){
      var cur=links[0].getAttribute('href').slice(1);
      for(var i=0;i<links.length;i++){
        var el=document.getElementById(links[i].getAttribute('href').slice(1));
        if(el&&el.getBoundingClientRect().top<=110) cur=links[i].getAttribute('href').slice(1);
      }
      links.forEach(function(a){a.classList.toggle('active',a.getAttribute('href')==='#'+cur)});
    }
  }
  window.addEventListener('scroll', onScroll, {passive:true});
  top.addEventListener('click', function(){window.scrollTo({top:0,behavior:'smooth'})});
  onScroll();
  // copy-code button (event delegation)
  document.addEventListener('click', function(e){
    var btn = e.target.closest ? e.target.closest('.copybtn') : null;
    if(!btn) return;
    var block = btn.closest('.codeblock');
    var pre = block ? block.querySelector('pre') : null;
    if(!pre) return;
    var text = pre.innerText;
    function done(ok){
      btn.classList.add('copied');
      btn.textContent = ok ? '✓ Copied!' : '✗ Error';
      setTimeout(function(){ btn.classList.remove('copied'); btn.textContent = 'Copy'; }, 1600);
    }
    function fallback(){
      var ta = document.createElement('textarea');
      ta.value = text; ta.style.position = 'fixed'; ta.style.opacity = '0';
      document.body.appendChild(ta); ta.select();
      var ok = false;
      try { ok = document.execCommand('copy'); } catch(err) {}
      document.body.removeChild(ta);
      done(ok);
    }
    if (navigator.clipboard && navigator.clipboard.writeText) {
      navigator.clipboard.writeText(text).then(function(){ done(true); }, fallback);
    } else { fallback(); }
  });
})();
"""

TPL = """<!DOCTYPE html>
<html lang="vi">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>{title}</title>
<style>
{pygments}
{css}
</style>
</head>
<body>
<div id="progress"></div>
<button id="backtop" aria-label="Lên đầu trang">↑</button>
{toc_html}
<nav class="crumb"><a href="{up}index.html">📚 Docs</a> &nbsp;›&nbsp; {folder}</nav>
<main class="wrap">
<article class="card">
{body}
</article>
<nav class="pager">
{prev}
{next}
</nav>
<footer>Generated from Markdown · <a href="{up}index.html">Về trang chủ docs</a></footer>
</main>
<script>
{script}
</script>
</body>
</html>
"""

TOC_TPL = ('<aside class="toc"><div class="toc-title">📋 Nội dung</div>{items}</aside>')


def first_h1(html_body: str) -> str:
    m = re.search(r"<h1[^>]*>(.*?)</h1>", html_body, re.S)
    if not m:
        return ""
    return re.sub(r"<[^>]+>", "", m.group(1)).strip()


def build_toc(body: str) -> str:
    items = re.findall(r'<h([23])\s[^>]*?id="([^"]+)"[^>]*>(.*?)</h\1>', body)
    if len(items) < 4:
        return ""
    links = []
    for level, hid, text in items:
        text = re.sub(r"<[^>]+>", "", text).strip()
        if len(text) > 60:
            text = text[:57] + "…"
        links.append(f'<a class="l{level}" href="#{hid}">{text}</a>')
    return TOC_TPL.format(items="".join(links))


SERIES_META = {
    "go-fundamentals": ("🐹", "Go Fundamentals",
                        "Nền tảng Go cho Node.js developer: mental model, syntax, struct/interface, concurrency, HTTP API, tooling & roadmap 7 ngày."),
    "nodejs-to-go": ("🔄", "Node.js → Go",
                     "Series chuyển đổi: phân tích project, so sánh Go vs Node.js, patterns thực tế và hướng dẫn Fiber framework hoàn chỉnh."),
}


def build_files():
    results = {}
    by_folder = {}
    for f in sorted(ROOT.rglob("*.md")):
        by_folder.setdefault(f.parent, []).append(f)

    for folder, files in by_folder.items():
        folder_name = folder.name
        converted = []
        for md in files:
            depth = len(md.parent.relative_to(ROOT).parts)  # 1 for docs/xxx/, 2 for docs/x/y/
            up = "../" * depth
            body = md_to_body(md.read_text(encoding="utf-8"))
            converted.append((md, body, first_h1(body) or md.stem, up))

        for i, (md, body, title, up) in enumerate(converted):
            prev_html = next_html = ""
            if len(converted) > 1:
                if i > 0:
                    pb, pt = converted[i - 1][0].stem, converted[i - 1][2]
                    prev_html = (f'<a href="{pb}.html"><span class="dir">← Bài trước</span>'
                                 f'{pt[:48]}</a>')
                if i < len(converted) - 1:
                    nb, nt = converted[i + 1][0].stem, converted[i + 1][2]
                    next_html = (f'<a class="next" href="{nb}.html">'
                                 f'<span class="dir">Bài sau →</span>{nt[:48]}</a>')

            html_out = TPL.format(
                title=f"{title} · {folder_name}",
                folder=folder_name,
                up=up,
                body=body,
                toc_html=build_toc(body),
                prev=prev_html,
                next=next_html,
                pygments=PYGMENTS_CSS,
                css=CSS,
                script=SCRIPT,
            )
            out = md.with_suffix(".html")
            out.write_text(html_out, encoding="utf-8")
            print(f"  ✓ {out.relative_to(ROOT)}")

        results[folder_name] = [(md.stem, t) for md, _, t, _ in converted]
    return results


def build_index(results):
    total = sum(len(v) for v in results.values())
    cards = ""
    for folder in sorted(results):
        icon, name, desc = SERIES_META.get(folder, ("📄", folder, ""))
        items = "".join(
            f'<li><a href="{folder}/{stem}.html"><span>{title[:70]}</span></a></li>'
            for stem, title in results[folder]
        )
        cards += f"""
<section class="series-card">
  <h2><span class="series-icon">{icon}</span> {name} <span class="badge">{len(results[folder])} bài</span></h2>
  <p class="desc">{desc}</p>
  <ul class="filelist">{items}</ul>
</section>"""

    index = f"""<!DOCTYPE html>
<html lang="vi">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>🐹 Learn Go — Tài liệu cho Node.js Developer</title>
<style>
{CSS}
</style>
</head>
<body>
<div id="progress"></div>
<button id="backtop" aria-label="Lên đầu trang">↑</button>
<main class="wrap wide">
<header class="hero">
  <div class="hero-emoji">🐹</div>
  <h1>Learn Go — Dành cho Node.js Developer</h1>
  <p class="sub">Series tài liệu học Go từ cơ bản đến nâng cao, viết riêng cho developer<br>có kinh nghiệm Node.js/TypeScript. Chọn một series để bắt đầu:</p>
  <div class="stats">
    <span class="stat">📚 2 series</span>
    <span class="stat">📄 {total} tài liệu</span>
    <span class="stat">⏱ ~12.000+ dòng</span>
  </div>
</header>
<div class="grid">
{cards}
</div>
<footer>Generated from Markdown · Light theme 🌞</footer>
</main>
<script>
{SCRIPT}
</script>
</body>
</html>"""
    (ROOT / "index.html").write_text(index, encoding="utf-8")
    print(f"  ✓ index.html ({total} docs)")


if __name__ == "__main__":
    print(f"Converting markdown -> HTML (light theme v2) · {ROOT}")
    res = build_files()
    build_index(res)
    print("Done.")
