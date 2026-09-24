import assert from 'node:assert/strict';
import { test } from 'node:test';
import {
  appendSessionToken,
  morphLoginHref,
  resolvePublicUrl,
  sessionCookieAttributes,
  sheetxEmbedUrl,
  usesLocalStartHint,
} from './publicUrl.ts';

test('runtime primary wins over alias and built values', () => {
  assert.equal(
    resolvePublicUrl({
      runtime: 'https://morph.example/api/',
      runtimeAlias: 'https://alias.example',
      built: 'https://built.example',
      builtAlias: 'https://built-alias.example',
      devFallback: 'http://localhost:19909',
      dev: false,
    }),
    'https://morph.example/api',
  );
});

test('blank runtime does not hide a built value', () => {
  assert.equal(
    resolvePublicUrl({
      runtime: '  ',
      runtimeAlias: '',
      built: 'https://built.example/',
      devFallback: 'http://localhost:19909',
      dev: true,
    }),
    'https://built.example',
  );
});

test('runtime alias is used when the primary is blank', () => {
  assert.equal(
    resolvePublicUrl({
      runtime: '',
      runtimeAlias: 'https://alias.example/',
      built: 'https://built.example',
      dev: false,
    }),
    'https://alias.example',
  );
});

test('dev localhost fallback is used only when every candidate is blank', () => {
  assert.equal(
    resolvePublicUrl({
      devFallback: 'http://localhost:19909/',
      dev: true,
    }),
    'http://localhost:19909',
  );
});

test('empty SheetX origin does not produce a truthy /events-info embed', () => {
  const embedUrl = sheetxEmbedUrl('');
  assert.equal(embedUrl, '');
  assert.equal(Boolean(embedUrl), false);
});

test('SheetX origin keeps the events-info path', () => {
  assert.equal(sheetxEmbedUrl('http://localhost:19909'), 'http://localhost:19909/events-info');
});

test('configured https origins are the embed sources', () => {
  assert.equal(sheetxEmbedUrl('https://logs.example'), 'https://logs.example/events-info');
  assert.equal(appendSessionToken('https://logs.example/events-info', 'jwt'), 'https://logs.example/events-info?userspanel_token=jwt');
  assert.equal(appendSessionToken('https://content.example', 'jwt'), 'https://content.example/?userspanel_token=jwt');
  assert.equal(appendSessionToken('https://data.example', ''), 'https://data.example');
  for (const src of [
    sheetxEmbedUrl('https://logs.example'),
    'https://content.example',
    'https://data.example',
    'https://project.example',
  ]) {
    assert.equal(src.includes('localhost'), false);
    assert.equal(src.includes('127.0.0.1'), false);
  }
});

test('session cookie is host-only Lax with no Domain', () => {
  const attrs = sessionCookieAttributes(3600);
  assert.match(attrs, /SameSite=Lax/);
  assert.equal(attrs.includes('Domain'), false);
  assert.match(attrs, /Path=\//);
});

test('production sign-in href uses Morph and ignores loopback', () => {
  assert.equal(
    morphLoginHref({ morphAi: '', morphApi: 'https://morph.example/', dev: false }),
    'https://morph.example',
  );
  assert.equal(
    morphLoginHref({ morphAi: 'https://ai.example', morphApi: 'https://morph.example', dev: false }),
    'https://ai.example',
  );
  assert.equal(
    morphLoginHref({ morphAi: '', morphApi: 'http://localhost:3031', dev: false }),
    '',
  );
  assert.equal(morphLoginHref({ morphAi: '', morphApi: '', dev: false }), '');
  assert.equal(morphLoginHref({ morphAi: '', morphApi: '', dev: true }), '');
  assert.equal(
    morphLoginHref({ morphAi: 'http://localhost:3031', morphApi: '', dev: true }),
    'http://localhost:3031',
  );
});

test('remote embed miss does not use the local start hint', () => {
  assert.equal(usesLocalStartHint('http://localhost:5178'), true);
  assert.equal(usesLocalStartHint('http://127.0.0.1:5179'), true);
  assert.equal(usesLocalStartHint('https://data.example'), false);
});

test('production stays empty when every candidate is blank', () => {
  assert.equal(
    resolvePublicUrl({
      runtime: '',
      built: undefined,
      devFallback: 'http://localhost:19909',
      dev: false,
    }),
    '',
  );
});
