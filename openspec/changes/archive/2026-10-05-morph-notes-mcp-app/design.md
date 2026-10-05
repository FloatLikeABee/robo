## Context

See proposal.md for why. Code and spec checked before this design:

- `morph/mcp.NewServer` registers `whoami`, `list_my_tasks`, `get_task`, and `create_note` on `github.com/modelcontextprotocol/go-sdk` v1.8.0. Capabilities set `Resources.ListChanged` false and leave the resource list empty. Every tool call runs `DecodeToken` again. Reads use the mode=ro pool. `create_note` uses a separate mode=rw connection and `ownerClause`. `mcp.ExposeRecord` is not an owner check and is not called.
- That SDK version has `Tool.Meta` and `Resource.Meta` (`_meta`), `Server.AddResource`, `ResourceContents.Text` / `MIMEType`, and `ServerCapabilities.AddExtension`. `AddResource` notifies `resources/list_changed` only when `Resources.ListChanged` is true. `url.Parse("ui://morph/notes")` yields a scheme, so the SDK accepts the URI.
- MCP Apps (SEP-1865, stable spec 2026-01-26, extension `io.modelcontextprotocol/ui`) is the official app surface. A tool sets `_meta.ui.resourceUri` to a `ui://` resource. `resources/read` returns HTML with MIME `text/html;profile=mcp-app`. The host renders it in a sandboxed iframe. The page talks JSON-RPC over `postMessage` (`ui/initialize`, then `tools/call`). Hosts that do not negotiate the extension ignore `_meta.ui` and show the normal tool result.
- The official MCP Apps overview, fetched for this change, lists current renderers as Claude, Claude Desktop, VS Code GitHub Copilot, Microsoft 365 Copilot, Goose, Postman, MCPJam, and Archestra.AI. The extension matrix tracks more clients, including Cursor and ChatGPT, and does not put them on that supported list. ChatGPT was reported elsewhere as missing UI-initiated `tools/call`.

## Goals / Non-Goals

**Goals:**

- An MCP-Apps host that loads this server can open one notes panel and create, list, and get notes through it.
- The panel is one static HTML document: no network, no token, text-only note rendering.
- Clients that ignore MCP Apps keep today's tool results.
- `resources.listChanged` stays false. Owner checks and auth stay where they are.

**Non-Goals:**

- Editing or deleting notes, todos in the panel, or a second app for `whoami`.
- A new auth scheme, HTTP transport, Render, the Morph SPA, skills, or `pkg/morphai`.
- Per-session tool lists that hide `_meta` from non-app clients.
- Bundling `@modelcontextprotocol/ext-apps` or any npm package into the server.

## Decisions

### 1. The app is one `ui://` resource on the existing server

`resources/list` returns `ui://morph/notes`. `resources/read` returns the embedded HTML with MIME `text/html;profile=mcp-app`. `list_my_tasks`, `get_task`, and `create_note` set `_meta.ui.resourceUri` to that URI and `_meta.ui.visibility` to `["model", "app"]`. The deprecated flat key `_meta["ui/resourceUri"]` is set to the same URI so a host that still reads the pre-nested key finds the page. `whoami` has no UI meta. Initialize adds extension `io.modelcontextprotocol/ui` with `mimeTypes: ["text/html;profile=mcp-app"]`. `Resources.ListChanged` stays false, which also suppresses the SDK's list-changed notification when the resource is registered.

The page, after `ui/initialize` and `ui/notifications/initialized`, calls `list_my_tasks` with `type: "note"`, renders that page, loads one row with `get_task`, and submits the form with `create_note`. It also paints a `ui/notifications/tool-result` the host pushes for the tool that opened the view. Note titles and bodies are `textContent` only.

Alternatives:

- (b) A documented client config or deeplink and no UI resource. Rejected. An MCP-Apps host loads a `ui://` document linked from tool `_meta`. Another `mcp.json` block is the raw tool wiring this story is adding a surface on top of. It does not make create, list, and get available *through an app*.
- (c) The UI resource plus a separate MCP bundle or manifest. Rejected. SEP-1865 does not require a bundle. The stdio server is the app entry. A second file would not be what Claude Desktop or VS Code fetch.
- Return HTML inside the tool result (older MCP-UI embedded resource) instead of a predeclared `ui://` resource. Rejected. The stable spec requires the resource to be registered and referenced from `_meta.ui.resourceUri`, and the MIME type is `text/html;profile=mcp-app`, not `text/html+mcp`.
- Point the resource at the human SPA on `:3031`. Rejected. Out of scope, needs a network origin, and would put the session in a browser context this server does not control.

Failure mode: a host that only reads the flat key still works because both keys are set. A host that only reads the nested key works too. A host that renders neither still gets the tool's structured content, because the handlers are unchanged.

### 2. The HTML is static, closed, and does not see the token

The document is `//go:embed` of one file. Nothing is interpolated from the environment. Its CSP meta tag is `default-src 'none'`, inline script and style only, `connect-src 'none'`, `frame-src 'none'`, `object-src 'none'`, `base-uri 'none'`. Resource `_meta.ui.csp` on both the listed resource and the read content sets empty `connectDomains`, `resourceDomains`, and `frameDomains`, and sets `prefersBorder` true. No camera, microphone, or clipboard permission is requested. Writes go through `create_note`, so the owner subquery and the token recheck stay in the existing tool.

`resources/read` does not re-run `DecodeToken`. The document has no note rows. The next `tools/call` still rechecks the token inside the existing handlers.

Alternatives:

- Load the ext-apps `App` class from a CDN. Rejected. The spec allows a hand-written `postMessage` client, and a CDN URL breaks the closed-document rule.
- Build the page with the Morph CRA app. Rejected. Out of scope, and it pulls a frontend toolchain into this server.
- Register the UI tools only after the client advertises the extension. Rejected. This SDK registers tools once, before initialize, on a server-global list. A second list handler would duplicate the SDK to hide metadata that non-app clients already ignore. Those clients must keep all four tools and their current results.

Failure modes:

- A note title or body contains `<img onerror>` or `<script>`. `textContent` stores that as text. A node test renders those strings and asserts the DOM text is the raw string and that no element was created from the markup. The Go test rejects `innerHTML`, `insertAdjacentHTML`, `outerHTML`, and `document.write` in the file. The URL check rejects `http:` and `https:` (and `url(`). It allows the CSP `http-equiv` attribute, which is not a network URL.
- Someone later concatenates the token into the HTML. The read test uses a live token and asserts the document does not contain it.
- A host applies CSP from `_meta.ui.csp` and blocks the inline script. Empty `resourceDomains` plus the spec's default `script-src 'unsafe-inline'` still allows the inline script. The document does not need a network fetch.

### 3. Plain-tool behavior stays on the same handlers

No new tool names. No change to `ownerClause`, limits, the `[morph-mcp]` prefix, the idempotent title+body check, or the read/write pools. `create_note` is still not read-only. The other three tools stay read-only. The panel's `tools/call` messages are ordinary tool calls, so a deleted or expired token fails inside the handler that already does that, and the error text still omits the token.

### 4. Docs name Claude Desktop for the app and keep Cursor as the plain-tools path

Claude Desktop is on the official supported-renderer list and already uses the `mcpServers` stdio shape this repo documents. The doc adds that block again under an MCP Apps heading, with `MORPH_MCP_TOKEN`, `JWT_SECRET`, and `TRAN_SQLITE_PATH`, and the existing login `curl` against local Morph (`http://127.0.0.1:9090/api/auth/login` or `http://localhost:3031/api/auth/login`). It states the panel lists notes, opens one, and creates one. It states that a client which does not render the extension, including Cursor with the existing snippet, still gets the tool results. Examples stay placeholders.

Other renderers named by the same overview page (Claude on the web, VS Code GitHub Copilot, Microsoft 365 Copilot, Goose, Postman, MCPJam, Archestra.AI) are listed as the same stdio server, without a second config dialect. Cursor and ChatGPT are not described as MCP Apps renderers.

## Review

Proposer: ship only the Claude Desktop config and call that the app entry. Reviewer: a config file does not satisfy "an MCP-app-capable client loads the Morph notes app." The host loads `ui://morph/notes` from tool metadata. The config only starts the process. Keep the resource. The config documents how to start it.

Proposer: omit the deprecated flat `_meta["ui/resourceUri"]`. Reviewer: the stable spec still documents that key as the compatibility read path, and at least one shipped host example checks only the flat key. Writing both keys is two fields. A host that understands only the nested key is unaffected.

Proposer: hide UI metadata unless the client sent the extension during initialize. Reviewer: tools are registered before initialize, on one server. Stripping `_meta` per session means a custom `tools/list`. Non-app clients already ignore unknown `_meta` and still need the structured result. Register the metadata always. The plain-result test calls the tools without reading the resource.

Proposer: let the page set `innerHTML` and escape with a replace. Reviewer: a missed character is a script in the host's iframe. Note text is untrusted. `textContent` cannot parse markup. The node test fails if a title containing `<script>` shows up as an element.

Proposer: recheck the JWT inside `resources/read`. Reviewer: the document is the same for every user and contains no rows. Rechecking there does not change who can create a note. The tool handlers already recheck. Leave `DecodeToken` and `ownerClause` alone.

Proposer: list todos in the panel too. Reviewer: the story is notes. `list_my_tasks` without a type returns todos, and the model can still call it that way. The panel asks for `type: "note"`. A host-pushed `get_task` result for a todo still renders that one row as text, because the detail view follows the task object shape, not the list filter.

Three failure modes that stay tested: the resource MIME and closed HTML (no external URL, no HTML sink); UI `_meta` on the three note tools and not on `whoami`; a tool call that never reads the resource still returns note data, not the HTML page.

## Risks / Trade-offs

- [Host never sends `ui/notifications/tool-result`] → The page calls `list_my_tasks` itself after initialize, so the list does not depend on the opening tool.
- [Iframe clipped with no size notification] → The page sends `ui/notifications/size-changed` from a `ResizeObserver` when that API exists. A missing API does not block the tool calls.
- [`postMessage` target `*`] → The spec's own non-SDK example uses `*`, because the sandbox origin is chosen by the host and is not known to the page. The page posts only to `window.parent`.
- [Deprecated meta key removed by a future host] → The nested `ui.resourceUri` remains the primary field.
- [Claude web connector directory] → Not required for local stdio. The doc is a localhost `mcpServers` entry, not a directory submission.
- [Marker and owner rules drift] → This change does not edit `tasks.go`.

## Migration Plan

No schema change and no data migration. Rollback is removing the resource, the tool `_meta`, and the extension advertisement. Existing notes stay. Tool names and arguments stay. Operators who already run the Cursor snippet do not change their config. The human UI remains `http://localhost:3031/`.

## Open Questions

None. The surface, the document rules, the capability flag, and the documented client are decided above.
