package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"idongivaflyinfa/config"
	"idongivaflyinfa/db"
	"idongivaflyinfa/internal/demoentitydetail"
)

func seedCap() int {
	return config.GetConfig().SeedTranCap
}

func detailJSONIsPresent(js string) bool {
	s := strings.TrimSpace(js)
	if s == "" || s == "{}" || strings.ToLower(s) == "null" {
		return false
	}
	var m map[string]interface{}
	if json.Unmarshal([]byte(s), &m) == nil && len(m) == 0 {
		return false
	}
	return true
}

func buildStaffTypeCycle() []string {
	out := make([]string, 0, 50)
	add := func(name string, n int) {
		for i := 0; i < n; i++ {
			out = append(out, name)
		}
	}
	add("Bus Driver", 15)
	add("Bus Assistant", 10)
	add("School Personnel", 15)
	add("Mechanic", 6)
	add("Dispatcher", 4)
	return out
}

func upsertStaffMongoDetails(ctx context.Context, mongoDB db.EntityDetailStore, id int, fn, ln string, gender int, employType string, dob time.Time) error {
	if mongoDB == nil {
		return nil
	}
	detail := demoentitydetail.StaffDemoDetailJSON(id, fn, ln, gender, employType, dob, false, nil)
	if err := mongoDB.SetEntityDetailJSON(ctx, "staff", id, detail); err != nil {
		return err
	}
	return mongoDB.SetEntityDetailJSON(ctx, "employee", id, detail)
}

func upsertMemberMongoDetails(ctx context.Context, mongoDB db.EntityDetailStore, id int, fn, ln string, gender int, facility, ptype string) error {
	if mongoDB == nil {
		return nil
	}
	detail := demoentitydetail.ParticipantLayoutDetailJSON(id, fn, ln, gender, facility, ptype)
	return mongoDB.SetEntityDetailJSON(ctx, "student", id, detail)
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

func execInIDs(db *sql.DB, query string, ids []int) error {
	for _, ch := range chunkInts(ids, 200) {
		if len(ch) == 0 {
			continue
		}
		q := fmt.Sprintf(query, placeHolders(len(ch)))
		args := make([]interface{}, len(ch))
		for i, v := range ch {
			args[i] = v
		}
		if _, err := db.Exec(q, args...); err != nil {
			return err
		}
	}
	return nil
}

func placeHolders(n int) string {
	b := make([]string, n)
	for i := range b {
		b[i] = "?"
	}
	return strings.Join(b, ",")
}

func deleteMongoForMembers(ctx context.Context, mongoDB db.EntityDetailStore, ids []int) {
	if mongoDB == nil {
		return
	}
	for _, id := range ids {
		_ = mongoDB.DeleteEntityDetail(ctx, "student", id)
	}
}

func deleteMongoForEmployees(ctx context.Context, mongoDB db.EntityDetailStore, ids []int) {
	if mongoDB == nil {
		return
	}
	for _, id := range ids {
		_ = mongoDB.DeleteEntityDetail(ctx, "staff", id)
		_ = mongoDB.DeleteEntityDetail(ctx, "employee", id)
	}
}

func deleteMongoForAssets(ctx context.Context, mongoDB db.EntityDetailStore, ids []int) {
	if mongoDB == nil {
		return
	}
	for _, id := range ids {
		_ = mongoDB.DeleteEntityDetail(ctx, "vehicle", id)
	}
}

func deleteMongoForContacts(ctx context.Context, mongoDB db.EntityDetailStore, ids []int) {
	if mongoDB == nil {
		return
	}
	for _, id := range ids {
		_ = mongoDB.DeleteEntityDetail(ctx, "contact", id)
	}
}

// deleteMembersByIDs removes members and dependent rows (no FK from MySQL to member on CaseTaskAssignee).
func deleteMembersByIDs(sqlDB *sql.DB, mongoDB db.EntityDetailStore, ctx context.Context, ids []int) error {
	if len(ids) == 0 {
		return nil
	}
	deleteMongoForMembers(ctx, mongoDB, ids)
	if err := execInIDs(sqlDB, `DELETE FROM ActivityParticipant WHERE member_id IN (%s)`, ids); err != nil {
		return err
	}
	if err := execInIDs(sqlDB, `DELETE FROM CaseTaskAssignee WHERE assignee_kind = 'member' AND assignee_id IN (%s)`, ids); err != nil {
		return err
	}
	if err := execInIDs(sqlDB, `DELETE FROM record_contact WHERE EntityType = 'student' AND RecordID IN (%s)`, ids); err != nil {
		return err
	}
	return execInIDs(sqlDB, `DELETE FROM `+"`member`"+` WHERE id IN (%s)`, ids)
}

func deleteEmployeesByIDs(sqlDB *sql.DB, mongoDB db.EntityDetailStore, ctx context.Context, ids []int) error {
	if len(ids) == 0 {
		return nil
	}
	deleteMongoForEmployees(ctx, mongoDB, ids)
	if err := execInIDs(sqlDB, `DELETE FROM ActivityEmployee WHERE employee_id IN (%s)`, ids); err != nil {
		return err
	}
	if err := execInIDs(sqlDB, `DELETE FROM CaseTaskAssignee WHERE assignee_kind = 'employee' AND assignee_id IN (%s)`, ids); err != nil {
		return err
	}
	if err := execInIDs(sqlDB, `DELETE FROM record_contact WHERE EntityType = 'staff' AND RecordID IN (%s)`, ids); err != nil {
		return err
	}
	return execInIDs(sqlDB, `DELETE FROM employee WHERE id IN (%s)`, ids)
}

func deleteAssetsByIDs(sqlDB *sql.DB, mongoDB db.EntityDetailStore, ctx context.Context, ids []int) error {
	if len(ids) == 0 {
		return nil
	}
	deleteMongoForAssets(ctx, mongoDB, ids)
	if err := execInIDs(sqlDB, `DELETE FROM ActivityAsset WHERE asset_id IN (%s)`, ids); err != nil {
		return err
	}
	if err := execInIDs(sqlDB, `DELETE FROM record_contact WHERE EntityType = 'vehicle' AND RecordID IN (%s)`, ids); err != nil {
		return err
	}
	return execInIDs(sqlDB, `DELETE FROM Asset WHERE ID IN (%s)`, ids)
}

func deleteContactsByIDs(sqlDB *sql.DB, mongoDB db.EntityDetailStore, ctx context.Context, ids []int) error {
	if len(ids) == 0 {
		return nil
	}
	deleteMongoForContacts(ctx, mongoDB, ids)
	if err := execInIDs(sqlDB, `DELETE FROM record_contact WHERE ContactID IN (%s)`, ids); err != nil {
		return err
	}
	return execInIDs(sqlDB, `DELETE FROM contact WHERE ID IN (%s)`, ids)
}

func queryIntIDs(sqlDB *sql.DB, query string) ([]int, error) {
	rows, err := sqlDB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		out = append(out, id)
	}
	return out, rows.Err()
}

// demoPrunePeopleWithoutDetails deletes member/employee rows with no (or empty) Mongo entity_details document.
func demoPrunePeopleWithoutDetails(ctx context.Context, sqlDB *sql.DB, mongoDB db.EntityDetailStore) error {
	if mongoDB == nil {
		log.Println("demo prune: entity details unavailable — skipping remove-without-details")
		return nil
	}

	memberIDs, err := queryIntIDs(sqlDB, `SELECT id FROM `+"`member`"+` ORDER BY id`)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "doesn't exist") {
			return nil
		}
		return err
	}
	var dropMembers []int
	for _, id := range memberIDs {
		js, err := mongoDB.GetEntityDetailJSON(ctx, "student", id)
		if err != nil {
			log.Printf("demo prune: mongo read student %d: %v", id, err)
			continue
		}
		if !detailJSONIsPresent(js) {
			dropMembers = append(dropMembers, id)
		}
	}
	if len(dropMembers) > 0 {
		log.Printf("demo prune: removing %d members without Mongo details", len(dropMembers))
		if err := deleteMembersByIDs(sqlDB, mongoDB, ctx, dropMembers); err != nil {
			return fmt.Errorf("delete members without detail: %w", err)
		}
	}

	empIDs, err := queryIntIDs(sqlDB, `SELECT id FROM employee ORDER BY id`)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "doesn't exist") {
			return nil
		}
		return err
	}
	var dropEmp []int
	for _, id := range empIDs {
		js, err := mongoDB.GetEntityDetailJSON(ctx, "staff", id)
		if err != nil {
			log.Printf("demo prune: mongo read staff %d: %v", id, err)
			continue
		}
		if !detailJSONIsPresent(js) {
			dropEmp = append(dropEmp, id)
		}
	}
	if len(dropEmp) > 0 {
		log.Printf("demo prune: removing %d employees without Mongo details", len(dropEmp))
		if err := deleteEmployeesByIDs(sqlDB, mongoDB, ctx, dropEmp); err != nil {
			return fmt.Errorf("delete employees without detail: %w", err)
		}
	}
	return nil
}

func demoTrimMembersToCap(ctx context.Context, sqlDB *sql.DB, mongoDB db.EntityDetailStore, cap int) error {
	all, err := queryIntIDs(sqlDB, `SELECT id FROM `+"`member`"+` ORDER BY id ASC`)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "doesn't exist") {
			return nil
		}
		return err
	}
	keep := map[int]struct{}{}
	for i := 0; i < cap && i < len(all); i++ {
		keep[all[i]] = struct{}{}
	}
	var remove []int
	for _, id := range all {
		if _, ok := keep[id]; !ok {
			remove = append(remove, id)
		}
	}
	if len(remove) == 0 {
		return nil
	}
	log.Printf("demo trim: removing %d excess members (keeping %d)", len(remove), cap)
	return deleteMembersByIDs(sqlDB, mongoDB, ctx, remove)
}

func demoTrimEmployeesToCap(ctx context.Context, sqlDB *sql.DB, mongoDB db.EntityDetailStore, cap int) error {
	all, err := queryIntIDs(sqlDB, `SELECT id FROM employee ORDER BY id ASC`)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "doesn't exist") {
			return nil
		}
		return err
	}
	if len(all) <= cap {
		return nil
	}
	keep := map[int]struct{}{}
	for i := 0; i < cap; i++ {
		keep[all[i]] = struct{}{}
	}
	var remove []int
	for _, id := range all {
		if _, ok := keep[id]; !ok {
			remove = append(remove, id)
		}
	}
	log.Printf("demo trim: removing %d excess employees (keeping %d)", len(remove), cap)
	return deleteEmployeesByIDs(sqlDB, mongoDB, ctx, remove)
}

func demoTrimAssetsToCap(ctx context.Context, sqlDB *sql.DB, mongoDB db.EntityDetailStore, cap int) error {
	all, err := queryIntIDs(sqlDB, `SELECT ID FROM Asset ORDER BY ID ASC`)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "doesn't exist") {
			return nil
		}
		return err
	}
	if len(all) <= cap {
		return nil
	}
	keep := map[int]struct{}{}
	for i := 0; i < cap; i++ {
		keep[all[i]] = struct{}{}
	}
	var remove []int
	for _, id := range all {
		if _, ok := keep[id]; !ok {
			remove = append(remove, id)
		}
	}
	log.Printf("demo trim: removing %d excess assets (keeping %d)", len(remove), cap)
	return deleteAssetsByIDs(sqlDB, mongoDB, ctx, remove)
}

func demoTrimContactsToCap(ctx context.Context, sqlDB *sql.DB, mongoDB db.EntityDetailStore, cap int) error {
	all, err := queryIntIDs(sqlDB, `SELECT ID FROM contact ORDER BY ID ASC`)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "doesn't exist") {
			return nil
		}
		return err
	}
	if len(all) <= cap {
		return nil
	}
	keep := map[int]struct{}{}
	for i := 0; i < cap; i++ {
		keep[all[i]] = struct{}{}
	}
	var remove []int
	for _, id := range all {
		if _, ok := keep[id]; !ok {
			remove = append(remove, id)
		}
	}
	log.Printf("demo trim: removing %d excess contacts (keeping %d)", len(remove), cap)
	return deleteContactsByIDs(sqlDB, mongoDB, ctx, remove)
}

// demoPruneAndCapToLimits removes rows without entity details, then trims main demo tables to SEED_TRAN_CAP (default 50).
// Set SEED_TRAN_SKIP_PRUNE=1 to skip. Uses detailStore for the "without details" pass; trimming uses SQL IDs only.
func demoPruneAndCapToLimits(sqlDB *sql.DB) error {
	if config.GetConfig().SeedTranSkipPrune {
		log.Println("SEED_TRAN_SKIP_PRUNE enabled — skipping demo prune/trim")
		return nil
	}
	capN := seedCap()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	mongoDB := detailStore
	if mongoDB == nil {
		log.Println("demo prune: entity details unavailable — trim-only path")
	}

	if mongoDB != nil {
		if err := demoPrunePeopleWithoutDetails(ctx, sqlDB, mongoDB); err != nil {
			return err
		}
	}

	if err := demoTrimMembersToCap(ctx, sqlDB, mongoDB, capN); err != nil {
		return err
	}
	if err := demoTrimEmployeesToCap(ctx, sqlDB, mongoDB, capN); err != nil {
		return err
	}
	if err := demoTrimAssetsToCap(ctx, sqlDB, mongoDB, capN); err != nil {
		return err
	}
	if err := demoDedupeContactsByName(ctx, sqlDB, mongoDB); err != nil {
		return err
	}
	if err := demoTrimContactsToCap(ctx, sqlDB, mongoDB, demoentitydetail.ContactDemoCap); err != nil {
		return err
	}
	return nil
}

func demoDedupeContactsByName(ctx context.Context, sqlDB *sql.DB, mongoDB db.EntityDetailStore) error {
	rows, err := sqlDB.Query(`SELECT ID, COALESCE(FirstName,''), COALESCE(LastName,'') FROM contact ORDER BY ID ASC`)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "doesn't exist") {
			return nil
		}
		return err
	}
	defer rows.Close()
	seen := map[string]int{}
	var drop []int
	for rows.Next() {
		var id int
		var fn, ln string
		if err := rows.Scan(&id, &fn, &ln); err != nil {
			return err
		}
		key := demoentitydetail.ContactDisplayKey(fn, ln)
		if key == "" {
			continue
		}
		if kept, ok := seen[key]; ok {
			drop = append(drop, id)
			log.Printf("demo dedupe contact %q: drop id=%d keep id=%d", key, id, kept)
			continue
		}
		seen[key] = id
	}
	if err := rows.Err(); err != nil {
		return err
	}
	if len(drop) == 0 {
		return nil
	}
	log.Printf("demo dedupe: removing %d duplicate-name contacts", len(drop))
	return deleteContactsByIDs(sqlDB, mongoDB, ctx, drop)
}

// loadStaffGlobalsFromDB repopulates staffAll, staffDrivers, staffAssistants after skip-insert path.
func loadStaffGlobalsFromDB(sqlDB *sql.DB) error {
	staffAll = staffAll[:0]
	staffDrivers = staffDrivers[:0]
	staffAssistants = staffAssistants[:0]
	rows, err := sqlDB.Query(`
		SELECT e.id, st.StaffTypeName
		FROM employee e
		INNER JOIN StaffStaffType sst ON sst.StaffID = e.id
		INNER JOIN StaffType st ON st.StaffTypeID = sst.StaffTypeID
		ORDER BY e.id, sst.PrimaryFlag DESC, st.StaffTypeName`)
	if err != nil {
		return err
	}
	defer rows.Close()
	seen := map[int]struct{}{}
	for rows.Next() {
		var id int
		var typeName string
		if err := rows.Scan(&id, &typeName); err != nil {
			return err
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		info := staffInfo{ID: id, TypeName: typeName}
		staffAll = append(staffAll, info)
		switch typeName {
		case "Bus Driver":
			staffDrivers = append(staffDrivers, info)
		case "Bus Assistant":
			staffAssistants = append(staffAssistants, info)
		}
	}
	return rows.Err()
}

// loadStudentsFromDB sets students var from first `limit` member ids (ordered by id).
func loadStudentsFromDB(sqlDB *sql.DB, limit int) error {
	students = students[:0]
	rows, err := sqlDB.Query(`SELECT id FROM `+"`member`"+` ORDER BY id ASC LIMIT ?`, limit)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return err
		}
		students = append(students, studentInfo{ID: id})
	}
	return rows.Err()
}
