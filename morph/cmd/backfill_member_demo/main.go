// Backfills realistic demo data for Tran `member` (participants): middle_name, gender,
// participant_type, description (SQL), and JSON detail documents (entity_details).
//
// Usage:
//
//	cd morph
//	go run ./cmd/backfill_member_demo
//	STORAGE_BACKEND=legacy go run ./cmd/backfill_member_demo
//
// SQL-only (skips entity detail): pass -mysql-only
package main

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"flag"
	"fmt"
	"log"
	"time"

	"idongivaflyinfa/config"
	"idongivaflyinfa/db"
	"idongivaflyinfa/internal/demoentitydetail"
)

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
	log.Printf("backfill_member_demo backend=%s", stores.Backend)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Minute)
	defer cancel()

	var details db.EntityDetailStore
	if !*sqlOnly {
		details = stores.Details
	}

	rows, err := sqlDB.QueryContext(ctx,
		`SELECT id, COALESCE(last_name,''), COALESCE(first_name,''), COALESCE(facility,'') FROM `+"`member`"+` ORDER BY id`)
	if err != nil {
		log.Fatalf("select members: %v", err)
	}
	defer rows.Close()

	type row struct {
		id        int
		lastName  string
		firstName string
		facility  string
	}
	var list []row
	for rows.Next() {
		var r row
		if err := rows.Scan(&r.id, &r.lastName, &r.firstName, &r.facility); err != nil {
			log.Fatalf("scan: %v", err)
		}
		list = append(list, r)
	}
	if err := rows.Err(); err != nil {
		log.Fatalf("rows: %v", err)
	}
	log.Printf("Found %d members", len(list))

	types := []string{"student", "passenger", "adult", "visitor"}
	n := 0
	for _, r := range list {
		h := sha256.Sum256([]byte(fmt.Sprintf("member-demo-%d", r.id)))
		gender := 1 + int(binary.BigEndian.Uint16(h[0:2])%2)
		pt := types[int(h[2])%len(types)]
		mn := ""
		if h[3]%5 == 0 {
			mn = string(rune('A' + int(h[4]%26)))
		}
		desc := fmt.Sprintf(
			"%s %s — %s participant attached to %s. Demo routing notes refreshed for MorphData seed.",
			r.firstName, r.lastName, pt, nullFacility(r.facility),
		)

		_, err := sqlDB.ExecContext(ctx, `
			UPDATE `+"`member`"+` SET middle_name = NULLIF(?, ''), gender = ?, participant_type = ?, description = ?
			WHERE id = ?`, mn, gender, pt, desc, r.id)
		if err != nil {
			log.Fatalf("update member %d: %v", r.id, err)
		}

		if details != nil {
			body := demoentitydetail.ParticipantLayoutDetailJSON(r.id, r.firstName, r.lastName, gender, r.facility, pt)
			if err := details.SetEntityDetailJSON(ctx, "student", r.id, body); err != nil {
				log.Fatalf("detail student %d: %v", r.id, err)
			}
		}
		n++
		if n%100 == 0 {
			log.Printf("… %d / %d", n, len(list))
		}
	}

	msg := fmt.Sprintf("Done: updated %d members", n)
	if details != nil {
		msg += "; upserted entity_details for entity=student"
	}
	log.Print(msg)
}

func nullFacility(s string) string {
	if s == "" {
		return "unassigned facility"
	}
	return s
}
