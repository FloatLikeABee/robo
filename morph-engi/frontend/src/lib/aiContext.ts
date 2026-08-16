import type { PageId } from './nav'

/** Snapshot of live app data sent to Morph Engi AI on every message. */
export type EngiAiSnapshot = {
  page: PageId
  activeProjectId: number | null
  organization: unknown
  projects: unknown[]
  tasks: unknown[]
  siteLogs: unknown[]
  resourceFiles: unknown[]
  contractors: unknown[]
}

export function buildAiStateExtra(snapshot: EngiAiSnapshot) {
  const pid = snapshot.activeProjectId
  const project = pid ? snapshot.projects.find((p: { id?: number }) => p?.id === pid) : null

  return {
    fields: {
      active_project_id: pid != null ? String(pid) : '',
      current_page: snapshot.page,
      active_project: project ? JSON.stringify(project) : '',
      organization: JSON.stringify(snapshot.organization ?? {}),
      app_context: JSON.stringify({
        page: snapshot.page,
        active_project_id: pid,
        active_project: project,
        organization: snapshot.organization,
        projects: snapshot.projects,
        tasks: snapshot.tasks,
        site_logs: snapshot.siteLogs,
        resource_files: snapshot.resourceFiles,
        contractors: snapshot.contractors,
      }),
    },
  }
}
