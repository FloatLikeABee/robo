import React, { act } from 'react';
import { createRoot } from 'react-dom/client';
import HeaderMoreMenu from './HeaderMoreMenu';

global.IS_REACT_ACT_ENVIRONMENT = true;

function renderMenu(props) {
  const el = document.createElement('div');
  document.body.appendChild(el);
  const root = createRoot(el);
  act(() => {
    root.render(<HeaderMoreMenu {...props} />);
  });
  return {
    cleanup() {
      act(() => root.unmount());
      el.remove();
    },
  };
}

function click(el) {
  act(() => {
    el.dispatchEvent(new MouseEvent('click', { bubbles: true }));
  });
}

const items = [
  { id: 'skills', label: 'Skills', onClick: jest.fn() },
  { id: 'bk', label: 'AI tools', onClick: jest.fn() },
  { id: 'morphdata', label: 'MorphNotes', href: '/morphdata' },
];

function menuItems() {
  return [...document.querySelectorAll('[role="menuitem"]')].map((el) => el.textContent);
}

test('more menu lists labeled app chips and clear, and omits MorphUtils when it is not configured', () => {
  items[0].onClick.mockClear();
  const view = renderMenu({ items, onClear: jest.fn() });
  click(document.querySelector('[aria-label="More apps"]'));
  expect(menuItems()).toEqual(['Skills', 'AI tools', 'MorphNotes', 'Clear chat']);
  click(document.querySelector('[role="menuitem"]'));
  expect(items[0].onClick).toHaveBeenCalledTimes(1);
  expect(document.querySelector('[role="menu"]')).toBeNull();
  view.cleanup();
});

test('backdrop and escape close the more menu', () => {
  const view = renderMenu({ items: items.slice(0, 1), onClear: () => {} });
  click(document.querySelector('[aria-label="More apps"]'));
  click(document.querySelector('[aria-label="Close more apps"]'));
  expect(document.querySelector('[role="menu"]')).toBeNull();

  click(document.querySelector('[aria-label="More apps"]'));
  act(() => {
    document.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape', bubbles: true }));
  });
  expect(document.querySelector('[role="menu"]')).toBeNull();
  view.cleanup();
});
