import {
  initialWorkspaceOpen,
  readWorkspaceOpen,
  readWorkspaceTab,
  resolveRestoredSessionId,
  workspaceTabStorageKey,
} from './AgentWorkspace';

jest.mock('../notesTodos/NotesTodosContent', () => () => null);
jest.mock('../../HybridContextDrawer', () => () => null);

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

test('unset phone workspace starts closed and desktop stays open', () => {
  expect(initialWorkspaceOpen({ stored: null, phone: true })).toBe(false);
  expect(initialWorkspaceOpen({ phone: true })).toBe(false);
  expect(initialWorkspaceOpen({ stored: null, phone: false })).toBe(true);
});

test('unset storage still reads as open for the desktop helper', () => {
  localStorage.removeItem('morphai-workspace-open');
  expect(readWorkspaceOpen()).toBe(true);
});

test('saved workspace choice wins on a phone', () => {
  expect(initialWorkspaceOpen({ stored: '1', phone: true })).toBe(true);
  expect(initialWorkspaceOpen({ stored: '0', phone: true })).toBe(false);
  expect(initialWorkspaceOpen({ stored: '0', phone: false })).toBe(false);
});

test('readWorkspaceTab maps leftover files tab to knowledge', () => {
  localStorage.setItem(workspaceTabStorageKey('s1'), 'files');
  expect(readWorkspaceTab('s1')).toBe('knowledge');
  localStorage.setItem(workspaceTabStorageKey('s1'), 'notes');
  expect(readWorkspaceTab('s1')).toBe('notes');
});
