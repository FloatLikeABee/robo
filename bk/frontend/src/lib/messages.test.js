import { en, zh } from './messages';

test('every english catalog key has simplified chinese', () => {
  expect(Object.keys(zh).sort()).toEqual(Object.keys(en).sort());
  for (const key of Object.keys(en)) {
    expect(en[key].trim()).not.toBe('');
    expect(zh[key].trim()).not.toBe('');
  }
  expect(zh.morphTools).toBe('MorphTools');
});
