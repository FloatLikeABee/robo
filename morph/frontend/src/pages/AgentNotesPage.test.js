import React, { act } from 'react';
import { createRoot } from 'react-dom/client';
import { RouterProvider, createMemoryRouter } from 'react-router-dom';
import fs from 'fs';
import path from 'path';
import ProtectedLayout from '../components/ProtectedLayout';
import AgentNotesPage from './AgentNotesPage';
import { getMorphToken } from '../auth/morphSession';
import { tranApi } from '../api/tranClient';

global.IS_REACT_ACT_ENVIRONMENT = true;

jest.mock('../auth/morphSession', () => ({
  getMorphToken: jest.fn(() => ''),
}));

jest.mock('../api/tranClient', () => ({
  tranApi: { get: jest.fn() },
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

function renderAt(pathname) {
  const router = createMemoryRouter(
    [
      { path: '/login', element: <div>Sign in screen</div> },
      {
        element: <ProtectedLayout />,
        children: [{ path: '/agent-notes', element: <AgentNotesPage /> }],
      },
    ],
    { initialEntries: [pathname] }
  );
  const el = document.createElement('div');
  document.body.appendChild(el);
  const root = createRoot(el);
  act(() => {
    root.render(<RouterProvider router={router} />);
  });
  return {
    router,
    cleanup() {
      act(() => root.unmount());
      el.remove();
    },
  };
}

test('logged-out agent notes route goes to login and does not load notes', () => {
  getMorphToken.mockReturnValue('');
  tranApi.get.mockClear();
  const view = renderAt('/agent-notes');
  expect(view.router.state.location.pathname).toBe('/login');
  expect(document.body.textContent).toContain('Sign in screen');
  expect(tranApi.get).not.toHaveBeenCalled();
  view.cleanup();
});

test('signed-in list shows title time and status, and detail shows the body', async () => {
  getMorphToken.mockReturnValue('token');
  tranApi.get.mockResolvedValue({
    data: [note, { ...note, id: 4, title: '[morph-mcp] Closed', body: 'source: morph-mcp\n\ndone', completed: true }],
  });
  const view = renderAt('/agent-notes');
  await act(async () => {
    await Promise.resolve();
  });
  expect(tranApi.get).toHaveBeenCalledWith('/api/tran/agent-notes');
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
  getMorphToken.mockReturnValue('token');
  tranApi.get.mockResolvedValue({ data: [] });
  const view = renderAt('/agent-notes');
  await act(async () => {
    await Promise.resolve();
  });
  expect(document.body.textContent).toContain('Morph note tool');
  expect(document.body.textContent).toContain('Nothing from an agent is here yet');
  view.cleanup();
});

test('agent notes page fits a phone column and keeps a 44px back control', () => {
  const css = fs.readFileSync(path.join(__dirname, '../App.css'), 'utf8');
  const page = css.slice(css.indexOf('.agent-notes-page'));
  expect(css.indexOf('.agent-notes-page')).toBeGreaterThan(-1);
  expect(page).toMatch(/safe-area-inset-top/);
  expect(page).toMatch(/safe-area-inset-bottom/);
  expect(page).toMatch(/overflow-wrap:\s*anywhere/);
  expect(page).toMatch(/text-wrap:\s*pretty/);
  expect(page).toMatch(/text-wrap:\s*balance/);
  expect(page).toMatch(/min-height:\s*44px/);
  const phone = css.slice(css.lastIndexOf('@media (max-width: 768px)'));
  expect(phone).toMatch(/\.agent-notes-layout\s*\{[^}]*grid-template-columns:\s*minmax\(0,\s*1fr\)/);
  expect(phone).toMatch(/\.agent-notes-back\s*\{[^}]*min-width:\s*44px/);
  expect(phone).toMatch(/\.agent-notes-back\s*\{[^}]*min-height:\s*44px/);
  const router = fs.readFileSync(path.join(__dirname, '../appRouter.js'), 'utf8');
  const agent = router.indexOf("path: 'agent-notes'");
  const catchAll = router.indexOf("path: '*'", agent);
  expect(agent).toBeGreaterThan(-1);
  expect(catchAll).toBeGreaterThan(agent);
  const chat = fs.readFileSync(path.join(__dirname, '../SkoolAiChat.js'), 'utf8');
  expect(chat).toContain("label: 'Agent notes'");
  expect(chat).toContain("'/agent-notes'");
});
