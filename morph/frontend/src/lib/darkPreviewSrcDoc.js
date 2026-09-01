const DARK_SCROLL_STYLE = `<style>
:root {
  color-scheme: dark;
  --scrollbar-thumb: #64748b;
  --scrollbar-track: #0b1220;
  scrollbar-width: thin;
  scrollbar-color: var(--scrollbar-thumb) var(--scrollbar-track);
}
html, body { margin: 0; min-height: 100%; }
@supports not (scrollbar-color: auto) {
  *::-webkit-scrollbar { width: 12px; height: 12px; }
  *::-webkit-scrollbar-thumb { background: var(--scrollbar-thumb); }
  *::-webkit-scrollbar-track { background: var(--scrollbar-track); }
}
</style>`;

/** Wrap or inject dark color-scheme + themed scrollbars into HTML preview srcDoc. */
export function withDarkPreviewSrcDoc(html) {
  const src = String(html || '');
  const inject = `<meta name="color-scheme" content="dark">${DARK_SCROLL_STYLE}`;
  if (!src.trim()) {
    return `<!DOCTYPE html><html><head>${inject}</head><body></body></html>`;
  }
  if (/scrollbar-color/i.test(src) && /color-scheme/i.test(src)) {
    return src;
  }
  if (/<head[\s>]/i.test(src)) {
    return src.replace(/<head([^>]*)>/i, `<head$1>${inject}`);
  }
  return `<!DOCTYPE html><html><head>${inject}</head><body>${src}</body></html>`;
}

export const darkPreviewIframeSx = {
  width: '100%',
  height: '100%',
  border: 0,
  display: 'block',
  bgcolor: '#0b1220',
  colorScheme: 'dark',
};
