import { assistantHasBody } from './assistantHasBody';

test('an empty streaming reply is not a bubble', () => {
  expect(assistantHasBody({ type: 'assistant', content: '', reasoning: '' })).toBe(false);
  expect(assistantHasBody({ type: 'assistant', content: '   ', reasoning: '' })).toBe(false);
  expect(assistantHasBody({ type: 'assistant', content: 'Hi', reasoning: '' })).toBe(true);
  expect(assistantHasBody({ type: 'assistant', content: '', reasoning: 'plan' })).toBe(true);
});
