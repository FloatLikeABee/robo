import assert from 'node:assert/strict'
import fs from 'node:fs'
import test from 'node:test'

const panel = fs.readFileSync(new URL('../components/ComposePublishPanel.svelte', import.meta.url), 'utf8')

test('source field grows and actions stack with data last', () => {
  const field = panel.slice(panel.indexOf('.publish-source-box textarea'))
  assert.match(field.slice(0, 280), /flex:\s*1/)
  assert.match(field.slice(0, 280), /min-height:\s*12rem/)
  assert.doesNotMatch(field.slice(0, 400), /5\.6rem/)

  const actions = panel.slice(panel.indexOf('publish-source-actions'), panel.indexOf('composer-ai-rail'))
  const upload = actions.indexOf("'Upload'")
  const fromText = actions.indexOf('From text')
  const fromData = actions.indexOf('From data')
  const primary = actions.indexOf('primary-action')
  assert.ok(upload >= 0 && upload < fromText && fromText < primary && primary < fromData)
})

test('secondary publish actions do not use the panel accent stroke', () => {
  const secondary = panel.slice(panel.indexOf('.btn-secondary {'), panel.indexOf('.btn-ghost {'))
  assert.match(secondary, /background:\s*var\(--color-surface\)/)
  assert.match(secondary, /border:\s*1px solid transparent/)
  assert.doesNotMatch(secondary, /--color-border-subtle/)

  const send = panel.slice(panel.indexOf('.btn-send-fancy {'), panel.indexOf('.btn-send-fancy :global'))
  assert.match(send, /background:\s*var\(--color-primary\)/)
  assert.doesNotMatch(send, /29,\s*78,\s*216|37,\s*99,\s*235/)

  const records = fs.readFileSync(new URL('../components/ComposePublishRecordsPanel.svelte', import.meta.url), 'utf8')
  const refresh = records.slice(records.indexOf('.btn-secondary {'), records.indexOf('.btn-ghost {'))
  assert.match(refresh, /background:\s*var\(--color-surface\)/)
  assert.doesNotMatch(refresh, /--color-border-subtle/)
})
