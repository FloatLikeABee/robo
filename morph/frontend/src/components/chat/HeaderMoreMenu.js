import React, { useEffect, useState } from 'react';

function Chip({ item, inMenu, onDone }) {
  const style = item.color ? { '--header-app-color': item.color } : undefined;
  const className = `header-app-link${item.href ? '' : ' header-app-link--button'}`;
  const content = (
    <>
      {item.icon ? <img src={item.icon} alt="" className="header-app-link-icon" aria-hidden /> : null}
      <span className="header-app-link-label">{item.label}</span>
    </>
  );
  if (item.href) {
    return (
      <a
        href={item.href}
        target="_blank"
        rel="noopener noreferrer"
        className={className}
        style={style}
        role={inMenu ? 'menuitem' : undefined}
        title={item.label}
        onClick={onDone}
      >
        {content}
      </a>
    );
  }
  return (
    <button
      type="button"
      className={className}
      style={style}
      title={item.label}
      role={inMenu ? 'menuitem' : undefined}
      aria-haspopup={inMenu ? undefined : item.hasPopup}
      aria-expanded={inMenu ? undefined : item.expanded}
      onClick={() => {
        item.onClick?.();
        onDone?.();
      }}
    >
      {content}
    </button>
  );
}

export default function HeaderMoreMenu({ items = [], onClear }) {
  const [open, setOpen] = useState(false);

  useEffect(() => {
    if (!open) return undefined;
    const onKey = (event) => {
      if (event.key === 'Escape') setOpen(false);
    };
    document.addEventListener('keydown', onKey);
    return () => document.removeEventListener('keydown', onKey);
  }, [open]);

  return (
    <>
      <div className="header-app-links header-app-links--bar" aria-label="App links">
        {items.map((item) => (
          <Chip key={item.id} item={item} />
        ))}
      </div>
      <div className="header-more">
        <button
          type="button"
          className="chat-icon-button header-more-button"
          aria-label="More apps"
          aria-expanded={open}
          aria-haspopup="menu"
          onClick={() => setOpen((value) => !value)}
        >
          <span aria-hidden="true">⋯</span>
        </button>
        {open ? (
          <>
            <button
              type="button"
              className="header-more-backdrop"
              aria-label="Close more apps"
              onClick={() => setOpen(false)}
            />
            <div className="header-more-panel" role="menu">
              {items.map((item) => (
                <Chip key={item.id} item={item} inMenu onDone={() => setOpen(false)} />
              ))}
              {typeof onClear === 'function' ? (
                <button
                  type="button"
                  role="menuitem"
                  className="header-more-item header-app-link header-app-link--button"
                  onClick={() => {
                    setOpen(false);
                    onClear();
                  }}
                >
                  Clear chat
                </button>
              ) : null}
            </div>
          </>
        ) : null}
      </div>
    </>
  );
}
