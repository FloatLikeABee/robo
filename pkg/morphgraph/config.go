package morphgraph

import (
	"os"
	"strconv"
	"strings"
)

// Config for GraphRAG / Neo4j / embeddings.
type Config struct {
	Enabled        bool
	Neo4jURI       string
	Neo4jUser      string
	Neo4jPassword  string
	Neo4jDatabase  string
	EmbeddingModel string
	OpenAIAPIKey   string
	OpenAIBaseURL  string
}

func LoadFromEnv() Config {
	enabled := strings.EqualFold(strings.TrimSpace(os.Getenv("MORPH_GRAPH_ENABLED")), "true") ||
		os.Getenv("MORPH_GRAPH_ENABLED") == "1"
	return Config{
		Enabled:        enabled,
		Neo4jURI:       firstEnv("NEO4J_URI", "neo4j://127.0.0.1:7687"),
		Neo4jUser:      firstEnv("NEO4J_USER", "neo4j"),
		Neo4jPassword:  os.Getenv("NEO4J_PASSWORD"),
		Neo4jDatabase:  firstEnv("NEO4J_DATABASE", "neo4j"),
		EmbeddingModel: firstEnv("MORPH_GRAPH_EMBEDDING_MODEL", "text-embedding-3-small"),
		OpenAIAPIKey: firstNonEmpty(
			os.Getenv("TRAN_OPENAI_API_KEY"),
			os.Getenv("MORPH_AI_API_KEY"),
			os.Getenv("OPENAI_API_KEY"),
		),
		OpenAIBaseURL: firstNonEmpty(
			os.Getenv("TRAN_OPENAI_BASE_URL"),
			os.Getenv("MORPH_AI_BASE_URL"),
			"https://dashscope.aliyuncs.com/compatible-mode/v1",
		),
	}
}

func firstEnv(key, def string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return def
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}

// UID builds a stable graph node key: source:type:id
func UID(source, typ, id string) string {
	return strings.ToLower(strings.TrimSpace(source)) + ":" +
		strings.ToLower(strings.TrimSpace(typ)) + ":" +
		strings.TrimSpace(id)
}

func EnvBool(key string, def bool) bool {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return def
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return def
	}
	return b
}
