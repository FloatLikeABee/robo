import { replyExceedsLineLimit } from './replyEnlarge';

test('a tall rendered body qualifies even when the text has no newlines', () => {
  const text = 'one long wrapped paragraph without line breaks';
  expect(text.includes('\n')).toBe(false);
  expect(replyExceedsLineLimit(220, 20)).toBe(true);
});

test('a short rendered body does not qualify even with many newlines', () => {
  const text = Array.from({ length: 40 }, () => 'x').join('\n');
  expect(text.split('\n').length).toBeGreaterThan(10);
  expect(replyExceedsLineLimit(100, 20)).toBe(false);
});

test('exactly ten line-heights does not qualify', () => {
  expect(replyExceedsLineLimit(200, 20, 10)).toBe(false);
  expect(replyExceedsLineLimit(201, 20, 10)).toBe(true);
});
