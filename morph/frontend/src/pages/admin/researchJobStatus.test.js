import {
  researchAlert,
  researchCanPublish,
  researchEmptyThesisMessage,
  researchListSecondary,
  researchShowsThesisEditor,
} from './researchJobStatus';

test('failed job lists the cause and hides an empty thesis', () => {
  const job = {
    status: 'failed',
    error_text: 'AI is not configured (set MORPH_AI_API_KEY)',
    markdown_content: '',
    current_round: 0,
    round_count: 5,
  };
  expect(researchListSecondary(job)).toContain('failed');
  expect(researchListSecondary(job)).toContain('AI is not configured');
  expect(researchAlert(job)).toEqual({
    severity: 'error',
    message: 'AI is not configured (set MORPH_AI_API_KEY)',
  });
  expect(researchShowsThesisEditor(job)).toBe(false);
  expect(researchEmptyThesisMessage).toBe('No thesis was produced.');
  expect(researchCanPublish(job)).toBe(false);
});

test('complete job with failed rounds shows the thesis and a warning', () => {
  const job = {
    status: 'complete',
    error_text: '2 of 5 rounds failed (2, 4)',
    markdown_content: '# Partial thesis',
    current_round: 5,
    round_count: 5,
  };
  expect(researchAlert(job)).toEqual({
    severity: 'warning',
    message: '2 of 5 rounds failed (2, 4)',
  });
  expect(researchShowsThesisEditor(job)).toBe(true);
  expect(researchCanPublish(job)).toBe(true);
  expect(researchListSecondary(job)).toContain('complete');
  expect(researchListSecondary(job)).toContain('2 of 5');
  expect(researchListSecondary(job)).toContain('5/5');
});

test('successful job has no failure alert', () => {
  const job = {
    status: 'complete',
    markdown_content: '# Thesis',
    current_round: 5,
    round_count: 5,
  };
  expect(researchAlert(job)).toBeNull();
  expect(researchShowsThesisEditor(job)).toBe(true);
  expect(researchCanPublish(job)).toBe(true);
  const line = researchListSecondary(job);
  expect(line).toContain('complete');
  expect(line).not.toContain('failed');
});
