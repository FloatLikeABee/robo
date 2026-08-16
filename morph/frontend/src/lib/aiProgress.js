/** Lightweight copy of platform-chat/aiProgress for Morph AI (CRA). */

function truncate(text, max = 56) {
  const t = String(text ?? '').trim().replace(/\s+/g, ' ');
  if (t.length <= max) return t;
  return `${t.slice(0, max - 1)}…`;
}

function mentionsWebSearch(text) {
  return /\b(search the web|web search|look up online|google|research online|latest news|browse the web)\b/i.test(text);
}

function extractSearchTopic(text) {
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

export function inferMorphProgressSteps(ctx = {}) {
  const text = String(ctx.userText ?? '').trim();
  const low = text.toLowerCase();
  const steps = ['Reading your question…'];

  if (ctx.analyzeFile || ctx.hasFile) steps.push('Parsing uploaded file…');
  if (ctx.hasHybridContext) steps.push('Checking hybrid context…');

  const web = ctx.webSearch ?? mentionsWebSearch(text);
  if (web) {
    const topic = extractSearchTopic(text);
    steps.push(topic ? `Searching the web for “${topic}”…` : 'Searching the web…');
  }

  if (/\b(list|show|what are my)\b.*\b(forms?|events?|members?|employees?|tasks?|notes?)\b/i.test(text)) {
    steps.push('Fetching workspace data…');
  } else if (/\b(create|add|new|make)\b/i.test(text)) {
    steps.push('Setting up the create flow…');
  } else if (/\b(update|edit|change|delete|remove)\b/i.test(text)) {
    steps.push('Locating the target record…');
  } else if (/\b(analyz|summar|report|chart|sql|query|export)\b/i.test(low)) {
    steps.push('Analyzing data…');
  } else {
    steps.push('Running Morph tools…');
  }

  steps.push('Organizing the answer…');

  const out = [];
  for (const step of steps) {
    if (out[out.length - 1] !== step) out.push(step);
  }
  return out;
}

export function startProgressTicker(steps, onUpdate, signal) {
  const list = (steps || []).filter(Boolean);
  if (!list.length) {
    onUpdate('Working…');
    return () => {};
  }
  let idx = 0;
  onUpdate(list[0]);
  const timers = [];
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
  return () => timers.forEach(clearTimeout);
}

export function inferTextAssistProgressSteps(ctx = {}) {
  const mode = String(ctx.mode ?? '').toLowerCase();
  const kind = String(ctx.kind ?? 'text').toLowerCase();
  const steps = ['Reading your draft…'];

  if (mode === 'improve') {
    steps.push('Polishing wording and structure…');
  } else if (mode.includes('todo')) {
    steps.push('Drafting task details…');
  } else if (mode.includes('note')) {
    steps.push('Drafting note content…');
  } else if (mode === 'task_chain_step') {
    steps.push('Running task chain step…');
  } else {
    steps.push('Generating text…');
  }

  steps.push('Applying suggestions…');

  const out = [];
  for (const step of steps) {
    if (out[out.length - 1] !== step) out.push(step);
  }
  return out;
}

export function runAiProgress(ctx, onUpdate, signal) {
  return startProgressTicker(inferMorphProgressSteps(ctx), onUpdate, signal);
}

export function runTextAssistProgress(ctx, onUpdate, signal) {
  return startProgressTicker(inferTextAssistProgressSteps(ctx), onUpdate, signal);
}
