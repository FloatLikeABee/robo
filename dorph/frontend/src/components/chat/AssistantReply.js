import React, { useLayoutEffect, useRef, useState } from 'react';
import { replyExceedsLineLimit } from '../../lib/replyEnlarge';

function lineHeightPx(el) {
  const style = getComputedStyle(el);
  const parsed = parseFloat(style.lineHeight);
  if (Number.isFinite(parsed) && parsed > 0) return parsed;
  const font = parseFloat(style.fontSize);
  return Number.isFinite(font) && font > 0 ? font * 1.5 : 24;
}

export default function AssistantReply({ text, modal, children }) {
  const bubbleRef = useRef(null);
  const closeRef = useRef(null);
  const openerRef = useRef(null);
  const [tall, setTall] = useState(false);
  const [open, setOpen] = useState(false);

  useLayoutEffect(() => {
    const el = bubbleRef.current;
    if (!el) return undefined;
    const measure = () => {
      setTall(replyExceedsLineLimit(el.scrollHeight, lineHeightPx(el)));
    };
    measure();
    if (typeof ResizeObserver !== 'function') return undefined;
    const ro = new ResizeObserver(measure);
    ro.observe(el);
    return () => ro.disconnect();
  }, [text]);

  useLayoutEffect(() => {
    if (!open) return undefined;
    closeRef.current?.focus();
    const onKey = (e) => {
      if (e.key !== 'Escape') return;
      setOpen(false);
      openerRef.current?.focus();
    };
    window.addEventListener('keydown', onKey);
    return () => window.removeEventListener('keydown', onKey);
  }, [open]);

  const close = () => {
    setOpen(false);
    openerRef.current?.focus();
  };

  return (
    <>
      <div className="message-bubble assistant-bubble" ref={bubbleRef}>
        {tall ? (
          <button
            ref={openerRef}
            type="button"
            className="reply-enlarge-btn"
            onClick={() => setOpen(true)}
          >
            Enlarge
          </button>
        ) : null}
        {children}
      </div>
      {open ? (
        <div
          className="visual-lightbox-overlay"
          role="presentation"
          onClick={(e) => {
            if (e.target === e.currentTarget) close();
          }}
        >
          <div
            className="visual-lightbox"
            role="dialog"
            aria-modal="true"
            aria-label="Enlarged reply"
            onClick={(e) => e.stopPropagation()}
          >
            <button ref={closeRef} type="button" className="visual-lightbox-close" aria-label="Close" onClick={close}>
              ✕
            </button>
            <div className="reply-lightbox-body">{modal}</div>
          </div>
        </div>
      ) : null}
    </>
  );
}
