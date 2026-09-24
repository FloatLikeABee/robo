import { act } from 'react';
import { createRoot } from 'react-dom/client';
import Research from './Research';
import { tranApi } from '../../api/tranClient';

jest.mock('../../api/tranClient', () => ({
  tranApi: {
    get: jest.fn(),
    post: jest.fn(),
    delete: jest.fn(),
    patch: jest.fn(),
  },
  tranEndpoints: {
    research: '/api/tran/research',
    researchItem: (id) => `/api/tran/research/${id}`,
    researchPublish: (id) => `/api/tran/research/${id}/publish`,
    researchCancel: (id) => `/api/tran/research/${id}/cancel`,
  },
}));

jest.mock('../../components/ConfirmDialog', () => ({
  useConfirm: () => ({ confirm: async () => false }),
}));

jest.mock('../../components/admin/MarkdownEditor', () => ({
  __esModule: true,
  default: function MarkdownEditor({ value }) {
    return <textarea readOnly value={value || ''} />;
  },
}));

beforeAll(() => {
  global.IS_REACT_ACT_ENVIRONMENT = true;
});

const failed = {
  id: 7,
  title: 'Graphene',
  status: 'failed',
  error_text: 'AI is not configured (set MORPH_AI_API_KEY)',
  markdown_content: '',
  html_content: '',
  current_round: 0,
  round_count: 5,
  pieces: [{ round_index: 1, status: 'error', markdown: 'Round 1 failed: AI is not configured', verification: '' }],
};

test('opening a failed research job shows the cause and no thesis', async () => {
  tranApi.get.mockImplementation((url) => {
    if (String(url).endsWith('/7')) return Promise.resolve({ data: failed });
    return Promise.resolve({ data: [failed] });
  });

  const div = document.createElement('div');
  document.body.appendChild(div);
  const root = createRoot(div);
  await act(async () => {
    root.render(<Research />);
  });
  await act(async () => {
    await Promise.resolve();
  });

  expect(div.textContent).toContain('AI is not configured (set MORPH_AI_API_KEY)');
  expect(div.textContent).toContain('No thesis was produced.');
  expect(div.textContent).toContain('failed');
  expect(div.textContent).not.toContain('Round 1 failed');
  const publish = div.querySelector('button[aria-label="Publish"]');
  expect(publish).toBeTruthy();
  expect(publish.disabled).toBe(true);

  await act(async () => {
    root.unmount();
  });
  div.remove();
});
