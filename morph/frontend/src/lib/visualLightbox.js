import { createContext, useCallback, useContext, useEffect, useRef, useState } from 'react';

const VisualLightboxContext = createContext(null);

export function useVisualLightbox() {
  return useContext(VisualLightboxContext);
}

/** Drop mermaid’s bubble width/height so the clone can fill the overlay. */
export function scaleSvgHtml(html) {
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

export function enlargeBind(open, getItem) {
  if (!open) return {};
  const activate = (e) => {
    if (e.target.closest?.('a')) return;
    const item = getItem(e);
    if (item) open(item);
  };
  return {
    role: 'button',
    tabIndex: 0,
    onClick: activate,
    onKeyDown: (e) => {
      if (e.key === 'Enter' || e.key === ' ') {
        e.preventDefault();
        activate(e);
      }
    },
  };
}

export function EnlargeImg({ src, alt = '', className = '' }) {
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

export function VisualLightboxProvider({ children }) {
  const [item, setItem] = useState(null);
  const closeRef = useRef(null);
  const prevFocus = useRef(null);

  const open = useCallback((next) => {
    if (!next || (!next.html && !next.src)) return;
    prevFocus.current = document.activeElement;
    setItem(next.html ? { ...next, html: scaleSvgHtml(next.html) } : next);
  }, []);

  const close = useCallback(() => {
    setItem(null);
    const el = prevFocus.current;
    prevFocus.current = null;
    if (el && typeof el.focus === 'function') el.focus();
  }, []);

  useEffect(() => {
    if (!item) return undefined;
    closeRef.current?.focus();
    const onKey = (e) => {
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
