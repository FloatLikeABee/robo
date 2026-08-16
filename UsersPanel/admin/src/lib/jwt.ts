export function parseJwtPayload(token: string): Record<string, unknown> | null {
  try {
    const parts = token.split('.')
    if (parts.length !== 3) return null
    const pad = parts[1].length % 4 === 0 ? '' : '='.repeat(4 - (parts[1].length % 4))
    const json = atob(parts[1].replace(/-/g, '+').replace(/_/g, '/') + pad)
    return JSON.parse(json) as Record<string, unknown>
  } catch {
    return null
  }
}

export function rolesFromToken(token: string | null): string[] {
  if (!token) return []
  const p = parseJwtPayload(token)
  const r = p?.roles
  return Array.isArray(r) ? (r as string[]) : []
}
