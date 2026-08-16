export type UtilsModuleId = 'sheetx' | 'composerx' | 'datax' | 'projects';

export type UtilsModule = {
  id: UtilsModuleId;
  label: string;
  shortLabel: string;
  description: string;
  accent: string;
  icon: string;
  embedUrl?: string;
};

const sheetxUrl = (
  import.meta.env.VITE_SHEETX_URL ?? import.meta.env.VITE_FORMSX_URL ?? 'http://localhost:19909'
).replace(/\/$/, '');
const composerxUrl = (import.meta.env.VITE_COMPOSERX_URL ?? 'http://localhost:8044').replace(/\/$/, '');
export const DATAX_URL = (import.meta.env.VITE_DATAX_URL ?? 'http://localhost:5178').replace(/\/$/, '');
const projectsUrl = (
  import.meta.env.VITE_PROJECTS_URL ?? import.meta.env.VITE_MORPH_ENGI_URL ?? 'http://localhost:5179'
).replace(/\/$/, '');

export const MORPH_AI_URL = (import.meta.env.VITE_MORPH_AI_URL ?? 'http://localhost:3031').replace(/\/$/, '');

export const UTILS_MODULES: UtilsModule[] = [
  {
    id: 'sheetx',
    label: 'Survey Maker',
    shortLabel: 'Surveys',
    description: 'AI surveys and published survey links.',
    accent: '#0ea5e9',
    icon: '/icons/sheetx-icon.svg',
    embedUrl: `${sheetxUrl}/survey-bot`,
  },
  {
    id: 'composerx',
    label: 'Content Maker',
    shortLabel: 'Content',
    description: 'Compose and publish markdown and HTML pages for outside readers.',
    accent: '#16a34a',
    icon: '/icons/composerx-icon.svg',
    embedUrl: composerxUrl,
  },
  {
    id: 'datax',
    label: 'Data Access',
    shortLabel: 'Data',
    description: 'Data tables and file-based data reports.',
    accent: '#2563eb',
    icon: '/icons/datax-icon.svg',
    embedUrl: DATAX_URL,
  },
  {
    id: 'projects',
    label: 'Project',
    shortLabel: 'Project',
    description: 'AI project documents from files or paste, plus a files library.',
    accent: '#2dd4bf',
    icon: '/icons/projects-icon.svg',
    embedUrl: projectsUrl,
  },
];

/** Map legacy Utils paths to current module ids. */
export function normalizeModuleId(id: string | undefined): UtilsModuleId | undefined {
  if (!id) return undefined;
  if (id === 'academi' || id === 'docs') return 'datax';
  if (id === 'booki') return 'projects';
  if (UTILS_MODULES.some((m) => m.id === id)) return id as UtilsModuleId;
  return undefined;
}

export function moduleById(id: string | undefined): UtilsModule {
  const normalized = normalizeModuleId(id);
  return UTILS_MODULES.find((m) => m.id === normalized) ?? UTILS_MODULES[0];
}
