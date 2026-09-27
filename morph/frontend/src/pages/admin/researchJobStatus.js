export const researchEmptyThesisMessage = 'No thesis was produced.';

function jobStatus(job) {
  return String(job?.status || '');
}

function jobCause(job) {
  return String(job?.error_text || '').trim();
}

export function researchAlert(job) {
  const message = jobCause(job);
  if (!message) return null;
  return {
    severity: jobStatus(job) === 'failed' ? 'error' : 'warning',
    message,
  };
}

export function researchCanPublish(job) {
  return jobStatus(job) !== 'failed';
}

export function researchShowsThesisEditor(job) {
  if (jobStatus(job) === 'failed' && !String(job?.markdown_content || '').trim()) {
    return false;
  }
  return true;
}

export function researchListSecondary(job) {
  const status = jobStatus(job);
  const round = job?.current_round ? ` · ${job.current_round}/${job.round_count || 5}` : '';
  const published = job?.published_path ? ' · published' : '';
  const line = `${status}${round}${published}`;
  const message = jobCause(job);
  if (!message || (status !== 'failed' && status !== 'complete')) {
    return line;
  }
  const short = message.length > 80 ? `${message.slice(0, 77)}...` : message;
  return `${line} · ${short}`;
}
