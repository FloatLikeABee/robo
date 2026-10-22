package importcol

import (
	"context"
	"database/sql"
	"strconv"
	"strings"

	"github.com/google/uuid"
)

type RowResult struct {
	RowRef   string `json:"row_ref"`
	Success  bool   `json:"success"`
	Message  string `json:"message"`
	RecordID *int64 `json:"record_id,omitempty"`
}

func field(row map[string]string, keys ...string) string {
	for _, k := range keys {
		for rk, rv := range row {
			if strings.EqualFold(rk, k) {
				return strings.TrimSpace(rv)
			}
		}
	}
	return ""
}

func InsertRecord(ctx context.Context, db *sql.DB, kind EntityKind, row map[string]string, rowRef string) RowResult {
	var (
		res sql.Result
		err error
	)
	switch kind {
	case EntityDistrict:
		name := field(row, "name", "Name", "title")
		if name == "" {
			name = "Imported district"
		}
		district := field(row, "district", "District")
		if district == "" {
			district = name
		}
		did := field(row, "district_id", "DistrictID")
		desc := field(row, "description")
		if did != "" {
			res, err = db.ExecContext(ctx, `INSERT INTO District (DBID, DistrictID, District, Name, description) VALUES (1, ?, ?, ?, ?)`, did, district, name, nullStr(desc))
		} else {
			res, err = db.ExecContext(ctx, `INSERT INTO District (DBID, District, Name, description) VALUES (1, ?, ?, ?)`, district, name, nullStr(desc))
		}
	case EntityFacility:
		name := field(row, "name", "Name", "title")
		if name == "" {
			name = "Imported facility"
		}
		code := field(row, "facility_code")
		var districtID any
		if s := field(row, "district_id"); s != "" {
			if n, e := strconv.Atoi(s); e == nil {
				districtID = n
			}
		}
		res, err = db.ExecContext(ctx,
			`INSERT INTO facility (facility_code, name, district_id, facility_type, description, location) VALUES (?, ?, ?, ?, ?, ?)`,
			nullStr(code), name, districtID, nullStr(field(row, "facility_type")), nullStr(field(row, "description")), nullStr(field(row, "location")))
	case EntityMember:
		last := field(row, "last_name", "LastName", "title")
		if last == "" {
			last = "Imported"
		}
		var gender any
		if s := field(row, "gender"); s != "" {
			if n, e := strconv.Atoi(s); e == nil {
				gender = n
			}
		}
		res, err = db.ExecContext(ctx,
			`INSERT INTO member (last_name, first_name, middle_name, dob, entry_date, facility, gender, email, participant_type, description) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			last, nullStr(field(row, "first_name", "FirstName")), nullStr(field(row, "middle_name")),
			nullStr(field(row, "dob")), nullStr(field(row, "entry_date")), nullStr(field(row, "facility")),
			gender, nullStr(field(row, "email")), nullStr(field(row, "participant_type")), nullStr(field(row, "description")))
	case EntityEmployee:
		last := field(row, "last_name", "LastName", "title")
		if last == "" {
			last = "Imported"
		}
		var facilityID any
		if s := field(row, "facility_id"); s != "" {
			if n, e := strconv.Atoi(s); e == nil {
				facilityID = n
			}
		}
		res, err = db.ExecContext(ctx,
			`INSERT INTO employee (last_name, first_name, middle_name, staff_guid, email, phone_number, facility_id, employ_type, description) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			last, nullStr(field(row, "first_name", "FirstName")), nullStr(field(row, "middle_name")),
			nullStr(field(row, "staff_guid")), nullStr(field(row, "email")), nullStr(field(row, "phone_number", "phone")),
			facilityID, nullStr(field(row, "employ_type")), nullStr(field(row, "description")))
	case EntityAsset:
		tag := field(row, "asset_tag", "AssetID", "title")
		if tag == "" {
			tag = "ASSET-" + strings.ToUpper(uuid.NewString()[:8])
		}
		res, err = db.ExecContext(ctx,
			`INSERT INTO Asset (asset_tag, description, AssetType, ContractorID) VALUES (?, ?, ?, ?)`,
			tag, nullStr(field(row, "description")), nullStr(field(row, "AssetType", "asset_type")), nullStr(field(row, "ContractorID", "contractor_id")))
	case EntityActivity:
		name := field(row, "Name", "name", "title")
		if name == "" {
			name = "Imported activity"
		}
		res, err = db.ExecContext(ctx,
			`INSERT INTO Activity (Name, ActivityType, start_date, end_date, location, GUID, description) VALUES (?, ?, ?, ?, ?, ?, ?)`,
			name, nullStr(field(row, "ActivityType", "activity_type")), nullStr(field(row, "start_date")),
			nullStr(field(row, "end_date")), nullStr(field(row, "location")), nullStr(field(row, "GUID", "guid")), nullStr(field(row, "description")))
	case EntityContact:
		last := field(row, "LastName", "last_name", "title")
		if last == "" {
			last = "Imported"
		}
		res, err = db.ExecContext(ctx,
			`INSERT INTO contact (LastName, FirstName, Email, Phone, Mobile, description) VALUES (?, ?, ?, ?, ?, ?)`,
			last, nullStr(field(row, "FirstName", "first_name")), nullStr(field(row, "Email", "email")),
			nullStr(field(row, "Phone", "phone")), nullStr(field(row, "Mobile", "mobile")), nullStr(field(row, "description")))
	case EntityUser:
		login := field(row, "login_id", "title")
		if login == "" {
			login = "import_" + uuid.NewString()[:8]
		}
		first := field(row, "first_name", "title")
		if first == "" {
			first = "Imported"
		}
		last := field(row, "last_name")
		if last == "" {
			last = "User"
		}
		admin := 0
		switch strings.ToLower(field(row, "administrator")) {
		case "1", "true", "yes":
			admin = 1
		}
		res, err = db.ExecContext(ctx,
			`INSERT INTO User (LoginID, FirstName, LastName, Email, Phone, Title, Administrator, Deactivated) VALUES (?, ?, ?, ?, ?, ?, ?, 0)`,
			login, first, last, nullStr(field(row, "email")), nullStr(field(row, "phone")), nullStr(field(row, "title")), admin)
	default:
		return RowResult{RowRef: rowRef, Success: false, Message: "unknown entity"}
	}
	if err != nil {
		return RowResult{RowRef: rowRef, Success: false, Message: err.Error()}
	}
	id, _ := res.LastInsertId()
	rid := id
	return RowResult{RowRef: rowRef, Success: true, Message: "Imported successfully", RecordID: &rid}
}

func nullStr(s string) any {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	return s
}

func Summarize(results []RowResult) (imported, failed int) {
	for _, r := range results {
		if r.Success {
			imported++
		} else {
			failed++
		}
	}
	return
}
