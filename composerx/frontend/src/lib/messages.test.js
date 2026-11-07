import assert from 'node:assert/strict';
import test from 'node:test';
import { en, zh } from './messages.js';

test('publish chrome is named Content', () => {
  assert.equal(en.contentMaker, 'Content');
  assert.equal(zh.contentMaker, '内容');
  assert.equal(en.signInSub.includes('Content Maker'), false);
});

test('every english catalog key has simplified chinese', () => {
  assert.deepEqual(Object.keys(zh).sort(), Object.keys(en).sort());
  for (const key of Object.keys(en)) {
    assert.ok(en[key].trim());
    assert.ok(zh[key].trim());
  }
});
