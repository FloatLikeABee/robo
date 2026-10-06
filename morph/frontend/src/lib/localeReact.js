import { useSyncExternalStore } from 'react';
import { getLocale, subscribeLocale, translate } from './locale';

export function useLocale() {
  return useSyncExternalStore(subscribeLocale, getLocale, () => 'en');
}

export function useT() {
  const locale = useLocale();
  return (key, vars) => translate(locale, key, vars);
}
