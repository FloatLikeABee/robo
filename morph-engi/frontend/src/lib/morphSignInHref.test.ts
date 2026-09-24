import assert from 'node:assert/strict';
import { test } from 'node:test';
import { missingMorphSessionReason, morphSignInHref } from './morphSignInHref.ts';

test('production Project sign-in is not localhost when unset', () => {
  assert.equal(morphSignInHref('', false), '');
  assert.equal(morphSignInHref('http://127.0.0.1:3031', false), '');
  assert.equal(missingMorphSessionReason(false).includes('localhost'), false);
});

test('configured Project sign-in uses the Morph origin', () => {
  assert.equal(morphSignInHref('https://morph.example/', false), 'https://morph.example');
});

test('dev Project sign-in stays on localhost when unset', () => {
  assert.equal(morphSignInHref('', true), 'http://localhost:3031');
  assert.match(missingMorphSessionReason(true), /localhost:3031/);
});
