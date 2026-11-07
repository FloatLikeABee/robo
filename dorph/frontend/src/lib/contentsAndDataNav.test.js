import fs from 'fs';
import path from 'path';
import { en, zh } from './messages';

function read(rel) {
  return fs.readFileSync(path.join(__dirname, rel), 'utf8');
}

test('drawer is one Contents & Data item without a Generic data row', () => {
  const drawer = read('../components/admin/AppDrawer.js');
  expect(drawer).toContain("t('contentMaker')");
  expect(drawer).toContain("'/contents-data'");
  expect(drawer).not.toMatch(/t\('genericData'\)/);
  expect(drawer).not.toMatch(/base \+ '\/generic-data'\)\s*\n\s*onClick/);
});

test('labels rename Content Maker to Contents & Data', () => {
  expect(en.contentMaker).toBe('Contents & Data');
  expect(zh.contentMaker).toBe('内容与数据');
});

test('old generic-data and content-maker routes redirect into contents-data', () => {
  const router = read('../appRouter.js');
  expect(router).toMatch(/index: true, element: <Navigate to="contents-data" replace \/>/);
  expect(router).toMatch(/path: 'contents-data', element: <ContentsAndData \/>/);
  expect(router).toMatch(
    /path: 'generic-data', element: <Navigate to=\{`\$\{ADMIN_BASE_PATH\}\/contents-data`\} replace \/>/
  );
  expect(router).toMatch(
    /path: 'content-maker', element: <Navigate to=\{`\$\{ADMIN_BASE_PATH\}\/contents-data\?tab=publish`\} replace \/>/
  );
  expect(router).not.toMatch(/path: 'generic-data', element: <GenericData/);
});

test('header title uses light ink not clipped gradient fill', () => {
  const css = read('../App.css');
  const start = css.indexOf('\n.app-title {');
  const end = css.indexOf('\n.app-title::before {');
  const title = css.slice(start, end);
  expect(title).toMatch(/color:\s*var\(--chat-text\)/);
  expect(title).not.toMatch(/-webkit-text-fill-color:\s*transparent/);
  expect(title).not.toMatch(/background-clip:\s*text/);
});

test('generic data publishes through composerx publishes', () => {
  const page = read('../pages/admin/GenericData.js');
  expect(page).toContain("tranApi.post('/api/composerx/publishes'");
  expect(page).toContain('buildPublishPayload');
  expect(page).toContain('Publish as HTML');
});
