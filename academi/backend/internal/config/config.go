package config

import (
	"os"
	"strings"
	"sync"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Server            ServerConfig
	Database          DatabaseConfig
	JWT               JWTConfig
	AI                AIConfig
	Cache             CacheConfig
	CORS              CORSConfig
	UsersPanelBaseURL string
}

type ServerConfig struct {
	Port       string
	Mode       string
	UploadDir  string
}

type DatabaseConfig struct {
	Path               string
	ValueLogMaxEntries int
}

type JWTConfig struct {
	Secret      string
	ExpiryHours int
}

type AIProvider struct {
	Name        string
	APIKey      string
	BaseURL     string
	Model       string
	MaxTokens   int
	Temperature float64
}

type AIConfig struct {
	DefaultProvider string
	Providers      map[string]AIProvider
}

type CacheConfig struct {
	DefaultTTL int
}

type CORSConfig struct {
	Origins []string
}

var (
	instance *Config
	once     sync.Once
)

func Load() *Config {
	once.Do(func() {
		godotenv.Load()
		instance = &Config{
			Server: ServerConfig{
				Port:      getEnv("SERVER_PORT", "8080"),
				Mode:      getEnv("SERVER_MODE", "debug"),
				UploadDir: getEnv("UPLOAD_DIR", "./data/uploads"),
			},
			Database: DatabaseConfig{
				Path:               getEnv("DB_PATH", "./data/badger"),
				ValueLogMaxEntries: 1000000,
			},
			JWT: JWTConfig{
				Secret:      getEnv("JWT_SECRET", "academi-secret-key"),
				ExpiryHours: 24,
			},
			AI: AIConfig{
				DefaultProvider: getEnv("AI_DEFAULT_PROVIDER", "openai"),
				Providers: map[string]AIProvider{
					"openai": {
						Name:        "OpenAI",
						APIKey:      getEnv("OPENAI_API_KEY", ""),
						BaseURL:     getEnv("OPENAI_BASE_URL", "https://api.openai.com/v1"),
						Model:       getEnv("OPENAI_MODEL", "gpt-4"),
						MaxTokens:   2048,
						Temperature: 0.7,
					},
					"anthropic": {
						Name:        "Anthropic",
						APIKey:      getEnv("ANTHROPIC_API_KEY", ""),
						BaseURL:     getEnv("ANTHROPIC_BASE_URL", "https://api.anthropic.com/v1"),
						Model:       getEnv("ANTHROPIC_MODEL", "claude-3-opus-20240229"),
						MaxTokens:   4096,
						Temperature: 0.7,
					},
					"azure": {
						Name:        "Azure OpenAI",
						APIKey:      getEnv("AZURE_OPENAI_API_KEY", ""),
						BaseURL:     getEnv("AZURE_OPENAI_BASE_URL", ""),
						Model:       getEnv("AZURE_OPENAI_MODEL", "gpt-4"),
						MaxTokens:   2048,
						Temperature: 0.7,
					},
					"ollama": {
						Name:        "Ollama",
						APIKey:      getEnv("OLLAMA_API_KEY", ""),
						BaseURL:     getEnv("OLLAMA_BASE_URL", "http://localhost:11434/v1"),
						Model:       getEnv("OLLAMA_MODEL", "llama2"),
						MaxTokens:   2048,
						Temperature: 0.7,
					},
					"qwen": {
						Name:        "Qwen",
						APIKey:      getEnv("QWEN_API_KEY", ""),
						BaseURL:     getEnv("QWEN_BASE_URL", "https://dashscope.aliyuncs.com/compatible-mode/v1"),
						Model:       getEnv("QWEN_MODEL", "qwen-turbo"),
						MaxTokens:   2048,
						Temperature: 0.7,
					},
					// SiliconFlow (OpenAI-compatible). Uses SILICONFLOW_* or OPENAI_* env vars.
					"siliconflow": {
						Name:        "SiliconFlow",
						APIKey:      firstEnv("SILICONFLOW_API_KEY", "OPENAI_API_KEY"),
						BaseURL:     envDefault("https://api.siliconflow.cn/v1", "SILICONFLOW_BASE_URL", "OPENAI_BASE_URL"),
						Model:       envDefault("Qwen/Qwen2.5-7B-Instruct", "SILICONFLOW_MODEL", "OPENAI_MODEL"),
						MaxTokens:   2048,
						Temperature: 0.7,
					},
				},
			},
			Cache: CacheConfig{
				DefaultTTL: 3600,
			},
			CORS: CORSConfig{
				Origins: []string{"*"},
			},
			UsersPanelBaseURL: getEnv("USERS_PANEL_BASE_URL", "http://127.0.0.1:5001"),
		}
	})
	return instance
}

func getEnv(key, fallback string) string {
	if val := strings.TrimSpace(os.Getenv(key)); val != "" {
		return val
	}
	return fallback
}

// firstEnv returns the first non-empty trimmed value from the given env var names.
func firstEnv(keys ...string) string {
	for _, k := range keys {
		if v := strings.TrimSpace(os.Getenv(k)); v != "" {
			return v
		}
	}
	return ""
}

// envDefault tries each env var in order, then returns fallback if all are unset.
func envDefault(fallback string, keys ...string) string {
	if v := firstEnv(keys...); v != "" {
		return v
	}
	return fallback
}

func (c *JWTConfig) ExpiryDuration() time.Duration {
	return time.Duration(c.ExpiryHours) * time.Hour
}
