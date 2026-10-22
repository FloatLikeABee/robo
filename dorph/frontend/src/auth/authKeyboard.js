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

function visibleBand(visualViewport, gap) {
  const top = Number(visualViewport.offsetTop) || 0;
  const height = Number(visualViewport.height);
  if (!Number.isFinite(height)) return null;
  return { visibleTop: top + gap, visibleBottom: top + height - gap };
}

/** True when a control sits outside the visible visual viewport. */
export function controlHidden(rect, visualViewport, gap = 8) {
  if (!rect || !visualViewport) return false;
  const band = visibleBand(visualViewport, gap);
  if (!band) return false;
  return rect.bottom > band.visibleBottom || rect.top < band.visibleTop;
}

/** Pixels to scroll so the control sits inside the visible visual viewport. */
export function scrollDelta(rect, visualViewport, gap = 8) {
  if (!rect || !visualViewport) return 0;
  const band = visibleBand(visualViewport, gap);
  if (!band) return 0;
  if (rect.bottom > band.visibleBottom) return rect.bottom - band.visibleBottom;
  if (rect.top < band.visibleTop) return rect.top - band.visibleTop;
  return 0;
}
