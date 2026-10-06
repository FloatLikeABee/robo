import { HEADER_APP_ICONS, morphUtilsBaseURL } from './headerAppLinks';

test('header icons are the public svg paths the server must serve', () => {
  expect(HEADER_APP_ICONS.bk).toMatch(/\/icons\/bk-icon\.svg$/);
  expect(HEADER_APP_ICONS.morphdata).toMatch(/\/icons\/morph-data-icon\.svg$/);
  expect(HEADER_APP_ICONS.morphutils).toMatch(/\/icons\/morph-utils-icon\.svg$/);
});

test('a public MorphUtils URL does not produce a header target', () => {
  expect(morphUtilsBaseURL('production', 'https://utils.example.com/app')).toBe('');
  expect(morphUtilsBaseURL('development', undefined)).toBe('');
  expect(morphUtilsBaseURL('development', 'http://localhost:3040')).toBe('');
});
