const KEY = 'morph-engi-browser-store'
const TOKEN = 'local-preview'

type Project = {
  id: number
  code: string
  name: string
  status: string
  description: string
  markdown_content: string
  html_content: string
  source_summary: string
  published_slug?: string | null
  published_path?: string | null
  created_at: string
  updated_at: string
}

type ResourceFile = {
  id: number
  name: string
  source_type: string
  file_url: string
  file_name: string
  description: string
  created_at: string
}

type Store = {
  nextId: number
  projects: Project[]
  files: ResourceFile[]
}

export function isBrowserStore(): boolean {
  if (import.meta.env.VITE_STORAGE === 'local') return true
  if (typeof window === 'undefined') return false
  const host = window.location.hostname
  return host.endsWith('.vercel.app') || host.endsWith('.vercel.sh')
}

function now() {
  return new Date().toISOString().replace('T', ' ').slice(0, 19)
}

function load(): Store {
  try {
    const raw = localStorage.getItem(KEY)
    if (raw) return JSON.parse(raw) as Store
  } catch {
    /* ignore */
  }
  return { nextId: 1, projects: [], files: [] }
}

function save(store: Store) {
  localStorage.setItem(KEY, JSON.stringify(store))
}

function alloc(store: Store) {
  const id = store.nextId
  store.nextId += 1
  return id
}

function slugify(title: string) {
  const s = title
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, '-')
    .replace(/^-|-$/g, '')
    .slice(0, 60)
  return s || 'project'
}

function codeFrom(name: string) {
  const base = name.trim().toUpperCase().replace(/[^A-Z0-9]+/g, '-').replace(/^-|-$/g, '')
  return base.slice(0, 16) || `PRJ-${Date.now().toString().slice(-6)}`
}

function markdownToHtml(title: string, md: string) {
  const body = md
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/^### (.+)$/gm, '<h3>$1</h3>')
    .replace(/^## (.+)$/gm, '<h2>$1</h2>')
    .replace(/^# (.+)$/gm, '<h1>$1</h1>')
    .replace(/\n\n/g, '</p><p>')
  return `<!doctype html><html><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"><title>${title.replace(/</g, '')}</title>
<style>body{font-family:Inter,system-ui,sans-serif;max-width:44rem;margin:2rem auto;padding:0 1rem;line-height:1.5;color:#111}pre{white-space:pre-wrap}</style></head><body><p>${body}</p></body></html>`
}

function jsonError(msg: string): never {
  throw new Error(msg)
}

async function readFileText(file: File): Promise<string> {
  const name = file.name.toLowerCase()
  if (name.endsWith('.pdf')) {
    return `(PDF “${file.name}” — text extraction is not available in this Vercel preview. Paste the contents or use a .txt/.md/.csv file.)`
  }
  try {
    return await file.text()
  } catch {
    return `(Could not read ${file.name})`
  }
}

export function previewToken(): string {
  return TOKEN
}

export async function handleApi<T = unknown>(path: string, init: RequestInit = {}): Promise<T> {
  const method = (init.method || 'GET').toUpperCase()
  const urlPath = path.split('?')[0]
  const store = load()

  if (urlPath === '/api/v1/auth/me' && method === 'GET') {
    return { user_id: 1, organization_id: 1, role: 'owner' } as T
  }
  if (
    (urlPath === '/api/v1/auth/preview-login' || urlPath === '/api/v1/auth/login' || urlPath === '/api/v1/auth/dev-login') &&
    method === 'POST'
  ) {
    return { access_token: TOKEN, token_type: 'Bearer', user_id: 1, organization_id: 1, role: 'owner' } as T
  }
  if (urlPath === '/api/v1/organization' && method === 'GET') {
    return { id: 1, name: 'Preview', country: '', currency: 'USD' } as T
  }
  if (urlPath === '/api/v1/projects' && method === 'GET') {
    return { projects: store.projects } as T
  }
  if (urlPath === '/api/v1/resource-files' && method === 'GET') {
    return { resource_files: store.files } as T
  }
  if (urlPath === '/api/v1/resource-files' && method === 'POST') {
    const body = JSON.parse(String(init.body || '{}')) as Partial<ResourceFile>
    if (!String(body.name || '').trim()) jsonError('Name is required')
    const rec: ResourceFile = {
      id: alloc(store),
      name: String(body.name).trim(),
      source_type: body.source_type || 'url',
      file_url: body.file_url || '',
      file_name: body.file_name || '',
      description: body.description || '',
      created_at: now(),
    }
    store.files.unshift(rec)
    save(store)
    return { resource_file: { id: rec.id } } as T
  }

  const pub = urlPath.match(/^\/api\/v1\/projects\/(\d+)\/publish$/)
  if (pub && method === 'POST') {
    const id = Number(pub[1])
    const p = store.projects.find((x) => x.id === id)
    if (!p) jsonError('project not found')
    const slug = `${slugify(p.name)}-${p.id}`
    p.published_slug = slug
    p.published_path = `blob:${slug}`
    p.updated_at = now()
    save(store)
    return { ...p, published_path: p.published_path } as T
  }

  const proj = urlPath.match(/^\/api\/v1\/projects\/(\d+)$/)
  if (proj && method === 'DELETE') {
    const id = Number(proj[1])
    store.projects = store.projects.filter((x) => x.id !== id)
    save(store)
    return { deleted: true } as T
  }

  const fileDel = urlPath.match(/^\/api\/v1\/resource-files\/(\d+)$/)
  if (fileDel && method === 'DELETE') {
    const id = Number(fileDel[1])
    store.files = store.files.filter((x) => x.id !== id)
    save(store)
    return { deleted: true } as T
  }

  if (urlPath === '/api/v1/assistant/chat' && method === 'POST') {
    const body = JSON.parse(String(init.body || '{}')) as { messages?: { role: string; content: string }[] }
    const last = [...(body.messages || [])].reverse().find((m) => m.role === 'user')?.content || ''
    const n = store.projects.length
    const f = store.files.length
    return {
      assistant_message: last
        ? `This Vercel preview keeps data in your browser.\n\n**${n}** project(s), **${f}** file(s).\n\nCreate a document from paste or a text file on the Projects tab.`
        : 'How can I help with your project documents?',
      completed: true,
    } as T
  }

  jsonError(`Unsupported in this preview: ${method} ${urlPath}`)
}

export async function handleUploadFile(file: File): Promise<{ file_url: string; file_name: string; source_type: string }> {
  const url = URL.createObjectURL(file)
  return { file_url: url, file_name: file.name, source_type: 'upload' }
}

export async function handleGenerateDocument(
  files: File[],
  fields: Record<string, string>
): Promise<{ project: Project }> {
  const paste = (fields.paste || '').trim()
  const titleHint = (fields.title || '').trim()
  if (!files.length && !paste) jsonError('Provide at least one source: upload a file or paste content.')

  const store = load()
  const parts: { name: string; text: string }[] = []
  for (const file of files.slice(0, 5)) {
    const text = await readFileText(file)
    parts.push({ name: file.name, text })
    const rec: ResourceFile = {
      id: alloc(store),
      name: file.name,
      source_type: 'upload',
      file_url: URL.createObjectURL(file),
      file_name: file.name,
      description: 'Uploaded as a project source',
      created_at: now(),
    }
    store.files.unshift(rec)
  }
  if (paste) {
    parts.push({ name: 'Pasted content', text: paste })
    const rec: ResourceFile = {
      id: alloc(store),
      name: `paste-${now().replace(/[: ]/g, '-')}.txt`,
      source_type: 'upload',
      file_url: `data:text/plain;charset=utf-8,${encodeURIComponent(paste)}`,
      file_name: 'paste.txt',
      description: 'Saved from paste',
      created_at: now(),
    }
    store.files.unshift(rec)
  }

  const title = titleHint || parts[0]?.name.replace(/\.[^.]+$/, '') || 'Untitled project'
  let markdown = `# ${title}\n\n## Overview\n\nOrganized from source material in this Vercel preview.\n\n`
  for (const part of parts) {
    markdown += `## ${part.name}\n\n${part.text}\n\n`
  }
  const html = markdownToHtml(title, markdown)
  const stamp = now()
  const project: Project = {
    id: alloc(store),
    code: codeFrom(title),
    name: title,
    status: 'planning',
    description: parts.map((p) => p.name).join(', '),
    markdown_content: markdown,
    html_content: html,
    source_summary: parts.map((p) => p.name).join(', '),
    published_slug: null,
    published_path: null,
    created_at: stamp,
    updated_at: stamp,
  }
  store.projects.unshift(project)
  save(store)
  return { project }
}
