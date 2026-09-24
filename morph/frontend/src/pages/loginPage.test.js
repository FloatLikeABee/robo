const fs = require('fs');
const path = require('path');

const frontend = path.join(__dirname, '..', '..');

function read(rel) {
  return fs.readFileSync(path.join(frontend, rel), 'utf8');
}

test('login stylesheet top-aligns the phone card and gives Sign in a 44px target', () => {
  const cssPath = path.join(frontend, 'src/pages/LoginPage.css');
  expect(fs.existsSync(cssPath)).toBe(true);
  const css = fs.readFileSync(cssPath, 'utf8');
  const phone = css.slice(css.indexOf('@media (max-width: 480px)'));
  expect(phone).toMatch(/align-content:\s*start/);
  expect(css).toMatch(/min-height:\s*100dvh/);
  expect(css).toMatch(/font-size:\s*16px/);
  expect(css).toMatch(/min-height:\s*44px/);
  expect(css).toMatch(/min-width:\s*44px/);
  expect(css).toMatch(/overflow-wrap:\s*anywhere/);
  expect(css).toMatch(/safe-area-inset-top/);
  expect(css).toMatch(/safe-area-inset-bottom/);
  expect(css).toMatch(/--auth-keyboard-inset/);
});

test('login page alerts errors and does not change the shared viewport meta', () => {
  const page = read('src/pages/LoginPage.js');
  expect(page).toMatch(/role=["']alert["']/);
  expect(page).toMatch(/authKeyboardOverlap/);
  expect(page).toMatch(/name=["']username["']/);
  expect(page).toMatch(/autoComplete=["']current-password["']/);
  const html = read('public/index.html');
  const meta = html.match(/<meta\s+name="viewport"\s+content="([^"]+)"/);
  expect(meta[1]).toContain('viewport-fit=cover');
  expect(meta[1]).toContain('interactive-widget=resizes-content');
  // The shared meta sets interactive-widget for the chat composer. Login must not write it.
  expect(page).not.toMatch(/interactive-widget/);
});
