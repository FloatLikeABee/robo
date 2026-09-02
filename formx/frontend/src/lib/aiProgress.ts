export type AiProgressApp =
  | 'morph'
  | 'formsx'
  | 'sheetx'
  | 'composerx'
  | 'composerx-publish'
  | 'datax'
  | 'booki'
  | 'generic';

export type AiProgressContext = {
  userText?: string;
  app?: AiProgressApp;
  webSearch?: boolean;
  hasFile?: boolean;
  hasAttachments?: boolean;
  hasDataTables?: boolean;
  hasReferenceDocs?: boolean;
  analyzeFile?: boolean;
};

function truncate(text: string, max = 56): string {
  const t = text.trim().replace(/\s+/g, ' ');
  if (t.length <= max) return t;
  return `${t.slice(0, max - 1)}…`;
}

function mentionsWebSearch(text: string): boolean {
  return /\b(search the web|web search|look up online|google|research online|latest news|browse the web)\b/i.test(
    text,
  );
}

function extractSearchTopic(text: string): string | null {
  const patterns = [
    /(?:search(?:\s+the\s+web)?\s+for|look up|research)\s+["']?([^"'\n.!?]{3,60})/i,
    /web search:\s*["']?([^"'\n.!?]{3,60})/i,
  ];
  for (const re of patterns) {
    const m = text.match(re);
    if (m?.[1]) return truncate(m[1]);
  }
  return null;
}

/** Short, contextual status lines shown while the model is working. */
export function inferProgressSteps(ctx: AiProgressContext): string[] {
  const text = (ctx.userText ?? '').trim();
  const low = text.toLowerCase();
  const steps: string[] = [];

  steps.push('Reading your question…');

  if (ctx.analyzeFile || ctx.hasFile) {
    steps.push('Parsing uploaded file…');
  }

  if (ctx.hasDataTables) {
    steps.push('Querying attached tables…');
  }

  if (ctx.hasAttachments && !ctx.hasDataTables) {
    steps.push('Reviewing attachments…');
  }

  if (ctx.hasReferenceDocs) {
    steps.push('Pulling reference excerpts…');
  }

  const web = ctx.webSearch ?? mentionsWebSearch(text);
  if (web) {
    const topic = extractSearchTopic(text);
    steps.push(topic ? `Searching the web for “${topic}”…` : 'Searching the web…');
  }

  if (/\b(list|show|what are my)\b.*\b(forms?|events?|templates?|contacts?|tasks?|notes?|records?)\b/i.test(text)) {
    steps.push('Fetching workspace data…');
  } else if (/\b(create|add|new|make)\b/i.test(text)) {
    steps.push('Setting up the create flow…');
  } else if (/\b(update|edit|change|delete|remove)\b/i.test(text)) {
    steps.push('Locating the target record…');
  } else if (/\b(analyz|summar|report|chart|sql|query|export)\b/i.test(low)) {
    steps.push('Analyzing data…');
  } else if (ctx.app === 'composerx' || ctx.app === 'composerx-publish') {
    steps.push(web ? 'Drafting with research…' : 'Drafting content…');
  } else if (
    (ctx.app === 'formsx' || ctx.app === 'sheetx') &&
    /\bform|survey|question|template|landing|sheet/i.test(text)
  ) {
    steps.push('Designing form layout…');
  } else if (ctx.app === 'morph') {
    steps.push('Running Morph tools…');
  } else if (ctx.app === 'booki') {
    steps.push('Reviewing ledger context…');
  } else if (ctx.app === 'datax') {
    steps.push('Building the data view…');
  }

  steps.push('Organizing the answer…');

  const out: string[] = [];
  for (const step of steps) {
    if (out[out.length - 1] !== step) out.push(step);
  }
  return out;
}

/** Cycle through steps with short delays; stays on the last step until stopped. */
export function startProgressTicker(
  steps: string[],
  onUpdate: (status: string) => void,
  signal?: AbortSignal,
): () => void {
  const list = steps.filter(Boolean);
  if (!list.length) {
    onUpdate('Working…');
    return () => {};
  }

  let idx = 0;
  onUpdate(list[0]);
  const timers: ReturnType<typeof setTimeout>[] = [];

  const schedule = () => {
    if (signal?.aborted || idx >= list.length - 1) return;
    const delay = idx === 0 ? 700 : idx >= list.length - 2 ? 2200 : 1300;
    const timer = setTimeout(() => {
      if (signal?.aborted) return;
      idx += 1;
      onUpdate(list[idx]);
      schedule();
    }, delay);
    timers.push(timer);
  };
  schedule();

  return () => {
    for (const timer of timers) clearTimeout(timer);
  };
}

export function runAiProgress(
  ctx: AiProgressContext,
  onUpdate: (status: string) => void,
  signal?: AbortSignal,
): () => void {
  return startProgressTicker(inferProgressSteps(ctx), onUpdate, signal);
}
