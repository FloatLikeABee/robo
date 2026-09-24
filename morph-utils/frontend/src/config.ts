import { readRuntimeConfig, resolvePublicUrl } from './publicUrl';

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

const runtimeConfig = readRuntimeConfig();

const sheetxUrl = resolvePublicUrl({
  runtime: runtimeConfig.sheetxUrl,
  runtimeAlias: runtimeConfig.formsxUrl,
  built: import.meta.env.VITE_SHEETX_URL,
  builtAlias: import.meta.env.VITE_FORMSX_URL,
  devFallback: import.meta.env.DEV ? 'http://localhost:19909' : '',
  dev: import.meta.env.DEV,
});
const composerxUrl = resolvePublicUrl({
  runtime: runtimeConfig.composerxUrl,
  built: import.meta.env.VITE_COMPOSERX_URL,
  devFallback: import.meta.env.DEV ? 'http://localhost:8044' : '',
  dev: import.meta.env.DEV,
});
export const DATAX_URL = resolvePublicUrl({
  runtime: runtimeConfig.dataxUrl,
  built: import.meta.env.VITE_DATAX_URL,
  devFallback: import.meta.env.DEV ? 'http://localhost:5178' : '',
  dev: import.meta.env.DEV,
});
const projectsUrl = resolvePublicUrl({
  runtime: runtimeConfig.projectsUrl,
  runtimeAlias: runtimeConfig.morphEngiUrl,
  built: import.meta.env.VITE_PROJECTS_URL,
  builtAlias: import.meta.env.VITE_MORPH_ENGI_URL,
  devFallback: import.meta.env.DEV ? 'http://localhost:5179' : '',
  dev: import.meta.env.DEV,
});

export const MORPH_AI_URL = resolvePublicUrl({
  runtime: runtimeConfig.morphAiUrl,
  built: import.meta.env.VITE_MORPH_AI_URL,
  devFallback: import.meta.env.DEV ? 'http://localhost:3031' : '',
  dev: import.meta.env.DEV,
});

export const UTILS_MODULES: UtilsModule[] = [
  {
    id: 'sheetx',
    label: 'Event Logs',
    shortLabel: 'Logs',
    description: 'Events & Info log.',
    accent: '#0ea5e9',
    icon: '/icons/sheetx-icon.svg',
    embedUrl: `${sheetxUrl}/events-info`,
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
    accent: '#3b82f6',
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
