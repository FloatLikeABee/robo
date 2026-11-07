import { providerSaveReady, modelIdsFromScan } from './chatProviderForm';

const fs = require('fs');
const path = require('path');

test('session rail keeps the provider settings control at the bottom', () => {
  const src = fs.readFileSync(path.join(__dirname, '../SkoolAiChat.js'), 'utf8');
  const aside = src.slice(src.indexOf('<aside'), src.indexOf('</aside>'));
  expect(aside).toMatch(/sidebar-provider-settings/);
  expect(aside.indexOf('sidebar-split-top')).toBeLessThan(aside.indexOf('sidebar-provider-settings'));
});

test('Ollama can be saved without a key', () => {
  const providers = [
    { id: 'openai', requires_key: true },
    { id: 'ollama', requires_key: false },
  ];
  expect(providerSaveReady({
    provider: 'ollama', model: 'llama3', apiKey: '', keySet: false, providers,
  })).toBe(true);
  expect(providerSaveReady({
    provider: 'openai', model: 'gpt-4.1', apiKey: '', keySet: false, providers,
  })).toBe(false);
  expect(providerSaveReady({
    provider: 'openai', model: 'gpt-4.1', apiKey: '', keySet: true, providers,
  })).toBe(true);
});

test('scan ids come from the model list', () => {
  expect(modelIdsFromScan({ models: [{ id: 'llama3' }, { id: 'qwen2.5' }] })).toEqual(['llama3', 'qwen2.5']);
});
