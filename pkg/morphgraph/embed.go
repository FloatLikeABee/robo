package morphgraph

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Embedder creates embedding vectors via an OpenAI-compatible API.
type Embedder struct {
	APIKey  string
	BaseURL string
	Model   string
	Client  *http.Client
}

func NewEmbedder(cfg Config) *Embedder {
	base := strings.TrimRight(cfg.OpenAIBaseURL, "/")
	return &Embedder{
		APIKey:  cfg.OpenAIAPIKey,
		BaseURL: base,
		Model:   cfg.EmbeddingModel,
		Client:  &http.Client{Timeout: 45 * time.Second},
	}
}

func (e *Embedder) Configured() bool {
	return e != nil && strings.TrimSpace(e.APIKey) != ""
}

type embedRequest struct {
	Model string   `json:"model"`
	Input []string `json:"input"`
}

type embedResponse struct {
	Data []struct {
		Embedding []float32 `json:"embedding"`
		Index     int       `json:"index"`
	} `json:"data"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

// Embed returns one vector per input string.
func (e *Embedder) Embed(ctx context.Context, inputs []string) ([][]float32, error) {
	if !e.Configured() {
		return nil, fmt.Errorf("embeddings not configured (set TRAN_OPENAI_API_KEY or MORPH_AI_API_KEY)")
	}
	if len(inputs) == 0 {
		return nil, nil
	}
	body, _ := json.Marshal(embedRequest{Model: e.Model, Input: inputs})
	url := e.BaseURL + "/embeddings"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+e.APIKey)
	res, err := e.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(res.Body, 8<<20))
	if res.StatusCode >= 300 {
		return nil, fmt.Errorf("embeddings HTTP %d: %s", res.StatusCode, TruncateRunes(string(raw), 400))
	}
	var parsed embedResponse
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return nil, err
	}
	if parsed.Error != nil && parsed.Error.Message != "" {
		return nil, fmt.Errorf("embeddings: %s", parsed.Error.Message)
	}
	out := make([][]float32, len(inputs))
	for _, d := range parsed.Data {
		if d.Index >= 0 && d.Index < len(out) {
			out[d.Index] = d.Embedding
		}
	}
	for i, v := range out {
		if v == nil {
			return nil, fmt.Errorf("missing embedding for input %d", i)
		}
	}
	return out, nil
}
