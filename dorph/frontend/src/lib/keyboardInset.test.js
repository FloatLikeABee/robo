const { keyboardCoveredPx } = require('./keyboardInset');

test('keyboard overlap is the layout height the visual viewport does not show', () => {
  expect(keyboardCoveredPx(800, 500, 0)).toBe(300);
});

test('keyboard overlap is zero when the layout viewport already shrank', () => {
  expect(keyboardCoveredPx(500, 500, 0)).toBe(0);
});

test('keyboard overlap is zero when the visual viewport scroll already clears the keyboard', () => {
  expect(keyboardCoveredPx(800, 500, 300)).toBe(0);
});

test('keyboard overlap ignores sub-pixel noise and non-numbers', () => {
  expect(keyboardCoveredPx(800, 799.4, 0)).toBe(0);
  expect(keyboardCoveredPx(undefined, 500, 0)).toBe(0);
});
