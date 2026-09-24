const fs = require('fs');
const path = require('path');

function read(rel) {
  return fs.readFileSync(path.join(__dirname, rel), 'utf8');
}

function lastPhone(css) {
  return css.slice(css.lastIndexOf('@media (max-width: 768px)'));
}

test('phone body copy is 16px with 1.5 line-height before and after CssBaseline', () => {
  const indexPhone = lastPhone(read('index.css'));
  expect(indexPhone).toMatch(/font-size:\s*16px/);
  expect(indexPhone).toMatch(/line-height:\s*1\.5/);

  const theme = read('theme.js');
  expect(theme).toMatch(/MuiCssBaseline:[\s\S]*@media \(max-width:768px\)[\s\S]*fontSize: '16px'/);
  expect(theme).toMatch(/MuiCssBaseline:[\s\S]*@media \(max-width:768px\)[\s\S]*lineHeight: 1\.5/);
  expect(theme).toMatch(/body2:[\s\S]*@media \(max-width:768px\)[\s\S]*fontSize: '1rem'/);
});

test('phone chat, welcome, and skills empty copy is 16px and the title ellipsizes', () => {
  const phone = lastPhone(read('App.css'));
  const copy = phone.slice(phone.indexOf('.message-bubble,'), phone.indexOf('.skills-panel-empty'));
  expect(copy).toMatch(/font-size:\s*16px/);
  expect(copy).toMatch(/line-height:\s*1\.5/);
  expect(phone).toMatch(/\.welcome-message\s*\{[^}]*max-width:\s*100%/);
  expect(phone).toMatch(/\.welcome-message\s*\{[^}]*overflow-wrap:\s*anywhere/);
  const empty = phone.slice(phone.indexOf('.skills-panel-empty'));
  expect(empty).toMatch(/\.skills-panel-empty,\s*\.skills-picker-empty\s*\{[^}]*font-size:\s*1rem/);
  expect(empty).toMatch(/\.skills-panel-empty,\s*\.skills-picker-empty\s*\{[^}]*line-height:\s*1\.5/);
  const title = phone.slice(phone.indexOf('.chat-header-title-stack .app-title'));
  expect(title).toMatch(/width:\s*100%/);
  expect(title).toMatch(/text-overflow:\s*ellipsis/);
  expect(phone).toMatch(/\.chat-header-title-stack\s*\{[^}]*align-items:\s*stretch/);
});

test('phone error dismiss controls are at least 44px', () => {
  const phone = lastPhone(read('App.css'));
  expect(phone).toMatch(/\.error-bubble\s*\{[^}]*font-size:\s*16px/);
  expect(phone).toMatch(/\.error-bubble\s*\{[^}]*line-height:\s*1\.5/);
  expect(phone).toMatch(/\.error-bubble-dismiss\s*\{[^}]*min-width:\s*44px/);
  expect(phone).toMatch(/\.error-bubble-dismiss\s*\{[^}]*min-height:\s*44px/);

  const theme = read('theme.js');
  expect(theme).toMatch(/MuiAlert:[\s\S]*minHeight: 44/);
  expect(theme).toMatch(/MuiSnackbar:[\s\S]*safe-area-inset-bottom/);
  const button = theme.slice(theme.indexOf('MuiButton:'), theme.indexOf('MuiAlert:'));
  const root = button.slice(button.indexOf('root:'), button.indexOf('containedPrimary:'));
  expect(root).toMatch(/@media \(max-width:768px\)/);
  expect(root).toMatch(/minHeight: 44/);
  expect(theme).toMatch(/sizeSmall:[\s\S]*minHeight: 44/);
});

test('sign-in labels and errors are body sized and the error can be dismissed', () => {
  const page = read('pages/LoginPage.js');
  const css = read('pages/LoginPage.css');
  expect(page).toMatch(/className="login-shell"/);
  expect(page).toMatch(/authKeyboardOverlap/);
  expect(page).toMatch(/role="alert"/);
  expect(page).toMatch(/aria-label="Dismiss error"/);
  expect(page).toMatch(/className="login-error-dismiss"/);
  expect(css).toMatch(/\.login-brand p\s*\{[^}]*font-size:\s*16px/);
  expect(css).toMatch(/\.login-brand p\s*\{[^}]*line-height:\s*1\.5/);
  expect(css).toMatch(/\.login-field\s*\{[^}]*font-size:\s*16px/);
  expect(css).toMatch(/\.login-field\s*\{[^}]*line-height:\s*1\.5/);
  expect(css).toMatch(/\.login-error\s*\{[^}]*font-size:\s*16px/);
  expect(css).toMatch(/\.login-error\s*\{[^}]*line-height:\s*1\.5/);
  expect(css).toMatch(/\.login-error-dismiss\s*\{[^}]*min-width:\s*44px/);
  expect(css).toMatch(/\.login-error-dismiss\s*\{[^}]*min-height:\s*44px/);
});

test('chat and notes error banners expose a dismiss control', () => {
  const chat = read('SkoolAiChat.js');
  expect(chat).toMatch(/className="error-bubble-dismiss"/);
  expect(chat).toMatch(/aria-label="Dismiss error"/);
  expect(chat).toMatch(/bindKeyboardInset/);

  for (const rel of ['pages/admin/CaseTasks.js', 'pages/admin/StoryBoard.js']) {
    const tags = read(rel).match(/<Alert severity="error"[\s\S]*?>/g) || [];
    expect(tags.length).toBeGreaterThan(0);
    for (const tag of tags) {
      expect(tag).toMatch(/onClose=/);
    }
  }
});
