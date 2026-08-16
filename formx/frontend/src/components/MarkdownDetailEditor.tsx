import { useState } from 'react';
import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';

type MarkdownDetailEditorProps = {
  value: string;
  onChange: (value: string) => void;
  placeholder?: string;
  disabled?: boolean;
  rows?: number;
  isPublic?: boolean;
};

export function MarkdownDetailEditor({
  value,
  onChange,
  placeholder = 'Write in Markdown…',
  disabled = false,
  rows = 5,
  isPublic = false,
}: MarkdownDetailEditorProps) {
  const [tab, setTab] = useState<'write' | 'preview'>('write');

  const tabClass = (active: boolean) =>
    active
      ? isPublic
        ? 'bg-violet-600 text-white'
        : 'bg-violet-600 text-white dark:bg-violet-500'
      : isPublic
        ? 'text-slate-400 hover:text-slate-200'
        : 'text-slate-500 hover:text-slate-800 dark:text-slate-400 dark:hover:text-slate-100';

  const inputClass = isPublic
    ? 'w-full px-3 py-2.5 rounded-lg bg-slate-800 border border-slate-600 text-slate-100 text-sm font-mono'
    : 'w-full px-2.5 py-2 rounded-md bg-white dark:bg-slate-800 border border-slate-300 dark:border-slate-600 text-slate-900 dark:text-white text-sm font-mono';

  const previewClass = isPublic
    ? 'rounded-lg border border-slate-600 bg-slate-800/70 px-3 py-2.5 text-sm text-slate-100 prose prose-invert max-w-none prose-p:my-2 prose-headings:my-2'
    : 'rounded-md border border-slate-300 dark:border-slate-600 bg-white dark:bg-slate-800 px-2.5 py-2 text-sm text-slate-900 dark:text-slate-100 prose dark:prose-invert max-w-none prose-p:my-2 prose-headings:my-2';

  return (
    <div className="space-y-2">
      <div className="flex items-center gap-1">
        <button type="button" className={`rounded-md px-2 py-0.5 text-xs font-medium ${tabClass(tab === 'write')}`} onClick={() => setTab('write')}>
          Write
        </button>
        <button type="button" className={`rounded-md px-2 py-0.5 text-xs font-medium ${tabClass(tab === 'preview')}`} onClick={() => setTab('preview')}>
          Preview
        </button>
        <span className={`ml-auto text-[10px] ${isPublic ? 'text-slate-500' : 'text-slate-400 dark:text-slate-500'}`}>Markdown supported</span>
      </div>
      {tab === 'write' ? (
        <textarea
          value={value}
          onChange={(e) => onChange(e.target.value)}
          rows={rows}
          disabled={disabled}
          className={`${inputClass} resize-y ${isPublic ? 'min-h-[9rem]' : 'min-h-[6rem]'}`}
          placeholder={placeholder}
        />
      ) : (
        <div className={`${previewClass} min-h-[6rem] overflow-auto`}>
          {value.trim() ? (
            <ReactMarkdown remarkPlugins={[remarkGfm]}>{value}</ReactMarkdown>
          ) : (
            <span className={isPublic ? 'text-slate-500' : 'text-slate-400 dark:text-slate-500'}>Nothing to preview yet.</span>
          )}
        </div>
      )}
    </div>
  );
}

export function MarkdownDetailView({ value, className = '' }: { value: string; className?: string }) {
  if (!value.trim()) {
    return <span className="text-slate-400">No detail.</span>;
  }
  return (
    <div className={`prose dark:prose-invert max-w-none prose-p:my-2 prose-headings:my-2 prose-a:text-violet-600 dark:prose-a:text-violet-300 ${className}`}>
      <ReactMarkdown remarkPlugins={[remarkGfm]}>{value}</ReactMarkdown>
    </div>
  );
}
