// Prune member, employee, and Asset rows whose entity_details JSON is missing or empty
// (same notion as handlers.isEmptyMongoDetail: "", "{}", or whitespace only).
//
// Related rows (comments, record_contact, activity links) are removed first
// where this command knows about them. ActivityParticipant / ActivityEmployee / ActivityAsset use
// ON DELETE CASCADE when linked via FK.
//
// Usage:
//
//	cd morph && go run ./cmd/prune_empty_detail -dry-run
//	go run ./cmd/prune_empty_detail
//	STORAGE_BACKEND=legacy go run ./cmd/prune_empty_detail
//
package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"strings"
	"time"

	"idongivaflyinfa/config"
	"idongivaflyinfa/db"
)

func isEmptyDetail(js string) bool {
	s := strings.TrimSpace(js)
	if s == "" || s == "{}" {
		return true
	}
	var m map[string]interface{}
	if json.Unmarshal([]byte(s), &m) == nil && len(m) == 0 {
		return true
	}
	return false
}

func detailMissing(store db.EntityDetailStore, ctx context.Context, primaryEntity, altEntity string, id int) (bool, error) {
	js, err := store.GetEntityDetailJSON(ctx, primaryEntity, id)
	if err != nil {
		return false, err
	}
	if !isEmptyDetail(js) {
		return false, nil
	}
	if altEntity == "" {
		return true, nil
	}
	js2, err := store.GetEntityDetailJSON(ctx, altEntity, id)
	if err != nil {
		return false, err
	}
	return isEmptyDetail(js2), nil
}

func chunkIDs(ids []int, n int) [][]int {
	if n <= 0 {
		n = 400
	}
	var out [][]int
	for i := 0; i < len(ids); i += n {
		j := i + n
		if j > len(ids) {
			j = len(ids)
		}
		out = append(out, ids[i:j])
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

type sqlExecer interface {
	ExecContext(context.Context, string, ...interface{}) (sql.Result, error)
}

func deleteByInts(ctx context.Context, d sqlExecer, query string, ids []int) (int64, error) {
	if len(ids) == 0 {
		return 0, nil
	}
	var total int64
	for _, chunk := range chunkIDs(ids, 300) {
		args := make([]interface{}, len(chunk))
		for i, v := range chunk {
			args[i] = v
		}
		r, err := d.ExecContext(ctx, fmt.Sprintf(query, placeholders(len(chunk))), args...)
		if err != nil {
			return total, err
		}
		n, _ := r.RowsAffected()
		total += n
	}
	return total, nil
}

func main() {
	dry := flag.Bool("dry-run", false, "list counts only; do not delete")
	flag.Parse()

	cfg := config.GetConfig()
	stores, err := db.OpenCLIStores(cfg)
	if err != nil {
		log.Fatalf("open stores: %v", err)
	}
	defer stores.Close()
	sqlDB := stores.SQL
	details := stores.Details
	if details == nil {
		log.Fatal("entity details store is required")
	}
	log.Printf("prune_empty_detail backend=%s", stores.Backend)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	var studentScheduleExists int
	if err := sqlDB.QueryRowContext(ctx, `SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='StudentSchedule'`).Scan(&studentScheduleExists); err != nil {
		// MySQL legacy: information_schema
		_ = sqlDB.QueryRowContext(ctx,
			`SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = DATABASE() AND table_name = 'StudentSchedule'`).Scan(&studentScheduleExists)
	}

	// --- Members (entity "student") ---
	memberRows, err := sqlDB.QueryContext(ctx, `SELECT id FROM `+"`member`"+` ORDER BY id`)
	if err != nil {
		log.Fatalf("member ids: %v", err)
	}
	defer memberRows.Close()
	var memberIDs []int
	for memberRows.Next() {
		var id int
		if err := memberRows.Scan(&id); err != nil {
			log.Fatalf("scan member: %v", err)
		}
		missing, err := detailMissing(details, ctx, "student", "", id)
		if err != nil {
			log.Fatalf("detail student %d: %v", id, err)
		}
		if missing {
			memberIDs = append(memberIDs, id)
		}
	}
	if err := memberRows.Err(); err != nil {
		log.Fatalf("member rows: %v", err)
	}
	log.Printf("members without detail: %d", len(memberIDs))

	// --- Employees (entity "staff", mirror "employee") ---
	empRows, err := sqlDB.QueryContext(ctx, `SELECT id FROM `+"`employee`"+` ORDER BY id`)
	if err != nil {
		log.Fatalf("employee ids: %v", err)
	}
	defer empRows.Close()
	var empIDs []int
	for empRows.Next() {
		var id int
		if err := empRows.Scan(&id); err != nil {
			log.Fatalf("scan employee: %v", err)
		}
		missing, err := detailMissing(details, ctx, "staff", "employee", id)
		if err != nil {
			log.Fatalf("detail staff %d: %v", id, err)
		}
		if missing {
			empIDs = append(empIDs, id)
		}
	}
	if err := empRows.Err(); err != nil {
		log.Fatalf("employee rows: %v", err)
	}
	log.Printf("employees without detail: %d", len(empIDs))

	// --- Assets (entity "vehicle") ---
	assetRows, err := sqlDB.QueryContext(ctx, `SELECT ID FROM Asset ORDER BY ID`)
	if err != nil {
		log.Fatalf("asset ids: %v", err)
	}
	defer assetRows.Close()
	var assetIDs []int
	for assetRows.Next() {
		var id int
		if err := assetRows.Scan(&id); err != nil {
			log.Fatalf("scan asset: %v", err)
		}
		missing, err := detailMissing(details, ctx, "vehicle", "", id)
		if err != nil {
			log.Fatalf("detail vehicle %d: %v", id, err)
		}
		if missing {
			assetIDs = append(assetIDs, id)
		}
	}
	if err := assetRows.Err(); err != nil {
		log.Fatalf("asset rows: %v", err)
	}
	log.Printf("assets without detail: %d", len(assetIDs))

	if *dry {
		log.Println("dry-run: no changes")
		return
	}

	tx, err := sqlDB.BeginTx(ctx, nil)
	if err != nil {
		log.Fatalf("begin tx: %v", err)
	}
	defer func() { _ = tx.Rollback() }()

	exec := func(q string, args ...interface{}) {
		if _, err := tx.ExecContext(ctx, q, args...); err != nil {
			log.Fatalf("exec: %v\nquery: %s", err, q)
		}
	}

	if len(memberIDs) > 0 {
		if studentScheduleExists > 0 {
			for _, chunk := range chunkIDs(memberIDs, 300) {
				ph := placeholders(len(chunk))
				args := make([]interface{}, len(chunk))
				for i, v := range chunk {
					args[i] = v
				}
				if _, err := tx.ExecContext(ctx, `DELETE FROM StudentSchedule WHERE member_id IN (`+ph+`)`, args...); err != nil {
					log.Fatalf("StudentSchedule: %v", err)
				}
			}
		}
		n64, err := deleteByInts(ctx, tx, `DELETE FROM record_contact WHERE EntityType = 'student' AND RecordID IN (%s)`, memberIDs)
		if err != nil {
			log.Fatalf("record_contact student: %v", err)
		}
		log.Printf("deleted record_contact (student) rows: %d", n64)
		n64, err = deleteByInts(ctx, tx, `DELETE FROM comment WHERE EntityType = 'student' AND RecordID IN (%s)`, memberIDs)
		if err != nil {
			log.Fatalf("comment student: %v", err)
		}
		log.Printf("deleted comment (student) rows: %d", n64)
		for _, id := range memberIDs {
			exec(`DELETE FROM `+"`member`"+` WHERE id = ?`, id)
		}
		log.Printf("deleted member rows: %d", len(memberIDs))
	}

	if len(empIDs) > 0 {
		n64, err := deleteByInts(ctx, tx, `DELETE FROM record_contact WHERE EntityType = 'staff' AND RecordID IN (%s)`, empIDs)
		if err != nil {
			log.Fatalf("record_contact staff: %v", err)
		}
		log.Printf("deleted record_contact (staff) rows: %d", n64)
		n64, err = deleteByInts(ctx, tx, `DELETE FROM comment WHERE EntityType = 'staff' AND RecordID IN (%s)`, empIDs)
		if err != nil {
			log.Fatalf("comment staff: %v", err)
		}
		log.Printf("deleted comment (staff) rows: %d", n64)
		for _, id := range empIDs {
			exec(`DELETE FROM `+"`employee`"+` WHERE id = ?`, id)
		}
		log.Printf("deleted employee rows: %d", len(empIDs))
	}

	if len(assetIDs) > 0 {
		n64, err := deleteByInts(ctx, tx, `DELETE FROM record_contact WHERE EntityType = 'vehicle' AND RecordID IN (%s)`, assetIDs)
		if err != nil {
			log.Fatalf("record_contact vehicle: %v", err)
		}
		log.Printf("deleted record_contact (vehicle) rows: %d", n64)
		n64, err = deleteByInts(ctx, tx, `DELETE FROM comment WHERE EntityType = 'vehicle' AND RecordID IN (%s)`, assetIDs)
		if err != nil {
			log.Fatalf("comment vehicle: %v", err)
		}
		log.Printf("deleted comment (vehicle) rows: %d", n64)
		for _, id := range assetIDs {
			exec(`DELETE FROM Asset WHERE ID = ?`, id)
		}
		log.Printf("deleted Asset rows: %d", len(assetIDs))
	}

	if err := tx.Commit(); err != nil {
		log.Fatalf("commit: %v", err)
	}

	for _, id := range memberIDs {
		if err := details.DeleteEntityDetail(ctx, "student", id); err != nil {
			log.Printf("warn: detail delete student %d: %v", id, err)
		}
	}
	for _, id := range empIDs {
		_ = details.DeleteEntityDetail(ctx, "staff", id)
		_ = details.DeleteEntityDetail(ctx, "employee", id)
	}
	for _, id := range assetIDs {
		if err := details.DeleteEntityDetail(ctx, "vehicle", id); err != nil {
			log.Printf("warn: detail delete vehicle %d: %v", id, err)
		}
	}

	log.Println("done")
}
