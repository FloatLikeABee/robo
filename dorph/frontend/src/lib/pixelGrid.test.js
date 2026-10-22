import { parsePixelGrid, PIXEL_MAX } from './pixelGrid';

test('parses hex and dot hash cells', () => {
  const g = parsePixelGrid('#ff0000 . #\n. #38bdf8 .');
  expect(g).toEqual({
    width: 3,
    height: 2,
    rows: [
      ['#ff0000', null, '#38bdf8'],
      [null, '#38bdf8', null],
    ],
  });
});

test('rejects oversize grids', () => {
  const line = Array(PIXEL_MAX + 1)
    .fill('#')
    .join(' ');
  expect(parsePixelGrid(line)).toBeNull();
  const tall = Array(PIXEL_MAX + 1)
    .fill('#')
    .join('\n');
  expect(parsePixelGrid(tall)).toBeNull();
});

test('rejects ragged rows and junk tokens', () => {
  expect(parsePixelGrid('# #\n#')).toBeNull();
  expect(parsePixelGrid('red green')).toBeNull();
});
