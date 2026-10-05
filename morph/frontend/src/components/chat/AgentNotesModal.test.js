import React, { act, useCallback, useState } from 'react';
import { createRoot } from 'react-dom/client';
import { RouterProvider, createMemoryRouter } from 'react-router-dom';
import fs from 'fs';
import path from 'path';
import AgentNotesModal from './AgentNotesModal';
import SkoolAiChat from '../../SkoolAiChat';
import { agentNotesRoute } from '../../agentNotesRoute';
import ProtectedLayout from '../../components/ProtectedLayout';
import { ConfirmProvider } from '../ConfirmDialog';
import { getMorphToken } from '../../auth/morphSession';
import { tranApi } from '../../api/tranClient';

global.IS_REACT_ACT_ENVIRONMENT = true;

jest.mock('../../ChatMarkdown', () => () => null);
jest.mock('../notesTodos/NotesTodosContent', () => () => null);

jest.mock('../../auth/morphSession', () => ({
  getMorphToken: jest.fn(() => ''),
  clearMorphSession: jest.fn(),
}));

jest.mock('../../api/tranClient', () => ({
  tranApi: {
    get: jest.fn(() => Promise.resolve({ data: [] })),
    post: jest.fn(() => Promise.resolve({ data: {} })),
    delete: jest.fn(() => Promise.resolve({ data: {} })),
  },
  tranEndpoints: {
    agentNotes: '/api/tran/agent-notes',
    agentNote: (id) => `/api/tran/agent-notes/${id}`,
  },
}));

const note = {
  id: 3,
  title: '[morph-mcp] Shift report',
  body: 'source: morph-mcp\n\ndock 4 is clear',
  completed: false,
  created_on: '2026-10-01T12:00:00Z',
  item_type: 'note',
};

function mount(node) {
  const el = document.createElement('div');
  document.body.appendChild(el);
  const root = createRoot(el);
  act(() => {
    root.render(node);
  });
  return {
    cleanup() {
      act(() => root.unmount());
      el.remove();
    },
  };
}

async function flush() {
  await act(async () => {
    await Promise.resolve();
    await Promise.resolve();
  });
}

function Harness({ restoreFocusRef, fallbackFocusRef }) {
  const [open, setOpen] = useState(true);
  const close = useCallback(() => setOpen(false), []);
  if (!open) return null;
  return (
    <AgentNotesModal
      onClose={close}
      restoreFocusRef={restoreFocusRef}
      fallbackFocusRef={fallbackFocusRef}
    />
  );
}

beforeAll(() => {
  if (typeof window.matchMedia !== 'function') {
    window.matchMedia = () => ({
      matches: false,
      addEventListener() {},
      removeEventListener() {},
      addListener() {},
      removeListener() {},
    });
  }
  if (typeof Element.prototype.scrollIntoView !== 'function') {
    Element.prototype.scrollIntoView = () => {};
  }
});

beforeEach(() => {
  getMorphToken.mockReturnValue('');
  tranApi.get.mockReset();
  tranApi.get.mockResolvedValue({ data: [] });
  document.body.style.overflow = '';
});

test('signed-in list shows title time and status, and detail shows the body', async () => {
  tranApi.get.mockResolvedValue({
    data: [note, { ...note, id: 4, title: '[morph-mcp] Closed', body: 'source: morph-mcp\n\ndone', completed: true }],
  });
  const view = mount(<Harness />);
  await flush();
  expect(tranApi.get).toHaveBeenCalledWith('/api/tran/agent-notes');
  expect(tranApi.get.mock.calls.some((call) => String(call[0]).includes('/agent-notes/'))).toBe(false);
  expect(document.body.textContent).toContain('[morph-mcp] Shift report');
  expect(document.body.textContent).toMatch(/2026/);
  expect(document.body.textContent).toContain('Open');
  expect(document.body.textContent).toContain('Done');
  expect(document.body.textContent).not.toContain('dock 4 is clear');
  const row = [...document.querySelectorAll('button')].find((el) => el.textContent.includes('Shift report'));
  act(() => {
    row.dispatchEvent(new MouseEvent('click', { bubbles: true }));
  });
  expect(document.body.textContent).toContain('dock 4 is clear');
  expect(document.body.textContent).toContain('source: morph-mcp');
  view.cleanup();
});

test('empty agent notes explain how agents post them', async () => {
  tranApi.get.mockResolvedValue({ data: [] });
  const view = mount(<Harness />);
  await flush();
  expect(document.body.textContent).toContain('Morph note tool');
  expect(document.body.textContent).toContain('Nothing from an agent is here yet');
  view.cleanup();
});

test('a failed load shows the error and no note rows', async () => {
  tranApi.get.mockRejectedValue({ response: { data: { error: 'session does not map to one notes user' } } });
  const view = mount(<Harness />);
  await flush();
  expect(document.body.textContent).toContain('session does not map to one notes user');
  expect(document.body.textContent).not.toContain('Shift report');
  expect(document.querySelectorAll('.agent-notes-row')).toHaveLength(0);
  view.cleanup();
});

test('escape and close dismiss the modal, return focus, and do not bubble escape', async () => {
  const opener = document.createElement('button');
  opener.textContent = 'opener';
  document.body.appendChild(opener);
  opener.focus();
  const restoreFocusRef = { current: opener };
  document.body.style.overflow = 'auto';
  const bubble = jest.fn();
  document.addEventListener('keydown', bubble);
  const view = mount(<Harness restoreFocusRef={restoreFocusRef} />);
  await flush();
  const close = document.querySelector('[aria-label="Close agent notes"]');
  expect(close).toBeTruthy();
  expect(document.activeElement).toBe(close);
  expect(document.body.style.overflow).toBe('hidden');
  expect(document.querySelector('[role="dialog"]')).toBeTruthy();

  act(() => {
    window.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape', bubbles: true, cancelable: true }));
  });
  expect(document.querySelector('[role="dialog"]')).toBeNull();
  expect(document.activeElement).toBe(opener);
  expect(bubble).not.toHaveBeenCalled();
  expect(document.body.style.overflow).toBe('auto');

  opener.focus();
  const again = mount(<Harness restoreFocusRef={restoreFocusRef} />);
  await flush();
  act(() => {
    document.querySelector('[aria-label="Close agent notes"]').dispatchEvent(new MouseEvent('click', { bubbles: true }));
  });
  expect(document.querySelector('[role="dialog"]')).toBeNull();
  expect(document.activeElement).toBe(opener);
  document.removeEventListener('keydown', bubble);
  opener.remove();
  view.cleanup();
  again.cleanup();
});

function renderApp(entry) {
  const router = createMemoryRouter(
    [
      { path: '/login', element: <div>Sign in screen</div> },
      {
        element: <ProtectedLayout />,
        children: [
          agentNotesRoute,
          { path: '*', element: <SkoolAiChat variant="page" enableFileUpload /> },
        ],
      },
    ],
    { initialEntries: [entry] }
  );
  const view = mount(
    <ConfirmProvider>
      <RouterProvider router={router} />
    </ConfirmProvider>
  );
  return { router, ...view };
}

test('logged-out agent-notes path goes to login and does not load notes', async () => {
  getMorphToken.mockReturnValue('');
  const view = renderApp('/agent-notes');
  await flush();
  expect(view.router.state.location.pathname).toBe('/login');
  expect(decodeURIComponent(view.router.state.location.search)).toContain('/agent-notes');
  expect(document.body.textContent).toContain('Sign in');
  expect(tranApi.get).not.toHaveBeenCalled();
  view.cleanup();
});

test('signed-in agent-notes path opens the chat modal and drops the query', async () => {
  getMorphToken.mockReturnValue('token');
  tranApi.get.mockResolvedValue({ data: [note] });
  const view = renderApp('/agent-notes');
  await flush();
  expect(view.router.state.location.pathname).toBe('/');
  expect(view.router.state.location.search).toBe('');
  expect(document.querySelector('[role="dialog"]')).toBeTruthy();
  expect(document.body.textContent).toContain('[morph-mcp] Shift report');
  expect(tranApi.get).toHaveBeenCalledWith('/api/tran/agent-notes');
  view.cleanup();
});

test('header Agent notes opens the dialog and stays on the chat', async () => {
  getMorphToken.mockReturnValue('token');
  const view = renderApp('/');
  await flush();
  expect(document.querySelector('[role="dialog"]')).toBeNull();
  const bar = document.querySelector('.header-app-links--bar');
  const button = [...bar.querySelectorAll('button')].find((el) => el.textContent.includes('Agent notes'));
  act(() => {
    button.focus();
    button.click();
  });
  await flush();
  expect(document.querySelector('[role="dialog"]')).toBeTruthy();
  expect(view.router.state.location.pathname).toBe('/');
  expect(view.router.state.location.search).not.toContain('agent-notes');
  act(() => {
    window.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape', bubbles: true, cancelable: true }));
  });
  expect(document.querySelector('[role="dialog"]')).toBeNull();
  expect(document.activeElement).toBe(button);
  view.cleanup();
});

test('more-menu Agent notes returns focus to More apps', async () => {
  getMorphToken.mockReturnValue('token');
  const view = renderApp('/');
  await flush();
  const more = document.querySelector('[aria-label="More apps"]');
  act(() => {
    more.click();
  });
  const item = [...document.querySelectorAll('[role="menuitem"]')].find((el) => el.textContent.includes('Agent notes'));
  act(() => {
    item.focus();
    item.click();
  });
  await flush();
  expect(document.querySelector('[role="menu"]')).toBeNull();
  expect(document.querySelector('[role="dialog"]')).toBeTruthy();
  act(() => {
    window.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape', bubbles: true, cancelable: true }));
  });
  expect(document.activeElement).toBe(more);
  view.cleanup();
});

test('agent notes panel fits a phone column and keeps a 44px back control', () => {
  const css = fs.readFileSync(path.join(__dirname, '../../App.css'), 'utf8');
  const panel = css.slice(css.indexOf('.agent-notes-modal'));
  expect(css.indexOf('.agent-notes-modal')).toBeGreaterThan(-1);
  expect(css).not.toMatch(/\.agent-notes-page\b/);
  expect(panel).toMatch(/safe-area-inset-top/);
  expect(panel).toMatch(/safe-area-inset-bottom/);
  expect(panel).toMatch(/overflow-wrap:\s*anywhere/);
  expect(panel).toMatch(/text-wrap:\s*pretty/);
  expect(panel).toMatch(/text-wrap:\s*balance/);
  expect(panel).toMatch(/min-height:\s*44px/);
  const phone = css.slice(css.lastIndexOf('@media (max-width: 768px)'));
  expect(phone).toMatch(/\.agent-notes-layout\s*\{[^}]*grid-template-columns:\s*minmax\(0,\s*1fr\)/);
  expect(phone).toMatch(/\.agent-notes-back\s*\{[^}]*min-width:\s*44px/);
  expect(phone).toMatch(/\.agent-notes-back\s*\{[^}]*min-height:\s*44px/);
});
