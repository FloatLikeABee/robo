import { SESSION_SWATCH_COLORS, sessionColorsFor } from './sessionSwatch';

test('session tiles stay in the twilight palette and neighbors differ', () => {
  const ids = ['a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i'];
  const colors = sessionColorsFor(ids);
  colors.forEach((color, i) => {
    expect(SESSION_SWATCH_COLORS).toContain(color);
    if (i > 0) expect(color).not.toBe(colors[i - 1]);
  });
  expect(new Set(colors).size).toBeLessThan(ids.length);
});
