import assert from 'node:assert/strict'
import test from 'node:test'
import { previewCanEnlarge } from './previewEnlarge.js'

test('starter preview does not enlarge', () => {
  assert.equal(previewCanEnlarge(''), false)
  assert.equal(previewCanEnlarge('<body data-composerx-starter="1"></body>'), false)
})

test('a real page can enlarge', () => {
  assert.equal(previewCanEnlarge('<!doctype html><html><body><h1>Launch</h1></body></html>'), true)
})
