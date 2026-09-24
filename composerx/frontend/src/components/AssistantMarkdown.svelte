<script>
  import { tick } from 'svelte'
  import { renderMarkdownHtml } from '../lib/contentMarkdown'
  import { runMermaidIn } from '../lib/runMermaidIn'

  /** @type {{ text?: string }} */
  let { text = '' } = $props()
  const html = $derived(renderMarkdownHtml(text))
  let wrap = $state(null)
  /** @type {{ html?: string, src?: string, alt?: string } | null} */
  let item = $state(null)
  /** @type {HTMLButtonElement | null} */
  let closeBtn = $state(null)
  /** @type {HTMLElement | null} */
  let prevFocus = null

  $effect(() => {
    html
    let cancelled = false
    const ac = new AbortController()
    void tick().then(async () => {
      await runMermaidIn(wrap)
      if (cancelled || !wrap) return
      wrap.querySelectorAll('.mermaid svg, img').forEach((el) => {
        el.setAttribute('tabindex', '0')
        el.setAttribute('role', 'button')
        el.setAttribute('aria-label', 'Enlarge visual')
        el.classList.add('chat-visual-enlarge')
        el.addEventListener('click', activate, { signal: ac.signal })
        el.addEventListener('keydown', onKey, { signal: ac.signal })
      })
    })
    return () => {
      cancelled = true
      ac.abort()
    }
  })

  $effect(() => {
    if (!item) return
    closeBtn?.focus()
    const onKey = (e) => {
      if (e.key === 'Escape') close()
    }
    window.addEventListener('keydown', onKey)
    const prev = document.body.style.overflow
    document.body.style.overflow = 'hidden'
    return () => {
      window.removeEventListener('keydown', onKey)
      document.body.style.overflow = prev
    }
  })

  function scaleSvgHtml(raw) {
    try {
      const doc = new DOMParser().parseFromString(String(raw || ''), 'image/svg+xml')
      const svgEl = doc.documentElement
      if (!svgEl || svgEl.tagName.toLowerCase() !== 'svg') return raw
      const w = parseFloat((svgEl.getAttribute('width') || '').replace(/px$/i, ''))
      const h = parseFloat((svgEl.getAttribute('height') || '').replace(/px$/i, ''))
      const percent = /%/.test(svgEl.getAttribute('width') || '') || /%/.test(svgEl.getAttribute('height') || '')
      if (!svgEl.getAttribute('viewBox') && !percent && w > 0 && h > 0) {
        svgEl.setAttribute('viewBox', `0 0 ${w} ${h}`)
      }
      svgEl.removeAttribute('width')
      svgEl.removeAttribute('height')
      svgEl.setAttribute('preserveAspectRatio', 'xMidYMid meet')
      svgEl.style.width = '100%'
      svgEl.style.height = '100%'
      svgEl.style.maxWidth = '100%'
      svgEl.style.maxHeight = '100%'
      return new XMLSerializer().serializeToString(svgEl)
    } catch {
      return raw
    }
  }

  function close() {
    item = null
    prevFocus?.focus?.()
    prevFocus = null
  }

  function activate(e) {
    if (e.target.closest?.('a')) return
    const svg = e.target.closest?.('svg')
    const img = e.target.closest?.('img')
    if (svg && wrap?.contains(svg)) {
      prevFocus = document.activeElement
      item = { html: scaleSvgHtml(svg.outerHTML) }
      return
    }
    if (img && wrap?.contains(img)) {
      prevFocus = document.activeElement
      item = { src: img.currentSrc || img.src, alt: img.alt || 'Enlarged visual' }
    }
  }

  function portal(node) {
    document.body.appendChild(node)
    return {
      destroy() {
        node.remove()
      },
    }
  }

  function onKey(e) {
    if (e.key !== 'Enter' && e.key !== ' ') return
    if (!e.target.closest?.('svg, img')) return
    e.preventDefault()
    activate(e)
  }
</script>

<div class="assistant-md" bind:this={wrap}>{@html html}</div>

{#if item}
  <div class="visual-lightbox-overlay" role="presentation" onclick={close} use:portal>
    <div
      class="visual-lightbox"
      role="dialog"
      aria-modal="true"
      aria-label="Enlarged visual"
      tabindex="-1"
      onclick={(e) => e.stopPropagation()}
      onkeydown={(e) => {
        if (e.key === 'Escape') close()
      }}
    >
      <button bind:this={closeBtn} type="button" class="visual-lightbox-close" aria-label="Close" onclick={close}>✕</button>
      {#if item.html}
        <div class="visual-lightbox-svg">{@html item.html}</div>
      {:else}
        <img class="visual-lightbox-img" src={item.src} alt={item.alt || 'Enlarged visual'} />
      {/if}
    </div>
  </div>
{/if}

<style>
  :global(.assistant-md svg[role='button']),
  :global(.assistant-md img[role='button']) {
    cursor: zoom-in;
  }
  .visual-lightbox-overlay {
    position: fixed;
    inset: 0;
    z-index: 1300;
    display: flex;
    align-items: center;
    justify-content: center;
    padding: clamp(12px, 3vw, 28px);
    box-sizing: border-box;
    background: rgba(2, 4, 8, 0.78);
  }
  .visual-lightbox {
    position: relative;
    width: min(96vw, 1800px);
    height: min(92dvh, 1600px);
    display: flex;
    flex-direction: column;
    overflow: hidden;
    padding: 48px 16px 16px;
    box-sizing: border-box;
    border-radius: 16px;
    border: 1px solid #334155;
    background: #0f172a;
    color: #e2e8f0;
  }
  .visual-lightbox-close {
    position: absolute;
    top: 8px;
    right: 8px;
    width: 36px;
    height: 36px;
    border-radius: 10px;
    border: 1px solid #334155;
    background: #1e293b;
    color: #e2e8f0;
    font-size: 1.25rem;
    cursor: pointer;
  }
  .visual-lightbox-svg,
  .visual-lightbox-img {
    flex: 1;
    min-width: 0;
    min-height: 0;
    width: 100%;
    height: 100%;
  }
  .visual-lightbox-svg {
    display: flex;
    align-items: center;
    justify-content: center;
  }
  .visual-lightbox-svg :global(svg) {
    display: block;
    width: 100%;
    height: 100%;
    max-width: 100%;
    max-height: 100%;
  }
  .visual-lightbox-img {
    object-fit: contain;
  }
</style>
