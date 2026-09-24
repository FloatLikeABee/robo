import assert from 'node:assert/strict';
import { test } from 'node:test';
import { resolvePublicUrl, sheetxEmbedUrl } from './publicUrl.ts';

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
