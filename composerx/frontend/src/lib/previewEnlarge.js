export function previewCanEnlarge(html) {
  const raw = String(html || '')
  if (!raw.trim()) return false
  return !raw.includes('data-composerx-starter')
}
