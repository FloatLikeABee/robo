import React, { act, useState } from 'react';
import { createRoot } from 'react-dom/client';
import GrillModal from './GrillModal';
import { createGrillHost } from './grillHost';
import { tranApi } from '../../api/tranClient';

global.IS_REACT_ACT_ENVIRONMENT = true;

jest.mock('../../api/tranClient', () => ({
  tranApi: {
    get: jest.fn(),
    post: jest.fn(),
  },
  tranEndpoints: {
    grill: '/api/tran/grill',
    grillUI: '/api/tran/grill/ui',
  },
}));

const page = '<!DOCTYPE html><html><head><style>body{font-family:Arial;background:#070b14}</style></head><body><h1>AI assist</h1></body></html>';

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

function buttonNamed(name) {
  return [...document.body.querySelectorAll('button')].find((node) => node.textContent === name);
}

beforeEach(() => {
  tranApi.get.mockReset();
  tranApi.post.mockReset();
  tranApi.get.mockResolvedValue({ data: page });
});

test('cancel leaves the note body unchanged', async () => {
  const applied = [];
  function Harness() {
    const [body, setBody] = useState('old body');
    const [open, setOpen] = useState(true);
    return (
      <div>
        <textarea readOnly value={body} />
        {open ? (
          <GrillModal
            kind="note"
            title="buy milk"
            description=""
            onClose={() => setOpen(false)}
            onApply={(draft) => {
              applied.push(draft);
              setBody(draft);
            }}
          />
        ) : null}
      </div>
    );
  }
  const view = mount(<Harness />);
  await flush();
  const cancel = buttonNamed('Cancel');
  expect(cancel).toBeTruthy();
  await act(async () => {
    cancel.click();
  });
  expect(applied).toEqual([]);
  expect(document.querySelector('textarea').value).toBe('old body');
  expect(document.querySelector('[aria-labelledby="grill-title"]') || document.getElementById('grill-title')).toBeFalsy();
  view.cleanup();
});

test('confirm inserts the server draft and ignores a page-invented draft', async () => {
  tranApi.post.mockResolvedValue({
    data: {
      id: 'g1',
      phase: 'review',
      draft: 'Milk, 2%',
      title: 'buy milk',
      kind: 'note',
      asked: 1,
      planned: 1,
    },
  });
  const posted = [];
  let deliver;
  const createHost = (opts) => {
    const host = createGrillHost({
      ...opts,
      post: (msg) => posted.push(msg),
    });
    deliver = host;
    return host;
  };
  function Harness() {
    const [body, setBody] = useState('old body');
    return (
      <div>
        <textarea readOnly value={body} />
        <GrillModal
          kind="note"
          title="buy milk"
          description=""
          createHost={createHost}
          onClose={() => {}}
          onApply={(draft) => setBody(draft)}
        />
      </div>
    );
  }
  const view = mount(<Harness />);
  for (let i = 0; i < 6 && !deliver; i += 1) {
    await flush();
  }
  expect(deliver).toBeTruthy();
  const frame = document.querySelector('iframe[title="AI assist"]');
  expect(frame).toBeTruthy();
  expect(frame.getAttribute('srcdoc')).toContain('Arial');
  const useThis = buttonNamed('Use this');
  expect(useThis.disabled).toBe(true);

  await act(async () => {
    await deliver({
      data: {
        jsonrpc: '2.0',
        id: 7,
        method: 'tools/call',
        params: { name: 'note_grill', arguments: { action: 'finish', id: 'g1' } },
      },
    });
  });
  await flush();
  expect(tranApi.post).toHaveBeenCalledWith('/api/tran/grill', expect.objectContaining({
    action: 'finish',
    id: 'g1',
    title: 'buy milk',
  }));
  expect(buttonNamed('Use this').disabled).toBe(false);
  await act(async () => {
    await deliver({
      data: {
        jsonrpc: '2.0',
        method: 'ui/notifications/grill-apply',
        params: { draft: "Great! I'm Dorph AI" },
      },
    });
  });
  expect(document.querySelector('textarea').value).toBe('Milk, 2%');
  expect(posted.some((msg) => msg.result && msg.result.structuredContent && msg.result.structuredContent.draft === 'Milk, 2%')).toBe(true);
  view.cleanup();
});

test('a failed model call does not insert a draft', async () => {
  const applied = [];
  const host = createGrillHost({
    kind: 'note',
    title: 'buy milk',
    description: '',
    callGrill: async () => {
      const err = new Error('model down');
      err.response = { data: { error: 'model down' } };
      throw err;
    },
    onApply: (draft) => applied.push(draft),
    onReview: () => {},
    onError: () => {},
    post: () => {},
  });
  await host({
    data: {
      jsonrpc: '2.0',
      id: 1,
      method: 'tools/call',
      params: { name: 'note_grill', arguments: { action: 'start' } },
    },
  });
  await host({
    data: { jsonrpc: '2.0', method: 'ui/notifications/grill-apply', params: { draft: 'nope' } },
  });
  expect(applied).toEqual([]);
});
