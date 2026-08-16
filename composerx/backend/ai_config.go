package main

import (
	"encoding/json"
	"log"
	"os"
	"strings"

	openai "github.com/sashabaranov/go-openai"
)

const (
	defaultQwenBaseURL   = "https://dashscope.aliyuncs.com/compatible-mode/v1"
	defaultQwenChatModel = "qwen3-max"
)

type aiConfigFile struct {
	QwenAPIKey       string `json:"qwen_api_key"`
	QwenBaseURL      string `json:"qwen_base_url"`
	QwenDefaultModel string `json:"qwen_default_model"`
}

// mergeMorphAIEnv applies unified MorphAI env vars (MORPH_AI_*), with TRAN_QWEN_* legacy fallbacks.
func mergeMorphAIEnv(cfg *Config) {
	if s := strings.TrimSpace(os.Getenv("MORPH_AI_API_KEY")); s != "" {
		cfg.QwenAPIKey = s
	} else if s := strings.TrimSpace(os.Getenv("TRAN_QWEN_API_KEY")); s != "" {
		cfg.QwenAPIKey = s
	}
	if s := strings.TrimSpace(os.Getenv("MORPH_AI_BASE_URL")); s != "" {
		cfg.QwenBaseURL = strings.TrimRight(s, "/")
	} else if s := strings.TrimSpace(os.Getenv("TRAN_QWEN_BASE_URL")); s != "" {
		cfg.QwenBaseURL = strings.TrimRight(s, "/")
	}
	if s := strings.TrimSpace(os.Getenv("MORPH_AI_MODEL")); s != "" {
		cfg.QwenDefaultModel = s
	} else if s := strings.TrimSpace(os.Getenv("TRAN_QWEN_MODEL")); s != "" {
		cfg.QwenDefaultModel = s
	}
}

// mergeAIConfigFromFile loads optional JSON (path from TRAN_AI_CONFIG_PATH, default ai.config.json).
// Missing file is ignored; invalid JSON is logged. Env vars from mergeMorphAIEnv take precedence.
func mergeAIConfigFromFile(cfg *Config, path string) {
	path = strings.TrimSpace(path)
	if path == "" {
		return
	}
	b, err := os.ReadFile(path)
	if err != nil {
		if !os.IsNotExist(err) {
			log.Printf("ai config: read %s: %v", path, err)
		}
		return
	}
	var f aiConfigFile
	if err := json.Unmarshal(b, &f); err != nil {
		log.Printf("ai config: invalid JSON in %s: %v", path, err)
		return
	}
	if s := strings.TrimSpace(f.QwenAPIKey); s != "" {
		cfg.QwenAPIKey = s
	}
	if s := strings.TrimSpace(f.QwenBaseURL); s != "" {
		cfg.QwenBaseURL = strings.TrimRight(s, "/")
	}
	if s := strings.TrimSpace(f.QwenDefaultModel); s != "" {
		cfg.QwenDefaultModel = s
	}
}

func normalizeQwenDefaults(cfg *Config) {
	if strings.TrimSpace(cfg.QwenAPIKey) == "" {
		return
	}
	if strings.TrimSpace(cfg.QwenBaseURL) == "" {
		cfg.QwenBaseURL = strings.TrimRight(defaultQwenBaseURL, "/")
	}
	if strings.TrimSpace(cfg.QwenDefaultModel) == "" {
		cfg.QwenDefaultModel = defaultQwenChatModel
	}
}

// chatCompletionClient returns Qwen compatible-mode client when api key file/env is set, else OpenAI.
func (a *App) chatCompletionClient() (*openai.Client, string, bool) {
	if k := strings.TrimSpace(a.cfg.QwenAPIKey); k != "" {
		c := openai.DefaultConfig(k)
		base := strings.TrimSpace(a.cfg.QwenBaseURL)
		if base == "" {
			base = strings.TrimRight(defaultQwenBaseURL, "/")
		} else {
			base = strings.TrimRight(base, "/")
		}
		c.BaseURL = base
		client := openai.NewClientWithConfig(c)
		model := strings.TrimSpace(a.cfg.QwenDefaultModel)
		if model == "" {
			model = defaultQwenChatModel
		}
		return client, model, true
	}
	if k := strings.TrimSpace(a.cfg.OpenAIAPIKey); k != "" {
		model := strings.TrimSpace(a.cfg.OpenAIChatModel)
		if model == "" {
			model = openai.GPT4oMini
		}
		return openai.NewClient(k), model, true
	}
	return nil, "", false
}
