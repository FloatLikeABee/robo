import { marked } from 'marked'

marked.setOptions({ gfm: true, breaks: true })

export function renderMarkdownHtml(markdown: string): string {
  return marked.parse(String(markdown || ''), { async: false }) as string
}
