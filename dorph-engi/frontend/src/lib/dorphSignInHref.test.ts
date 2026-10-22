import assert from 'node:assert/strict';
import { test } from 'node:test';
import { missingDorphSessionReason, dorphSignInHref } from './dorphSignInHref.ts';

test('production Project sign-in is not localhost when unset', () => {
  assert.equal(dorphSignInHref('', false), '');
  assert.equal(dorphSignInHref('http://127.0.0.1:3031', false), '');
  assert.equal(missingDorphSessionReason(false).includes('localhost'), false);
});

test('configured Project sign-in uses the Dorph origin', () => {
  assert.equal(dorphSignInHref('https://dorph.example/', false), 'https://dorph.example');
});

test('dev Project sign-in stays on localhost when unset', () => {
  assert.equal(dorphSignInHref('', true), 'http://localhost:3031');
  assert.match(missingDorphSessionReason(true), /localhost:3031/);
});
