import { marked } from 'marked'

marked.setOptions({ gfm: true, breaks: true })

/** @param {string} markdown */
export function renderMarkdownHtml(markdown) {
  return marked.parse(String(markdown || ''), { async: false })
}

/** @param {string} name @param {string} [fallback] */
export function safeMarkdownFilename(name, fallback = 'content') {
  const base = String(name || fallback)
    .trim()
    .replace(/[^\w\s.-]/g, '')
    .replace(/\s+/g, '-')
    .slice(0, 80)
  const stem = base || fallback
  return stem.toLowerCase().endsWith('.md') ? stem : `${stem}.md`
}

/** @param {string} filename @param {string} markdown */
export function downloadMarkdownFile(filename, markdown) {
  const blob = new Blob([markdown || ''], { type: 'text/markdown;charset=utf-8' })
  const url = URL.createObjectURL(blob)
  const anchor = document.createElement('a')
  anchor.href = url
  anchor.download = safeMarkdownFilename(filename)
  anchor.click()
  URL.revokeObjectURL(url)
}

/** @param {{ markdown_content?: string, html_content?: string } | null | undefined} record */
export function savedContentMarkdown(record) {
  return String(record?.markdown_content || record?.html_content || '')
}
