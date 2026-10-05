## 1. Failing tests

- [x] 1.1 Update the in-memory handshake test so `resources/list` expects `ui://morph/notes` with MIME `text/html;profile=mcp-app`, `listChanged` stays false, and initialize advertises extension `io.modelcontextprotocol/ui` with that MIME type. Assert `list_my_tasks`, `get_task`, and `create_note` carry `_meta.ui.resourceUri` (and the flat `ui/resourceUri` key) for that URI with model and app visibility, and `whoami` does not.
- [x] 1.2 Assert `resources/read` of `ui://morph/notes` returns that MIME type and HTML with no `http:` or `https:` URL, no `url(`, and no `innerHTML`, `insertAdjacentHTML`, `outerHTML`, or `document.write`. Assert the HTML names the three note tools and does not contain the session token. Assert a tool call that does not read the resource still returns note data rather than HTML.
- [x] 1.3 Add a node test that loads the page script with a DOM shim, renders a title and body containing `<script>` and `<img onerror>`, and asserts the text is unchanged and no element was parsed from that markup.
- [x] 1.4 Add a go-sdk stdio test that lists tools, checks the UI `_meta`, reads `ui://morph/notes`, checks the MIME type and the closed HTML, and logs a redacted transcript. Run the new tests and confirm they fail because the resource and metadata are absent.

## 2. Notes app

- [x] 2.1 Add one embedded HTML document: CSP with no external origins, `textContent` rendering, `ui/initialize` then `tools/call` for `list_my_tasks` (`type` note), `get_task`, and `create_note`. No token interpolation and no network URL.
- [x] 2.2 Register the resource and the tool `_meta` from `NewServer`. Advertise the UI extension. Leave `Resources.ListChanged` false. Do not change `ownerClause`, token checks, tool handlers, or limits.
- [x] 2.3 Re-run the tests from section 1 and the node DOM test until they pass.

## 3. Docs

- [x] 3.1 Update `docs/agents/14-morph-mcp.md` and `morph/cmd/morph-mcp/README.md`: Claude Desktop localhost attach with `MORPH_MCP_TOKEN`, `JWT_SECRET`, and `TRAN_SQLITE_PATH`, how to get the token from local Morph login, what the notes panel does, the other renderers named by the MCP Apps overview, and the plain-tools fallback. No real secrets. Update the main spec purpose so it mentions the notes app.

## 4. Verification

- [ ] 4.1 Run `cd morph && go build ./... && go vet ./... && go test -race -count=1 ./...` and the node DOM test. Keep the redacted stdio transcript.
