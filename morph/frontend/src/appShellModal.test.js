const fs = require('fs');
const path = require('path');

const css = fs.readFileSync(path.join(__dirname, 'App.css'), 'utf8');
const aiTools = fs.readFileSync(path.join(__dirname, 'components/chat/AiToolsWorkspaceDrawer.js'), 'utf8');
const notes = fs.readFileSync(path.join(__dirname, 'components/chat/MorphNotesModal.js'), 'utf8');

function ruleBody(source, selector) {
  const start = source.indexOf(selector);
  if (start < 0) return '';
  const brace = source.indexOf('{', start);
  let depth = 0;
  for (let i = brace; i < source.length; i += 1) {
    if (source[i] === '{') depth += 1;
    else if (source[i] === '}') {
      depth -= 1;
      if (depth === 0) return source.slice(brace + 1, i);
    }
  }
  return '';
}

test('shell modals share 96vw by 96dvh and are not capped at 1200px', () => {
  const body = ruleBody(css, '.hybrid-drawer.app-shell-modal');
  expect(body).toMatch(/width:\s*96vw/);
  expect(body).toMatch(/height:\s*96dvh/);
  expect(body).not.toMatch(/1200px/);
  expect(css).not.toMatch(/min\(96vw,\s*1200px\)/);
  expect(aiTools).toMatch(/app-shell-modal--full/);
  expect(aiTools).not.toMatch(/1200px/);
  expect(notes).toMatch(/app-shell-modal/);
  const full = ruleBody(css, '.hybrid-drawer.app-shell-modal.app-shell-modal--full');
  expect(full).toMatch(/height:\s*100dvh/);
  expect(notes).toMatch(/app-shell-modal--full/);
});
