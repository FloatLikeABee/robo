import { useEffect, useState } from 'react';

type SettingsState = {
  userUsername: string;
  userName: string;
  userEmail: string;
  userPhone: string;
  smtpHost: string;
  smtpPort: string;
  smtpUser: string;
  smtpFrom: string;
  exportFormsDir: string;
  exportAnswersDir: string;
};

const STORAGE_KEY = 'sheetx-settings-v1';
const LEGACY_STORAGE_KEY = 'formsx-settings-v1';

const DEFAULTS: SettingsState = {
  userUsername: '',
  userName: '',
  userEmail: '',
  userPhone: '',
  smtpHost: '',
  smtpPort: '587',
  smtpUser: '',
  smtpFrom: '',
  exportFormsDir: '~/exports/forms',
  exportAnswersDir: '~/exports/form-answers',
};

export function Settings() {
  const [settings, setSettings] = useState<SettingsState>(DEFAULTS);
  const [saved, setSaved] = useState(false);

  useEffect(() => {
    try {
      const raw =
        window.localStorage.getItem(STORAGE_KEY) ?? window.localStorage.getItem(LEGACY_STORAGE_KEY);
      if (!raw) return;
      const parsed = JSON.parse(raw) as Partial<SettingsState>;
      setSettings({ ...DEFAULTS, ...parsed });
    } catch {
      // ignore malformed local data
    }
  }, []);

  const patch = (key: keyof SettingsState, value: string) =>
    setSettings((prev) => ({ ...prev, [key]: value }));

  const save = () => {
    window.localStorage.setItem(STORAGE_KEY, JSON.stringify(settings));
    setSaved(true);
    window.setTimeout(() => setSaved(false), 1800);
  };

  const reset = () => {
    setSettings(DEFAULTS);
    window.localStorage.removeItem(STORAGE_KEY);
    setSaved(false);
  };

  const inputClass =
    'w-full px-3 py-2 rounded-lg bg-white dark:bg-slate-800 border border-slate-300 dark:border-slate-600 text-slate-900 dark:text-white text-sm';
  const labelClass = 'block text-[11px] text-slate-500 dark:text-slate-400 mb-1';

  return (
    <div className="flex flex-col flex-1 min-h-0 overflow-auto">
      <div className="w-full max-w-5xl mx-auto px-2 sm:px-4 lg:px-6">
      <div className="shrink-0 flex items-center justify-between gap-3 mb-3">
        <h1 className="text-xl font-semibold text-slate-900 dark:text-white">Settings</h1>
        <div className="flex items-center gap-2">
          {saved && (
            <span className="text-xs text-emerald-700 dark:text-emerald-300 bg-emerald-100 dark:bg-emerald-900/30 px-2 py-1 rounded-md">
              Saved
            </span>
          )}
          <button
            type="button"
            onClick={reset}
            className="px-3 py-1.5 rounded-lg bg-slate-200 dark:bg-slate-700 text-slate-800 dark:text-slate-100 text-sm"
          >
            Reset
          </button>
          <button
            type="button"
            onClick={save}
            className="px-3 py-1.5 rounded-lg bg-violet-600 hover:bg-violet-500 text-white text-sm"
          >
            Save settings
          </button>
        </div>
      </div>

      <div className="space-y-3">
        <section className="rounded-xl border border-slate-200 dark:border-slate-700/50 bg-white dark:bg-slate-900/25 p-3">
          <h2 className="text-sm font-medium text-slate-900 dark:text-white mb-2">User info</h2>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
            <div>
              <label className={labelClass}>Username</label>
              <input
                value={settings.userUsername}
                onChange={(e) => patch('userUsername', e.target.value)}
                className={inputClass}
                placeholder="username"
              />
            </div>
            <div>
              <label className={labelClass}>Full name</label>
              <input
                value={settings.userName}
                onChange={(e) => patch('userName', e.target.value)}
                className={inputClass}
                placeholder="Your full name"
              />
            </div>
            <div>
              <label className={labelClass}>Email</label>
              <input
                value={settings.userEmail}
                onChange={(e) => patch('userEmail', e.target.value)}
                className={inputClass}
                placeholder="you@example.com"
              />
            </div>
            <div>
              <label className={labelClass}>Phone</label>
              <input
                value={settings.userPhone}
                onChange={(e) => patch('userPhone', e.target.value)}
                className={inputClass}
                placeholder="+1 234 567 890"
              />
            </div>
          </div>
        </section>

        <section className="rounded-xl border border-slate-200 dark:border-slate-700/50 bg-white dark:bg-slate-900/25 p-3">
          <h2 className="text-sm font-medium text-slate-900 dark:text-white mb-2">Email service info</h2>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
            <div>
              <label className={labelClass}>SMTP host</label>
              <input
                value={settings.smtpHost}
                onChange={(e) => patch('smtpHost', e.target.value)}
                className={inputClass}
                placeholder="smtp.example.com"
              />
            </div>
            <div>
              <label className={labelClass}>SMTP port</label>
              <input
                value={settings.smtpPort}
                onChange={(e) => patch('smtpPort', e.target.value)}
                className={inputClass}
                placeholder="587"
              />
            </div>
            <div>
              <label className={labelClass}>SMTP user</label>
              <input
                value={settings.smtpUser}
                onChange={(e) => patch('smtpUser', e.target.value)}
                className={inputClass}
                placeholder="smtp-user"
              />
            </div>
            <div>
              <label className={labelClass}>SMTP from email</label>
              <input
                value={settings.smtpFrom}
                onChange={(e) => patch('smtpFrom', e.target.value)}
                className={inputClass}
                placeholder="no-reply@example.com"
              />
            </div>
          </div>
        </section>

        <section className="rounded-xl border border-slate-200 dark:border-slate-700/50 bg-white dark:bg-slate-900/25 p-3">
          <h2 className="text-sm font-medium text-slate-900 dark:text-white mb-2">Export file storage</h2>
          <div className="grid grid-cols-1 gap-3">
            <div>
              <label className={labelClass}>Exported forms files path</label>
              <input
                value={settings.exportFormsDir}
                onChange={(e) => patch('exportFormsDir', e.target.value)}
                className={inputClass}
                placeholder="~/exports/forms"
              />
            </div>
            <div>
              <label className={labelClass}>Exported form answers files path</label>
              <input
                value={settings.exportAnswersDir}
                onChange={(e) => patch('exportAnswersDir', e.target.value)}
                className={inputClass}
                placeholder="~/exports/form-answers"
              />
            </div>
          </div>
        </section>
      </div>
      </div>
    </div>
  );
}
