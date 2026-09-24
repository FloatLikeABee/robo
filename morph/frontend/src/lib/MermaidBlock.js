import { useEffect, useRef, useState } from 'react';
import { enlargeBind, useVisualLightbox } from './visualLightbox';

let mermaidReady = null;

function loadMermaid() {
  if (!mermaidReady) {
    mermaidReady = import('mermaid').then((mod) => {
      const mermaid = mod.default;
      mermaid.initialize({
        startOnLoad: false,
        theme: 'dark',
        securityLevel: 'strict',
        suppressErrorRendering: true,
      });
      return mermaid;
    });
  }
  return mermaidReady;
}

export default function MermaidBlock({ source }) {
  const ref = useRef(null);
  const [fallback, setFallback] = useState(null);
  const open = useVisualLightbox();

  useEffect(() => {
    let cancelled = false;
    setFallback(null);
    const text = String(source || '').trim();
    if (!text) {
      setFallback('');
      return undefined;
    }
    (async () => {
      try {
        const mermaid = await loadMermaid();
        const id = `mmd-${Math.random().toString(36).slice(2, 10)}`;
        const { svg } = await mermaid.render(id, text);
        if (cancelled || !ref.current) return;
        ref.current.innerHTML = svg;
      } catch {
        if (!cancelled) setFallback(text);
      }
    })();
    return () => {
      cancelled = true;
    };
  }, [source]);

  if (fallback != null) {
    return <pre className="chat-mermaid-fallback">{fallback}</pre>;
  }
  const bind = enlargeBind(open, () => {
    const svg = ref.current?.querySelector('svg');
    return svg ? { html: svg.outerHTML } : null;
  });
  return (
    <div
      className={open ? 'chat-mermaid chat-visual-enlarge' : 'chat-mermaid'}
      ref={ref}
      aria-label={open ? 'Enlarge diagram' : undefined}
      {...bind}
    />
  );
}
