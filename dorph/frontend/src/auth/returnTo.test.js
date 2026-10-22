import { safeReturnPath } from './returnTo';

test('keeps a DorphNotes path including search and hash', () => {
  expect(safeReturnPath('/dorphdata/research')).toBe('/dorphdata/research');
  expect(safeReturnPath('/dorphdata/research?tab=1#draft')).toBe('/dorphdata/research?tab=1#draft');
});

test('rejects absolute and protocol-relative return targets', () => {
  expect(safeReturnPath('https://evil.example/phish')).toBe('');
  expect(safeReturnPath('//evil.example')).toBe('');
  expect(safeReturnPath('/\\evil.example')).toBe('');
  expect(safeReturnPath('javascript:alert(1)')).toBe('');
  expect(safeReturnPath('/dorphdata/\nresearch')).toBe('');
});
