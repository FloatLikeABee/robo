/** @param {unknown} raw */
export function normalizeDraftToMarkdown(raw) {
  const s = String(raw ?? '').trim()
  if (!s) return ''
  if (looksLikeHTML(s)) return htmlToMarkdown(s)
  return s
}

/** @param {string} s */
function looksLikeHTML(s) {
  return s.includes('<') && /<[a-z][\s\S]*>/i.test(s)
}

/** @param {string} html */
export function htmlToMarkdown(html) {
  const s = String(html ?? '').trim()
  if (!s) return ''
  if (!looksLikeHTML(s)) return s

  if (typeof DOMParser !== 'undefined') {
    try {
      const doc = new DOMParser().parseFromString(s, 'text/html')
      doc.querySelectorAll('style, script').forEach((el) => el.remove())
      const text = walkNode(doc.body).trim()
      if (text) return collapseBlankLines(text)
    } catch {
      // fall through to regex strip
    }
  }

  return collapseBlankLines(
    s
      .replace(/<style[\s\S]*?<\/style>/gi, '')
      .replace(/<script[\s\S]*?<\/script>/gi, '')
      .replace(/<br\s*\/?>/gi, '\n')
      .replace(/<\/(p|div|tr|h[1-6]|li|table|section|article|td)>/gi, '\n\n')
      .replace(/<[^>]+>/g, '')
      .replace(/&nbsp;/gi, ' ')
      .replace(/&amp;/gi, '&')
      .replace(/&lt;/gi, '<')
      .replace(/&gt;/gi, '>')
      .trim()
  )
}

/** @param {Node | null} node */
function walkNode(node) {
  if (!node) return ''
  if (node.nodeType === Node.TEXT_NODE) return node.textContent ?? ''
  if (node.nodeType !== Node.ELEMENT_NODE) return ''

  const el = /** @type {HTMLElement} */ (node)
  const tag = el.tagName.toLowerCase()
  const childText = Array.from(el.childNodes)
    .map((c) => walkNode(c))
    .join('')
    .trim()

  switch (tag) {
    case 'h1':
    case 'h2':
    case 'h3':
    case 'h4':
    case 'h5':
    case 'h6': {
      const level = Number(tag.slice(1)) || 1
      return `\n${'#'.repeat(level)} ${childText}\n`
    }
    case 'li':
      return `\n- ${childText}\n`
    case 'br':
      return '\n'
    case 'p':
    case 'div':
    case 'tr':
    case 'td':
    case 'section':
    case 'article':
      return childText ? `\n${childText}\n` : ''
    default:
      return childText
  }
}

/** @param {string} text */
function collapseBlankLines(text) {
  return text
    .split('\n')
    .map((line) => line.trimEnd())
    .join('\n')
    .replace(/\n{3,}/g, '\n\n')
    .trim()
}

/** @param {string} text */
export function extractMarkdownFence(text) {
  const s = String(text ?? '').trim()
  if (!s) return ''
  const m = s.match(/```(?:markdown|md)?\s*([\s\S]*?)```/i)
  return m ? m[1].trim() : ''
}

/** @param {Record<string, unknown> | null | undefined} res */
export function resolveComposerProposedDraft(res) {
  if (!res || typeof res !== 'object') return ''

  const candidates = [
    res.proposed_markdown,
    res.proposedMarkdown,
    res.proposed_email_html,
    res.proposedEmailHtml,
    res.document,
    res.markdown,
    res.content,
    res.draft,
    res.body,
  ]

  for (const c of candidates) {
    const md = normalizeDraftToMarkdown(c)
    if (md) return md
  }

  return extractMarkdownFence(res.assistant_message ?? res.assistantMessage ?? '')
}
