import { contextFingerprint, inferSubAgents } from './agentContext';

test('contextFingerprint is notes and knowledge only (files slot off)', () => {
  expect(contextFingerprint({ includeNotes: true, includeKnowledge: true })).toBe('0|1|1|');
  expect(contextFingerprint({ includeNotes: false, includeKnowledge: true })).toBe('0|0|1|');
});

test('inferSubAgents routes knowledge and notes without a files worker', () => {
  expect(inferSubAgents('summarize this according to the library', { hasKnowledge: true })).toEqual(['knowledge']);
  expect(inferSubAgents('add a todo for the meeting', { hasNotes: true })).toEqual(['notes']);
  expect(inferSubAgents('open the folder workspace files', { hasKnowledge: false, hasNotes: false })).toEqual([]);
});
