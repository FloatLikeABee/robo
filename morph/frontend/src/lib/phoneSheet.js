const FOCUSABLE =
  'a[href], button:not([disabled]), input:not([disabled]), textarea:not([disabled]), select:not([disabled]), [tabindex]:not([tabindex="-1"])';

function sheetFocusables(root) {
  if (!root || typeof root.querySelectorAll !== 'function') return [];
  return [...root.querySelectorAll(FOCUSABLE)].filter((el) => el.getAttribute('aria-hidden') !== 'true');
}

export function onSheetKeyDown(event, root, onClose) {
  if (!event) return;
  if (event.key === 'Escape') {
    event.preventDefault();
    if (typeof onClose === 'function') onClose();
    return;
  }
  if (event.key !== 'Tab') return;
  const items = sheetFocusables(root);
  if (items.length === 0) return;
  const index = items.indexOf(document.activeElement);
  const next = event.shiftKey
    ? items[index <= 0 ? items.length - 1 : index - 1]
    : items[index === -1 || index >= items.length - 1 ? 0 : index + 1];
  event.preventDefault();
  next.focus();
}
