package demoentitydetail

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

func hashFacility(seed int) uint64 {
	return hashDemoDetail(seed ^ 0xFAC111) // distinct salt from staff/participant
}

func inferCampusLevel(facilityCode string) string {
	c := strings.ToUpper(strings.TrimSpace(facilityCode))
	switch {
	case strings.Contains(c, "-ES"):
		return "elementary"
	case strings.Contains(c, "-MS"):
		return "middle"
	case strings.Contains(c, "-HS"):
		return "high"
	default:
		return "campus"
	}
}

// FacilityDemoDetailJSON builds Mongo entity_details.body JSON for entity "school"
// (MySQL `facility`; REST /api/tran/facilities). Suitable for demos and seeds.
func FacilityDemoDetailJSON(id int, facilityCode, displayName, districtName, facilityType string) string {
	code := strings.TrimSpace(facilityCode)
	name := strings.TrimSpace(displayName)
	if name == "" {
		name = fmt.Sprintf("Facility %d", id)
	}
	district := strings.TrimSpace(districtName)
	if district == "" {
		district = "Unassigned district"
	}
	ftype := strings.TrimSpace(facilityType)
	if ftype == "" {
		ftype = "school"
	}

	h := hashFacility(id + 2048)
	level := inferCampusLevel(code)

	lecPrefix := strings.ToUpper(strings.ReplaceAll(code, "-", ""))
	if len(lecPrefix) > 8 {
		lecPrefix = lecPrefix[:8]
	}
	if lecPrefix == "" {
		lecPrefix = "SITE"
	}

	principals := []string{"Dr. Avery Kim", "Ms. Jordan Reeves", "Mr. Carlos Mendez", "Dr. Priya Narayan", "Ms. Helen Okafor"}
	liaisons := []string{"R. Santoro — Transportation", "M. Patel — SPED routing", "T. Nguyen — Safety coordinator", "L. Frost — Ops liaison"}
	streetBases := []string{"1842 Maple Creek Rd", "415 Riverbend Ave", "88 Summit Ridge Dr", "1200 Orchard Lane", "2601 Prairie View Blvd"}
	cities := []string{"Auburn Vale", "Cedar Mills", "Riverton", "Oakridge Crossing", "Summit Heights"}

	principal := principals[int(h%uint64(len(principals)))]
	liaison := liaisons[int((h>>8)%uint64(len(liaisons)))]
	street := streetBases[int((h>>12)%uint64(len(streetBases)))]
	city := cities[int((h>>16)%uint64(len(cities)))]
	state := []string{"CA", "OR", "WA", "ID"}[h%4]
	zip := fmt.Sprintf("%05d", 97000+int(h%899))

	var bellStart, bellEnd string
	var dismissStagger string
	switch level {
	case "elementary":
		bellStart, bellEnd = "08:40", "15:10"
		dismissStagger = "kinder 2m early release Fri; grades 1–5 standard stagger by grade (speaker 1–5)"
	case "middle":
		bellStart, bellEnd = "08:05", "15:05"
		dismissStagger = "7th west lot / 8th east circle — 4 min offset on late-bus days"
	case "high":
		bellStart, bellEnd = "07:45", "14:55"
		dismissStagger = "activities block 15:15–17:30; athletics buses load zone D"
	default:
		bellStart, bellEnd = "08:00", "15:00"
		dismissStagger = "Standard single bell; confirm district calendar for minimum days"
	}

	radioPairs := []struct{ bus, ops string }{
		{"Ch 2 — campus simple", "452.900 MHz digital"},
		{"Ch 4 — north loop", "453.050 MHz simplex backup"},
		{"Ch 1 — primary", "151.415 MHz licensed"},
		{"Ch 3 — SPED staging", "452.875 MHz"},
	}
	rp := radioPairs[int(h%uint64(len(radioPairs)))]

	loadZones := []map[string]interface{}{
		{
			"code": "A", "label": "Main circle — general education", "curb_length_ft": 280,
			"gps_hint": "north lot; avoid fire lane hash marks",
		},
		{
			"code": "B", "label": "Gymnasium side — activity / athletics", "curb_length_ft": 190,
			"gps_hint": "east turnaround; buses over 40 ft use sweep pattern",
		},
		{
			"code": "C", "label": "SPED / lift-assist staging", "curb_length_ft": 120,
			"gps_hint": "signed concrete pad; 8-minute dwell cap unless nurse notified",
		},
	}

	earlyRelease := []string{
		"2026-11-25 minimum day — dismiss 12:30",
		"2026-06-05 last day — K/staged 11:40",
		"",
	}[int(h%3)]

	constructionNote := []string{
		"",
		"West fence line utility trench through May — substitute staging on Maple side.",
		"Crossing guard relocated to Pine & 4th until signal upgrade completes.",
		"",
	}[int((h >> 4) % 4)]

	obj := map[string]interface{}{
		"site_identity": map[string]interface{}{
			"facility_code":             code,
			"display_name":              name,
			"district_official_name":    district,
			"facility_type_code":        ftype,
			"campus_level":              level,
			"nces_id_style_placeholder": fmt.Sprintf("06-%05d", 10000+(id%89999)),
			"local_education_number":    fmt.Sprintf("LEN-%s-%04d", lecPrefix, id%10000),
			"time_zone":                 "America/Los_Angeles",
			"mailing_address": map[string]interface{}{
				"street": street, "city": city, "state": state, "postal_code": zip,
			},
		},
		"leadership_transport": map[string]interface{}{
			"principal_display":                  principal,
			"transportation_liaison":            liaison,
			"front_office_main":                 fmt.Sprintf("+1-555-%03d-%04d", 200+int(h%50), 2100+id%7900),
			"after_hours_facilities_emergency":  fmt.Sprintf("+1-555-%03d-%04d", 700+int((h>>6)%40), 3100+id%6900),
			"visitor_check_in_required":          true,
		},
		"bell_and_calendar": map[string]interface{}{
			"regular_day_instruction_start_local": bellStart,
			"regular_day_instruction_end_local":   bellEnd,
			"dismissal_stagger_policy":             dismissStagger,
			"early_release_or_minimum_days_note":   earlyRelease,
			"state_testing_blackout_weeks_demo":    []string{"2026-04-20w", "2026-05-04w"},
		},
		"student_transport_staging": map[string]interface{}{
			"primary_load_zones":               loadZones,
			"dispatch_radio_channel_hint":      rp.bus,
			"operations_talkaround":            rp.ops,
			"max_buses_on_campus_simultaneous": 14 + int(h%8),
			"wheelchair_lift_zone":             "Zone C — pad must stay clear until ramp cycle complete",
			"kindergarten_sibling_policy":      "K pickup requires adult with matching placard after first 2 weeks",
		},
		"routing_dispatcher_notes": map[string]interface{}{
			"construction_or_hazards":        constructionNote,
			"rail_crossing_near_campus":      (h % 5) == 0,
			"known_pm_congestion":            []string{"Oak & 12th (15:15–15:45)", "Highway 9 off-ramp Fridays"}[int(h%2)],
			"sped_shuttle_handoff_location":  []string{"Zone C east bollards", "Health office side door"}[int((h>>10)%2)],
			"field_trip_bus_cap_notice":      "Charter overflow: staging shifts to zone B; notify dispatch 24h when >2 buses.",
		},
		"safety_and_compliance": map[string]interface{}{
			"single_point_entry_w_intake_raptor_demo": (h % 7) != 0,
			"reunification_drill_last_completed":    time.Now().AddDate(0, -int(h%6)-1, -int(h%20)).Format("2006-01-02"),
			"aed_locations_for_drivers":             []string{"Main office hallway", "Gym equipment room"}[int(h%2)],
			"student_supervision_at_curb_until":     "bus stop supervisor release or driver acknowledgement on manifest app",
		},
		"food_and_programs": map[string]interface{}{
			"usda_meal_program_site":      (h % 11) != 0,
			"title_i_support_site_demo":   level == "elementary" && (h%3) == 0,
			"after_school_program_vendor": []string{"YMCA Spectrum", "Boys & Girls Clubs — River chapter", "District Extended Learning", ""}[int((h>>14)%4)],
		},
		"data_admin": map[string]interface{}{
			"demo_seed_hash":             fmt.Sprintf("fac-%d-%016x", id, h),
			"last_reviewed_by_transport": time.Now().AddDate(0, -int((h>>18)%5)-1, -int(h%27)).Format("2006-01-02"),
			"detail_schema_hint":         "tran_demo_facility_v1",
		},
	}

	b, err := json.MarshalIndent(obj, "", "  ")
	if err != nil {
		return "{}"
	}
	return string(b)
}
