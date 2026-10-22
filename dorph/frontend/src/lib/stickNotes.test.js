import { STICKER_PALETTE, STICKER_TITLE_PALETTE, filterStickNotes, stickerColorMap, stickerTitleColorMap } from './stickNotes';

function rows(n) {
  return Array.from({ length: n }, (_, i) => ({
    id: i + 1,
    title: i === 3 ? 'Alpha title' : `Note ${i + 1}`,
    description: i === 5 ? 'unique body phrase' : 'plain',
  }));
}

test('first twelve stickers use different stable colors', () => {
  const list = rows(12);
  const first = stickerColorMap(list);
  const again = stickerColorMap([...list].reverse());
  const colors = list.map((row) => first.get(row.id));
  expect(new Set(colors).size).toBe(12);
  expect(STICKER_PALETTE.length).toBeGreaterThanOrEqual(12);
  list.forEach((row) => {
    expect(again.get(row.id)).toBe(first.get(row.id));
  });
});

test('a search subset keeps the color from the full id order', () => {
  const list = rows(14);
  const full = stickerColorMap(list);
  const shown = filterStickNotes(list, 'alpha');
  expect(shown).toHaveLength(1);
  expect(full.get(shown[0].id)).toBe(STICKER_PALETTE[3]);
  expect(stickerTitleColorMap(list).get(shown[0].id)).toBe(STICKER_TITLE_PALETTE[3]);
  expect(stickerTitleColorMap(list).get(shown[0].id)).not.toBe(full.get(shown[0].id));
  expect(stickerColorMap(list).get(shown[0].id)).toBe(full.get(shown[0].id));
  expect(filterStickNotes(list, 'no-such-note')).toEqual([]);
  expect(filterStickNotes(list, 'UNIQUE BODY').map((row) => row.id)).toEqual([6]);
  expect(filterStickNotes(list, '').length).toBe(14);
});
