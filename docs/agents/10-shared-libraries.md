# 10 — Shared Libraries (`pkg/`)

Go modules use `replace` in `go.mod`. Rust uses a path dependency on `pkg/dorphai-rs`.

```
pkg/
├── dorphai/          Go multi-provider client (DashScope by default)
├── dorphai-rs/       Rust DashScope client
├── repoenv/          Load repo-root `.env` (walk up to start-all.sh)
├── assistmd/         Markdown helpers
├── docextract/       PDF/text extract for prompts
├── dorphgraph/       GraphRAG types + Neo4j helpers
└── webresearch/      Wikipedia, DuckDuckGo, arXiv
```

## `pkg/dorphai` (Go)

**Module:** `github.com/robo/dorphai`

**Used by:** dorph, formx, composerx.

```go
client := dorphai.NewClientFromEnv()
response, err := client.ChatCompletion(ctx, messages)
```

Env: `DORPH_AI_API_KEY`, `DORPH_AI_MODEL` (default `qwen3-max`), `DORPH_AI_API_URL` / `DORPH_AI_BASE_URL`. Named providers and their env keys are listed in [`pkg/dorphai/README.md`](../../pkg/dorphai/README.md). A missing named-provider key is an error, not a silent fallback.

## `pkg/dorphai-rs` (Rust)

**Used by:** Data Access (`SharpReport`), Project (`dorph-engi`). Same `DORPH_AI_*` keys.

## `pkg/repoenv`

Call `repoenv.Load()` at process start so nested working directories still read the **repo-root** `.env`. Nested leftover `.env` files must not blank shared keys.

## `pkg/dorphgraph`

Optional Neo4j types/ops for GraphRAG. Worker: `dorphgraph-worker/`. See [`docs/DORPH_GRAPH_OPS.md`](../DORPH_GRAPH_OPS.md).
