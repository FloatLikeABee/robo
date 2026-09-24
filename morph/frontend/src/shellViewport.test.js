const fs = require('fs');
const path = require('path');

const frontend = path.join(__dirname, '..');

function read(rel) {
  return fs.readFileSync(path.join(frontend, rel), 'utf8');
}

function mediaBlock(css, header) {
  const start = css.indexOf(header);
  if (start < 0) throw new Error(`missing ${header}`);
  const next = css.indexOf('@media', start + header.length);
  return css.slice(start, next === -1 ? css.length : next);
}

test('viewport meta fits the device and covers the safe area', () => {
  const html = read('public/index.html');
  const meta = html.match(/<meta\s+name="viewport"\s+content="([^"]+)"/);
  expect(meta).not.toBeNull();
  expect(meta[1]).toContain('width=device-width');
  expect(meta[1]).toContain('viewport-fit=cover');
});

test('document shell clips horizontal overflow on first paint', () => {
  const css = read('src/index.css');
  expect(css).toMatch(/overflow-x:\s*clip/);
  expect(css).toMatch(/max-width:\s*100%/);
});

test('phone header actions wrap instead of scrolling sideways', () => {
  const phone = mediaBlock(read('src/App.css'), '@media (max-width: 768px)');
  const actionsStart = phone.indexOf('.header-actions');
  expect(actionsStart).toBeGreaterThan(-1);
  const actions = phone.slice(actionsStart, phone.indexOf('.header-app-links'));
  expect(actions).not.toMatch(/overflow-x:\s*auto/);
  expect(actions).toMatch(/flex-wrap:\s*wrap/);
  expect(actions).toMatch(/min-width:\s*0/);
});

test('skills page padding uses safe-area insets', () => {
  const css = read('src/App.css');
  const start = css.indexOf('.skills-panel-page {');
  expect(start).toBeGreaterThan(-1);
  const block = css.slice(start, css.indexOf('}', start));
  expect(block).toMatch(/safe-area-inset-top/);
  expect(block).toMatch(/safe-area-inset-bottom/);
});

test('MorphNotes home-indicator spacer is CSS, not a media-query hook', () => {
  const layout = read('src/AdminLayout.js');
  expect(layout).not.toMatch(/useMediaQuery/);
  expect(layout).toMatch(/calc\(56px \+ env\(safe-area-inset-top\)\)/);
  expect(layout).toMatch(/xs:\s*'block'/);
  expect(layout).toMatch(/sm:\s*'none'/);
  expect(layout).toMatch(/env\(safe-area-inset-bottom\)/);
});

test('phone drawers use the layout viewport instead of 100vw', () => {
  for (const rel of ['src/pages/admin/CaseTasks.js', 'src/pages/admin/StoryBoard.js']) {
    expect(read(rel)).not.toMatch(/100vw/);
  }
});
