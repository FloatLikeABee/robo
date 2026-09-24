import { useEffect, useRef, useState, type ReactNode } from 'react';
import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';
import { EnlargeImg, enlargeBind, useVisualLightbox } from './visualLightbox';

let mermaidReady: Promise<typeof import('mermaid').default> | null = null;

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

function MermaidBlock({ source }: { source: string }) {
  const ref = useRef<HTMLDivElement>(null);
  const [fallback, setFallback] = useState<string | null>(null);
  const open = useVisualLightbox();
  useEffect(() => {
    let cancelled = false;
    setFallback(null);
    const text = String(source || '').trim();
    if (!text) {
      setFallback('');
      return undefined;
    }
    void (async () => {
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
  if (fallback != null) return <pre>{fallback}</pre>;
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

function fenceLang(className?: string) {
  const m = /language-([^\s]+)/.exec(className || '');
  return (m?.[1] || '').toLowerCase();
}

function CodeBlock({ className, children, ...props }: { className?: string; children?: ReactNode }) {
  const lang = fenceLang(className);
  const text = String(children || '').replace(/\n$/, '');
  if (lang === 'mermaid') return <MermaidBlock source={text} />;
  return (
    <code className={className} {...props}>
      {children}
    </code>
  );
}

export default function VisualMarkdown({ text }: { text: string }) {
  return (
    <ReactMarkdown
      remarkPlugins={[remarkGfm]}
      components={{
        a: ({ href, children, ...props }) => (
          <a href={href} target="_blank" rel="noopener noreferrer" {...props}>
            {children}
          </a>
        ),
        img: ({ src, alt }) => <EnlargeImg src={src} alt={alt || ''} />,
        code: CodeBlock,
      }}
    >
      {text || ''}
    </ReactMarkdown>
  );
}
