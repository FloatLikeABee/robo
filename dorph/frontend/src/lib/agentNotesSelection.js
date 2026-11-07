/** Select the first note only when nothing is selected yet. Back leaves null. */
export function noteSelectionAfterLoad(rows, selectedId) {
  if (selectedId != null) return selectedId;
  const list = Array.isArray(rows) ? rows : [];
  return list.length ? list[0].id : null;
}
