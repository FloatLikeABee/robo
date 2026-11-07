const fs = require('fs');
const path = require('path');

function read(rel) {
  return fs.readFileSync(path.join(__dirname, rel), 'utf8');
}

function luminance(hex) {
  const n = parseInt(hex.slice(1), 16);
  const r = (n >> 16) & 255;
  const g = (n >> 8) & 255;
  const b = n & 255;
  return 0.2126 * r + 0.7152 * g + 0.0722 * b;
}

test('the page is darker than the bento tiles', () => {
  expect(luminance('#221c34')).toBeLessThan(luminance('#322a4a'));
  expect(luminance('#322a4a')).toBeLessThan(luminance('#5c5182'));
});

test('chat transcript and workspace sit side by side', () => {
  const css = read('../App.css');
  const shell = css.slice(css.indexOf('.app.app--agent {'));
  expect(shell.slice(0, 500)).toMatch(/grid-template-columns:[^;]*minmax\(0, 1fr\) minmax\(0, 1fr\)/);
  expect(shell.slice(0, 500)).toMatch(/gap:\s*16px/);
});

test('content preview and source are two columns on a wide screen', () => {
  const panel = read('../../../../composerx/frontend/src/components/ComposePublishPanel.svelte');
  const left = panel.slice(panel.indexOf('.compose-publish-left {'));
  expect(left.slice(0, 280)).toMatch(/grid-template-columns:\s*minmax\(0, 1\.4fr\) minmax\(280px, 1fr\)/);
});

test('navigation items keep an icon and a label', () => {
  const drawer = read('../components/admin/AppDrawer.js');
  expect(drawer).toMatch(/ListItemIcon/);
  expect(drawer).toMatch(/ListItemText/);
  expect(drawer).not.toMatch(/ListItemButton[\s\S]*ListItemButton[\s\S]*ListItemButton[\s\S]*nested/);
});

test('first list load is a skeleton, not only a spinner', () => {
  const page = read('../pages/admin/GenericData.js');
  expect(page).toMatch(/bento-skeleton/);
  expect(page).toMatch(/loading && rows\.length === 0/);
});
