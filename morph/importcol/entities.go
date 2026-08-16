package importcol

import (
	"encoding/json"
	"strings"
)

type EntityKind string

const (
	EntityDistrict EntityKind = "district"
	EntityFacility EntityKind = "facility"
	EntityMember   EntityKind = "member"
	EntityEmployee EntityKind = "employee"
	EntityAsset    EntityKind = "asset"
	EntityActivity EntityKind = "activity"
	EntityContact  EntityKind = "contact"
	EntityUser     EntityKind = "user"
)

func ParseEntityKind(s string) (EntityKind, bool) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "district", "districts":
		return EntityDistrict, true
	case "facility", "facilities", "school", "schools", "place", "places":
		return EntityFacility, true
	case "member", "members", "participant", "participants", "student", "students", "people-member", "people":
		return EntityMember, true
	case "employee", "employees", "staff", "people-employee":
		return EntityEmployee, true
	case "asset", "assets", "vehicle", "vehicles":
		return EntityAsset, true
	case "activity", "activities", "trip", "trips":
		return EntityActivity, true
	case "contact", "contacts":
		return EntityContact, true
	case "user", "users":
		return EntityUser, true
	default:
		return "", false
	}
}

func (k EntityKind) Label() string {
	switch k {
	case EntityDistrict:
		return "District"
	case EntityFacility:
		return "Places"
	case EntityMember:
		return "People"
	case EntityEmployee:
		return "People — employees"
	case EntityAsset:
		return "Assets"
	case EntityActivity:
		return "Activities"
	case EntityContact:
		return "Contact"
	case EntityUser:
		return "Morph User"
	default:
		return string(k)
	}
}

func AllKinds() []EntityKind {
	return []EntityKind{
		EntityMember, EntityAsset, EntityActivity,
	}
}

type EntitySpec struct {
	Kind             string   `json:"kind"`
	Label            string   `json:"label"`
	Description      string   `json:"description"`
	RequiredFields   []string `json:"required_fields"`
	OptionalFields   []string `json:"optional_fields"`
	TemplateHeaders  []string `json:"template_headers"`
	CSVTemplate      string   `json:"csv_template"`
	CSVExample       string   `json:"csv_example"`
	JSONExample      string   `json:"json_example"`
	Instructions     []string `json:"instructions"`
}

func SpecFor(kind EntityKind) EntitySpec {
	var optional, headers []string
	var csvEx, jsonEx, desc string
	switch kind {
	case EntityDistrict:
		optional = []string{"district_id", "district", "description"}
		headers = []string{"name", "district_id", "district", "description"}
		csvEx = "Central Region,DR-001,Central Region,Regional district for central facilities"
		jsonEx = `[{"name":"Central Region","district_id":"DR-001","district":"Central Region","description":"Regional district"}]`
		desc = "Import district records into MorphData District table."
	case EntityFacility:
		optional = []string{"facility_code", "district_id", "facility_type", "description", "location"}
		headers = []string{"name", "facility_code", "district_id", "facility_type", "description", "location"}
		csvEx = "Main Campus,FC-100,1,school,Primary campus building,"
		jsonEx = `[{"name":"Main Campus","facility_code":"FC-100","district_id":"1","facility_type":"school","description":"Primary campus"}]`
		desc = "Import places (sites) linked to districts. Columns: name, code (facility_code), type (facility_type)."
	case EntityMember:
		optional = []string{"first_name", "middle_name", "dob", "entry_date", "facility", "gender", "email", "participant_type", "description"}
		headers = []string{"last_name", "first_name", "middle_name", "dob", "entry_date", "facility", "gender", "email", "participant_type", "description"}
		csvEx = "Doe,Jane,,2005-03-15,2024-09-01,Main Campus,1,jane@example.com,student,Grade 5 participant"
		jsonEx = `[{"last_name":"Doe","first_name":"Jane","email":"jane@example.com","facility":"Main Campus","description":"Grade 5 participant"}]`
		desc = "Import people who are members/participants. The facility column is a place name or code."
	case EntityEmployee:
		optional = []string{"first_name", "middle_name", "staff_guid", "email", "phone_number", "facility_id", "employ_type", "description"}
		headers = []string{"last_name", "first_name", "middle_name", "staff_guid", "email", "phone_number", "facility_id", "employ_type", "description"}
		csvEx = "Smith,John,,EMP-42,john@example.com,555-0100,1,full-time,Transport coordinator"
		jsonEx = `[{"last_name":"Smith","first_name":"John","email":"john@example.com","facility_id":"1","description":"Transport coordinator"}]`
		desc = "Import people who are employees. facility_id is the place record id."
	case EntityAsset:
		optional = []string{"description", "AssetType", "ContractorID"}
		headers = []string{"asset_tag", "description", "AssetType", "ContractorID"}
		csvEx = "BUS-12,School bus type C,bus,"
		jsonEx = `[{"asset_tag":"BUS-12","description":"School bus type C","AssetType":"bus"}]`
		desc = "Import assets (vehicles, equipment)."
	case EntityActivity:
		optional = []string{"ActivityType", "start_date", "end_date", "location", "GUID", "description"}
		headers = []string{"Name", "ActivityType", "start_date", "end_date", "location", "GUID", "description"}
		csvEx = "Field Trip 2026,excursion,2026-05-01,2026-05-01,Museum,ACT-001,Annual museum visit"
		jsonEx = `[{"Name":"Field Trip 2026","ActivityType":"excursion","start_date":"2026-05-01","description":"Annual museum visit"}]`
		desc = "Import activities/trips."
	case EntityContact:
		optional = []string{"FirstName", "Email", "Phone", "Mobile", "description"}
		headers = []string{"LastName", "FirstName", "Email", "Phone", "Mobile", "description"}
		csvEx = "Doe,Jane,jane@example.com,555-0101,,Billing contact"
		jsonEx = `[{"LastName":"Doe","FirstName":"Jane","Email":"jane@example.com","description":"Billing contact"}]`
		desc = "Import contacts."
	case EntityUser:
		optional = []string{"email", "phone", "administrator"}
		headers = []string{"login_id", "first_name", "last_name", "email", "phone", "administrator"}
		csvEx = "jdoe,Jane,Doe,jane@example.com,555-0100,false"
		jsonEx = `[{"login_id":"jdoe","first_name":"Jane","last_name":"Doe","email":"jane@example.com","administrator":"false"}]`
		desc = "Import Morph User profiles."
	}
	headerLine := strings.Join(headers, ",")
	return EntitySpec{
		Kind:            string(kind),
		Label:           kind.Label(),
		Description:     desc,
		RequiredFields:  []string{},
		OptionalFields:  optional,
		TemplateHeaders: headers,
		CSVTemplate:     headerLine + "\n",
		CSVExample:      headerLine + "\n" + csvEx + "\n",
		JSONExample:     jsonEx + "\n",
		Instructions: []string{
			"Download the CSV or JSON example and add your rows.",
			"Using the template: template columns import into MorphData.",
			"Validate your file, then start the background import.",
		},
	}
}

func AllSpecs() []EntitySpec {
	out := make([]EntitySpec, 0, len(AllKinds()))
	for _, k := range AllKinds() {
		out = append(out, SpecFor(k))
	}
	return out
}

// MustEncode is a tiny helper for tests/debug.
func MustEncode(v any) string {
	b, _ := json.Marshal(v)
	return string(b)
}
