# 10 — Shared Libraries (`pkg/`)

Go modules use `replace` in `go.mod`. Rust uses a path dependency on `pkg/morphai-rs`.

```
pkg/
├── morphai/          Go DashScope client
├── morphai-rs/       Rust DashScope client
├── repoenv/          Load repo-root `.env` (walk up to start-all.sh)
├── assistmd/         Markdown helpers
├── docextract/       PDF/text extract for prompts
├── morphgraph/       GraphRAG types + Neo4j helpers
└── webresearch/      Wikipedia, DuckDuckGo, arXiv
```

## `pkg/morphai` (Go)

**Module:** `github.com/robo/morphai`

**Used by:** morph, formx, composerx.

```go
client := morphai.NewClientFromEnv()
response, err := client.ChatCompletion(ctx, messages)
```

Env: `MORPH_AI_API_KEY`, `MORPH_AI_MODEL` (default `qwen3-max`), `MORPH_AI_API_URL` / `MORPH_AI_BASE_URL`. See [`pkg/morphai/README.md`](../../pkg/morphai/README.md).

## `pkg/morphai-rs` (Rust)

**Used by:** Data Access (`SharpReport`), Project (`morph-engi`). Same `MORPH_AI_*` keys.

## `pkg/repoenv`

Call `repoenv.Load()` at process start so nested working directories still read the **repo-root** `.env`. Nested leftover `.env` files must not blank shared keys.

## `pkg/morphgraph`

Optional Neo4j types/ops for GraphRAG. Worker: `morphgraph-worker/`. See [`docs/MORPH_GRAPH_OPS.md`](../MORPH_GRAPH_OPS.md).
