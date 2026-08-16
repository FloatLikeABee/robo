package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
	"github.com/robo/morphgraph"
)

func main() {
	_ = godotenv.Load()
	_ = godotenv.Load(".env")
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}
	cmd := os.Args[1]
	cfg := morphgraph.LoadFromEnv()
	dsn := firstNonEmpty(os.Getenv("TRAN_MYSQL_DSN"), os.Getenv("DATABASE_URL"))
	if dsn == "" {
		log.Fatal("TRAN_MYSQL_DSN required")
	}
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	var store *morphgraph.Store
	if cfg.Enabled {
		store, err = morphgraph.OpenStore(cfg)
		if err != nil {
			log.Fatalf("neo4j: %v", err)
		}
		defer store.Close(ctx)
	}

	switch cmd {
	case "bootstrap-schema":
		if store == nil {
			log.Fatal("set MORPH_GRAPH_ENABLED=true and Neo4j credentials")
		}
		if err := store.BootstrapSchema(ctx); err != nil {
			log.Fatal(err)
		}
		log.Println("schema bootstrap done")
	case "run":
		runWorker(ctx, db, store, cfg)
	case "backfill":
		source := "all"
		for _, a := range os.Args[2:] {
			if strings.HasPrefix(a, "--source=") {
				source = strings.TrimPrefix(a, "--source=")
			}
			if a == "--all" {
				source = "all"
			}
		}
		if err := backfill(ctx, db, store, cfg, source); err != nil {
			log.Fatal(err)
		}
		log.Println("backfill done")
	case "sync":
		mode := "daily"
		for _, a := range os.Args[2:] {
			if strings.HasPrefix(a, "--mode=") {
				mode = strings.TrimPrefix(a, "--mode=")
			}
		}
		if err := processOutbox(ctx, db, store, cfg, 500); err != nil {
			log.Printf("outbox: %v", err)
		}
		if mode == "daily" || mode == "full" {
			if err := backfill(ctx, db, store, cfg, "all"); err != nil {
				log.Fatal(err)
			}
		}
		log.Println("sync done")
	case "status":
		var pending int64
		_ = db.QueryRow(`SELECT COUNT(*) FROM graph_sync_outbox WHERE processed_at IS NULL`).Scan(&pending)
		fmt.Printf("outbox_pending=%d neo4j_enabled=%v\n", pending, store != nil)
	default:
		usage()
		os.Exit(2)
	}
}

func usage() {
	fmt.Println(`morphgraph-worker commands:
  bootstrap-schema
  run                 # poll outbox forever
  backfill [--source=morph|formsx|composerx|knowledge|all]
  sync --mode=daily   # drain outbox + full backfill
  status`)
}

func runWorker(ctx context.Context, db *sql.DB, store *morphgraph.Store, cfg morphgraph.Config) {
	log.Println("morphgraph-worker running (outbox poll 3s)")
	for {
		if err := processOutbox(ctx, db, store, cfg, 50); err != nil {
			log.Printf("outbox: %v", err)
		}
		time.Sleep(3 * time.Second)
	}
}

func processOutbox(ctx context.Context, db *sql.DB, store *morphgraph.Store, cfg morphgraph.Config, limit int) error {
	rows, err := db.QueryContext(ctx, `
		SELECT id, source, entity_type, entity_id, op
		FROM graph_sync_outbox
		WHERE processed_at IS NULL AND available_at <= NOW(3)
		ORDER BY id ASC LIMIT ?`, limit)
	if err != nil {
		return err
	}
	defer rows.Close()
	type item struct {
		id                     int64
		source, etype, eid, op string
	}
	var items []item
	for rows.Next() {
		var it item
		if err := rows.Scan(&it.id, &it.source, &it.etype, &it.eid, &it.op); err != nil {
			return err
		}
		items = append(items, it)
	}
	for _, it := range items {
		err := applyOutbox(ctx, db, store, cfg, it.source, it.etype, it.eid, it.op)
		if err != nil {
			_, _ = db.ExecContext(ctx, `
				UPDATE graph_sync_outbox SET attempts=attempts+1, last_error=?, available_at=DATE_ADD(NOW(3), INTERVAL LEAST(attempts+1,30) MINUTE)
				WHERE id=?`, truncate(err.Error(), 500), it.id)
			continue
		}
		_, _ = db.ExecContext(ctx, `UPDATE graph_sync_outbox SET processed_at=NOW(3), last_error=NULL WHERE id=?`, it.id)
	}
	return nil
}

func applyOutbox(ctx context.Context, db *sql.DB, store *morphgraph.Store, cfg morphgraph.Config, source, etype, eid, op string) error {
	uid := morphgraph.UID(source, etype, eid)
	if op == "delete" {
		if store != nil {
			return store.DeleteEntity(ctx, uid)
		}
		return nil
	}
	switch {
	case source == "morph" && etype == "knowledge_file":
		return syncKnowledgeFile(ctx, db, store, cfg, eid)
	case source == "morph":
		return syncMorphEntity(ctx, db, store, etype, eid)
	case source == "formsx" && etype == "form":
		return syncForm(ctx, db, store, eid)
	case source == "composerx" && etype == "email_template":
		return syncEmailTemplate(ctx, db, store, eid)
	default:
		// best-effort generic node
		if store == nil {
			return nil
		}
		return store.UpsertEntity(ctx, morphgraph.NodeProps{
			UID: uid, Source: source, Type: etype, SourceID: eid,
			Title: etype + " " + eid, UpdatedAt: time.Now().UTC().Format(time.RFC3339),
		})
	}
}

func backfill(ctx context.Context, db *sql.DB, store *morphgraph.Store, cfg morphgraph.Config, source string) error {
	if source == "all" || source == "morph" {
		log.Println("backfill morph…")
		if err := backfillMorph(ctx, db, store); err != nil {
			return err
		}
	}
	if source == "all" || source == "formsx" {
		log.Println("backfill formsx…")
		if err := backfillForms(ctx, db, store); err != nil {
			return err
		}
	}
	if source == "all" || source == "composerx" {
		log.Println("backfill composerx…")
		if err := backfillTemplates(ctx, db, store); err != nil {
			return err
		}
	}
	if source == "all" || source == "knowledge" {
		log.Println("backfill knowledge…")
		if err := backfillKnowledge(ctx, db, store, cfg); err != nil {
			return err
		}
	}
	return nil
}

func backfillMorph(ctx context.Context, db *sql.DB, store *morphgraph.Store) error {
	type row struct {
		table, typ, idCol, titleExpr string
	}
	jobs := []row{
		{"District", "district", "id", "COALESCE(name, district, CAST(id AS CHAR))"},
		{"facility", "facility", "id", "COALESCE(name, facility_code, CAST(id AS CHAR))"},
		{"member", "member", "id", "TRIM(CONCAT(COALESCE(first_name,''),' ',COALESCE(last_name,'')))"},
		{"employee", "employee", "id", "TRIM(CONCAT(COALESCE(first_name,''),' ',COALESCE(last_name,'')))"},
		{"Contact", "contact", "ID", "TRIM(CONCAT(COALESCE(FirstName,''),' ',COALESCE(LastName,'')))"},
		{"Asset", "asset", "ID", "COALESCE(asset_tag, description, CAST(ID AS CHAR))"},
		{"Activity", "activity", "ID", "COALESCE(name, CAST(ID AS CHAR))"},
		{"CaseTask", "case_task", "ID", "COALESCE(title, CAST(ID AS CHAR))"},
	}
	for _, j := range jobs {
		q := fmt.Sprintf(`SELECT %s, %s FROM %s`, j.idCol, j.titleExpr, j.table)
		rows, err := db.QueryContext(ctx, q)
		if err != nil {
			log.Printf("skip %s: %v", j.table, err)
			continue
		}
		for rows.Next() {
			var id int64
			var title sql.NullString
			if err := rows.Scan(&id, &title); err != nil {
				continue
			}
			_ = upsertNamed(ctx, store, "morph", j.typ, strconv.FormatInt(id, 10), nullStr(title), j.typ)
		}
		rows.Close()
	}
	// relationships: facility -> district
	rows, err := db.QueryContext(ctx, `SELECT id, district_id FROM facility WHERE district_id IS NOT NULL`)
	if err == nil {
		for rows.Next() {
			var fid, did int64
			if rows.Scan(&fid, &did) == nil && store != nil {
				_ = store.UpsertRel(ctx,
					morphgraph.UID("morph", "facility", strconv.FormatInt(fid, 10)),
					morphgraph.UID("morph", "district", strconv.FormatInt(did, 10)),
					"IN_DISTRICT", nil)
			}
		}
		rows.Close()
	}
	return nil
}

func backfillForms(ctx context.Context, db *sql.DB, store *morphgraph.Store) error {
	rows, err := db.QueryContext(ctx, `SELECT id, name, COALESCE(slug,''), COALESCE(description,'') FROM forms`)
	if err != nil {
		log.Printf("forms table: %v", err)
		return nil
	}
	defer rows.Close()
	for rows.Next() {
		var id int64
		var name, slug, desc string
		if err := rows.Scan(&id, &name, &slug, &desc); err != nil {
			continue
		}
		_ = upsertNamed(ctx, store, "formsx", "form", strconv.FormatInt(id, 10), name, desc+" "+slug)
	}
	return nil
}

func backfillTemplates(ctx context.Context, db *sql.DB, store *morphgraph.Store) error {
	rows, err := db.QueryContext(ctx, `SELECT id, name, COALESCE(tag,''), COALESCE(description,'') FROM email_templates`)
	if err != nil {
		log.Printf("email_templates: %v", err)
		return nil
	}
	defer rows.Close()
	for rows.Next() {
		var id int64
		var name, tag, desc string
		if err := rows.Scan(&id, &name, &tag, &desc); err != nil {
			continue
		}
		_ = upsertNamed(ctx, store, "composerx", "email_template", strconv.FormatInt(id, 10), name, tag+" "+desc)
	}
	return nil
}

func backfillKnowledge(ctx context.Context, db *sql.DB, store *morphgraph.Store, cfg morphgraph.Config) error {
	rows, err := db.QueryContext(ctx, `SELECT id FROM morph_knowledge_files`)
	if err != nil {
		return nil
	}
	defer rows.Close()
	for rows.Next() {
		var id int64
		if rows.Scan(&id) == nil {
			_ = syncKnowledgeFile(ctx, db, store, cfg, strconv.FormatInt(id, 10))
		}
	}
	return nil
}

func syncKnowledgeFile(ctx context.Context, db *sql.DB, store *morphgraph.Store, cfg morphgraph.Config, eid string) error {
	id, _ := strconv.ParseInt(eid, 10, 64)
	var title, filename, kind, excerpt string
	err := db.QueryRowContext(ctx, `
		SELECT title, filename, kind, COALESCE(text_excerpt,'') FROM morph_knowledge_files WHERE id=?`, id).
		Scan(&title, &filename, &kind, &excerpt)
	if err != nil {
		return err
	}
	uid := morphgraph.UID("morph", "knowledge_file", eid)
	if store != nil {
		_ = store.UpsertEntity(ctx, morphgraph.NodeProps{
			UID: uid, Source: "morph", Type: "knowledge_file", SourceID: eid,
			Title: title, Summary: morphgraph.TruncateRunes(excerpt, 400),
			Labels: []string{"KnowledgeFile"}, UpdatedAt: time.Now().UTC().Format(time.RFC3339),
		})
	}
	crows, err := db.QueryContext(ctx, `
		SELECT chunk_index, text_content, COALESCE(embedding_json,'') FROM morph_knowledge_chunks WHERE file_id=? ORDER BY chunk_index`, id)
	if err != nil {
		return nil
	}
	defer crows.Close()
	for crows.Next() {
		var idx int
		var text, embJSON string
		if crows.Scan(&idx, &text, &embJSON) != nil {
			continue
		}
		var emb []float32
		if embJSON != "" {
			_ = json.Unmarshal([]byte(embJSON), &emb)
		}
		cuid := morphgraph.UID("morph", "knowledge_chunk", fmt.Sprintf("%d_%d", id, idx))
		if store != nil {
			_ = store.UpsertChunk(ctx, cuid, uid, text, emb)
		}
	}
	return nil
}

func syncMorphEntity(ctx context.Context, db *sql.DB, store *morphgraph.Store, etype, eid string) error {
	return upsertNamed(ctx, store, "morph", etype, eid, etype+" "+eid, "")
}

func syncForm(ctx context.Context, db *sql.DB, store *morphgraph.Store, eid string) error {
	var name, slug, desc string
	err := db.QueryRowContext(ctx, `SELECT name, COALESCE(slug,''), COALESCE(description,'') FROM forms WHERE id=?`, eid).
		Scan(&name, &slug, &desc)
	if err != nil {
		return err
	}
	return upsertNamed(ctx, store, "formsx", "form", eid, name, desc+" "+slug)
}

func syncEmailTemplate(ctx context.Context, db *sql.DB, store *morphgraph.Store, eid string) error {
	var name, tag, desc string
	err := db.QueryRowContext(ctx, `SELECT name, COALESCE(tag,''), COALESCE(description,'') FROM email_templates WHERE id=?`, eid).
		Scan(&name, &tag, &desc)
	if err != nil {
		return err
	}
	return upsertNamed(ctx, store, "composerx", "email_template", eid, name, tag+" "+desc)
}

func upsertNamed(ctx context.Context, store *morphgraph.Store, source, typ, id, title, summary string) error {
	if store == nil {
		return nil
	}
	if strings.TrimSpace(title) == "" {
		title = typ + " " + id
	}
	label := "Node"
	if len(typ) > 0 {
		label = strings.ToUpper(typ[:1]) + typ[1:]
	}
	return store.UpsertEntity(ctx, morphgraph.NodeProps{
		UID: morphgraph.UID(source, typ, id), Source: source, Type: typ, SourceID: id,
		Title: title, Summary: morphgraph.TruncateRunes(summary, 500),
		Labels: []string{label}, UpdatedAt: time.Now().UTC().Format(time.RFC3339),
	})
}

func nullStr(s sql.NullString) string {
	if s.Valid {
		return strings.TrimSpace(s.String)
	}
	return ""
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}
