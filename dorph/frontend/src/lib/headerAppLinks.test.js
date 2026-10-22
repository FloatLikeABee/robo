import fs from 'fs';
import path from 'path';
import { HEADER_APP_ICONS } from './headerAppLinks';

test('header icons are the public svg paths the server must serve', () => {
  expect(HEADER_APP_ICONS.bk).toMatch(/\/icons\/bk-icon\.svg$/);
  expect(HEADER_APP_ICONS.dorphdata).toMatch(/\/icons\/dorph-data-icon\.svg$/);
  expect(HEADER_APP_ICONS.dorphutils).toBeUndefined();
});

test('Dorph AI never mounts a DorphUtils header chip', () => {
  const chat = fs.readFileSync(path.join(__dirname, '../SkoolAiChat.js'), 'utf8');
  const links = fs.readFileSync(path.join(__dirname, 'headerAppLinks.js'), 'utf8');
  expect(chat).not.toMatch(/DorphUtils/);
  expect(chat).not.toMatch(/dorphUtilsBaseURL/);
  expect(chat).not.toMatch(/REACT_APP_DORPH_UTILS_URL/);
  expect(links).not.toMatch(/dorphUtilsBaseURL/);
  expect(links).not.toMatch(/dorphutils:/);
});
