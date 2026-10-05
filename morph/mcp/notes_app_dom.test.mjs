import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import path from "node:path";
import { test } from "node:test";
import { fileURLToPath } from "node:url";
import vm from "node:vm";

const htmlPath = path.join(path.dirname(fileURLToPath(import.meta.url)), "notes_app.html");

function createElement(tag) {
  return {
    tag,
    children: [],
    attrs: {},
    text: null,
    set textContent(value) {
      this.text = String(value);
      this.children = [];
    },
    get textContent() {
      return this.text;
    },
    set innerHTML(_value) {
      throw new Error("innerHTML");
    },
    appendChild(child) {
      this.children.push(child);
      return child;
    },
    removeChild(child) {
      const index = this.children.indexOf(child);
      if (index >= 0) this.children.splice(index, 1);
      return child;
    },
    setAttribute(key, value) {
      this.attrs[key] = String(value);
    },
    get firstChild() {
      return this.children[0] || null;
    },
  };
}

function walk(node, visit) {
  visit(node);
  for (const child of node.children || []) walk(child, visit);
}

test("note title and body stay text", () => {
  const html = readFileSync(htmlPath, "utf8");
  const scripts = [...html.matchAll(/<script>([\s\S]*?)<\/script>/g)].map((match) => match[1]);
  assert.equal(scripts.length, 1);

  const sandbox = { document: { createElement } };
  vm.createContext(sandbox);
  vm.runInContext(scripts[0], sandbox);

  const markup = "<script>alert(1)</script><img src=x onerror=alert(1)>";
  const list = createElement("ul");
  sandbox.morphNotesRenderList(list, [{ id: 7, title: markup }]);
  const title = list.children[0].children[0].children[0];
  assert.equal(title.text, markup);
  walk(list, (node) => {
    assert.notEqual(node.tag, "script");
    assert.notEqual(node.tag, "img");
  });

  const detail = createElement("section");
  sandbox.morphNotesRenderDetail(detail, { id: 7, title: markup, body: markup + " body" });
  assert.equal(detail.children[0].text, markup);
  assert.equal(detail.children[1].text, markup + " body");
  walk(detail, (node) => {
    assert.notEqual(node.tag, "script");
    assert.notEqual(node.tag, "img");
  });
});
