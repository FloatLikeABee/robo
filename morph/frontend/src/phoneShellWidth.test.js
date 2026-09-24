const fs = require('fs');
const path = require('path');

const css = fs.readFileSync(path.join(__dirname, 'App.css'), 'utf8');

function braceBlock(source, openIndex) {
  let depth = 0;
  for (let i = openIndex; i < source.length; i += 1) {
    if (source[i] === '{') depth += 1;
    else if (source[i] === '}') {
      depth -= 1;
      if (depth === 0) return source.slice(openIndex + 1, i);
    }
  }
  throw new Error('unbalanced braces');
}

function ruleBodies(source, selector) {
  const bodies = [];
  let from = 0;
  while (from < source.length) {
    const start = source.indexOf(selector, from);
    if (start < 0) break;
    const brace = source.indexOf('{', start);
    const between = source.slice(start + selector.length, brace);
    if (/^\s*$/.test(between)) bodies.push(braceBlock(source, brace));
    from = start + selector.length;
  }
  return bodies;
}

function mediaBodies(source, header) {
  const bodies = [];
  let from = 0;
  while (from < source.length) {
    const start = source.indexOf(header, from);
    if (start < 0) break;
    const brace = source.indexOf('{', start);
    bodies.push(braceBlock(source, brace));
    from = brace + 1;
  }
  return bodies;
}

test('phone collapsed workspace uses one flexible column', () => {
  const unscoped = ruleBodies(css, '.app.app--agent.app--workspace-collapsed')[0];
  expect(unscoped).toMatch(/grid-template-columns:\s*var\(--agent-sessions\)/);

  const phone = mediaBodies(css, '@media (max-width: 768px)').join('\n');
  const collapsed = ruleBodies(phone, '.app.app--agent.app--workspace-collapsed');
  const oneColumn = collapsed.find((body) => /grid-template-areas:[\s\S]*'head'\s*'chat'/.test(body));
  expect(oneColumn).toBeTruthy();
  expect(oneColumn).toMatch(/grid-template-columns:\s*minmax\(0,\s*1fr\)/);
});

test('phone hides composer Notes and Knowledge chips', () => {
  const base = ruleBodies(css, '.agent-include-bar')[0];
  expect(base).toMatch(/display:\s*flex/);
  const phone = mediaBodies(css, '@media (max-width: 768px)').join('\n');
  const hidden = ruleBodies(phone, '.agent-include-bar');
  expect(hidden.some((body) => /display:\s*none/.test(body))).toBe(true);
});
