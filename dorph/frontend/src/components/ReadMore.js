import React, { useState } from 'react';

export function cardNeedsReadMore(text, limit = 180) {
  return String(text || '').trim().length > limit;
}

export default function ReadMore({ text, limit = 180 }) {
  const [open, setOpen] = useState(false);
  const body = String(text || '');
  const long = cardNeedsReadMore(body, limit);
  const shown = !long || open ? body : `${body.slice(0, limit).trimEnd()}…`;
  return (
    <>
      <span style={{ whiteSpace: 'pre-wrap' }}>{shown}</span>
      {long ? (
        <button type="button" onClick={() => setOpen((value) => !value)} style={{ marginLeft: 8 }}>
          {open ? 'Show less' : 'Read more'}
        </button>
      ) : null}
    </>
  );
}
