import { test } from 'node:test'
import assert from 'node:assert/strict'
import { mergePublishedContents, slugifyPublishName } from './publishedContents.js'

test('slugifyPublishName matches backend identity', () => {
  assert.equal(slugifyPublishName('Summer Launch'), 'summer-launch')
})

test('merge keeps one row when draft and published share a name', () => {
  const rows = mergePublishedContents(
    [{ id: 10, name: 'Summer Launch', theme: 'dark', updated_at: '2026-09-01T01:00:00Z' }],
    [{ id: 3, name: 'Summer Launch', slug: 'summer-launch', theme: 'light', updated_at: '2026-09-01T02:00:00Z' }],
  )
  assert.equal(rows.length, 1)
  assert.equal(rows[0].status, 'Published')
  assert.equal(rows[0].path, '/public/p/summer-launch')
  assert.equal(rows[0].draftId, 10)
  assert.equal(rows[0].publishedId, 3)
  assert.equal(rows[0].slug, 'summer-launch')
  assert.equal(rows[0].canDelete, true)
})

test('saved-only rows keep draft actions and no public path', () => {
  const rows = mergePublishedContents(
    [{ id: 8, name: 'WIP', theme: 'default', updated_at: '2026-09-01T01:00:00Z' }],
    [],
  )
  assert.equal(rows.length, 1)
  assert.equal(rows[0].status, 'Saved')
  assert.equal(rows[0].path, '')
  assert.equal(rows[0].draftId, 8)
  assert.equal(rows[0].canOpen, false)
  assert.equal(rows[0].canView, true)
  assert.equal(rows[0].canDelete, true)
})

test('published-only rows can open and can delete', () => {
  const rows = mergePublishedContents(
    [],
    [{ id: 3, name: 'Live', slug: 'live', theme: 'default', updated_at: '2026-09-01T02:00:00Z' }],
  )
  assert.equal(rows.length, 1)
  assert.equal(rows[0].status, 'Published')
  assert.equal(rows[0].canOpen, true)
  assert.equal(rows[0].canDelete, true)
  assert.equal(rows[0].canView, false)
  assert.equal(rows[0].publishedId, 3)
  assert.equal(rows[0].draftId, null)
})
