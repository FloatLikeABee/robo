import { MORPH_NOTES_EMBEDS, resolveEmbedUrl, withSessionToken } from './morphNotesEmbeds';

test('embed list is Event Logs, Content Maker, and Project', () => {
  expect(MORPH_NOTES_EMBEDS.map((item) => item.label)).toEqual([
    'Event Logs',
    'Content Maker',
    'Project',
  ]);
  expect(MORPH_NOTES_EMBEDS.some((item) => /data access/i.test(item.label))).toBe(false);
});

test('Event Logs keeps /events-info and a blank production origin does not become a relative path', () => {
  expect(resolveEmbedUrl({ id: 'event-logs', nodeEnv: 'development', configured: '' })).toBe(
    'http://localhost:19909/events-info'
  );
  expect(resolveEmbedUrl({ id: 'event-logs', nodeEnv: 'production', configured: '' })).toBe('');
  expect(resolveEmbedUrl({ id: 'event-logs', nodeEnv: 'production', configured: 'http://127.0.0.1:19909' })).toBe('');
  expect(resolveEmbedUrl({
    id: 'event-logs',
    nodeEnv: 'production',
    configured: 'https://logs.example',
  })).toBe('https://logs.example/events-info');
});

test('embed URL carries the Morph bearer', () => {
  expect(withSessionToken('https://projects.example', 'abc')).toBe(
    'https://projects.example/?userspanel_token=abc'
  );
  expect(withSessionToken('', 'abc')).toBe('');
});
