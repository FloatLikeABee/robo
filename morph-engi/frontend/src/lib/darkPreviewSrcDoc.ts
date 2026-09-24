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
</style>`

const LEGACY_THEME_RE = /#2dd4bf|#134e4a|#10b981|#059669|#0f766e/i

const LEGACY_THEME_OVERRIDE = `<style id="morph-dark-theme-override">
:root {
  --ink: #e8eef7;
  --muted: #94a3b8;
  --line: #1e293b;
  --bg: #0b1220;
  --card: #111827;
  --accent: #38bdf8;
  --accent-purple: #818cf8;
}
body {
  background: radial-gradient(1200px 600px at 10% -10%, #1e1b4b 0%, #0b1220 55%) !important;
  color: #e8eef7 !important;
}
a, .prose a { color: #38bdf8 !important; }
</style>`

function needsLegacyThemeOverride(html: string) {
  return LEGACY_THEME_RE.test(String(html || ''))
}

function injectBeforeHeadClose(html: string, snippet: string) {
  if (/<\/head>/i.test(html)) {
    return html.replace(/<\/head>/i, `${snippet}</head>`)
  }
  return `${snippet}${html}`
}

function htmlHasMermaid(html: string) {
  return /language-mermaid|class=["']mermaid["']|```mermaid/i.test(String(html || ''))
}

const MERMAID_BOOTSTRAP = `<script type="module">
import mermaid from "https://cdn.jsdelivr.net/npm/mermaid@11/dist/mermaid.esm.min.mjs";
mermaid.initialize({ startOnLoad: false, theme: "dark", securityLevel: "strict", suppressErrorRendering: true });
document.querySelectorAll("code.language-mermaid").forEach((el) => {
  const pre = document.createElement("pre");
  pre.className = "mermaid";
  pre.textContent = el.textContent || "";
  el.parentElement?.replaceWith(pre);
});
for (const el of document.querySelectorAll("pre.mermaid, .mermaid")) {
  const source = (el.textContent || "").trim();
  if (!source) continue;
  const id = "mmd-" + Math.random().toString(36).slice(2, 10);
  try {
    const { svg } = await mermaid.render(id, source);
    el.innerHTML = svg;
  } catch (e) {
    el.textContent = source;
    console.warn("mermaid", e);
  }
}
</script>`

function injectBeforeBodyClose(html: string, snippet: string) {
  if (/<\/body>/i.test(html)) {
    return html.replace(/<\/body>/i, `${snippet}</body>`)
  }
  return `${html}${snippet}`
}

/** Wrap or inject dark color-scheme + themed scrollbars into HTML preview srcDoc. */
export function withDarkPreviewSrcDoc(html: string) {
  const src = String(html || '')
  const inject = `<meta name="color-scheme" content="dark">${DARK_SCROLL_STYLE}`
  const override = needsLegacyThemeOverride(src) ? LEGACY_THEME_OVERRIDE : ''

  if (!src.trim()) {
    return `<!DOCTYPE html><html><head>${inject}</head><body></body></html>`
  }

  let out = src
  const hasScroll = /scrollbar-color/i.test(src)
  const hasScheme = /color-scheme/i.test(src)

  if (!hasScroll || !hasScheme) {
    if (/<head[\s>]/i.test(out)) {
      out = out.replace(/<head([^>]*)>/i, `<head$1>${inject}`)
    } else {
      out = `<!DOCTYPE html><html><head>${inject}</head><body>${out}</body></html>`
    }
  }

  if (override) {
    out = injectBeforeHeadClose(out, override)
  }

  if (htmlHasMermaid(out) && !/mermaid\.esm\.min\.mjs/i.test(out)) {
    out = injectBeforeBodyClose(out, MERMAID_BOOTSTRAP)
  }

  return out
}
