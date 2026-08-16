// Backfills realistic JSON detail for Tran facilities (SQL `facility`) into entity_details
// with entity "school" (see handlers.entityKeySchool).
//
//	cd morph && go run ./cmd/backfill_facility_demo
//	STORAGE_BACKEND=legacy go run ./cmd/backfill_facility_demo
//	go run ./cmd/backfill_facility_demo -dry-run
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"idongivaflyinfa/config"
	"idongivaflyinfa/db"
	"idongivaflyinfa/internal/demoentitydetail"
)

func main() {
	dryRun := flag.Bool("dry-run", false, "list facilities only; do not write entity details")
	flag.Parse()

	cfg := config.GetConfig()
	stores, err := db.OpenCLIStores(cfg)
	if err != nil {
		log.Fatalf("open stores: %v", err)
	}
	defer stores.Close()
	sqlDB := stores.SQL
	log.Printf("backfill_facility_demo backend=%s", stores.Backend)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	rows, err := sqlDB.QueryContext(ctx, `
		SELECT f.id, f.facility_code, COALESCE(f.name,''), COALESCE(f.facility_type,''), COALESCE(d.Name,'')
		FROM `+"`facility`"+` f
		LEFT JOIN District d ON d.id = f.district_id
		ORDER BY f.id`)
	if err != nil {
		log.Fatalf("select facilities: %v", err)
	}
	defer rows.Close()

	type rec struct {
		id         int
		code, name string
		ftype      string
		district   string
	}
	var list []rec
	for rows.Next() {
		var r rec
		if err := rows.Scan(&r.id, &r.code, &r.name, &r.ftype, &r.district); err != nil {
			log.Fatalf("scan: %v", err)
		}
		list = append(list, r)
	}
	if err := rows.Err(); err != nil {
		log.Fatalf("rows: %v", err)
	}

	log.Printf("Found %d facility rows", len(list))
	if *dryRun {
		for _, r := range list {
			fmt.Printf("%d\t%s\t%s\n", r.id, r.code, r.name)
		}
		return
	}

	details := stores.Details
	if details == nil {
		log.Fatal("entity details store not available")
	}

	n := 0
	for _, r := range list {
		detailJSON := demoentitydetail.FacilityDemoDetailJSON(r.id, r.code, r.name, r.district, r.ftype)
		if err := details.SetEntityDetailJSON(ctx, "school", r.id, detailJSON); err != nil {
			log.Fatalf("detail school id=%d: %v", r.id, err)
		}
		n++
		if n%50 == 0 {
			log.Printf("… upserted %d / %d", n, len(list))
		}
	}
	log.Printf("Done: entity_details for entity=school count=%d", n)
}
