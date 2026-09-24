import { authKeyboardOverlap, controlHidden } from './authKeyboard';

test('reports no keyboard overlap when the sign-in field is not focused', () => {
  expect(
    authKeyboardOverlap({
      innerHeight: 844,
      visualViewport: { height: 430, offsetTop: 0 },
      focused: false,
    })
  ).toBe(0);
});

test('reports no keyboard overlap when the visual viewport is missing', () => {
  expect(
    authKeyboardOverlap({
      innerHeight: 844,
      visualViewport: null,
      focused: true,
    })
  ).toBe(0);
});

test('ignores a browser-chrome overlap under 120px', () => {
  expect(
    authKeyboardOverlap({
      innerHeight: 844,
      visualViewport: { height: 725, offsetTop: 0 },
      focused: true,
    })
  ).toBe(0);
});

test('returns a portrait keyboard overlap in CSS pixels', () => {
  expect(
    authKeyboardOverlap({
      innerHeight: 844,
      visualViewport: { height: 430, offsetTop: 0 },
      focused: true,
    })
  ).toBe(414);
});

test('subtracts the visual viewport offset from the keyboard overlap', () => {
  expect(
    authKeyboardOverlap({
      innerHeight: 844,
      visualViewport: { height: 500, offsetTop: 40 },
      focused: true,
    })
  ).toBe(304);
});

test('treats the 120px overlap boundary as a keyboard', () => {
  expect(
    authKeyboardOverlap({
      innerHeight: 844,
      visualViewport: { height: 724, offsetTop: 0 },
      focused: true,
    })
  ).toBe(120);
});

test('treats a control below the visible bottom as hidden', () => {
  expect(
    controlHidden({ top: 360, bottom: 404 }, { height: 380, offsetTop: 0 })
  ).toBe(true);
});

test('treats a control inside the visible viewport as shown', () => {
  expect(
    controlHidden({ top: 80, bottom: 124 }, { height: 430, offsetTop: 0 })
  ).toBe(false);
});
