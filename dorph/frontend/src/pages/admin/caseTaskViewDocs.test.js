import { buildCaseTaskHTML } from './caseTaskViewDocs';

test('mermaid fence keeps language class for HTML preview', () => {
  const html = buildCaseTaskHTML({
    title: 'Onboard',
    markdown: '# Onboard\n\n```mermaid\nflowchart LR\n  A-->B\n```\n',
  });
  expect(html).toContain('language-mermaid');
  expect(html).toContain('flowchart LR');
  expect(html).toContain('Arial');
  expect(html).toContain('#a78bfa');
  expect(html).toContain('#34d399');
  expect(html).not.toContain('Georgia');
});
