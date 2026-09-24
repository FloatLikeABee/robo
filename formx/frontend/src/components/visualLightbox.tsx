import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useRef,
  useState,
  type KeyboardEvent,
  type MouseEvent,
  type ReactNode,
} from 'react';

type LightboxItem = { html?: string; src?: string; alt?: string };

function scaleSvgHtml(html: string) {
  try {
    const doc = new DOMParser().parseFromString(String(html || ''), 'image/svg+xml');
    const svg = doc.documentElement;
    if (!svg || svg.tagName.toLowerCase() !== 'svg') return html;
    const w = parseFloat((svg.getAttribute('width') || '').replace(/px$/i, ''));
    const h = parseFloat((svg.getAttribute('height') || '').replace(/px$/i, ''));
    const percent = /%/.test(svg.getAttribute('width') || '') || /%/.test(svg.getAttribute('height') || '');
    if (!svg.getAttribute('viewBox') && !percent && w > 0 && h > 0) {
      svg.setAttribute('viewBox', `0 0 ${w} ${h}`);
    }
    svg.removeAttribute('width');
    svg.removeAttribute('height');
    svg.setAttribute('preserveAspectRatio', 'xMidYMid meet');
    svg.style.width = '100%';
    svg.style.height = '100%';
    svg.style.maxWidth = '100%';
    svg.style.maxHeight = '100%';
    return new XMLSerializer().serializeToString(svg);
  } catch {
    return html;
  }
}

const VisualLightboxContext = createContext<((item: LightboxItem) => void) | null>(null);

export function useVisualLightbox() {
  return useContext(VisualLightboxContext);
}

export function enlargeBind(open: ((item: LightboxItem) => void) | null, getItem: () => LightboxItem | null) {
  if (!open) return {};
  const activate = (e: MouseEvent | KeyboardEvent) => {
    const t = e.target as HTMLElement | null;
    if (t?.closest?.('a')) return;
    const item = getItem();
    if (item) open(item);
  };
  return {
    role: 'button' as const,
    tabIndex: 0,
    onClick: activate,
    onKeyDown: (e: KeyboardEvent) => {
      if (e.key === 'Enter' || e.key === ' ') {
        e.preventDefault();
        activate(e);
      }
    },
  };
}

export function EnlargeImg({ src, alt = '', className = '' }: { src?: string; alt?: string; className?: string }) {
  const open = useVisualLightbox();
  if (!src) return null;
  const bind = enlargeBind(open, () => ({ src, alt }));
  return (
    <img
      src={src}
      alt={alt}
      className={open ? `${className} chat-visual-enlarge`.trim() : className}
      {...bind}
    />
  );
}

export function VisualLightboxProvider({ children }: { children: ReactNode }) {
  const [item, setItem] = useState<LightboxItem | null>(null);
  const closeRef = useRef<HTMLButtonElement>(null);
  const prevFocus = useRef<HTMLElement | null>(null);

  const open = useCallback((next: LightboxItem) => {
    if (!next || (!next.html && !next.src)) return;
    prevFocus.current = document.activeElement as HTMLElement | null;
    setItem(next.html ? { ...next, html: scaleSvgHtml(next.html) } : next);
  }, []);

  const close = useCallback(() => {
    setItem(null);
    const el = prevFocus.current;
    prevFocus.current = null;
    el?.focus?.();
  }, []);

  useEffect(() => {
    if (!item) return undefined;
    closeRef.current?.focus();
    const onKey = (e: globalThis.KeyboardEvent) => {
      if (e.key === 'Escape') close();
    };
    window.addEventListener('keydown', onKey);
    const prev = document.body.style.overflow;
    document.body.style.overflow = 'hidden';
    return () => {
      window.removeEventListener('keydown', onKey);
      document.body.style.overflow = prev;
    };
  }, [item, close]);

  return (
    <VisualLightboxContext.Provider value={open}>
      {children}
      {item ? (
        <div className="visual-lightbox-overlay" role="presentation" onClick={close}>
          <div
            className="visual-lightbox"
            role="dialog"
            aria-modal="true"
            aria-label="Enlarged visual"
            onClick={(e) => e.stopPropagation()}
          >
            <button
              ref={closeRef}
              type="button"
              className="visual-lightbox-close"
              aria-label="Close"
              onClick={close}
            >
              ✕
            </button>
            {item.html ? (
              <div className="visual-lightbox-svg" dangerouslySetInnerHTML={{ __html: item.html }} />
            ) : (
              <img className="visual-lightbox-img" src={item.src} alt={item.alt || 'Enlarged visual'} />
            )}
          </div>
        </div>
      ) : null}
    </VisualLightboxContext.Provider>
  );
}
