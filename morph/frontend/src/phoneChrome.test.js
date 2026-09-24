const fs = require('fs');
const path = require('path');

function read(rel) {
  return fs.readFileSync(path.join(__dirname, rel), 'utf8');
}

test('phone header overflow does not scroll and targets are 44px', () => {
  const css = read('App.css');
  const phone = css.slice(css.lastIndexOf('@media (max-width: 768px)'));
  expect(phone).toMatch(/\.header-app-links--bar[\s\S]*?display:\s*none/);
  expect(phone).toMatch(/\.header-actions\s*\{[^}]*overflow:\s*visible/);
  expect(phone).toMatch(/\.chat-header \.chat-icon-button[\s\S]*?min-width:\s*44px/);
  expect(phone).toMatch(/\.chat-header \.chat-icon-button[\s\S]*?min-height:\s*44px/);
  expect(phone).toMatch(/\.header-more-panel \.header-app-link-label\s*\{[^}]*display:\s*inline/);
  expect(phone).toMatch(/\.header-more-panel \.header-app-link[\s\S]*?min-height:\s*44px/);
});

test('phone drawers are full-width sheets with a 44px close control', () => {
  const css = read('App.css');
  const phone = css.slice(css.lastIndexOf('@media (max-width: 768px)'));
  expect(phone).toMatch(/\.app\.app--agent \.chat-sidebar[\s\S]*?width:\s*100%/);
  expect(phone).not.toMatch(/88vw/);
  expect(phone).toMatch(/\.sidebar-sheet-close[\s\S]*?min-width:\s*44px/);
  expect(phone).toMatch(/\.hybrid-drawer-close[\s\S]*?min-height:\s*44px/);
  expect(phone).toMatch(/\.sidebar-session-main[\s\S]*?min-height:\s*44px/);
  expect(phone).toMatch(/overflow-x:\s*clip/);

  const drawer = read('components/admin/AppDrawer.js');
  expect(drawer).not.toMatch(/86vw/);
  expect(drawer).toMatch(/width:\s*'100%'/);
  expect(drawer).toMatch(/minHeight:\s*44/);

  expect(read('pages/admin/CaseTasks.js')).not.toMatch(/100vw/);
  expect(read('pages/admin/StoryBoard.js')).not.toMatch(/100vw/);
});

test('sessions sheet is a dialog that uses the focus helper', () => {
  const chat = read('SkoolAiChat.js');
  expect(chat).toMatch(/role=\{sidebarNavOpen \? 'dialog' : undefined\}/);
  expect(chat).toMatch(/sidebar-sheet-close/);
  expect(chat).toMatch(/onSheetKeyDown/);
  expect(chat).toMatch(/HeaderMoreMenu/);
});
