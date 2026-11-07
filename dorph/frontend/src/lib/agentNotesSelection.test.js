import { noteSelectionAfterLoad } from './agentNotesSelection';

test('load selects the first note and a later back stays cleared', () => {
  const rows = [{ id: 7 }, { id: 8 }];
  const selected = noteSelectionAfterLoad(rows, null);
  expect(selected).toBe(7);
  const afterBack = null;
  expect(afterBack).toBe(null);
  expect(noteSelectionAfterLoad(rows, 8)).toBe(8);
  expect(noteSelectionAfterLoad([], null)).toBe(null);
});
