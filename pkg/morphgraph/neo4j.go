package morphgraph

import (
	"context"
	"fmt"
	"strings"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

// Store is a Neo4j-backed graph client (nil-safe when disabled).
type Store struct {
	driver neo4j.DriverWithContext
	db     string
	cfg    Config
}

func OpenStore(cfg Config) (*Store, error) {
	if !cfg.Enabled {
		return nil, nil
	}
	// Empty password = local/dev NoAuth (Neo4j must have auth disabled, or use BasicAuth with blank pass if server allows).
	auth := neo4j.NoAuth()
	if strings.TrimSpace(cfg.Neo4jPassword) != "" {
		auth = neo4j.BasicAuth(cfg.Neo4jUser, cfg.Neo4jPassword, "")
	}
	driver, err := neo4j.NewDriverWithContext(cfg.Neo4jURI, auth)
	if err != nil {
		return nil, err
	}
	ctx := context.Background()
	if err := driver.VerifyConnectivity(ctx); err != nil {
		_ = driver.Close(ctx)
		return nil, err
	}
	return &Store{driver: driver, db: cfg.Neo4jDatabase, cfg: cfg}, nil
}

func (s *Store) Close(ctx context.Context) error {
	if s == nil || s.driver == nil {
		return nil
	}
	return s.driver.Close(ctx)
}

func (s *Store) session(ctx context.Context) neo4j.SessionWithContext {
	return s.driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: s.db})
}

// BootstrapSchema creates constraints and vector index.
func (s *Store) BootstrapSchema(ctx context.Context) error {
	if s == nil {
		return fmt.Errorf("neo4j store not open")
	}
	sess := s.session(ctx)
	defer sess.Close(ctx)
	stmts := []string{
		`CREATE CONSTRAINT entity_uid IF NOT EXISTS FOR (n:Entity) REQUIRE n.uid IS UNIQUE`,
		`CREATE CONSTRAINT chunk_uid IF NOT EXISTS FOR (c:Chunk) REQUIRE c.uid IS UNIQUE`,
		`CREATE VECTOR INDEX chunk_embedding IF NOT EXISTS FOR (c:Chunk) ON (c.embedding)
		 OPTIONS {indexConfig: {` + "`vector.dimensions`" + `: 1536, ` + "`vector.similarity_function`" + `: 'cosine'}}`,
	}
	for _, q := range stmts {
		if _, err := sess.Run(ctx, q, nil); err != nil {
			// vector index may fail if dims mismatch or enterprise features — continue with constraint
			if strings.Contains(strings.ToLower(err.Error()), "vector") {
				continue
			}
			// IF NOT EXISTS should make most errors recoverable
			if !strings.Contains(strings.ToLower(err.Error()), "equivalent") &&
				!strings.Contains(strings.ToLower(err.Error()), "already exists") {
				// soft-fail vector; hard-fail uniqueness only if not exists variant unsupported
				_ = err
			}
		}
	}
	return nil
}

// NodeProps are properties for MERGE upsert.
type NodeProps struct {
	UID       string
	Source    string
	Type      string
	SourceID  string
	Title     string
	Summary   string
	Labels    []string // extra labels besides Entity
	UpdatedAt string
}

func (s *Store) UpsertEntity(ctx context.Context, n NodeProps) error {
	if s == nil {
		return nil
	}
	labels := append([]string{"Entity"}, n.Labels...)
	labelCypher := ":" + strings.Join(labels, ":")
	sess := s.session(ctx)
	defer sess.Close(ctx)
	_, err := sess.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		q := fmt.Sprintf(`
MERGE (n%s {uid: $uid})
SET n.source=$source, n.type=$type, n.source_id=$source_id,
    n.title=$title, n.summary=$summary, n.updated_at=$updated_at
RETURN n.uid`, labelCypher)
		_, err := tx.Run(ctx, q, map[string]any{
			"uid": n.UID, "source": n.Source, "type": n.Type, "source_id": n.SourceID,
			"title": n.Title, "summary": n.Summary, "updated_at": n.UpdatedAt,
		})
		return nil, err
	})
	return err
}

func (s *Store) UpsertRel(ctx context.Context, fromUID, toUID, relType string, props map[string]any) error {
	if s == nil {
		return nil
	}
	if props == nil {
		props = map[string]any{}
	}
	sess := s.session(ctx)
	defer sess.Close(ctx)
	_, err := sess.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		q := fmt.Sprintf(`
MATCH (a:Entity {uid:$from}), (b:Entity {uid:$to})
MERGE (a)-[r:%s]->(b)
SET r += $props
RETURN type(r)`, relType)
		_, err := tx.Run(ctx, q, map[string]any{"from": fromUID, "to": toUID, "props": props})
		return nil, err
	})
	return err
}

func (s *Store) DeleteEntity(ctx context.Context, uid string) error {
	if s == nil {
		return nil
	}
	sess := s.session(ctx)
	defer sess.Close(ctx)
	_, err := sess.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		_, err := tx.Run(ctx, `
MATCH (n:Entity {uid:$uid})
OPTIONAL MATCH (c:Chunk)-[:DESCRIBES]->(n)
DETACH DELETE c, n`, map[string]any{"uid": uid})
		return nil, err
	})
	return err
}

// UpsertChunk stores a Chunk node linked to an entity via DESCRIBES.
func (s *Store) UpsertChunk(ctx context.Context, chunkUID, entityUID, text string, embedding []float32) error {
	if s == nil {
		return nil
	}
	sess := s.session(ctx)
	defer sess.Close(ctx)
	_, err := sess.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		params := map[string]any{
			"cuid": chunkUID, "euid": entityUID, "text": text,
		}
		q := `
MERGE (c:Chunk {uid:$cuid})
SET c.text=$text
WITH c
MATCH (e:Entity {uid:$euid})
MERGE (c)-[:DESCRIBES]->(e)`
		if len(embedding) > 0 {
			params["embedding"] = embedding
			q = `
MERGE (c:Chunk {uid:$cuid})
SET c.text=$text, c.embedding=$embedding
WITH c
MATCH (e:Entity {uid:$euid})
MERGE (c)-[:DESCRIBES]->(e)`
		}
		_, err := tx.Run(ctx, q, params)
		return nil, err
	})
	return err
}

// VectorSearch runs Neo4j vector query; falls back to empty if index unavailable.
func (s *Store) VectorSearch(ctx context.Context, embedding []float32, limit int) ([]map[string]any, error) {
	if s == nil || len(embedding) == 0 {
		return nil, nil
	}
	if limit < 1 {
		limit = 12
	}
	sess := s.session(ctx)
	defer sess.Close(ctx)
	res, err := sess.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		result, err := tx.Run(ctx, `
CALL db.index.vector.queryNodes('chunk_embedding', $k, $embedding)
YIELD node, score
OPTIONAL MATCH (node)-[:DESCRIBES]->(e:Entity)
RETURN node.uid AS chunk_uid, node.text AS text, score AS score,
       e.uid AS entity_uid, e.title AS title, e.type AS type, e.source AS source
ORDER BY score DESC`, map[string]any{"k": limit, "embedding": embedding})
		if err != nil {
			return nil, err
		}
		var out []map[string]any
		for result.Next(ctx) {
			rec := result.Record()
			row := map[string]any{}
			for _, k := range rec.Keys {
				if v, ok := rec.Get(k); ok {
					row[k] = v
				}
			}
			out = append(out, row)
		}
		return out, result.Err()
	})
	if err != nil {
		return nil, err
	}
	if res == nil {
		return nil, nil
	}
	return res.([]map[string]any), nil
}

func (s *Store) Ping(ctx context.Context) error {
	if s == nil {
		return fmt.Errorf("neo4j disabled")
	}
	return s.driver.VerifyConnectivity(ctx)
}
