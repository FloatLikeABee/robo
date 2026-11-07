const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const test = require('node:test');

const root = path.join(__dirname, '..');

test('dropped module leftovers are gone and live trees stay', () => {
  assert.equal(fs.existsSync(path.join(root, 'platform-chat')), false);
  assert.equal(
    fs.existsSync(path.join(root, 'bk/frontend/src/pages/GraphicDocumentGenerator.js')),
    false,
  );
  const api = fs.readFileSync(path.join(root, 'bk/src/api.py'), 'utf8');
  assert.match(api, /["']\/graphic-document\/generate["']/);
  assert.equal(fs.existsSync(path.join(root, 'dorph-utils')), true);
  assert.equal(fs.existsSync(path.join(root, 'dorph-engi/README.md')), true);
});
