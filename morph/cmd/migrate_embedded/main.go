// migrate_embedded copies Morph-owned Tran data from MySQL + Mongo into embedded SQLite + Badger.
//
// Source (read-only): TRAN_MYSQL_DSN, TRAN_MONGO_URI, TRAN_MONGO_DB
// Dest: TRAN_SQLITE_PATH, ENTITY_DETAILS_BADGER
//
//	go run ./cmd/migrate_embedded --dry-run
//	go run ./cmd/migrate_embedded --apply
//	go run ./cmd/migrate_embedded --apply --force   # overwrite non-empty dest
//	go run ./cmd/migrate_embedded --verify
package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"idongivaflyinfa/config"
	"idongivaflyinfa/db"

	_ "github.com/go-sql-driver/mysql"
	_ "modernc.org/sqlite"
)

func main() {
	dryRun := flag.Bool("dry-run", false, "report planned copy without writing destination")
	apply := flag.Bool("apply", false, "copy MySQL tables and Mongo entity_details into embedded stores")
	verify := flag.Bool("verify", false, "compare source vs destination row/document counts")
	force := flag.Bool("force", false, "allow apply when destination already has data")
	flag.Parse()

	modes := 0
	if *dryRun {
		modes++
	}
	if *apply {
		modes++
	}
	if *verify {
		modes++
	}
	if modes != 1 {
		log.Fatal("specify exactly one of --dry-run, --apply, or --verify")
	}

	cfg := config.GetConfig()
	if strings.TrimSpace(cfg.TranMySQLDSN) == "" {
		log.Fatal("TRAN_MYSQL_DSN is required (migration source)")
	}
	if strings.TrimSpace(cfg.TranMongoURI) == "" {
		log.Fatal("TRAN_MONGO_URI is required (migration source)")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Hour)
	defer cancel()

	srcSQL, err := openMySQLSource(cfg.TranMySQLDSN)
	if err != nil {
		log.Fatalf("source MySQL: %v", err)
	}
	defer srcSQL.Close()

	srcMongo, err := db.NewTranMongo(cfg.TranMongoURI, cfg.TranMongoDB)
	if err != nil {
		log.Fatalf("source Mongo: %v", err)
	}
	defer func() { _ = srcMongo.Close(context.Background()) }()

	switch {
	case *dryRun:
		if err := runDryRun(ctx, srcSQL, srcMongo, cfg); err != nil {
			log.Fatalf("dry-run: %v", err)
		}
	case *apply:
		if err := runApply(ctx, srcSQL, srcMongo, cfg, *force); err != nil {
			log.Fatalf("apply: %v", err)
		}
	case *verify:
		if err := runVerify(ctx, srcSQL, srcMongo, cfg); err != nil {
			log.Fatalf("verify: %v", err)
		}
	}
}

func openMySQLSource(dsn string) (*sql.DB, error) {
	// Read-only intent: do not run ensureTranMySQLSchema (would mutate source).
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, err
	}
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(2)
	return db, nil
}

func runDryRun(ctx context.Context, srcSQL *sql.DB, srcMongo *db.TranMongo, cfg config.Config) error {
	log.Println("dry-run: no destination writes")
	for _, table := range db.MorphOwnedTablesFKOrder {
		exists, err := mysqlTableExists(ctx, srcSQL, table)
		if err != nil {
			return err
		}
		if !exists {
			log.Printf("  SKIP %s (missing on source)", table)
			continue
		}
		n, err := countRows(ctx, srcSQL, table, true)
		if err != nil {
			return fmt.Errorf("%s count: %w", table, err)
		}
		log.Printf("  COPY %s → sqlite (%d rows)", table, n)
	}
	mongoN, err := srcMongo.CountEntityDetails(ctx)
	if err != nil {
		return fmt.Errorf("mongo count: %w", err)
	}
	log.Printf("  COPY entity_details → badger (%d docs) → %s", mongoN, cfg.EntityDetailsBadger)
	log.Printf("destination would be: sqlite=%s badger=%s", cfg.TranSQLitePath, cfg.EntityDetailsBadger)
	return nil
}

func runApply(ctx context.Context, srcSQL *sql.DB, srcMongo *db.TranMongo, cfg config.Config, force bool) error {
	if !force {
		nonEmpty, reason, err := destNonEmpty(ctx, cfg)
		if err != nil {
			return err
		}
		if nonEmpty {
			return fmt.Errorf("destination is not empty (%s); re-run with --force to overwrite, or wipe dest paths first", reason)
		}
	} else {
		log.Println("--force: allowing non-empty destination")
	}

	dstSQL, err := db.NewTranSQL(cfg.TranSQLitePath)
	if err != nil {
		return fmt.Errorf("open sqlite: %w", err)
	}
	defer dstSQL.Close()

	dstDetails, err := db.NewBadgerEntityDetails(cfg.EntityDetailsBadger)
	if err != nil {
		return fmt.Errorf("open badger: %w", err)
	}
	defer dstDetails.Close()

	for _, table := range db.MorphOwnedTablesFKOrder {
		exists, err := mysqlTableExists(ctx, srcSQL, table)
		if err != nil {
			return err
		}
		if !exists {
			log.Printf("warn: skip missing source table %s", table)
			continue
		}
		n, err := copyTable(ctx, srcSQL, dstSQL.DB, table)
		if err != nil {
			if strings.Contains(err.Error(), "no overlapping columns") {
				log.Printf("warn: skip %s (%v)", table, err)
				continue
			}
			return fmt.Errorf("copy %s: %w", table, err)
		}
		log.Printf("copied %s: %d rows", table, n)
	}

	copied, err := copyEntityDetails(ctx, srcMongo, dstDetails)
	if err != nil {
		return fmt.Errorf("entity_details: %w", err)
	}
	log.Printf("copied entity_details: %d docs", copied)
	log.Println("apply complete (sources unchanged)")
	return nil
}

func runVerify(ctx context.Context, srcSQL *sql.DB, srcMongo *db.TranMongo, cfg config.Config) error {
	dstSQL, err := db.NewTranSQL(cfg.TranSQLitePath)
	if err != nil {
		return fmt.Errorf("open sqlite: %w", err)
	}
	defer dstSQL.Close()

	dstDetails, err := db.NewBadgerEntityDetails(cfg.EntityDetailsBadger)
	if err != nil {
		return fmt.Errorf("open badger: %w", err)
	}
	defer dstDetails.Close()

	mismatches := 0
	for _, table := range db.MorphOwnedTablesFKOrder {
		srcExists, err := mysqlTableExists(ctx, srcSQL, table)
		if err != nil {
			return err
		}
		dstExists, err := sqliteTableExistsLocal(ctx, dstSQL.DB, table)
		if err != nil {
			return err
		}
		if !srcExists {
			log.Printf("verify SKIP %s (missing on source)", table)
			continue
		}
		srcN, err := countRows(ctx, srcSQL, table, true)
		if err != nil {
			return err
		}
		var dstN int64
		if dstExists {
			dstN, err = countRows(ctx, dstSQL.DB, table, false)
			if err != nil {
				return err
			}
		}
		if srcN != dstN {
			log.Printf("MISMATCH %s: source=%d dest=%d", table, srcN, dstN)
			mismatches++
		} else {
			log.Printf("OK %s: %d", table, srcN)
		}
	}

	srcDocs, err := srcMongo.CountEntityDetails(ctx)
	if err != nil {
		return err
	}
	dstDocs, err := dstDetails.CountEntityDetails()
	if err != nil {
		return err
	}
	if int64(dstDocs) != srcDocs {
		log.Printf("MISMATCH entity_details: source=%d dest=%d", srcDocs, dstDocs)
		mismatches++
	} else {
		log.Printf("OK entity_details: %d", srcDocs)
	}

	if mismatches > 0 {
		return fmt.Errorf("%d count mismatch(es)", mismatches)
	}
	log.Println("verify OK")
	return nil
}

func destNonEmpty(ctx context.Context, cfg config.Config) (bool, string, error) {
	if st, err := os.Stat(cfg.TranSQLitePath); err == nil && st.Size() > 0 {
		// Open and check row counts (schema-only file may be small but empty of data).
		tmp, err := sql.Open("sqlite", cfg.TranSQLitePath+"?mode=ro")
		if err == nil {
			defer tmp.Close()
			for _, table := range db.MorphOwnedTablesFKOrder {
				ok, err := sqliteTableExistsLocal(ctx, tmp, table)
				if err != nil {
					return false, "", err
				}
				if !ok {
					continue
				}
				n, err := countRows(ctx, tmp, table, false)
				if err != nil {
					continue
				}
				if n > 0 {
					return true, fmt.Sprintf("sqlite table %s has %d rows", table, n), nil
				}
			}
		} else if st.Size() > 4096 {
			return true, fmt.Sprintf("sqlite file size=%d", st.Size()), nil
		}
	}

	if st, err := os.Stat(cfg.EntityDetailsBadger); err == nil && st.IsDir() {
		b, err := db.NewBadgerEntityDetails(cfg.EntityDetailsBadger)
		if err != nil {
			// Directory exists but may be empty/corrupt — treat open failure as potentially non-empty.
			entries, _ := os.ReadDir(cfg.EntityDetailsBadger)
			if len(entries) > 0 {
				return true, "entity details badger dir not empty and failed to open", err
			}
		} else {
			n, err := b.CountEntityDetails()
			_ = b.Close()
			if err != nil {
				return false, "", err
			}
			if n > 0 {
				return true, fmt.Sprintf("badger has %d entity_detail keys", n), nil
			}
		}
	}
	return false, "", nil
}

func mysqlTableExists(ctx context.Context, db *sql.DB, table string) (bool, error) {
	var n int
	err := db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM information_schema.tables
		WHERE table_schema = DATABASE() AND table_name = ?`, table).Scan(&n)
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

func sqliteTableExistsLocal(ctx context.Context, db *sql.DB, table string) (bool, error) {
	var name string
	err := db.QueryRowContext(ctx,
		`SELECT name FROM sqlite_master WHERE type='table' AND name = ? LIMIT 1`, table,
	).Scan(&name)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return name != "", nil
}

func quoteMySQLIdent(name string) string {
	return "`" + strings.ReplaceAll(name, "`", "``") + "`"
}

func quoteSQLiteIdent(name string) string {
	return `"` + strings.ReplaceAll(name, `"`, `""`) + `"`
}

func countRows(ctx context.Context, db *sql.DB, table string, mysql bool) (int64, error) {
	q := "SELECT COUNT(*) FROM "
	if mysql {
		q += quoteMySQLIdent(table)
	} else {
		q += quoteSQLiteIdent(table)
	}
	var n int64
	err := db.QueryRowContext(ctx, q).Scan(&n)
	return n, err
}

func listColumns(ctx context.Context, db *sql.DB, table string, mysql bool) ([]string, error) {
	if mysql {
		rows, err := db.QueryContext(ctx, `
			SELECT COLUMN_NAME FROM information_schema.columns
			WHERE table_schema = DATABASE() AND table_name = ?
			ORDER BY ORDINAL_POSITION`, table)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		var cols []string
		for rows.Next() {
			var c string
			if err := rows.Scan(&c); err != nil {
				return nil, err
			}
			cols = append(cols, c)
		}
		return cols, rows.Err()
	}
	rows, err := db.QueryContext(ctx, fmt.Sprintf("PRAGMA table_info(%s)", quoteSQLiteIdent(table)))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var cols []string
	for rows.Next() {
		var cid int
		var name, ctype string
		var notnull, pk int
		var dflt sql.NullString
		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dflt, &pk); err != nil {
			return nil, err
		}
		cols = append(cols, name)
	}
	return cols, rows.Err()
}

func intersectCols(src, dst []string) []string {
	dstSet := make(map[string]string, len(dst))
	for _, c := range dst {
		dstSet[strings.ToLower(c)] = c
	}
	var out []string
	for _, c := range src {
		if d, ok := dstSet[strings.ToLower(c)]; ok {
			out = append(out, d) // use destination casing
		}
	}
	return out
}

func copyTable(ctx context.Context, src, dst *sql.DB, table string) (int64, error) {
	srcCols, err := listColumns(ctx, src, table, true)
	if err != nil {
		return 0, err
	}
	dstCols, err := listColumns(ctx, dst, table, false)
	if err != nil {
		return 0, err
	}
	cols := intersectCols(srcCols, dstCols)
	if len(cols) == 0 {
		return 0, fmt.Errorf("no overlapping columns for %s", table)
	}

	// Map dest names back to source names (case-insensitive).
	srcByLower := map[string]string{}
	for _, c := range srcCols {
		srcByLower[strings.ToLower(c)] = c
	}
	srcSelect := make([]string, len(cols))
	dstInsert := make([]string, len(cols))
	for i, dstName := range cols {
		dstInsert[i] = quoteSQLiteIdent(dstName)
		srcSelect[i] = quoteMySQLIdent(srcByLower[strings.ToLower(dstName)])
	}

	sel := fmt.Sprintf("SELECT %s FROM %s", strings.Join(srcSelect, ", "), quoteMySQLIdent(table))
	rows, err := src.QueryContext(ctx, sel)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	placeholders := make([]string, len(cols))
	for i := range placeholders {
		placeholders[i] = "?"
	}
	ins := fmt.Sprintf(
		"INSERT OR REPLACE INTO %s (%s) VALUES (%s)",
		quoteSQLiteIdent(table),
		strings.Join(dstInsert, ", "),
		strings.Join(placeholders, ", "),
	)

	tx, err := dst.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer func() { _ = tx.Rollback() }()

	stmt, err := tx.PrepareContext(ctx, ins)
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	raw := make([]interface{}, len(cols))
	ptrs := make([]interface{}, len(cols))
	for i := range raw {
		ptrs[i] = &raw[i]
	}

	var n int64
	for rows.Next() {
		if err := rows.Scan(ptrs...); err != nil {
			return n, err
		}
		args := make([]interface{}, len(raw))
		for i, v := range raw {
			args[i] = normalizeMySQLValue(v)
		}
		if _, err := stmt.ExecContext(ctx, args...); err != nil {
			return n, fmt.Errorf("insert row: %w", err)
		}
		n++
	}
	if err := rows.Err(); err != nil {
		return n, err
	}
	if err := tx.Commit(); err != nil {
		return n, err
	}
	return n, nil
}

func normalizeMySQLValue(v interface{}) interface{} {
	switch t := v.(type) {
	case []byte:
		return string(t)
	default:
		return v
	}
}

func copyEntityDetails(ctx context.Context, src *db.TranMongo, dst *db.BadgerEntityDetails) (int, error) {
	return src.ForEachEntityDetail(ctx, func(doc db.EntityDetailDoc) error {
		return dst.SetEntityDetailJSON(ctx, doc.Entity, doc.RecordID, doc.Body)
	})
}
