// Backfill MorphData contacts: remove duplicate names, trim to ~30 rows,
// and populate SQL description plus entity_details JSON (Badger by default).
//
// Usage:
//
//	cd morph
//	go run ./cmd/backfill_contact_demo
//	STORAGE_BACKEND=legacy go run ./cmd/backfill_contact_demo
package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"strings"
	"time"

	"idongivaflyinfa/config"
	"idongivaflyinfa/db"
	"idongivaflyinfa/internal/demoentitydetail"
)

type contactRow struct {
	id        int
	firstName string
	lastName  string
	email     string
	phone     string
	mobile    string
}

func main() {
	sqlOnly := flag.Bool("mysql-only", false, "only update SQL; skip entity detail upserts")
	flag.Parse()

	cfg := config.GetConfig()
	stores, err := db.OpenCLIStores(cfg)
	if err != nil {
		log.Fatalf("open stores: %v", err)
	}
	defer stores.Close()
	sqlDB := stores.SQL
	log.Printf("backfill_contact_demo backend=%s", stores.Backend)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	var details db.EntityDetailStore
	if !*sqlOnly {
		details = stores.Details
	}

	rows, err := loadContacts(ctx, sqlDB)
	if err != nil {
		log.Fatalf("load contacts: %v", err)
	}
	log.Printf("Found %d contacts before cleanup", len(rows))

	dupIDs := duplicateContactIDs(rows)
	if len(dupIDs) > 0 {
		log.Printf("Removing %d duplicate-name contacts", len(dupIDs))
		if err := deleteContacts(ctx, sqlDB, details, dupIDs); err != nil {
			log.Fatalf("delete duplicates: %v", err)
		}
	}

	rows, err = loadContacts(ctx, sqlDB)
	if err != nil {
		log.Fatalf("reload contacts: %v", err)
	}
	if len(rows) > demoentitydetail.ContactDemoCap {
		trimIDs := make([]int, 0, len(rows)-demoentitydetail.ContactDemoCap)
		for i := demoentitydetail.ContactDemoCap; i < len(rows); i++ {
			trimIDs = append(trimIDs, rows[i].id)
		}
		log.Printf("Trimming %d excess contacts (keeping %d)", len(trimIDs), demoentitydetail.ContactDemoCap)
		if err := deleteContacts(ctx, sqlDB, details, trimIDs); err != nil {
			log.Fatalf("trim contacts: %v", err)
		}
		rows = rows[:demoentitydetail.ContactDemoCap]
	}

	seeds := demoentitydetail.ContactDemoSeeds()
	updated := 0
	for i, r := range rows {
		seed := seeds[i%len(seeds)]
		fn := seed.FirstName
		ln := seed.LastName
		email := nullIfEmpty(r.email, seed.Email)
		phone := nullIfEmpty(r.phone, seed.Phone)
		mobile := nullIfEmpty(r.mobile, seed.Mobile)
		desc := demoentitydetail.ContactDescription(r.id, fn, ln, seed.Role)

		_, err := sqlDB.ExecContext(ctx,
			`UPDATE contact SET FirstName = ?, LastName = ?, Email = ?, Phone = ?, Mobile = ?, description = ? WHERE ID = ?`,
			fn, ln, email, phone, mobile, desc, r.id,
		)
		if err != nil {
			log.Fatalf("update contact %d: %v", r.id, err)
		}

		if details != nil {
			body := demoentitydetail.ContactLayoutDetailJSON(r.id, fn, ln, email, phone, mobile, seed.Role)
			if err := details.SetEntityDetailJSON(ctx, "contact", r.id, body); err != nil {
				log.Fatalf("mongo contact %d: %v", r.id, err)
			}
		}
		updated++
	}

	// Insert missing rows up to cap using seed slots not yet represented.
	cur := len(rows)
	if cur < demoentitydetail.ContactDemoCap {
		log.Printf("Inserting %d contacts to reach cap %d", demoentitydetail.ContactDemoCap-cur, demoentitydetail.ContactDemoCap)
	}
	for i := cur; i < demoentitydetail.ContactDemoCap; i++ {
		seed := seeds[i%len(seeds)]
		desc := demoentitydetail.ContactDescription(i+10000, seed.FirstName, seed.LastName, seed.Role)
		res, err := sqlDB.ExecContext(ctx,
			`INSERT INTO contact (LastName, FirstName, Email, Phone, Mobile, description) VALUES (?, ?, ?, ?, ?, ?)`,
			seed.LastName, seed.FirstName, seed.Email, seed.Phone, seed.Mobile, desc,
		)
		if err != nil {
			log.Fatalf("insert contact: %v", err)
		}
		id64, err := res.LastInsertId()
		if err != nil || id64 <= 0 {
			log.Fatalf("insert contact id: %v", err)
		}
		id := int(id64)
		if details != nil {
			body := demoentitydetail.ContactLayoutDetailJSON(id, seed.FirstName, seed.LastName, seed.Email, seed.Phone, seed.Mobile, seed.Role)
			if err := details.SetEntityDetailJSON(ctx, "contact", id, body); err != nil {
				log.Fatalf("mongo new contact %d: %v", id, err)
			}
		}
		updated++
	}

	log.Printf("Done: %d contacts with description%s", updated, detailSuffix(details != nil))
}

func detailSuffix(ok bool) string {
	if ok {
		return " and entity detail JSON"
	}
	return ""
}

func nullIfEmpty(v, fallback string) string {
	if strings.TrimSpace(v) == "" {
		return fallback
	}
	return strings.TrimSpace(v)
}

func loadContacts(ctx context.Context, sqlDB *sql.DB) ([]contactRow, error) {
	rs, err := sqlDB.QueryContext(ctx,
		`SELECT ID, COALESCE(FirstName,''), COALESCE(LastName,''), COALESCE(Email,''), COALESCE(Phone,''), COALESCE(Mobile,'')
		 FROM contact ORDER BY ID ASC`)
	if err != nil {
		return nil, err
	}
	defer rs.Close()
	var out []contactRow
	for rs.Next() {
		var r contactRow
		if err := rs.Scan(&r.id, &r.firstName, &r.lastName, &r.email, &r.phone, &r.mobile); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rs.Err()
}

func duplicateContactIDs(rows []contactRow) []int {
	seen := map[string]int{}
	var drop []int
	for _, r := range rows {
		key := demoentitydetail.ContactDisplayKey(r.firstName, r.lastName)
		if key == "" {
			key = fmt.Sprintf("id:%d", r.id)
		}
		if kept, ok := seen[key]; ok {
			drop = append(drop, r.id)
			log.Printf("  duplicate %q: drop id=%d keep id=%d", key, r.id, kept)
			continue
		}
		seen[key] = r.id
	}
	return drop
}

func deleteContacts(ctx context.Context, sqlDB *sql.DB, details db.EntityDetailStore, ids []int) error {
	if len(ids) == 0 {
		return nil
	}
	if details != nil {
		for _, id := range ids {
			_ = details.DeleteEntityDetail(ctx, "contact", id)
		}
	}
	for _, chunk := range chunkInts(ids, 200) {
		q := fmt.Sprintf(`DELETE FROM record_contact WHERE ContactID IN (%s)`, placeholders(len(chunk)))
		args := intArgs(chunk)
		if _, err := sqlDB.ExecContext(ctx, q, args...); err != nil {
			return err
		}
		q = fmt.Sprintf(`DELETE FROM contact WHERE ID IN (%s)`, placeholders(len(chunk)))
		if _, err := sqlDB.ExecContext(ctx, q, args...); err != nil {
			return err
		}
	}
	return nil
}

func chunkInts(ids []int, size int) [][]int {
	if size <= 0 {
		size = 200
	}
	var out [][]int
	for len(ids) > 0 {
		end := size
		if end > len(ids) {
			end = len(ids)
		}
		out = append(out, ids[:end])
		ids = ids[end:]
	}
	return out
}

func placeholders(n int) string {
	b := make([]string, n)
	for i := range b {
		b[i] = "?"
	}
	return strings.Join(b, ",")
}

func intArgs(ids []int) []interface{} {
	out := make([]interface{}, len(ids))
	for i, v := range ids {
		out[i] = v
	}
	return out
}
