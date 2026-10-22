/**
 * MUI Drawer/Modal scroll lock and backdrop nodes sometimes survive route changes
 * (e.g. login redirect, React StrictMode double-mount), leaving a grey layer that
 * blocks clicks with no console error.
 */
export function releaseStuckOverlays() {
  if (typeof document === 'undefined') return;

  document.body.style.removeProperty('overflow');
  document.body.style.removeProperty('padding-right');
  document.documentElement.style.removeProperty('overflow');
  document.documentElement.style.removeProperty('padding-right');

  document.querySelectorAll('.MuiBackdrop-root').forEach((backdrop) => {
    const modal = backdrop.closest('.MuiModal-root');
    if (!modal) {
      backdrop.remove();
      return;
    }
    const style = window.getComputedStyle(modal);
    const hidden =
      modal.getAttribute('aria-hidden') === 'true' ||
      style.display === 'none' ||
      style.visibility === 'hidden';
    if (hidden) modal.remove();
  });

  document.querySelectorAll('.MuiModal-root[aria-hidden="true"]').forEach((modal) => {
    modal.remove();
  });

  const root = document.getElementById('root');
  if (root?.getAttribute('aria-hidden') === 'true') {
    const openModal = document.querySelector('.MuiModal-root:not([aria-hidden="true"])');
    if (!openModal) root.removeAttribute('aria-hidden');
  }
}
