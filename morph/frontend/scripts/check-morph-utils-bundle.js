#!/usr/bin/env node
// Fails unless a CRA production bundle inlined REACT_APP_MORPH_UTILS_URL.
// ponytail: scans build/static/js/*.js as text. Upgrade path is a webpack
// stats assertion if the bundle is split into chunks that this concat misses.
const fs = require('fs');
const path = require('path');

const mode = process.argv[2];
const url = process.argv[3];
const dir = path.join(__dirname, '..', 'build', 'static', 'js');
const envRead = 'process.env.REACT_APP_MORPH_UTILS_URL';

function fail(message) {
  console.error(message);
  process.exit(1);
}

if (mode !== 'unset' && mode !== 'set') {
  fail('usage: check-morph-utils-bundle.js unset|set <url>');
}
if (!fs.existsSync(dir)) {
  fail('missing ' + dir);
}
const files = fs.readdirSync(dir).filter((name) => name.endsWith('.js') && !name.endsWith('.js.map'));
if (!files.length) fail('no js bundles in ' + dir);
const text = files.map((name) => fs.readFileSync(path.join(dir, name), 'utf8')).join('\n');
if (text.includes(envRead)) fail('bundle still reads ' + envRead);

if (mode === 'unset') {
  for (const forbidden of ['https://utils.example.com', 'https://morph-utils.onrender.com']) {
    if (text.includes(forbidden)) fail('unset bundle contains ' + forbidden);
  }
  if (!text.includes('http://localhost:3040')) fail('unset bundle missing the dev default');
  console.log('morph utils bundle unset ok');
  process.exit(0);
}

if (!url || url.includes('localhost') || url.includes('127.0.0.1') || url.includes('[::1]')) {
  fail('set url must be a non-loopback URL');
}
if (!text.includes(url)) fail('set bundle missing ' + url);
console.log('morph utils bundle set ok');
