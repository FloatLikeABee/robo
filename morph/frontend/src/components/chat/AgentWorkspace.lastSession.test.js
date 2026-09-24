jest.mock('../notesTodos/NotesTodosContent', () => () => null);
jest.mock('../../HybridContextDrawer', () => () => null);

import {
  readWorkspaceTab,
  resolveRestoredSessionId,
  workspaceTabStorageKey,
} from './AgentWorkspace';

test('keeps a last id that still exists', () => {
  expect(
    resolveRestoredSessionId({
      lastId: 'a',
      sessionIds: ['default', 'a', 'b'],
    })
  ).toBe('a');
});

test('falls back to default when last id is gone', () => {
  expect(
    resolveRestoredSessionId({
      lastId: 'gone',
      sessionIds: ['default', 'b', 'c'],
    })
  ).toBe('default');
});

test('empty inputs restore default', () => {
  expect(resolveRestoredSessionId({})).toBe('default');
  expect(resolveRestoredSessionId({ lastId: '', sessionIds: [] })).toBe('default');
});

test('readWorkspaceTab maps leftover files tab to knowledge', () => {
  localStorage.setItem(workspaceTabStorageKey('s1'), 'files');
  expect(readWorkspaceTab('s1')).toBe('knowledge');
  localStorage.setItem(workspaceTabStorageKey('s1'), 'notes');
  expect(readWorkspaceTab('s1')).toBe('notes');
});
