/** True when a rendered reply is taller than `lines` line-heights. Newline count is not an input. */
export function replyExceedsLineLimit(scrollHeight, lineHeight, lines = 10) {
  const height = Number(scrollHeight);
  const line = Number(lineHeight);
  const count = Number(lines);
  if (!Number.isFinite(height) || !Number.isFinite(line) || line <= 0 || !Number.isFinite(count)) return false;
  return height > line * count;
}
