// Backfills realistic demo data for Tran `employee` (staff API): employ_type, date_of_birth,
// gender, description (SQL), active_flag + inactive_date (consistent), JSON detail (entity_details:
// entity "staff" and mirror "employee").
// Detail body uses staff-specific realistic JSON (internal/demoentitydetail.StaffDemoDetailJSON).
// Does not modify user_id.
//
// Rule: if inactive_date is non-NULL, active_flag must be false (0).
//
//	cd morph && go run ./cmd/backfill_employee_demo
//	STORAGE_BACKEND=legacy go run ./cmd/backfill_employee_demo
//	go run ./cmd/backfill_employee_demo -mysql-only
package main

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/binary"
	"flag"
	"fmt"
	"log"
	"strings"
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
	log.Printf("backfill_employee_demo backend=%s", stores.Backend)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Minute)
	defer cancel()

	var details db.EntityDetailStore
	if !*sqlOnly {
		details = stores.Details
	}

	rows, err := sqlDB.QueryContext(ctx,
		`SELECT id, COALESCE(last_name,''), COALESCE(first_name,'') FROM `+"`employee`"+` ORDER BY id`)
	if err != nil {
		log.Fatalf("select employees: %v", err)
	}
	defer rows.Close()

	type row struct {
		id        int
		lastName  string
		firstName string
	}
	var list []row
	for rows.Next() {
		var r row
		if err := rows.Scan(&r.id, &r.lastName, &r.firstName); err != nil {
			log.Fatalf("scan: %v", err)
		}
		list = append(list, r)
	}
	if err := rows.Err(); err != nil {
		log.Fatalf("rows: %v", err)
	}

	log.Printf("Found %d employee rows", len(list))

	hasDetailCol := employeeHasDetailJSONColumn(ctx, sqlDB)
	if hasDetailCol {
		log.Printf("employee.detail column present — backfill will write JSON detail in SQL")
	}

	employTypes := []struct {
		code string
		w    int
	}{
		{"full_time", 72},
		{"part_time", 20},
		{"contractor", 8},
	}

	descTemplates := []func(first, last string, id int, employType string, inactive bool) string{
		func(first, last string, id int, employType string, inactive bool) string {
			status := "Active"
			if inactive {
				status = "Former — separation processed through HR"
			}
			return fmt.Sprintf(
				"%s %s (%s). %s. Transportation certification on file; "+
					"annual motor vehicle record review ID MR-%05d.",
				first, last, employType, status, id%90000+10000)
		},
		func(first, last string, id int, employType string, inactive bool) string {
			role := "route operations"
			if employType == "contractor" {
				role = "contracted coverage"
			}
			return fmt.Sprintf(
				"Assigned to %s under district safety policy TS-%04d. "+
					"%s %s completes recurring defensive-driving refresher through the consortium LMS.",
				role, id%8000+1000, first, last)
		},
		func(first, last string, id int, employType string, inactive bool) string {
			return fmt.Sprintf(
				"Payroll classification: %s. "+
					"Supervisor approved flexible shift notes where aligned with bell schedules and DOT rest rules.",
				employType)
		},
	}

	n := 0
	for _, r := range list {
		h := hash64(r.id)

		employType := weightedEmployType(employTypes, h)
		gender := pickGenderStaff(h)
		dob := staffDOBFromHash(h)

		inactive := int(h%100) < 14 // ~14% former employees
		var inactiveDateSQL interface{}
		var inactiveDatePtr *string
		activeFlag := 1
		if inactive {
			activeFlag = 0
			years := 1 + int(h%7)
			months := int((h >> 8) % 12)
			days := int((h >> 16) % 28)
			s := time.Now().AddDate(-years, -months, -days).Format("2006-01-02")
			inactiveDateSQL = s
			inactiveDatePtr = &s
		} else {
			inactiveDateSQL = nil
		}

		desc := descTemplates[int(h%uint64(len(descTemplates)))](r.firstName, r.lastName, r.id, employType, inactive)

		detailJSON := demoentitydetail.StaffDemoDetailJSON(
			r.id, r.firstName, r.lastName, gender, employType, dob, inactive, inactiveDatePtr)

		var err error
		if hasDetailCol {
			_, err = sqlDB.ExecContext(ctx,
				`UPDATE `+"`employee`"+` SET employ_type = ?, date_of_birth = ?, gender = ?, description = ?, active_flag = ?, inactive_date = ?, detail = ? WHERE id = ?`,
				employType,
				dob.Format("2006-01-02"),
				gender,
				desc,
				activeFlag,
				inactiveDateSQL,
				detailJSON,
				r.id,
			)
		} else {
			_, err = sqlDB.ExecContext(ctx,
				`UPDATE `+"`employee`"+` SET employ_type = ?, date_of_birth = ?, gender = ?, description = ?, active_flag = ?, inactive_date = ? WHERE id = ?`,
				employType,
				dob.Format("2006-01-02"),
				gender,
				desc,
				activeFlag,
				inactiveDateSQL,
				r.id,
			)
		}
		if err != nil {
			log.Fatalf("update employee %d: %v", r.id, err)
		}
		n++

		if details != nil {
			if err := details.SetEntityDetailJSON(ctx, "staff", r.id, detailJSON); err != nil {
				log.Fatalf("detail staff %d: %v", r.id, err)
			}
			if err := details.SetEntityDetailJSON(ctx, "employee", r.id, detailJSON); err != nil {
				log.Fatalf("detail employee %d: %v", r.id, err)
			}
		}

		if n%200 == 0 {
			log.Printf("… processed %d / %d", n, len(list))
		}
	}

	log.Printf("Done: updated SQL for %d employees", n)
	if details != nil {
		log.Printf("entity_details: upserted entity=staff and entity=employee (mirror)")
	}
	if hasDetailCol {
		log.Printf("SQL: wrote JSON into employee.detail")
	}
}

func employeeHasDetailJSONColumn(ctx context.Context, sqlDB *sql.DB) bool {
	rows, err := sqlDB.QueryContext(ctx, `PRAGMA table_info(employee)`)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var cid int
			var name, ctype string
			var notnull, pk int
			var dflt sql.NullString
			if err := rows.Scan(&cid, &name, &ctype, &notnull, &dflt, &pk); err != nil {
				return false
			}
			if strings.EqualFold(name, "detail") {
				return true
			}
		}
		return false
	}
	var n int
	err = sqlDB.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = DATABASE() AND table_name = 'employee' AND column_name = 'detail'`).Scan(&n)
	return err == nil && n > 0
}

func hash64(id int) uint64 {
	sum := sha256.Sum256([]byte(fmt.Sprintf("employee-demo-%d", id)))
	return binary.BigEndian.Uint64(sum[:8])
}

func weightedEmployType(types []struct {
	code string
	w    int
}, h uint64) string {
	r := int(h % 100)
	sum := 0
	for _, t := range types {
		sum += t.w
		if r < sum {
			return t.code
		}
	}
	return types[0].code
}

func pickGenderStaff(h uint64) int {
	if h%100 < 48 {
		return 1
	}
	return 2
}

func staffDOBFromHash(h uint64) time.Time {
	age := 24 + int(h%40)
	month := time.Month(1 + int((h>>12)%12))
	day := 1 + int((h>>20)%28)
	y := time.Now().Year() - age
	return time.Date(y, month, day, 0, 0, 0, 0, time.UTC)
}
