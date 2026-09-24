import { scaleSvgHtml } from './visualLightbox';

test('scaleSvgHtml drops pixel size so mermaid can fill the overlay', () => {
  const html =
    '<svg xmlns="http://www.w3.org/2000/svg" width="120" height="80" viewBox="0 0 120 80"><rect width="120" height="80"/></svg>';
  const out = scaleSvgHtml(html);
  const open = out.match(/<svg[^>]*>/)[0];
  expect(open).not.toMatch(/\swidth=/);
  expect(open).not.toMatch(/\sheight=/);
  expect(open).toMatch(/viewBox="0 0 120 80"/);
  expect(open).toMatch(/preserveAspectRatio="xMidYMid meet"/);
});
