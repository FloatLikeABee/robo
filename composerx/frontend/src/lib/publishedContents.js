export function slugifyPublishName(name) {
  const s = String(name || '')
    .toLowerCase()
    .trim()
    .replace(/[^a-z0-9]+/g, '-')
    .replace(/^-+|-+$/g, '')
  return s || 'page'
}

function laterStamp(a, b) {
  const left = String(a || '')
  const right = String(b || '')
  return left >= right ? left : right
}

/**
 * Merge saved HTML drafts and published pages into one row per slugified name.
 * Published wins when both exist.
 * @param {Array<{ id: number, name: string, theme?: string, updated_at?: string }>} drafts
 * @param {Array<{ id: number, name: string, slug: string, theme?: string, updated_at?: string }>} published
 */
export function mergePublishedContents(drafts, published) {
  /** @type {Map<string, any>} */
  const bySlug = new Map()

  for (const draft of drafts || []) {
    const slug = slugifyPublishName(draft.name)
    const existing = bySlug.get(slug)
    const stamp = String(draft.updated_at || '')
    if (existing && String(existing.updated_at || '') >= stamp) continue
    bySlug.set(slug, {
      key: slug,
      name: draft.name,
      status: 'Saved',
      path: '',
      slug: '',
      theme: draft.theme || 'default',
      updated_at: draft.updated_at || '',
      draftId: draft.id,
      publishedId: null,
      canOpen: false,
      canView: true,
      canDelete: true,
    })
  }

  for (const page of published || []) {
    const slug = page.slug || slugifyPublishName(page.name)
    const existing = bySlug.get(slug)
    bySlug.set(slug, {
      key: slug,
      name: page.name,
      status: 'Published',
      path: `/public/p/${page.slug}`,
      slug: page.slug,
      theme: page.theme || existing?.theme || 'default',
      updated_at: laterStamp(page.updated_at, existing?.updated_at),
      draftId: existing?.draftId ?? null,
      publishedId: page.id,
      canOpen: true,
      canView: existing?.draftId != null,
      canDelete: true,
    })
  }

  return [...bySlug.values()].sort((a, b) => String(b.updated_at).localeCompare(String(a.updated_at)))
}
