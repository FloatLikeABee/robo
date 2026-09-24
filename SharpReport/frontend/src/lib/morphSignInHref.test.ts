import assert from 'node:assert/strict';
import { test } from 'node:test';
import { morphSignInHref } from './morphSignInHref.ts';

test('production Data Access sign-in uses a configured Morph origin', () => {
  assert.equal(morphSignInHref('https://morph.example/', false), 'https://morph.example');
});

test('production Data Access sign-in is not localhost when unset', () => {
  assert.equal(morphSignInHref('', false), '');
  assert.equal(morphSignInHref('http://localhost:3031', false), '');
  assert.equal(morphSignInHref(undefined, false).includes('localhost'), false);
});

test('dev Data Access sign-in stays on localhost when unset', () => {
  assert.equal(morphSignInHref('', true), 'http://localhost:3031');
});
