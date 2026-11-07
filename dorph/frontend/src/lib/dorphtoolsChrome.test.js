const fs = require('fs');
const path = require('path');

function read(rel) {
  return fs.readFileSync(path.join(__dirname, rel), 'utf8');
}

test('DORPHTOOLS title is solid light ink', () => {
  const src = read('../../../../bk/frontend/src/components/Header.js');
  expect(src).toMatch(/color:\s*'#f4f0fa'/);
  expect(src).not.toMatch(/WebkitTextFillColor:\s*'transparent'/);
});

test('Documents does not mount Graphic documents', () => {
  const src = read('../../../../bk/frontend/src/pages/PublishHub.js');
  expect(src).toMatch(/Articles/);
  expect(src).toMatch(/ScholarForge/);
  expect(src).not.toMatch(/Graphic documents/);
  expect(src).not.toMatch(/GraphicDocumentGenerator/);
});
