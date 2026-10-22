import React from 'react';
import { setLocale } from '../lib/locale';
import { useLocale, useT } from '../lib/localeReact';

export default function LocaleSwitch({ className = '' }) {
  const locale = useLocale();
  const t = useT();
  return (
    <div className={`locale-switch ${className}`.trim()} role="group" aria-label={t('language')}>
      <button type="button" aria-pressed={locale === 'en'} onClick={() => setLocale('en')}>
        English
      </button>
      <button type="button" aria-pressed={locale === 'zh'} onClick={() => setLocale('zh')}>
        中文
      </button>
    </div>
  );
}
