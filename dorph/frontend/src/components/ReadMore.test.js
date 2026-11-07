import React from 'react';
import { renderToStaticMarkup } from 'react-dom/server';
import ReadMore, { cardNeedsReadMore } from './ReadMore';

test('a long card shows Read more and a short one does not', () => {
  const long = 'word '.repeat(80);
  expect(cardNeedsReadMore(long)).toBe(true);
  expect(renderToStaticMarkup(<ReadMore text={long} />)).toContain('Read more');
  expect(renderToStaticMarkup(<ReadMore text="Short note." />)).not.toContain('Read more');
});
