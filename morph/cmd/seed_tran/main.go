package main

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/binary"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"

	"idongivaflyinfa/config"
	"idongivaflyinfa/db"
	"idongivaflyinfa/internal/demoentitydetail"
)

const (
	dbid = 1
)

// detailStore is set in main (SQLite/Badger by default, or MySQL/Mongo when STORAGE_BACKEND=legacy).
var detailStore db.EntityDetailStore

func main() {
	// Loads `.env` from the current directory (via config.GetConfig) so SEED_TRAN_CAP, etc. work from a file.
	cfg := config.GetConfig()
	stores, err := db.OpenCLIStores(cfg)
	if err != nil {
		log.Fatalf("open stores: %v", err)
	}
	defer stores.Close()
	sqlDB := stores.SQL
	detailStore = stores.Details
	log.Printf("seed_tran storage backend=%s", stores.Backend)

	mode := strings.ToLower(strings.TrimSpace(os.Getenv("SEED_TRAN_MODE")))
	if mode == "" {
		mode = "case_task_details"
	}

	if mode != "full" {
		log.Println("seed_tran mode: case_task_details (only upserting case/task detail JSON)")
		if err := seedCaseTaskMongoDetails(sqlDB); err != nil {
			log.Fatalf("seed case task details: %v", err)
		}
		log.Println("Seeding completed successfully.")
		return
	}

	log.Println("seed_tran mode: full (seeding all demo entities)")

	if err := demoPruneAndCapToLimits(sqlDB); err != nil {
		log.Fatalf("demo prune/cap: %v", err)
	}

	rand.Seed(time.Now().UnixNano())

	if err := seedReferenceData(sqlDB); err != nil {
		log.Fatalf("seed reference data: %v", err)
	}
	if err := seedDistrictsAndSchools(sqlDB); err != nil {
		log.Fatalf("seed districts/schools: %v", err)
	}
	if err := seedFacilityMongoDetails(sqlDB); err != nil {
		log.Fatalf("seed facility details: %v", err)
	}
	if err := seedStaffAndTypes(sqlDB); err != nil {
		log.Fatalf("seed staff: %v", err)
	}
	if err := seedVehicles(sqlDB); err != nil {
		log.Fatalf("seed vehicles: %v", err)
	}
	if err := seedTrips(sqlDB); err != nil {
		log.Fatalf("seed trips: %v", err)
	}
	if err := seedCaseTasks(sqlDB); err != nil {
		log.Fatalf("seed case tasks: %v", err)
	}
	if err := seedCaseTaskMongoDetails(sqlDB); err != nil {
		log.Fatalf("seed case task details: %v", err)
	}
	if err := seedStudentsAndSchedules(sqlDB); err != nil {
		log.Fatalf("seed students/schedules: %v", err)
	}
	if err := seedActivityRelations(sqlDB); err != nil {
		log.Fatalf("seed activity relations: %v", err)
	}
	if err := seedContacts(sqlDB); err != nil {
		log.Fatalf("seed contacts: %v", err)
	}
	if err := seedContactMongoDetails(sqlDB); err != nil {
		log.Fatalf("seed contact details: %v", err)
	}
	if err := seedRecordContacts(sqlDB); err != nil {
		log.Fatalf("seed record_contact: %v", err)
	}
	if err := backfillMissingSampleData(sqlDB); err != nil {
		log.Fatalf("backfill sample data: %v", err)
	}

	log.Println("Seeding completed successfully.")
}

// genVIN returns a 17-character VIN-like string (no I, O, Q).
func genVIN() string {
	const chars = "ABCDEFGHJKLMNPRSTUVWXYZ0123456789"
	b := make([]byte, 17)
	for i := range b {
		b[i] = chars[rand.Intn(len(chars))]
	}
	return string(b)
}

func randomStaffDOB() time.Time {
	age := 25 + rand.Intn(38) // 25–62
	return time.Now().AddDate(-age, -rand.Intn(12), -rand.Intn(28))
}

func seedReferenceData(db *sql.DB) error {
	log.Println("Seeding reference data (StaffType)...")

	// StaffType – realistic roles
	if _, err := db.Exec(`INSERT IGNORE INTO StaffType (StaffTypeName, StaffTypeDescription, IsSystemDefined) VALUES
		('Bus Driver', 'Operates school buses on daily routes', 0),
		('Bus Assistant', 'Assists driver and students during transportation', 0),
		('School Personnel', 'Principals, secretaries, school staff', 0),
		('Mechanic', 'Maintains and repairs vehicles', 0),
		('Dispatcher', 'Dispatch and routing personnel', 0)`); err != nil {
		return fmt.Errorf("stafftype: %w", err)
	}

	return nil
}

type districtInfo struct {
	ID       int
	District string
	Name     string
}

type schoolInfo struct {
	ID           int
	FacilityCode string
	Name         string
	DistrictID   int
	Level        string
}

var (
	districts []districtInfo
	schools   []schoolInfo
)

func seedDistrictsAndSchools(db *sql.DB) error {
	log.Println("Seeding districts and schools...")

	// Districts (two realistic ones)
	_, err := db.Exec(`INSERT INTO District (DBID, DistrictID, District, Name)
		VALUES
		(?, 100, 'EA01', 'East Valley School District'),
		(?, 200, 'WE01', 'West Hills School District')
		ON DUPLICATE KEY UPDATE Name = VALUES(Name)`, dbid, dbid)
	if err != nil {
		return fmt.Errorf("insert districts: %w", err)
	}

	rows, err := db.Query(`SELECT ID, District, Name FROM District WHERE DBID = ?`, dbid)
	if err != nil {
		return fmt.Errorf("select districts: %w", err)
	}
	defer rows.Close()

	districts = districts[:0]
	for rows.Next() {
		var d districtInfo
		if err := rows.Scan(&d.ID, &d.District, &d.Name); err != nil {
			return err
		}
		districts = append(districts, d)
	}

	if len(districts) == 0 {
		return fmt.Errorf("no districts found after insert")
	}

	// Schools per district: elementary, middle, high
	for _, d := range districts {
		prefix := "EV"
		if strings.HasPrefix(d.District, "WE") {
			prefix = "WH"
		}
		schoolDefs := []struct {
			codeSuffix string
			name       string
			level      string
		}{
			{"-ES", "Elementary School", "elementary"},
			{"-MS", "Middle School", "middle"},
			{"-HS", "High School", "high"},
		}
		for i, sd := range schoolDefs {
			schoolCode := fmt.Sprintf("%s%02d%s", prefix, i+1, sd.codeSuffix)
			_, err := db.Exec(`INSERT INTO `+"`facility`"+`
				(facility_code, name, district_id, capacity, facility_type)
				VALUES (?, ?, ?, ?, 'school')
				ON DUPLICATE KEY UPDATE name = VALUES(name), district_id = VALUES(district_id), capacity = VALUES(capacity), facility_type = VALUES(facility_type)`,
				schoolCode, d.Name+" "+sd.name, d.ID,
				350+rand.Intn(250))
			if err != nil {
				return fmt.Errorf("insert school %s: %w", schoolCode, err)
			}
		}
	}

	// Load schools back with level classification based on code suffix
	rows2, err := db.Query(`SELECT id, facility_code, name, district_id FROM ` + "`facility`")
	if err != nil {
		return fmt.Errorf("select schools: %w", err)
	}
	defer rows2.Close()

	schools = schools[:0]
	for rows2.Next() {
		var s schoolInfo
		if err := rows2.Scan(&s.ID, &s.FacilityCode, &s.Name, &s.DistrictID); err != nil {
			return err
		}
		if strings.Contains(s.FacilityCode, "-ES") {
			s.Level = "elementary"
		} else if strings.Contains(s.FacilityCode, "-MS") {
			s.Level = "middle"
		} else if strings.Contains(s.FacilityCode, "-HS") {
			s.Level = "high"
		} else {
			s.Level = "other"
		}
		schools = append(schools, s)
	}

	return nil
}

// seedFacilityMongoDetails writes realistic facility/school JSON into entity_details.
func seedFacilityMongoDetails(sqlDB *sql.DB) error {
	if detailStore == nil {
		log.Println("entity details unavailable; skip facilities")
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	rows, err := sqlDB.QueryContext(ctx, `
		SELECT f.id, f.facility_code, COALESCE(f.name,''), COALESCE(f.facility_type,''), COALESCE(d.Name,'')
		FROM `+"`facility`"+` f
		LEFT JOIN District d ON d.id = f.district_id
		ORDER BY f.id`)
	if err != nil {
		return fmt.Errorf("select facilities for details: %w", err)
	}
	defer rows.Close()

	n := 0
	for rows.Next() {
		var id int
		var code, name, ftype, district string
		if err := rows.Scan(&id, &code, &name, &ftype, &district); err != nil {
			return err
		}
		detail := demoentitydetail.FacilityDemoDetailJSON(id, code, name, district, ftype)
		if err := detailStore.SetEntityDetailJSON(ctx, "school", id, detail); err != nil {
			return fmt.Errorf("school detail %d: %w", id, err)
		}
		n++
	}
	if err := rows.Err(); err != nil {
		return err
	}
	log.Printf("entity_details: upserted %d facilities (entity=school)", n)
	return nil
}

type staffInfo struct {
	ID       int
	TypeName string
}

var (
	staffDrivers    []staffInfo
	staffAssistants []staffInfo
	staffAll        []staffInfo
)

func seedStaffAndTypes(sqlDB *sql.DB) error {
	log.Println("Seeding staff and staff types...")
	capN := seedCap()

	staffAll = staffAll[:0]
	staffDrivers = staffDrivers[:0]
	staffAssistants = staffAssistants[:0]

	var existing int
	if err := sqlDB.QueryRow(`SELECT COUNT(*) FROM employee`).Scan(&existing); err != nil {
		if !strings.Contains(strings.ToLower(err.Error()), "doesn't exist") {
			return fmt.Errorf("count employees: %w", err)
		}
	}
	if existing >= capN {
		log.Printf("employee count %d ≥ cap %d; loading existing staff for demo relations", existing, capN)
		return loadStaffGlobalsFromDB(sqlDB)
	}
	needed := capN - existing
	typeCycle := buildStaffTypeCycle()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
	defer cancel()
	mongoDB := detailStore
	if mongoDB == nil {
		log.Println("entity details unavailable; new employees will lack details until backfill")
	}

	firstNames := []string{"Michael", "Sarah", "James", "Emily", "David", "Laura", "Robert", "Linda", "Daniel", "Jessica"}
	lastNames := []string{"Smith", "Johnson", "Brown", "Williams", "Jones", "Garcia", "Miller", "Davis", "Martinez", "Wilson"}

	insertStaffOne := func(typeName string, facIdx int) error {
		var staffTypeID int
		if err := sqlDB.QueryRow(`SELECT StaffTypeID FROM StaffType WHERE StaffTypeName = ?`, typeName).Scan(&staffTypeID); err != nil {
			return fmt.Errorf("lookup stafftype %s: %w", typeName, err)
		}
		middleNames := []string{"", "J.", "M.", "A.", "R.", "L.", "K.", "S.", "C.", "D."}
		fn := firstNames[rand.Intn(len(firstNames))]
		ln := lastNames[rand.Intn(len(lastNames))]
		mn := middleNames[rand.Intn(len(middleNames))]
		email := fmt.Sprintf("%s.%s%d@example.edu", strings.ToLower(fn), strings.ToLower(ln), rand.Intn(1000))
		cellPhone := fmt.Sprintf("555-%04d", rand.Intn(10000))
		et := "full_time"
		if typeName == "Bus Assistant" {
			et = "part_time"
		}
		dob := randomStaffDOB()
		gender := 1 + rand.Intn(2)
		var facID interface{}
		if len(schools) > 0 {
			facID = schools[facIdx%len(schools)].ID
		}
		res, err := sqlDB.Exec(`INSERT INTO `+"`employee`"+`
			(last_name, first_name, middle_name, active_flag, email, phone_number, date_of_birth, gender, employ_type, facility_id)
			VALUES (?, ?, NULLIF(?, ''), 1, ?, ?, ?, ?, ?, ?)`,
			ln, fn, mn, email, cellPhone, dob, gender, et, facID)
		if err != nil {
			if !strings.Contains(err.Error(), "Duplicate entry") {
				return fmt.Errorf("insert staff: %w", err)
			}
			return nil
		}
		id64, _ := res.LastInsertId()
		id := int(id64)
		if id <= 0 {
			return nil
		}
		if _, err = sqlDB.Exec(`INSERT INTO StaffStaffType (StaffID, StaffTypeID, PrimaryFlag) VALUES (?, ?, 1)`, id, staffTypeID); err != nil && !strings.Contains(err.Error(), "Duplicate entry") {
			return fmt.Errorf("insert staffstafftype: %w", err)
		}
		if mongoDB != nil {
			if err := upsertStaffMongoDetails(ctx, mongoDB, id, fn, ln, gender, et, dob); err != nil {
				return fmt.Errorf("mongo staff detail %d: %w", id, err)
			}
		}
		info := staffInfo{ID: id, TypeName: typeName}
		staffAll = append(staffAll, info)
		switch typeName {
		case "Bus Driver":
			staffDrivers = append(staffDrivers, info)
		case "Bus Assistant":
			staffAssistants = append(staffAssistants, info)
		}
		return nil
	}

	for i := 0; i < needed; i++ {
		typeName := typeCycle[(existing+i)%len(typeCycle)]
		if err := insertStaffOne(typeName, existing+i); err != nil {
			return err
		}
	}

	if err := loadStaffGlobalsFromDB(sqlDB); err != nil {
		return fmt.Errorf("reload staff globals: %w", err)
	}
	log.Printf("Seeded staff up to cap %d (inserted ~%d); total loaded %d, %d drivers, %d assistants",
		capN, needed, len(staffAll), len(staffDrivers), len(staffAssistants))
	return nil
}

type vehicleInfo struct {
	ID int
}

var vehicles []vehicleInfo

func hashAssetProfileSeed(id int) uint64 {
	sum := sha256.Sum256([]byte(fmt.Sprintf("tran-asset-profile-%d", id)))
	return binary.BigEndian.Uint64(sum[:8])
}

// assetProfileForID picks a stable, varied description, fleet ID, and dictionary asset_type (bus | van | car | other).
func assetProfileForID(id int) (assetType string, assetID string, description string) {
	h := hashAssetProfileSeed(id)
	r := h % 100
	switch {
	case r < 52:
		assetType = "bus"
	case r < 80:
		assetType = "van"
	case r < 92:
		assetType = "car"
	default:
		assetType = "other"
	}

	var prefix string
	var seq int
	switch assetType {
	case "bus":
		prefix = "BUS"
		seq = 20000 + int((h>>8)%28000)
	case "van":
		prefix = "VAN"
		seq = 4100 + int((h>>8)%5900)
	case "car":
		prefix = "CAR"
		seq = 700 + int((h>>8)%3200)
	default:
		prefix = "AST"
		seq = 900 + int((h>>8)%9100)
	}
	assetID = fmt.Sprintf("%s-%05d", prefix, seq)

	busDesc := []string{
		fmt.Sprintf("%d Blue Bird Vision Type D, ~77 pax, rear engine. Lift-equipped; primary assignment East Valley high-school pyramid.", 2016+int((h>>10)%8)),
		fmt.Sprintf("%d Thomas Saf-T-Liner C2, 54 passengers. Parcel racks disabled for SPED shoulder room; camera DVR Gen4.", 2015+int((h>>12)%9)),
		fmt.Sprintf("%d IC Bus CE Series, conventional. Runs split AM/PM for Riverbend MS plus two elementary feeders.", 2017+int((h>>14)%7)),
		fmt.Sprintf("%d Blue Bird All American RE. Charter and athletics overflow; upgraded child-alert reminder buzzer 2024.", 2014+int((h>>11)%10)),
		fmt.Sprintf("%d Thomas HDX, Type D. Housed at West Hills satellite; winter chain kit staged Oct–Mar.", 2016+int((h>>9)%9)),
		fmt.Sprintf("%d IC RE Series dedicated SPED route; six wheelchair positions with QRT retractors.", 2018+int((h>>13)%6)),
		"Spare Type C — rotated through annual inspection rotation pool; odometer within district PM window.",
		"Retired interstate coach briefly — now restricted to activity trips under 85 mi per board policy.",
		"Low-mile spare acquired from neighboring co-op; VIN verified, mileage reconciled at central garage.",
		"Route 12 Oak Hollow primary bus; afternoon mirrored with Dispatch for tournament weekends.",
		"Propane autogas pilot unit; fuel card tracked separately; drivers trained TS-908 module.",
		"High-capacity 84-pax for consolidated high-school shuttles; staggered stop policy on file.",
	}
	vanDesc := []string{
		fmt.Sprintf("%d Ford Transit 350 HD — 14 passenger max, driver counts 13+1. Midday shuttles and late elementary sweep.", 2019+int((h>>10)%5)),
		fmt.Sprintf("%d RAM ProMaster extended; mobility lift (Ricon). Used for North clinic therapy runs Tue/Thu.", 2020+int((h>>11)%4)),
		fmt.Sprintf("%d Mercedes Sprinter 2500; supervisor occasionally signs out for field audits.", 2018+int((h>>12)%6)),
		fmt.Sprintf("%d Chevy Express 3500; pool vehicle for small-team athletics (limited interstate).", 2017+int((h>>9)%7)),
		"Mini-bus on loan from county ESD — marked district decals; insurance rider renewed each July.",
		"Aide shuttle — morning only, links developmental preschool to K–8 campus.",
	}
	carDesc := []string{
		fmt.Sprintf("%d Chevrolet Impala motor-pool sedan — routing supervisors, DMV road-test escorts.", 2020+int((h>>10)%4)),
		fmt.Sprintf("%d Ford Fusion hybrid — dispatch courier and state paperwork filings.", 2019+int((h>>11)%4)),
		"Unmarked supervisor vehicle — snow-route recon and parent complaint field visits.",
		"Nissan Altima — training checkout car for new hires (non-pupil transport).",
		"Subaru AWD — winter incident response; carries spill kit and cones.",
	}
	otherDesc := []string{
		"Chassis-cab with utility body — radios bench tests, parts runs between yards.",
		"Striping / signage truck — not for pupil transport; escrows commercial plates.",
		"Flatbed for bus shelter lumber and sign posts; CDL Class B district employee only.",
		"Generator trailer tow unit — event staging for graduation shuttle staging areas.",
		"Winter spreader loaner when municipal contract activates (rare).",
	}

	slot := int((h >> 16) % 256)
	switch assetType {
	case "bus":
		description = busDesc[slot%len(busDesc)]
	case "van":
		description = vanDesc[slot%len(vanDesc)]
	case "car":
		description = carDesc[slot%len(carDesc)]
	default:
		description = otherDesc[slot%len(otherDesc)]
	}
	return assetType, assetID, description
}

func seedVehicles(sqlDB *sql.DB) error {
	log.Println("Seeding assets...")

	var prior int
	if err := sqlDB.QueryRow(`SELECT COUNT(*) FROM Asset`).Scan(&prior); err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "doesn't exist") {
			log.Println("Asset table not found; skipping vehicles")
			return nil
		}
		return fmt.Errorf("count assets: %w", err)
	}
	capN := seedCap()
	toAdd := capN - prior
	if toAdd < 0 {
		toAdd = 0
	}
	for i := 0; i < toAdd; i++ {
		seedIdx := prior + i + 1
		vin := genVIN()
		atype, asset, desc := assetProfileForID(seedIdx)
		_, err := sqlDB.Exec(`INSERT INTO Asset
			(asset_tag, AssetID, description, AssetType)
			VALUES (?, ?, ?, ?)`,
			vin, asset, desc, atype)
		if err != nil && !strings.Contains(err.Error(), "Duplicate entry") {
			return fmt.Errorf("insert vehicle: %w", err)
		}
	}

	rows, err := sqlDB.Query(`SELECT ID FROM Asset`)
	if err != nil {
		return fmt.Errorf("select vehicles: %w", err)
	}
	defer rows.Close()

	vehicles = vehicles[:0]
	for rows.Next() {
		var v vehicleInfo
		if err := rows.Scan(&v.ID); err != nil {
			return err
		}
		vehicles = append(vehicles, v)
	}
	// Normalize / backfill: varied types, descriptions, fleet IDs; preserve non-empty asset_tag.
	for _, v := range vehicles {
		atype, asset, desc := assetProfileForID(v.ID)
		fallbackTag := genVIN()
		_, err := sqlDB.Exec(`UPDATE Asset SET
			asset_tag = COALESCE(NULLIF(TRIM(asset_tag), ''), ?),
			AssetID = ?,
			description = ?,
			AssetType = ?
			WHERE ID = ?`,
			fallbackTag, asset, desc, atype, v.ID)
		if err != nil {
			return fmt.Errorf("update vehicle %d: %w", v.ID, err)
		}
	}
	log.Printf("Loaded %d assets", len(vehicles))

	if detailStore == nil {
		log.Println("entity details unavailable; skip vehicles")
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	for _, v := range vehicles {
		var tag, aid, atype sql.NullString
		err := sqlDB.QueryRow(`SELECT asset_tag, AssetID, AssetType FROM Asset WHERE ID = ?`, v.ID).Scan(&tag, &aid, &atype)
		if err != nil {
			return fmt.Errorf("read asset row %d: %w", v.ID, err)
		}
		tagS, aidS, atypeS := "", "", ""
		if tag.Valid {
			tagS = tag.String
		}
		if aid.Valid {
			aidS = aid.String
		}
		if atype.Valid {
			atypeS = atype.String
		}
		detail := demoentitydetail.AssetDemoDetailJSON(v.ID, atypeS, aidS, tagS)
		if err := detailStore.SetEntityDetailJSON(ctx, "vehicle", v.ID, detail); err != nil {
			return fmt.Errorf("vehicle detail %d: %w", v.ID, err)
		}
	}
	log.Printf("entity_details: upserted %d vehicles", len(vehicles))
	return nil
}

type tripInfo struct {
	ID   int
	Name string
}

var trips []tripInfo

func seedTrips(db *sql.DB) error {
	log.Println("Seeding activities...")

	if len(schools) == 0 {
		return fmt.Errorf("schools not seeded")
	}

	for _, s := range schools {
		// Create a few activities per school.
		for i := 0; i < 3; i++ {
			name := fmt.Sprintf("%s Activity %d", s.FacilityCode, i+1)
			start := time.Date(2026, time.Month(1+rand.Intn(12)), 1+rand.Intn(20), 8+rand.Intn(8), 0, 0, 0, time.UTC)
			end := start.Add(time.Duration(2+rand.Intn(4)) * time.Hour)
			act := "standard"
			if i == 2 {
				act = "field_trip"
			}
			location := fmt.Sprintf(`{"location":"%s campus","coordx":%.4f,"coordy":%.4f}`, s.Name, 34.0+rand.Float64(), -118.0+rand.Float64())
			_, err := db.Exec(`INSERT INTO Activity
				(Name, start_date, end_date, location, ActivityType)
				VALUES (?, ?, ?, ?, ?)`,
				name, start, end, location, act)
			if err != nil && !strings.Contains(err.Error(), "Duplicate entry") {
				return fmt.Errorf("insert trip: %w", err)
			}
		}
	}

	rows, err := db.Query(`SELECT ID, Name FROM Activity`)
	if err != nil {
		return fmt.Errorf("select trips: %w", err)
	}
	defer rows.Close()

	trips = trips[:0]
	for rows.Next() {
		var t tripInfo
		if err := rows.Scan(&t.ID, &t.Name); err != nil {
			return err
		}
		trips = append(trips, t)
	}
	log.Printf("Loaded %d activities", len(trips))
	return nil
}

func seedCaseTasks(db *sql.DB) error {
	log.Println("Seeding case tasks...")

	var total int
	if err := db.QueryRow(`SELECT COUNT(*) FROM CaseTask`).Scan(&total); err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "doesn't exist") {
			log.Println("CaseTask table not found; run migrations first (skipping case task seed)")
			return nil
		}
		return fmt.Errorf("count case tasks: %w", err)
	}
	if total > 0 {
		log.Printf("CaseTask already has %d records; skipping seed", total)
		return nil
	}

	var firstEmployeeID, lastEmployeeID sql.NullInt64
	_ = db.QueryRow(`SELECT id FROM employee ORDER BY id LIMIT 1`).Scan(&firstEmployeeID)
	_ = db.QueryRow(`SELECT id FROM employee ORDER BY id DESC LIMIT 1`).Scan(&lastEmployeeID)
	var firstMemberID sql.NullInt64
	_ = db.QueryRow("SELECT id FROM `member` ORDER BY id LIMIT 1").Scan(&firstMemberID)

	if !firstEmployeeID.Valid && !firstMemberID.Valid {
		log.Println("No employees/members available yet; skipping case task seed")
		return nil
	}

	type seedItem struct {
		title        string
		description  string
		startOffsetH int
		endOffsetH   int
		locationJSON *string
		assigneeType string
		assigneeID   int64
	}

	areaLocation := `{"label":"East Valley corridor","area":[[34.142,-118.121],[34.137,-118.101],[34.121,-118.112],[34.126,-118.131]]}`
	locationDepot := `{"label":"Main depot area"}`
	items := []seedItem{}
	if firstEmployeeID.Valid {
		items = append(items, seedItem{
			title:        "Route delay follow-up",
			description:  "Investigate recurring AM route delay and assign mitigations.",
			startOffsetH: 24,
			endOffsetH:   26,
			locationJSON: &areaLocation,
			assigneeType: "employee",
			assigneeID:   firstEmployeeID.Int64,
		})
	}
	if firstMemberID.Valid {
		items = append(items, seedItem{
			title:        "Member assistance request",
			description:  "Case intake for transportation assistance change request.",
			startOffsetH: 48,
			endOffsetH:   49,
			locationJSON: nil,
			assigneeType: "member",
			assigneeID:   firstMemberID.Int64,
		})
	}
	if lastEmployeeID.Valid {
		items = append(items, seedItem{
			title:        "Fleet document review",
			description:  "Review maintenance docs and close outstanding validation task.",
			startOffsetH: 72,
			endOffsetH:   73,
			locationJSON: &locationDepot,
			assigneeType: "employee",
			assigneeID:   lastEmployeeID.Int64,
		})
	}

	now := time.Now().UTC().Truncate(time.Minute)
	for i, it := range items {
		startAt := now.Add(time.Duration(it.startOffsetH) * time.Hour)
		endAt := now.Add(time.Duration(it.endOffsetH) * time.Hour)
		res, err := db.Exec(`
			INSERT INTO CaseTask (title, description, start_at, end_at, location, assignee_type, assignee_id)
			VALUES (?, ?, ?, ?, ?, ?, ?)`,
			it.title, it.description, startAt, endAt, it.locationJSON, it.assigneeType, it.assigneeID,
		)
		if err != nil {
			return fmt.Errorf("insert case task %q: %w", it.title, err)
		}
		tid, ierr := res.LastInsertId()
		if ierr != nil || tid <= 0 {
			continue
		}
		type aj struct {
			kind string
			id   int64
		}
		ajoins := []aj{{it.assigneeType, it.assigneeID}}
		if i == 0 && firstMemberID.Valid && it.assigneeType != "member" {
			ajoins = append(ajoins, aj{"member", firstMemberID.Int64})
		}
		for _, a := range ajoins {
			if _, err := db.Exec(
				`INSERT INTO CaseTaskAssignee (case_task_id, assignee_kind, assignee_id) VALUES (?, ?, ?)`,
				tid, a.kind, a.id,
			); err != nil {
				if strings.Contains(strings.ToLower(err.Error()), "doesn't exist") {
					break
				}
				return fmt.Errorf("insert CaseTaskAssignee for %q: %w", it.title, err)
			}
		}
	}
	log.Printf("Seeded %d case tasks", len(items))
	return nil
}

func seedCaseTaskMongoDetails(sqlDB *sql.DB) error {
	if detailStore == nil {
		log.Println("entity details unavailable; skip case tasks")
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	rows, err := sqlDB.QueryContext(ctx, `
		SELECT ID, title, assignee_type, assignee_id, start_at, end_at
		FROM CaseTask
		ORDER BY ID`)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "doesn't exist") {
			log.Println("CaseTask table not found; skipping case_task details")
			return nil
		}
		return fmt.Errorf("select case tasks for details: %w", err)
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		var id int
		var title, assigneeType string
		var assigneeID int
		var startAt, endAt sql.NullTime
		if err := rows.Scan(&id, &title, &assigneeType, &assigneeID, &startAt, &endAt); err != nil {
			return err
		}
		var startPtr, endPtr *time.Time
		if startAt.Valid {
			t := startAt.Time
			startPtr = &t
		}
		if endAt.Valid {
			t := endAt.Time
			endPtr = &t
		}
		detail := demoentitydetail.CaseTaskDemoDetailJSON(id, title, assigneeType, assigneeID, startPtr, endPtr)
		if err := detailStore.SetEntityDetailJSON(ctx, "case_task", id, detail); err != nil {
			return fmt.Errorf("case_task detail %d: %w", id, err)
		}
		count++
	}
	if err := rows.Err(); err != nil {
		return err
	}
	log.Printf("entity_details: upserted %d case tasks (entity=case_task)", count)
	return nil
}

func seedActivityRelations(db *sql.DB) error {
	log.Println("Seeding activity relation tables...")
	if len(trips) == 0 {
		return nil
	}

	for _, t := range trips {
		// Link 1-2 employees to each activity.
		if len(staffAll) > 0 {
			n := 1
			if len(staffAll) > 1 && rand.Intn(100) < 35 {
				n = 2
			}
			for i := 0; i < n; i++ {
				emp := staffAll[(t.ID+i)%len(staffAll)]
				_, err := db.Exec(`INSERT IGNORE INTO ActivityEmployee (activity_id, employee_id) VALUES (?, ?)`, t.ID, emp.ID)
				if err != nil {
					return fmt.Errorf("insert activity employee: %w", err)
				}
			}
		}

		// Link 1-2 assets.
		if len(vehicles) > 0 {
			n := 1
			if len(vehicles) > 1 && rand.Intn(100) < 40 {
				n = 2
			}
			for i := 0; i < n; i++ {
				asset := vehicles[(t.ID+i)%len(vehicles)]
				_, err := db.Exec(`INSERT IGNORE INTO ActivityAsset (activity_id, asset_id) VALUES (?, ?)`, t.ID, asset.ID)
				if err != nil {
					return fmt.Errorf("insert activity asset: %w", err)
				}
			}
		}

		// Link 3-10 participants.
		if len(students) > 0 {
			n := 3 + rand.Intn(8)
			for i := 0; i < n; i++ {
				member := students[(t.ID+i)%len(students)]
				_, err := db.Exec(`INSERT IGNORE INTO ActivityParticipant (activity_id, member_id) VALUES (?, ?)`, t.ID, member.ID)
				if err != nil {
					return fmt.Errorf("insert activity participant: %w", err)
				}
			}
		}
	}
	return nil
}

type studentInfo struct {
	ID int
}

var students []studentInfo

func seedStudentsAndSchedules(sqlDB *sql.DB) error {
	log.Println("Seeding students...")
	if len(schools) == 0 {
		return fmt.Errorf("schools not seeded")
	}

	capN := seedCap()
	students = students[:0]

	var mcount int
	if err := sqlDB.QueryRow("SELECT COUNT(*) FROM `member`").Scan(&mcount); err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "doesn't exist") {
			log.Println("member table not found; skipping")
			return nil
		}
		return err
	}
	if mcount >= capN {
		log.Printf("member count %d ≥ cap %d; loading existing for activity links", mcount, capN)
		return loadStudentsFromDB(sqlDB, capN)
	}
	needed := capN - mcount

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
	defer cancel()
	mongoDB := detailStore
	if mongoDB == nil {
		log.Println("entity details unavailable; new members will lack details until backfill")
	}

	firstNames := []string{"Olivia", "Liam", "Emma", "Noah", "Ava", "Elijah", "Sophia", "Oliver", "Isabella", "Lucas"}
	lastNames := []string{"Anderson", "Thomas", "Jackson", "White", "Harris", "Martin", "Thompson", "Garcia", "Martinez", "Robinson"}

	for i := 0; i < needed; i++ {
		s := schools[i%len(schools)]
		fn := firstNames[rand.Intn(len(firstNames))]
		ln := lastNames[rand.Intn(len(lastNames))]
		email := fmt.Sprintf("%s.%s%d@student.example.edu", strings.ToLower(fn), strings.ToLower(ln), rand.Intn(10000))
		dob := randomDOBForLevel(s.Level)

		middleName := ""
		if rand.Float32() < 0.4 {
			middleName = string(rune('A' + rand.Intn(26)))
		}
		gender := 1 + rand.Intn(2)
		entryDate := time.Date(2020+rand.Intn(4), time.Month(8+rand.Intn(2)), 1+rand.Intn(20), 0, 0, 0, 0, time.UTC)
		res, err := sqlDB.Exec(`INSERT INTO `+"`member`"+`
			(last_name, first_name, middle_name, dob,
			 facility, email, gender, entry_date, participant_type)
			VALUES (?, ?, NULLIF(?, ''), ?, ?, ?, ?, ?, 'student')`,
			ln, fn, middleName, dob, s.FacilityCode, email, gender, entryDate)
		if err != nil {
			if strings.Contains(err.Error(), "Duplicate entry") {
				continue
			}
			return fmt.Errorf("insert student: %w", err)
		}
		id64, _ := res.LastInsertId()
		id := int(id64)
		if id <= 0 {
			continue
		}
		students = append(students, studentInfo{ID: id})
		if mongoDB != nil {
			if err := upsertMemberMongoDetails(ctx, mongoDB, id, fn, ln, gender, s.FacilityCode, "student"); err != nil {
				return fmt.Errorf("mongo member detail %d: %w", id, err)
			}
		}
	}

	if err := loadStudentsFromDB(sqlDB, capN); err != nil {
		return fmt.Errorf("reload students slice: %w", err)
	}
	log.Printf("Seeded members toward cap %d; loaded %d ids for links", capN, len(students))
	return nil
}

func seedContacts(db *sql.DB) error {
	log.Println("Seeding contacts...")
	capN := demoentitydetail.ContactDemoCap
	var cur int
	if err := db.QueryRow(`SELECT COUNT(*) FROM contact`).Scan(&cur); err != nil {
		return err
	}
	if cur >= capN {
		log.Printf("contact count %d ≥ cap %d; skipping insert", cur, capN)
		return nil
	}

	seeds := demoentitydetail.ContactDemoSeeds()
	existingKeys := map[string]struct{}{}
	rows, err := db.Query(`SELECT COALESCE(FirstName,''), COALESCE(LastName,'') FROM contact`)
	if err != nil {
		return err
	}
	for rows.Next() {
		var fn, ln string
		if err := rows.Scan(&fn, &ln); err != nil {
			rows.Close()
			return err
		}
		existingKeys[demoentitydetail.ContactDisplayKey(fn, ln)] = struct{}{}
	}
	rows.Close()

	inserted := 0
	for _, seed := range seeds {
		if cur+inserted >= capN {
			break
		}
		key := demoentitydetail.ContactDisplayKey(seed.FirstName, seed.LastName)
		if _, ok := existingKeys[key]; ok {
			continue
		}
		desc := demoentitydetail.ContactDescription(inserted+1, seed.FirstName, seed.LastName, seed.Role)
		_, err := db.Exec(`INSERT INTO contact (LastName, FirstName, Email, Phone, Mobile, description)
			VALUES (?, ?, ?, ?, ?, ?)`,
			seed.LastName, seed.FirstName, seed.Email, seed.Phone, seed.Mobile, desc)
		if err != nil {
			return fmt.Errorf("insert contact: %w", err)
		}
		existingKeys[key] = struct{}{}
		inserted++
	}
	log.Printf("Seeded %d contacts (target cap %d)", inserted, capN)
	return nil
}

func seedContactMongoDetails(sqlDB *sql.DB) error {
	if detailStore == nil {
		log.Println("entity details unavailable; skip contacts")
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	rows, err := sqlDB.Query(`SELECT ID, COALESCE(FirstName,''), COALESCE(LastName,''), COALESCE(Email,''), COALESCE(Phone,''), COALESCE(Mobile,'') FROM contact ORDER BY ID`)
	if err != nil {
		return err
	}
	defer rows.Close()

	seeds := demoentitydetail.ContactDemoSeeds()
	n := 0
	for rows.Next() {
		var id int
		var fn, ln, email, phone, mobile string
		if err := rows.Scan(&id, &fn, &ln, &email, &phone, &mobile); err != nil {
			return err
		}
		seed := seeds[n%len(seeds)]
		role := seed.Role
		desc := demoentitydetail.ContactDescription(id, fn, ln, role)
		if _, err := sqlDB.Exec(`UPDATE contact SET description = ? WHERE ID = ?`, desc, id); err != nil {
			return fmt.Errorf("update contact description %d: %w", id, err)
		}
		body := demoentitydetail.ContactLayoutDetailJSON(id, fn, ln, email, phone, mobile, role)
		if err := detailStore.SetEntityDetailJSON(ctx, "contact", id, body); err != nil {
			return fmt.Errorf("contact detail %d: %w", id, err)
		}
		n++
	}
	log.Printf("Upserted entity detail for %d contacts", n)
	return rows.Err()
}

// seedRecordContacts links contacts to students, staff, schools, vehicles, and trips via record_contact.
func seedRecordContacts(db *sql.DB) error {
	log.Println("Seeding record_contact (linking contacts to students, staff, schools, vehicles, trips)...")

	rows, err := db.Query(`SELECT ID FROM contact ORDER BY ID`)
	if err != nil {
		return fmt.Errorf("select contacts: %w", err)
	}
	defer rows.Close()
	var contactIDs []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return err
		}
		contactIDs = append(contactIDs, id)
	}
	if len(contactIDs) == 0 {
		log.Println("No contacts found; skipping record_contact seed")
		return nil
	}

	// Shuffle so we assign varied contacts
	rand.Shuffle(len(contactIDs), func(i, j int) { contactIDs[i], contactIDs[j] = contactIDs[j], contactIDs[i] })

	link := func(entityType string, recordID int, contactID int, relationship string, isPrimary int) error {
		_, err := db.Exec(`INSERT INTO record_contact (DBID, EntityType, RecordID, ContactID, Relationship, IsPrimary)
			VALUES (?, ?, ?, ?, ?, ?)`,
			dbid, entityType, recordID, contactID, relationship, isPrimary)
		return err
	}

	// Students: 1–2 contacts per student (Parent, Guardian, Emergency contact)
	studentRels := []string{"Parent", "Guardian", "Emergency contact"}
	for _, st := range students {
		n := 1 + rand.Intn(2)
		for i := 0; i < n && i < len(contactIDs); i++ {
			cid := contactIDs[(st.ID+i)%len(contactIDs)]
			rel := studentRels[i%len(studentRels)]
			primary := 0
			if i == 0 {
				primary = 1
			}
			if err := link("student", st.ID, cid, rel, primary); err != nil && !strings.Contains(err.Error(), "Duplicate entry") {
				return fmt.Errorf("record_contact student: %w", err)
			}
		}
	}
	log.Printf("  Linked contacts to %d students", len(students))

	// Staff: 0–1 contact per staff (Work contact, Emergency contact)
	for _, s := range staffAll {
		if rand.Float32() < 0.4 {
			cid := contactIDs[s.ID%len(contactIDs)]
			rel := "Work contact"
			if rand.Float32() < 0.3 {
				rel = "Emergency contact"
			}
			if err := link("staff", s.ID, cid, rel, 1); err != nil && !strings.Contains(err.Error(), "Duplicate entry") {
				return fmt.Errorf("record_contact staff: %w", err)
			}
		}
	}
	log.Printf("  Linked contacts to staff")

	// Schools: 1–3 contacts per school (Principal, Secretary, Nurse, etc.)
	schoolRels := []string{"Principal", "Secretary", "Nurse", "Office", "Transportation liaison"}
	for _, sch := range schools {
		n := 1 + rand.Intn(2)
		for i := 0; i < n && i < len(schoolRels); i++ {
			cid := contactIDs[(sch.ID+i*7)%len(contactIDs)]
			rel := schoolRels[i%len(schoolRels)]
			primary := 0
			if i == 0 {
				primary = 1
			}
			if err := link("school", sch.ID, cid, rel, primary); err != nil && !strings.Contains(err.Error(), "Duplicate entry") {
				return fmt.Errorf("record_contact school: %w", err)
			}
		}
	}
	log.Printf("  Linked contacts to %d schools", len(schools))

	// Vehicles: 0–1 contact per vehicle (Fleet contact, Maintenance)
	for _, v := range vehicles {
		if rand.Float32() < 0.35 {
			cid := contactIDs[v.ID%len(contactIDs)]
			rel := "Fleet contact"
			if rand.Float32() < 0.4 {
				rel = "Maintenance"
			}
			if err := link("vehicle", v.ID, cid, rel, 1); err != nil && !strings.Contains(err.Error(), "Duplicate entry") {
				return fmt.Errorf("record_contact vehicle: %w", err)
			}
		}
	}
	log.Printf("  Linked contacts to %d vehicles", len(vehicles))

	// Trips: 0–1 contact per trip (Route contact, Dispatcher)
	for _, t := range trips {
		if rand.Float32() < 0.4 {
			cid := contactIDs[t.ID%len(contactIDs)]
			rel := "Route contact"
			if rand.Float32() < 0.35 {
				rel = "Dispatcher"
			}
			if err := link("trip", t.ID, cid, rel, 1); err != nil && !strings.Contains(err.Error(), "Duplicate entry") {
				return fmt.Errorf("record_contact trip: %w", err)
			}
		}
	}
	log.Printf("  Linked contacts to %d trips", len(trips))

	log.Println("Seeded record_contact")
	return nil
}

func randomDOBForLevel(level string) time.Time {
	now := time.Now()
	var minAge, maxAge int
	switch level {
	case "elementary":
		minAge, maxAge = 5, 11
	case "middle":
		minAge, maxAge = 11, 14
	case "high":
		minAge, maxAge = 14, 19
	default:
		minAge, maxAge = 6, 18
	}
	age := minAge + rand.Intn(maxAge-minAge+1)
	// approximate DOB by subtracting years
	return now.AddDate(-age, -rand.Intn(12), -rand.Intn(28))
}

// backfillMissingSampleData is a no-op placeholder (legacy address backfill removed with migration 026).
func backfillMissingSampleData(db *sql.DB) error {
	_ = db
	return nil
}
