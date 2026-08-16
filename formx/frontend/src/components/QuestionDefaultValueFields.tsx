import type { QuestionConfig } from '../lib/api';
import { setConfigDefaultValue } from '../lib/defaultValue';

type Props = {
  questionType: string;
  config: QuestionConfig;
  /** Current default from config.default_value */
  value: unknown;
  compact?: boolean;
  onConfigChange: (next: QuestionConfig) => void;
};

export function QuestionDefaultValueFields({
  questionType,
  config,
  value,
  compact,
  onConfigChange,
}: Props) {
  const lbl = compact ? 'text-[10px] mb-0.5' : 'text-xs mb-1';
  const inp = compact
    ? 'px-2 py-1.5 rounded-md bg-white dark:bg-slate-800 border border-slate-300 dark:border-slate-600 text-slate-900 dark:text-white text-xs'
    : 'px-3 py-2 rounded-lg bg-slate-800 border border-slate-600 text-white text-sm';
  const opts = config.options || [];
  const fieldW = 'w-full max-w-xs';

  const patch = (def: unknown | undefined) => {
    onConfigChange(setConfigDefaultValue(config, def));
  };

  const block = (
    <div>
      <label className={`block text-slate-500 dark:text-slate-400 ${lbl}`}>Default value (optional)</label>
      <p className={`text-slate-500 ${compact ? 'text-[10px] mb-1' : 'text-xs mb-2'}`}>
        Shown to respondents; they can keep it or change it.
      </p>
    </div>
  );

  switch (questionType) {
    case 'text':
    case 'qrcode':
    case 'image':
    case 'document':
      return (
        <div className="space-y-1">
          {block}
          <input
            type="text"
            value={(value as string) ?? ''}
            onChange={(e) => patch(e.target.value === '' ? undefined : e.target.value)}
            className={`${fieldW} ${inp}`}
            placeholder="Preset text"
          />
        </div>
      );
    case 'integer':
      return (
        <div className="space-y-1">
          {block}
          <input
            type="number"
            step={1}
            value={value === undefined || value === null ? '' : String(value)}
            onChange={(e) => {
              const v = e.target.value;
              if (v === '') {
                patch(undefined);
                return;
              }
              const n = parseInt(v, 10);
              patch(Number.isNaN(n) ? undefined : n);
            }}
            className={`${fieldW} ${inp}`}
            placeholder="e.g. 42"
          />
        </div>
      );
    case 'float':
      return (
        <div className="space-y-1">
          {block}
          <input
            type="number"
            step="any"
            value={value === undefined || value === null ? '' : String(value)}
            onChange={(e) => {
              const v = e.target.value;
              if (v === '') {
                patch(undefined);
                return;
              }
              const n = parseFloat(v);
              patch(Number.isNaN(n) ? undefined : n);
            }}
            className={`${fieldW} ${inp}`}
            placeholder="e.g. 3.14"
          />
        </div>
      );
    case 'boolean':
      return (
        <div className="space-y-1">
          {block}
          <select
            value={value === true ? 'true' : value === false ? 'false' : ''}
            onChange={(e) => {
              const v = e.target.value;
              patch(v === '' ? undefined : v === 'true');
            }}
            className={`${fieldW} ${inp}`}
          >
            <option value="">No default</option>
            <option value="true">Yes (checked)</option>
            <option value="false">No (unchecked)</option>
          </select>
        </div>
      );
    case 'select':
      return (
        <div className="space-y-1">
          {block}
          <select
            value={value === undefined || value === null ? '' : String(value)}
            onChange={(e) => {
              const v = e.target.value;
              patch(v === '' ? undefined : parseInt(v, 10));
            }}
            className={`${fieldW} ${inp}`}
          >
            <option value="">No default</option>
            {opts.map((o) => (
              <option key={o.value} value={o.value}>
                {o.label}
              </option>
            ))}
          </select>
        </div>
      );
    case 'multiselect': {
      const selected = Array.isArray(value) ? (value as number[]) : [];
      const maxSel = config.max_selections;
      return (
        <div className="space-y-1">
          {block}
          <div className={`space-y-1.5 ${compact ? '' : 'space-y-2'}`}>
            {opts.map((o) => (
              <label key={o.value} className="flex items-center gap-2 cursor-pointer text-slate-700 dark:text-slate-300 text-xs">
                <input
                  type="checkbox"
                  className="rounded border-slate-600 bg-slate-800 text-violet-600"
                  checked={selected.includes(o.value)}
                  onChange={() => {
                    if (selected.includes(o.value)) {
                      const next = selected.filter((x) => x !== o.value);
                      patch(next.length ? next : undefined);
                    } else {
                      const next = [...selected, o.value];
                      if (maxSel && next.length > maxSel) return;
                      patch(next);
                    }
                  }}
                />
                {o.label}
              </label>
            ))}
          </div>
          {opts.length === 0 && (
            <p className="text-slate-500 text-xs">Add options above to choose defaults.</p>
          )}
        </div>
      );
    }
    case 'date':
      return (
        <div className="space-y-1">
          {block}
          <input
            type="date"
            value={(value as string) || ''}
            onChange={(e) => patch(e.target.value === '' ? undefined : e.target.value)}
            className={`${fieldW} ${inp}`}
          />
        </div>
      );
    case 'datetime':
      return (
        <div className="space-y-1">
          {block}
          <input
            type="datetime-local"
            value={(value as string) || ''}
            onChange={(e) => patch(e.target.value === '' ? undefined : e.target.value)}
            className={`${fieldW} ${inp}`}
          />
        </div>
      );
    default:
      return null;
  }
}
