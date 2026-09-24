// ponytail: overlap under 120px is browser chrome, not a keyboard. Portrait OSK is larger.
// Upgrade path: env(keyboard-inset-bottom) once iOS exposes it without a global viewport change.
const KEYBOARD_MIN_PX = 120;

/** CSS pixels of layout viewport covered by the on-screen keyboard, or 0. */
export function authKeyboardOverlap({ innerHeight, visualViewport, focused }) {
  if (!focused || !visualViewport) return 0;
  const height = Number(visualViewport.height);
  const offsetTop = Number(visualViewport.offsetTop) || 0;
  const layoutHeight = Number(innerHeight);
  if (!Number.isFinite(height) || !Number.isFinite(layoutHeight)) return 0;
  const overlap = Math.round(layoutHeight - height - offsetTop);
  if (overlap < KEYBOARD_MIN_PX) return 0;
  return overlap;
}

/** True when a control sits outside the visible visual viewport. */
export function controlHidden(rect, visualViewport, gap = 8) {
  if (!rect || !visualViewport) return false;
  const top = Number(visualViewport.offsetTop) || 0;
  const height = Number(visualViewport.height);
  if (!Number.isFinite(height)) return false;
  const visibleTop = top + gap;
  const visibleBottom = top + height - gap;
  return rect.bottom > visibleBottom || rect.top < visibleTop;
}
