import { uploadsUrl, type Question } from '../lib/api';

type Tone = 'public' | 'sheet';

/**
 * Prompt image/video shown with a question (edit form preview, public form, results).
 */
export function QuestionPromptAttachment({
  question,
  tone = 'sheet',
  compact = false,
}: {
  question: Question;
  tone?: Tone;
  compact?: boolean;
}) {
  const m = question.config?.question_prompt_media;
  if (!m?.relative_path) return null;
  const src = uploadsUrl(m.relative_path);
  const maxH = compact ? 'max-h-20' : 'max-h-64';
  const margin = compact ? 'my-0.5' : 'mb-3';
  const borderPublic = 'rounded-lg overflow-hidden border border-slate-600';
  const borderSheet =
    'rounded-lg overflow-hidden border border-slate-300 dark:border-slate-600';

  if (m.kind === 'video') {
    return (
      <div className={`${margin} ${tone === 'public' ? borderPublic : borderSheet} bg-black`}>
        <video src={src} controls className={`w-full ${maxH}`} playsInline />
      </div>
    );
  }

  const imgBg = tone === 'public' ? 'bg-slate-900' : 'bg-slate-100 dark:bg-slate-900';
  return (
    <div className={`${margin} ${tone === 'public' ? borderPublic : borderSheet} ${imgBg}`}>
      <img src={src} alt={`Media for: ${question.title}`} className={`w-full ${maxH} object-contain`} />
    </div>
  );
}
