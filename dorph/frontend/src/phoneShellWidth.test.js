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

test('phone agent shell is one full-width column with the workspace band', () => {
  const phone = mediaBodies(css, '@media (max-width: 768px)').join('\n');
  const open = ruleBodies(phone, '.app.app--agent');
  const column = open.find((body) => /grid-template-areas:[\s\S]*'workspace'/.test(body));
  expect(column).toBeTruthy();
  expect(column).toMatch(/grid-template-columns:\s*minmax\(0,\s*1fr\)/);
  expect(column).toMatch(/'head'/);
  expect(column).toMatch(/'chat'/);
});

test('composer tool panel is two rows of square buttons beside the field', () => {
  const grid = ruleBodies(css, '.composer-tools--six')[0];
  expect(grid).toMatch(/grid-template-columns:\s*repeat\(3,/);
  expect(grid).toMatch(/grid-template-rows:\s*repeat\(2,/);
  const field = ruleBodies(css, '.composer-field')[0];
  expect(field).toMatch(/flex:\s*1/);
  const send = ruleBodies(css, '.composer-field .send-button')[0];
  expect(send).toMatch(/align-self:\s*stretch/);
  const phone = mediaBodies(css, '@media (max-width: 768px)').join('\n');
  const hiddenTools = ruleBodies(phone, '.composer-tools');
  expect(hiddenTools.some((body) => /display:\s*none/.test(body))).toBe(false);
});
