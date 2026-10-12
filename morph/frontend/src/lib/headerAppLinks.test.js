import { HEADER_APP_ICONS, morphUtilsBaseURL } from './headerAppLinks';

test('header icons are the public svg paths the server must serve', () => {
  expect(HEADER_APP_ICONS.bk).toMatch(/\/icons\/bk-icon\.svg$/);
  expect(HEADER_APP_ICONS.morphdata).toMatch(/\/icons\/morph-data-icon\.svg$/);
  expect(HEADER_APP_ICONS.morphutils).toMatch(/\/icons\/morph-utils-icon\.svg$/);
});

test('production omits MorphUtils when the URL is unset or loopback', () => {
  expect(morphUtilsBaseURL('production', undefined)).toBe('');
  expect(morphUtilsBaseURL('production', '')).toBe('');
  expect(morphUtilsBaseURL('production', 'http://localhost:3040')).toBe('');
  expect(morphUtilsBaseURL('production', 'http://127.0.0.1:3040')).toBe('');
  expect(morphUtilsBaseURL('production', 'http://[::1]:3040')).toBe('');
  expect(morphUtilsBaseURL('production', 'http://LOCALHOST:9')).toBe('');
});

test('production keeps a configured public MorphUtils URL', () => {
  expect(morphUtilsBaseURL('production', '  https://utils.example.com/app  ')).toBe(
    'https://utils.example.com/app'
  );
});

test('development still defaults MorphUtils to localhost', () => {
  expect(morphUtilsBaseURL('development', undefined)).toBe('http://localhost:3040');
  expect(morphUtilsBaseURL('development', 'http://127.0.0.1:3040')).toBe('http://127.0.0.1:3040');
});
