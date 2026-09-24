import { safeReturnPath } from './returnTo';

test('keeps a MorphNotes path including search and hash', () => {
  expect(safeReturnPath('/morphdata/research')).toBe('/morphdata/research');
  expect(safeReturnPath('/morphdata/research?tab=1#draft')).toBe('/morphdata/research?tab=1#draft');
});

test('rejects absolute and protocol-relative return targets', () => {
  expect(safeReturnPath('https://evil.example/phish')).toBe('');
  expect(safeReturnPath('//evil.example')).toBe('');
  expect(safeReturnPath('/\\evil.example')).toBe('');
  expect(safeReturnPath('javascript:alert(1)')).toBe('');
  expect(safeReturnPath('/morphdata/\nresearch')).toBe('');
});
