/** Count all nodes in a comment tree returned by GET /api/tran/comments (nested replies). */
export function countCommentTree(nodes) {
  if (!Array.isArray(nodes) || nodes.length === 0) return 0;
  let n = 0;
  for (const c of nodes) {
    n += 1;
    if (c.replies && c.replies.length > 0) n += countCommentTree(c.replies);
  }
  return n;
}
