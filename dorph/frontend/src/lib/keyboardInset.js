/** Pixels of the layout viewport covered by the keyboard, after visual-viewport scroll. */
export function keyboardCoveredPx(layoutHeight, viewportHeight, offsetTop) {
  const covered = Number(layoutHeight) - Number(viewportHeight) - Number(offsetTop);
  if (!Number.isFinite(covered) || covered < 1) return 0;
  return Math.round(covered);
}

export function bindKeyboardInset(doc = document, win = window) {
  const root = doc.documentElement;
  const apply = () => {
    const vv = win.visualViewport;
    const px = vv ? keyboardCoveredPx(win.innerHeight, vv.height, vv.offsetTop) : 0;
    root.style.setProperty('--keyboard-inset', `${px}px`);
    if (px > 0) root.setAttribute('data-keyboard-open', '');
    else root.removeAttribute('data-keyboard-open');
  };
  apply();
  const vv = win.visualViewport;
  vv?.addEventListener('resize', apply);
  vv?.addEventListener('scroll', apply);
  win.addEventListener('resize', apply);
  return () => {
    vv?.removeEventListener('resize', apply);
    vv?.removeEventListener('scroll', apply);
    win.removeEventListener('resize', apply);
    root.style.removeProperty('--keyboard-inset');
    root.removeAttribute('data-keyboard-open');
  };
}
