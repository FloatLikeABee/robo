/** ponytail: cap 50 table rows in the generate prompt; upgrade: link the full record. */
export const DATA_TABLE_MAX_ROWS = 50

const TABLE_TYPES = new Set(['csv', 'json', 'xlsx', 'excel'])

export function dorphParentOrigin(referrer) {
  try {
    const origin = new URL(String(referrer || '')).origin
    if (!origin || origin === 'null') return ''
    return origin
  } catch {
    return ''
  }
}

function parseDetail(raw) {
  if (raw == null || raw === '') return null
  if (typeof raw === 'object') return raw
  try {
    return JSON.parse(String(raw))
  } catch {
    return null
  }
}

function articleText(title, analysis, detail) {
  const md = typeof detail?.content_markdown === 'string' ? detail.content_markdown.trim() : ''
  return [`Title: ${title}`, analysis && `Analysis:\n${analysis}`, md && `Body:\n${md}`].filter(Boolean).join('\n\n')
}

function tableText(title, analysis, detail) {
  const columns = Array.isArray(detail?.columns) ? detail.columns : []
  const rows = Array.isArray(detail?.rows) ? detail.rows : []
  const shown = rows.slice(0, DATA_TABLE_MAX_ROWS)
  const lines = [`Title: ${title}`, `Summary: ${rows.length} rows, ${columns.length} columns.`]
  if (analysis) lines.push(`Analysis:\n${analysis}`)
  if (columns.length) {
    lines.push(columns.join(' | '))
    for (const row of shown) lines.push(columns.map((col) => String(row?.[col] ?? '')).join(' | '))
    if (rows.length > shown.length) lines.push(`${rows.length - shown.length} more rows omitted.`)
  }
  return lines.join('\n')
}

export function buildDataSource(record) {
  const title = String(record?.title || 'Untitled data').trim() || 'Untitled data'
  const analysis = String(record?.ai_analysis || '').trim()
  const type = String(record?.source_type || '').toLowerCase()
  const detail = parseDetail(record?.detail)
  if (TABLE_TYPES.has(type)) {
    return { layout: 'table', title, text: tableText(title, analysis, detail) }
  }
  return { layout: 'article', title, text: articleText(title, analysis, detail) }
}
