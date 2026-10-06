import { en, zh } from './messages';
import { detectLocale, documentLang, normalizeLocale, resolveLocale, translate } from './locale';

test('the ui stays english when a saved choice or the browser is chinese', () => {
  expect(resolveLocale({ query: 'zh', saved: 'zh', languages: ['zh-CN', 'zh-TW'] })).toBe('en');
  expect(documentLang(resolveLocale())).toBe('en');
});

test('only en and zh are locales', () => {
  expect(normalizeLocale(' ZH ')).toBe('zh');
  expect(normalizeLocale('en')).toBe('en');
  expect(normalizeLocale('zh-CN')).toBe('');
  expect(normalizeLocale('fr')).toBe('');
});

test('chinese browsers including traditional tags detect as zh', () => {
  expect(detectLocale(['zh-Hant'])).toBe('zh');
  expect(detectLocale(['zh-HK', 'en'])).toBe('zh');
  expect(detectLocale(['en', 'de'])).toBe('en');
});

test('document language is zh-Hans or en', () => {
  expect(documentLang('zh')).toBe('zh-Hans');
  expect(documentLang('en')).toBe('en');
});

test('missing chinese strings fall back to english', () => {
  expect(translate('zh', 'skills')).toBe('技能');
  expect(translate('en', 'skills')).toBe('Skills');
  expect(translate('zh', 'not-a-key')).toBe('not-a-key');
});

test('every english catalog key has simplified chinese', () => {
  expect(Object.keys(zh).sort()).toEqual(Object.keys(en).sort());
  for (const key of Object.keys(en)) {
    expect(en[key].trim()).not.toBe('');
    expect(zh[key].trim()).not.toBe('');
  }
  expect(zh.morphAI).toBe('Morph AI');
  expect(zh.morphNotes).toBe('MorphNotes');
  expect(zh.morphTools).toBe('MorphTools');
});
