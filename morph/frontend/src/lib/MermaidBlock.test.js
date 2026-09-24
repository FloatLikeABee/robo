import { act } from 'react';
import { createRoot } from 'react-dom/client';
import mermaid from 'mermaid';
import MermaidBlock from './MermaidBlock';
import { VisualLightboxProvider } from './visualLightbox';

jest.mock('mermaid', () => ({
  __esModule: true,
  default: {
    initialize: jest.fn(),
    render: jest.fn(),
  },
}));

global.IS_REACT_ACT_ENVIRONMENT = true;

const FLOW = 'flowchart TD\nA-->B';
const SEQUENCE = 'sequenceDiagram\nAlice->>Bob: hi';

function mount(ui) {
  const el = document.createElement('div');
  document.body.appendChild(el);
  const root = createRoot(el);
  act(() => {
    root.render(ui);
  });
  return {
    el,
    unmount() {
      act(() => root.unmount());
      el.remove();
    },
  };
}

async function flush() {
  await act(async () => {
    await new Promise((resolve) => setTimeout(resolve, 0));
  });
}

beforeEach(() => {
  mermaid.initialize.mockClear();
  mermaid.render.mockReset();
});

test('lazy-loads mermaid and renders a flowchart svg', async () => {
  mermaid.render.mockResolvedValue({
    svg: '<svg id="flow" xmlns="http://www.w3.org/2000/svg"></svg>',
  });
  const view = mount(<MermaidBlock source={FLOW} />);
  await flush();
  expect(mermaid.initialize).toHaveBeenCalled();
  expect(mermaid.render).toHaveBeenCalledWith(expect.any(String), FLOW);
  expect(view.el.querySelector('svg#flow')).toBeTruthy();
  expect(view.el.querySelector('.chat-mermaid-fallback')).toBeNull();
  view.unmount();
});

test('renders a sequence diagram through the same lazy loader', async () => {
  mermaid.render.mockResolvedValue({
    svg: '<svg id="seq" xmlns="http://www.w3.org/2000/svg"></svg>',
  });
  const view = mount(<MermaidBlock source={SEQUENCE} />);
  await flush();
  expect(mermaid.render).toHaveBeenCalledWith(expect.any(String), SEQUENCE);
  expect(view.el.querySelector('svg#seq')).toBeTruthy();
  view.unmount();
});

test('falls back to the source when mermaid.render rejects', async () => {
  mermaid.render.mockRejectedValue(new Error('bad diagram'));
  const view = mount(<MermaidBlock source={FLOW} />);
  await flush();
  expect(view.el.querySelector('.chat-mermaid-fallback')?.textContent).toBe(FLOW);
  view.unmount();
});

test('clicking a diagram opens the enlarge lightbox', async () => {
  mermaid.render.mockResolvedValue({
    svg: '<svg xmlns="http://www.w3.org/2000/svg" width="40" height="20"><rect width="40" height="20"/></svg>',
  });
  const view = mount(
    <VisualLightboxProvider>
      <MermaidBlock source={FLOW} />
    </VisualLightboxProvider>
  );
  await flush();
  const diagram = view.el.querySelector('.chat-visual-enlarge');
  expect(diagram?.querySelector('svg')).toBeTruthy();
  act(() => {
    diagram.click();
  });
  const overlay = view.el.querySelector('.visual-lightbox');
  expect(overlay).toBeTruthy();
  expect(overlay.querySelector('svg')).toBeTruthy();
  expect(overlay.querySelector('svg').getAttribute('width')).toBeNull();
  view.unmount();
});
