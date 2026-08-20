import type { PageId } from './nav'

/** Snapshot of live app data sent to Projects AI on every message. */
export type EngiAiSnapshot = {
  page: PageId
  organization: unknown
  projects: unknown[]
  resourceFiles: unknown[]
}

export function buildAiStateExtra(snapshot: EngiAiSnapshot) {
  return {
    fields: {
      current_page: snapshot.page,
      organization: JSON.stringify(snapshot.organization ?? {}),
      app_context: JSON.stringify({
        page: snapshot.page,
        organization: snapshot.organization,
        projects: snapshot.projects,
        resource_files: snapshot.resourceFiles,
      }),
    },
  }
}
