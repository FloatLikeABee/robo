export type PageId = 'projects' | 'files' | 'settings'

export const NAV: { id: PageId; label: string; hint: string }[] = [
  { id: 'projects', label: 'Projects', hint: 'AI docs from files or paste' },
  { id: 'files', label: 'Files', hint: 'Uploaded and pasted content' },
]
