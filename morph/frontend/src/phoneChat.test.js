const fs = require('fs');
const path = require('path');

const frontend = path.join(__dirname, '..');

function read(rel) {
  return fs.readFileSync(path.join(frontend, rel), 'utf8');
}

function phoneBlock() {
  const css = read('src/App.css');
  return css.slice(css.lastIndexOf('@media (max-width: 768px)'));
}

test('phone transcript scrolls on the block axis and keeps wide blocks inside the column', () => {
  const phone = phoneBlock();
  const messages = phone.slice(phone.indexOf('.messages-container'));
  expect(messages).toMatch(/overflow-x:\s*clip/);
  expect(messages).toMatch(/overflow-y:\s*auto/);
  expect(phone).toMatch(/\.chat-markdown pre:has\(\.chat-mermaid\)[\s\S]*?max-width:\s*100%/);
  expect(phone).toMatch(/\.chat-markdown pre:has\(\.chat-mermaid\)[\s\S]*?overflow-x:\s*auto/);
  expect(phone).toMatch(/\.welcome-message[\s\S]*?overflow-wrap:\s*anywhere/);
  expect(phone).toMatch(/\.error-bubble[\s\S]*?overflow-wrap:\s*anywhere/);
});

test('phone composer clears side safe areas and uses 44px controls', () => {
  const phone = phoneBlock();
  const composer = phone.slice(phone.indexOf('.input-container'));
  expect(composer).toMatch(/safe-area-inset-left/);
  expect(composer).toMatch(/safe-area-inset-right/);
  expect(composer).toMatch(/flex-wrap:\s*wrap/);
  expect(phone).toMatch(/\.chat-icon-button--input[\s\S]*?min-width:\s*44px/);
  expect(phone).toMatch(/\.chat-icon-button--input[\s\S]*?min-height:\s*44px/);
  expect(phone).toMatch(/\.clear-input-button[\s\S]*?min-height:\s*44px/);
  expect(phone).toMatch(/\.send-button[\s\S]*?min-height:\s*44px/);
});

test('phone shell subtracts the keyboard inset and hides the workspace while it is open', () => {
  const phone = phoneBlock();
  expect(phone).toMatch(/height:\s*calc\(100dvh - var\(--keyboard-inset, 0px\)\)/);
  const shell = phone.slice(phone.indexOf('.app-outer'));
  expect(shell).toMatch(/min-height:\s*0/);
  expect(phone).toMatch(/html\[data-keyboard-open\] \.app\.app--agent \.agent-workspace\s*\{[^}]*display:\s*none/);
  expect(read('public/index.html')).toMatch(/interactive-widget=resizes-content/);
  expect(read('src/SkoolAiChat.js')).toMatch(/bindKeyboardInset/);
});
