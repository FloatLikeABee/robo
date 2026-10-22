import { onSheetKeyDown } from './phoneSheet';

function sheet() {
  document.body.innerHTML = '<div id="sheet"><button id="a">A</button><button id="b">B</button><button id="c">C</button></div>';
  return document.getElementById('sheet');
}

test('tab from the last control wraps inside the sheet', () => {
  const root = sheet();
  const last = document.getElementById('c');
  last.focus();
  const preventDefault = jest.fn();
  onSheetKeyDown({ key: 'Tab', shiftKey: false, preventDefault }, root, () => {});
  expect(preventDefault).toHaveBeenCalled();
  expect(document.activeElement).toBe(document.getElementById('a'));
});

test('shift-tab from the first control wraps to the last', () => {
  const root = sheet();
  document.getElementById('a').focus();
  onSheetKeyDown({ key: 'Tab', shiftKey: true, preventDefault: jest.fn() }, root, () => {});
  expect(document.activeElement).toBe(document.getElementById('c'));
});

test('escape closes the sheet', () => {
  const root = sheet();
  const onClose = jest.fn();
  const preventDefault = jest.fn();
  onSheetKeyDown({ key: 'Escape', preventDefault }, root, onClose);
  expect(preventDefault).toHaveBeenCalled();
  expect(onClose).toHaveBeenCalledTimes(1);
});
