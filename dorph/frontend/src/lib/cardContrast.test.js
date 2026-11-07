const fs = require('fs');
const path = require('path');

function read(rel) {
  return fs.readFileSync(path.join(__dirname, rel), 'utf8');
}

test('assistant cards are lighter than the page and user bubbles stay deeper', () => {
  const css = read('../App.css');
  const twilight = css.slice(css.indexOf("html[data-chat-theme='twilight']"), css.indexOf("html[data-chat-theme='dark']"));
  expect(twilight).toMatch(/--chat-page-bg:\s*#221c34/);
  expect(twilight).toMatch(/--chat-card:\s*#322a4a/);
  const assistant = css.slice(css.indexOf("html[data-chat-theme='twilight'] .assistant-bubble {"));
  expect(assistant.slice(0, 280)).toMatch(/background:\s*var\(--chat-card\)/);
  const user = css.slice(css.indexOf("html[data-chat-theme='twilight'] .user-bubble {"));
  expect(user.slice(0, 220)).toMatch(/rgba\(124,\s*58,\s*237/);
  expect(user.slice(0, 220)).not.toMatch(/#322a4a/);
});

test('card titles use light ink', () => {
  const css = read('../App.css');
  const titles = css.slice(css.indexOf("html[data-chat-theme='twilight'] .assistant-bubble h1"));
  expect(titles.slice(0, 400)).toMatch(/color:\s*#f4f0fa/);
  expect(titles.slice(0, 400)).toMatch(/font-weight:\s*700/);
});

test('notes, content, tools, and skills cards share the lighter fill', () => {
  expect(read('../theme.js')).toMatch(/paper:\s*isDark \? '#322a4a'/);
  expect(read('../../../../composerx/frontend/src/app.css')).toMatch(/--color-bg-elevated:\s*#322a4a/);
  expect(read('../../../../bk/frontend/src/theme.js')).toMatch(/paper:\s*'#322a4a'/);
  const skills = read('../App.css');
  const modal = skills.slice(skills.indexOf('.skills-modal {'), skills.indexOf('.skills-panel {'));
  expect(modal).toMatch(/var\(--chat-card,\s*#322a4a\)/);
});
