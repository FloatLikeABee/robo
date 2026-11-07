/** ponytail: cap 50 table rows in published HTML; upgrade: link to full CSV or paginate. */
export const PUBLISH_EXCERPT_MAX_ROWS = 50;

export function escapeHtml(value) {
  return String(value ?? '')
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;');
}

function parseDetail(raw) {
  if (raw == null || raw === '') return null;
  if (typeof raw === 'object') return raw;
  try {
    return JSON.parse(String(raw));
  } catch {
    return null;
  }
}

function analysisHtml(text) {
  const blocks = String(text || '')
    .trim()
    .split(/\n{2,}/)
    .map((p) => p.trim())
    .filter(Boolean);
  if (blocks.length === 0) return '<p><em>No analysis.</em></p>';
  return blocks.map((p) => `<p>${escapeHtml(p).replace(/\n/g, '<br>')}</p>`).join('\n');
}

function tableExcerptHtml(detail, maxRows = PUBLISH_EXCERPT_MAX_ROWS) {
  const columns = Array.isArray(detail?.columns) ? detail.columns : [];
  const rows = Array.isArray(detail?.rows) ? detail.rows : [];
  if (columns.length === 0 || rows.length === 0) return '';
  const shown = rows.slice(0, maxRows);
  const head = columns.map((c) => `<th>${escapeHtml(c)}</th>`).join('');
  const body = shown
    .map((row) => `<tr>${columns.map((c) => `<td>${escapeHtml(row?.[c] ?? '')}</td>`).join('')}</tr>`)
    .join('');
  const more =
    rows.length > maxRows
      ? `<p><em>${rows.length - maxRows} more rows omitted.</em></p>`
      : '';
  return `<h2>Data excerpt</h2><table><thead><tr>${head}</tr></thead><tbody>${body}</tbody></table>${more}`;
}

export function buildPublishHtml(record) {
  const title = String(record?.title || 'Untitled data').trim() || 'Untitled data';
  const analysis = String(record?.ai_analysis || '').trim();
  const detail = parseDetail(record?.detail);
  return `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>${escapeHtml(title)}</title>
<style>
body{font-family:Arial,"Helvetica Neue",Helvetica,sans-serif;max-width:52rem;margin:2rem auto;padding:0 1.25rem;line-height:1.55;background:#221c34;color:#f4f0fa}
h1{font-size:1.6rem;color:#a78bfa}h2{font-size:1.15rem;margin-top:1.75rem;color:#8b5cf6}
a{color:#f97316}
table{border-collapse:collapse;width:100%;font-size:.9rem}
th,td{border:1px solid #3d3458;padding:.4rem .55rem;text-align:left}
th{background:#322a4a}
</style>
</head>
<body>
<h1>${escapeHtml(title)}</h1>
<h2>Analysis</h2>
${analysisHtml(analysis)}
${tableExcerptHtml(detail)}
</body>
</html>`;
}

export function buildPublishPayload(record) {
  const name = String(record?.title || 'Untitled data').trim() || 'Untitled data';
  return {
    name,
    theme: 'dark',
    html_content: buildPublishHtml(record),
  };
}
