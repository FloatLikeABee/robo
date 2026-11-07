import { buildPublishHtml, buildPublishPayload, PUBLISH_EXCERPT_MAX_ROWS } from './contentsDataPublish';

const record = {
  title: 'Q1 <sales>',
  ai_analysis: 'Revenue rose.\n\nWatch returns.',
  detail: {
    columns: ['name', 'n'],
    rows: [{ name: '<x>', n: '1' }, { name: 'b', n: '2' }],
  },
};

test('publish payload is name, dark theme, and html_content', () => {
  const payload = buildPublishPayload(record);
  expect(payload.name).toBe('Q1 <sales>');
  expect(payload.theme).toBe('dark');
  expect(payload.html_content).toContain('background:#221c34');
  expect(payload.html_content).toContain('#322a4a');
  expect(payload.html_content).toContain('#8b5cf6');
  expect(payload.html_content).toContain('#a78bfa');
  expect(payload.html_content).toContain('#f97316');
  expect(payload.html_content).toContain('<h1>Q1 &lt;sales&gt;</h1>');
  expect(payload.html_content).toContain('Revenue rose.');
  expect(payload.html_content).toContain('Watch returns.');
  expect(payload.html_content).toContain('<td>&lt;x&gt;</td>');
  expect(payload.html_content).not.toContain('<td><x></td>');
});

test('caps table excerpt and notes omitted rows', () => {
  const rows = Array.from({ length: PUBLISH_EXCERPT_MAX_ROWS + 3 }, (_, i) => ({ a: String(i) }));
  const html = buildPublishHtml({
    title: 'Wide',
    ai_analysis: 'ok',
    detail: { columns: ['a'], rows },
  });
  expect(html).toContain(`${PUBLISH_EXCERPT_MAX_ROWS + 3 - PUBLISH_EXCERPT_MAX_ROWS} more rows omitted`);
  expect((html.match(/<tr>/g) || []).length).toBe(PUBLISH_EXCERPT_MAX_ROWS + 1);
});

test('parses detail JSON strings', () => {
  const html = buildPublishHtml({
    title: 'CSV',
    ai_analysis: 'fine',
    detail: JSON.stringify({ columns: ['a'], rows: [{ a: 'z' }] }),
  });
  expect(html).toContain('<td>z</td>');
});
