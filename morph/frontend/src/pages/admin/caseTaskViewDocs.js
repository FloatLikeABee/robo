const CASE_TASK_DETAIL_MAX_NESTING = 5;

function escapeHtml(s) {
  return String(s || '')
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;');
}

function prettyCaseTaskDetailLabel(key) {
  let s = String(key || '').trim().replace(/_/g, ' ');
  s = s.replace(/([a-z0-9])([A-Z])/g, '$1 $2');
  s = s.replace(/\s+/g, ' ').trim();
  return s
    .split(' ')
    .filter(Boolean)
    .map((w) => (w.toLowerCase() === 'id' ? 'ID' : w.charAt(0).toUpperCase() + w.slice(1).toLowerCase()))
    .join(' ');
}

function formatCaseTaskDetailScalar(value) {
  if (value === null || value === undefined) return '—';
  if (typeof value === 'boolean') return value ? 'Yes' : 'No';
  if (typeof value === 'number' && Number.isFinite(value) && Number.isInteger(value)) return String(value);
  if (typeof value === 'number') return String(value);
  if (typeof value === 'string') return value;
  try {
    return JSON.stringify(value);
  } catch {
    return String(value);
  }
}

function isPlainObject(v) {
  return v !== null && typeof v === 'object' && !Array.isArray(v) && !(v instanceof Date);
}

function isEmptyDetailValue(v) {
  if (v == null) return true;
  if (isPlainObject(v) && Object.keys(v).length === 0) return true;
  if (Array.isArray(v) && v.length === 0) return true;
  return false;
}

function parseCaseTaskDetail(detail) {
  if (detail == null || detail === '') {
    return { kind: 'empty' };
  }
  if (typeof detail === 'object' && !(detail instanceof Date)) {
    if (isEmptyDetailValue(detail)) return { kind: 'empty' };
    return { kind: 'value', value: detail };
  }
  const raw = String(detail).trim();
  if (!raw || raw === '{}' || raw === '[]') return { kind: 'empty', raw };
  try {
    const value = JSON.parse(raw);
    if (isEmptyDetailValue(value)) return { kind: 'empty', raw };
    return { kind: 'value', value, raw };
  } catch {
    return { kind: 'invalid', raw };
  }
}

function sortedDetailKeys(obj) {
  return Object.keys(obj).sort();
}

function caseTaskDetailHeading(depth, label) {
  const n = Math.min(6, Math.max(3, depth + 3));
  return `${'#'.repeat(n)} ${label}`;
}

function markdownList(arr, depth) {
  if (!arr.length) return 'Empty list\n\n';
  let out = '';
  for (const item of arr) {
    if (isPlainObject(item)) {
      const inner = markdownValue(item, depth + 1)
        .trim()
        .replace(/\n\n/g, '; ')
        .replace(/\n/g, ' ');
      out += `- ${inner}\n`;
      continue;
    }
    if (Array.isArray(item)) {
      out += `- \n${markdownList(item, depth + 1)}`;
      continue;
    }
    out += `- ${formatCaseTaskDetailScalar(item)}\n`;
  }
  return `${out}\n`;
}

function markdownValue(value, depth) {
  if (depth > CASE_TASK_DETAIL_MAX_NESTING) {
    return `${formatCaseTaskDetailScalar(value)}\n\n`;
  }
  if (isPlainObject(value)) {
    const keys = sortedDetailKeys(value);
    if (!keys.length) return 'Empty\n\n';
    let out = '';
    for (const k of keys) {
      const label = prettyCaseTaskDetailLabel(k);
      const child = value[k];
      if (isPlainObject(child)) {
        out += `${caseTaskDetailHeading(depth, label)}\n\n${markdownValue(child, depth + 1)}`;
        continue;
      }
      if (Array.isArray(child)) {
        out += `${caseTaskDetailHeading(depth, label)}\n\n${markdownList(child, depth + 1)}`;
        continue;
      }
      out += `**${label}:** ${formatCaseTaskDetailScalar(child)}\n\n`;
    }
    return out;
  }
  if (Array.isArray(value)) return markdownList(value, depth);
  return `${formatCaseTaskDetailScalar(value)}\n\n`;
}

function detailJSONToMarkdown(detail) {
  const parsed = parseCaseTaskDetail(detail);
  if (parsed.kind === 'empty') return '';
  if (parsed.kind === 'invalid') return `## Detail\n\n${parsed.raw}\n`;
  return `## Detail\n\n${markdownValue(parsed.value, 0)}`;
}

function htmlValue(value, depth) {
  if (depth > CASE_TASK_DETAIL_MAX_NESTING) {
    return `<p class="detail-value">${escapeHtml(formatCaseTaskDetailScalar(value))}</p>`;
  }
  if (isPlainObject(value)) {
    const keys = sortedDetailKeys(value);
    if (!keys.length) return '<p class="detail-muted">Empty</p>';
    if (depth === 0) {
      return keys
        .map((k) => {
          const label = prettyCaseTaskDetailLabel(k);
          return `<article class="detail-card"><h3>${escapeHtml(label)}</h3>${htmlValue(value[k], depth + 1)}</article>`;
        })
        .join('');
    }
    const inner = keys
      .map((k) => {
        const label = prettyCaseTaskDetailLabel(k);
        return `<div class="detail-field"><div class="detail-label">${escapeHtml(label)}</div>${htmlValue(value[k], depth + 1)}</div>`;
      })
      .join('');
    return `<div class="detail-nested">${inner}</div>`;
  }
  if (Array.isArray(value)) {
    if (!value.length) return '<p class="detail-muted">Empty list</p>';
    const items = value
      .map((item, i) => {
        const itemLabel = isPlainObject(item) ? `<div class="detail-item-label">Item ${i + 1}</div>` : '';
        return `<li>${itemLabel}${htmlValue(item, depth + 1)}</li>`;
      })
      .join('');
    return `<ul class="detail-list">${items}</ul>`;
  }
  return `<p class="detail-value">${escapeHtml(formatCaseTaskDetailScalar(value))}</p>`;
}

function detailJSONToHTMLFragment(detail) {
  const parsed = parseCaseTaskDetail(detail);
  if (parsed.kind === 'empty') return '';
  if (parsed.kind === 'invalid') {
    return `<section class="detail-ui"><h2>Detail</h2><pre class="detail-fallback">${escapeHtml(parsed.raw)}</pre></section>`;
  }
  return `<section class="detail-ui"><h2>Detail</h2>${htmlValue(parsed.value, 0)}</section>`;
}

function caseTaskDetailUICSS() {
  return `.detail-ui{margin-top:1.5rem}
.detail-ui h2{font:700 1.3rem/1.25 system-ui,sans-serif;color:#f8fafc;margin:0 0 .85rem}
.detail-card{background:var(--card);border:1px solid var(--line);border-radius:12px;padding:1rem 1.1rem;margin:0 0 .85rem}
.detail-card h3{font:700 .95rem/1.3 system-ui,sans-serif;color:var(--accent);margin:0 0 .5rem;letter-spacing:.02em}
.detail-label{font:700 11px/1.3 system-ui,sans-serif;color:var(--muted);margin:0 0 .2rem}
.detail-value{font:15px/1.45 system-ui,sans-serif;color:var(--ink);margin:0;word-break:break-word;white-space:pre-wrap}
.detail-muted{font:13px/1.4 system-ui,sans-serif;color:var(--muted);margin:0}
.detail-nested{border-left:2px solid var(--line);padding-left:.85rem;margin:.35rem 0}
.detail-field{margin:.55rem 0}
.detail-list{margin:.35rem 0 .2rem;padding-left:1.2em}
.detail-item-label{font:700 11px/1.3 system-ui,sans-serif;color:var(--accent);margin:.15rem 0 .25rem}
.detail-fallback{overflow:auto;padding:1em;border-radius:10px;background:var(--card);border:1px solid var(--line);font:13px/1.45 ui-monospace,SFMono-Regular,Menlo,Consolas,monospace;white-space:pre-wrap;word-break:break-word;color:var(--ink)}`;
}

function stripDetailSection(markdown) {
  const text = String(markdown || '');
  const m = text.match(/\n## Detail(?:\n|$)/);
  if (m && m.index >= 0) return text.slice(0, m.index).replace(/[ \t\r\n]+$/, '');
  if (/^## Detail(?:\n|$)/.test(text.trimStart())) return '';
  return text;
}

export function formatCaseTaskViewTime(value) {
  const s = String(value || '').trim();
  if (!s) return '';
  if (s.includes('T')) {
    const [d, rest] = s.split('T');
    const hm = (rest || '').slice(0, 5);
    if (d && hm) return `${d} ${hm}`;
  }
  return s.replace('T', ' ').slice(0, 16);
}

export function caseTaskLocationSummary(raw) {
  if (!raw) return '';
  const parsed =
    typeof raw === 'string'
      ? (() => {
          try {
            return JSON.parse(raw);
          } catch {
            return null;
          }
        })()
      : raw;
  if (!parsed || typeof parsed !== 'object') {
    return String(raw).trim();
  }
  const label =
    typeof parsed.label === 'string'
      ? parsed.label.trim()
      : typeof parsed.location === 'string'
        ? parsed.location.trim()
        : '';
  const area = Array.isArray(parsed.area) ? parsed.area : [];
  if (label && area.length > 0) return `${label} (${area.length} points)`;
  if (label) return label;
  if (area.length > 0) return `Area (${area.length} points)`;
  return '';
}

export function buildCaseTaskMarkdown({
  title = '',
  description = '',
  assignees = '',
  startAt = '',
  endAt = '',
  location = '',
  detail = '',
} = {}) {
  const heading = String(title || '').trim() || 'Case/task';
  const parts = [`# ${heading}`, ''];
  const a = String(assignees || '').trim();
  if (a) parts.push(`**Assignees:** ${a}`, '');
  const start = formatCaseTaskViewTime(startAt);
  if (start) parts.push(`**Start:** ${start}`, '');
  const end = formatCaseTaskViewTime(endAt);
  if (end) parts.push(`**End:** ${end}`, '');
  const loc = caseTaskLocationSummary(location);
  if (loc) parts.push(`**Location:** ${loc}`, '');
  const desc = String(description || '').trim();
  if (desc) parts.push(desc, '');
  const detailMd = detailJSONToMarkdown(detail);
  if (detailMd) parts.push(detailMd.replace(/\n$/, ''), '');
  return parts.join('\n');
}

function stripLeadingTitleHeading(markdown, title) {
  const md = String(markdown || '').replace(/^[ \t\r\n]+/, '');
  const t = String(title || '').trim();
  if (!md || !t) return markdown || '';
  const lines = md.split('\n');
  const first = (lines[0] || '').trim();
  if (first.startsWith('# ') && !first.startsWith('##')) {
    const heading = first.slice(1).trim();
    if (heading.toLowerCase() === t.toLowerCase()) {
      return lines.slice(1).join('\n').replace(/^[ \t\r\n]+/, '');
    }
  }
  return markdown || '';
}

function inlineMarkdown(s) {
  let e = escapeHtml(s);
  e = e.replace(/\*\*(.+?)\*\*/g, '<strong>$1</strong>');
  e = e.replace(/`([^`]+)`/g, '<code>$1</code>');
  return e;
}

function markdownToHTMLFragment(markdown) {
  const text = String(markdown || '')
    .replace(/\r\n/g, '\n')
    .trim();
  if (!text) return '';
  const codes = [];
  const withPh = text.replace(/```(\w*)\n([\s\S]*?)```/g, (_, _lang, body) => {
    const i = codes.length;
    codes.push(`<pre><code>${escapeHtml(String(body).replace(/\n$/, ''))}</code></pre>`);
    return `\n\n%%CODE${i}%%\n\n`;
  });
  return withPh
    .split(/\n{2,}/)
    .map((block) => {
      const t = block.trim();
      if (!t) return '';
      const codeM = t.match(/^%%CODE(\d+)%%$/);
      if (codeM) return codes[Number(codeM[1])] || '';
      if (t.startsWith('### ')) return `<h3>${inlineMarkdown(t.slice(4))}</h3>`;
      if (t.startsWith('## ')) return `<h2>${inlineMarkdown(t.slice(3))}</h2>`;
      if (t.startsWith('# ')) return `<h1>${inlineMarkdown(t.slice(2))}</h1>`;
      return `<p>${inlineMarkdown(t).replace(/\n/g, '<br/>')}</p>`;
    })
    .join('');
}

export function buildCaseTaskHTML({ title = '', markdown = '', detail } = {}) {
  const heading = String(title || '').trim() || 'Case/task';
  const mdBody = stripDetailSection(stripLeadingTitleHeading(markdown, heading));
  const rendered = markdownToHTMLFragment(mdBody);
  const detailFrag = detailJSONToHTMLFragment(detail);
  const proseCSS = `.prose{font-size:1.05rem;color:var(--ink)}
.prose>:first-child{margin-top:0}
.prose h1,.prose h2,.prose h3,.prose h4{line-height:1.25;margin:1.35em 0 .55em;font-weight:700;color:#f8fafc}
.prose h1{font-size:1.55rem}.prose h2{font-size:1.3rem}.prose h3{font-size:1.12rem}
.prose p,.prose ul,.prose ol,.prose blockquote,.prose pre,.prose table{margin:.85em 0}
.prose ul,.prose ol{padding-left:1.4em}
.prose li{margin:.35em 0}
.prose blockquote{padding:.35em 0 .35em 1em;border-left:3px solid var(--accent);color:var(--muted)}
.prose code{font-family:ui-monospace,SFMono-Regular,Menlo,Consolas,monospace;font-size:.9em;background:var(--card);padding:.12em .35em;border-radius:4px;border:1px solid var(--line)}
.prose pre{overflow:auto;padding:1em;border-radius:10px;background:var(--card);border:1px solid var(--line)}
.prose a{color:var(--accent)}
.prose hr{border:0;border-top:1px solid var(--line);margin:1.5em 0}
.prose strong{font-weight:700}
.prose table{border-collapse:collapse;width:100%;font-size:.95em}
.prose th,.prose td{border:1px solid var(--line);padding:.45em .6em;text-align:left}
.prose th{background:var(--card)}`;
  const css = `:root{color-scheme:dark;--ink:#e8eef7;--muted:#94a3b8;--line:#1e293b;--bg:#0b1220;--card:#111827;--accent:#38bdf8;--scrollbar-thumb:#64748b;--scrollbar-track:#0b1220;scrollbar-width:thin;scrollbar-color:var(--scrollbar-thumb) var(--scrollbar-track)}
html,body{height:100%;margin:0}
*{box-sizing:border-box}body{font-family:Georgia,"Times New Roman",serif;color:var(--ink);background:radial-gradient(1200px 600px at 10% -10%,#1e293b 0%,var(--bg) 55%);line-height:1.55}
.wrap{max-width:760px;margin:0 auto;padding:2rem 1.25rem 3rem}
.meta{font:12px/1.4 system-ui,sans-serif;color:var(--muted);text-transform:uppercase;letter-spacing:.06em;margin-bottom:.75rem}
h1.page-title{font-size:clamp(1.6rem,3vw,2.2rem);margin:0 0 1rem;line-height:1.2;color:#f8fafc}
${proseCSS}
${caseTaskDetailUICSS()}
.foot{font:11px/1.4 system-ui,sans-serif;color:var(--muted);margin-top:2rem}`;
  return `<!doctype html><html lang="en"><head><meta charset="utf-8"/><meta name="color-scheme" content="dark"/><meta name="viewport" content="width=device-width, initial-scale=1"/><title>${escapeHtml(heading)}</title><style>${css}</style></head><body><div class="wrap"><div class="meta">Case/task</div><h1 class="page-title">${escapeHtml(heading)}</h1><div class="prose">${rendered}</div>${detailFrag}<p class="foot">Generated Case/task</p></div></body></html>`;
}
