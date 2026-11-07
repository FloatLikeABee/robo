import assert from 'node:assert/strict'
import test from 'node:test'
import { DATA_TABLE_MAX_ROWS, buildDataSource, dorphParentOrigin } from './contentDataSource.js'

test('pdf becomes an article and csv becomes a capped table', () => {
  const article = buildDataSource({
    title: 'Brief',
    source_type: 'pdf',
    ai_analysis: 'Costs rose.',
    detail: { content_markdown: 'The fleet is small.' },
  })
  assert.equal(article.layout, 'article')
  assert.match(article.text, /Costs rose/)
  assert.match(article.text, /fleet is small/)
  assert.equal(article.text.includes(' | '), false)

  const rows = Array.from({ length: DATA_TABLE_MAX_ROWS + 2 }, (_, i) => ({ a: String(i) }))
  const table = buildDataSource({
    title: 'Fleet',
    source_type: 'csv',
    detail: JSON.stringify({ columns: ['a'], rows }),
  })
  assert.equal(table.layout, 'table')
  assert.match(table.text, /more rows omitted/)
  assert.equal((table.text.match(/\n/g) || []).length > DATA_TABLE_MAX_ROWS, true)
  const bodyRows = table.text.split('\n').filter((line) => /^\d+$/.test(line.trim()))
  assert.equal(bodyRows.length, DATA_TABLE_MAX_ROWS)
})

test('json is a table and a missing parent origin is empty', () => {
  const json = buildDataSource({ title: 'Dump', source_type: 'json', detail: { columns: ['k'], rows: [{ k: 'v' }] } })
  assert.equal(json.layout, 'table')
  assert.match(json.text, /^k$/m)
  assert.equal(dorphParentOrigin(''), '')
  assert.equal(dorphParentOrigin('http://127.0.0.1:3031/dorphdata/contents-data'), 'http://127.0.0.1:3031')
})
