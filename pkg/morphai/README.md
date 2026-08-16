# morphai

Shared MorphAI configuration and DashScope text-generation client for Go platform apps.

## Usage

```go
import "github.com/robo/morphai"

cfg := morphai.LoadFromEnv()
client := morphai.NewClient(cfg)
reply, err := client.ChatCompletion(ctx, []morphai.Message{
    {Role: "user", Content: "Hello"},
})
```

## Environment

| Variable | Default |
|----------|---------|
| `MORPH_AI_API_KEY` | _(required)_ |
| `MORPH_AI_MODEL` | `qwen3-max` |
| `MORPH_AI_API_URL` | DashScope text-generation endpoint |

Legacy: `GEMINI_API_KEY`, `GEMINI_MODEL`, `TRAN_QWEN_*`.

## Monorepo import

Add to consumer `go.mod`:

```
replace github.com/robo/morphai => ../../pkg/morphai
```
