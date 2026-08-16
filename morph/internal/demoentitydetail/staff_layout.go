package demoentitydetail

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

func hashStaff(seed int) uint64 {
	return hashDemoDetail(seed ^ 0x53544146) // STAF — distinct salt from participant layout
}

// StaffDemoDetailJSON generates realistic school-transport employee detail for Mongo entity_details (entity "staff")
// and optional MySQL employee.detail JSON column.
func StaffDemoDetailJSON(id int, firstName, lastName string, gender int, employType string, dob time.Time,
	inactive bool, inactiveDate *string,
) string {
	display := strings.TrimSpace(strings.TrimSpace(firstName) + " " + strings.TrimSpace(lastName))
	if display == "" {
		display = fmt.Sprintf("Employee #%d", id)
	}

	h := hashStaff(id + 911)

	departments := []string{
		"Transportation Operations",
		"Fleet Safety & Compliance",
		"Routing & Dispatch",
		"Vehicle Maintenance",
		"Training & Professional Standards",
	}
	jobTitles := []string{
		"School Bus Operator",
		"Transportation Specialist",
		"Bus Attendant / Monitor",
		"Dispatcher",
		"Fleet Technician (Bus)",
		"Routing Coordinator",
	}
	depots := []string{
		"East Annex Yard — Gate B",
		"Central Depot — 440 Industrial Way",
		"West Satellite Office",
		"District HQ — Transportation wing",
	}
	routeLabels := []string{"Route 12 — Oak Ridge", "Route 4 — Riverbend", "Route 18 — Summit", "SPED shuttle — North",
		"Late bus — secondary", "Activity / athletics block"}
	supervisors := []string{"M. Chen", "J. Ortega", "R. Patel", "T. Williams"}
	licenseStates := []string{"CA", "OR", "WA", "ID"}
	payPayGroups := []string{"HOURLY_TRANSPORT", "SALARY_EXEMPT_TRANSPORT", "CONTRACT_OPS"}

	title := jobTitles[int(h%uint64(len(jobTitles)))]
	if employType == "contractor" {
		title = "Contract Bus Operator"
	} else if employType == "part_time" && h%3 == 0 {
		title = "Part-Time Bus Attendant"
	}

	// Hire date: 1–14 years ago, spread from hash (after plausible working age).
	yearsEmployed := 1 + int((h>>4)%14)
	hireMonth := time.Month(1 + int((h>>8)%12))
	hireDay := 1 + int((h>>12)%28)
	hireYear := time.Now().Year() - yearsEmployed
	hire := time.Date(hireYear, hireMonth, hireDay, 0, 0, 0, 0, time.UTC)

	cdls := []string{"B", "B", "BP"}
	cdl := cdls[int(h%uint64(len(cdls)))]
	pEndorse := (h % 6) != 0
	meExpiry := time.Now().AddDate(0, int((h>>16)%9)+4, int(h%45))

	bgCleared := hire.AddDate(0, 0, int(7+h%120))
	drugPA := time.Now().AddDate(0, -int(h%36), int(h%20))

	emergencyNames := []string{"Jordan Ellis", "Sam Rivera", "Alex Morgan"}
	emergencyPhones := []string{"555-0142", "555-0298", "555-0731"}

	empStatus := "active"
	var sepIface interface{}
	if inactive {
		empStatus = "former"
		if inactiveDate != nil && strings.TrimSpace(*inactiveDate) != "" {
			sepIface = strings.TrimSpace(*inactiveDate)
		} else {
			sepIface = nil
		}
	} else {
		sepIface = nil
	}

	obj := map[string]interface{}{
		"employment": map[string]interface{}{
			"legal_name_preferred":   display,
			"gender_code":            gender,
			"employ_type":            employType,
			"job_title":              title,
			"department":             departments[int((h>>20)%uint64(len(departments)))],
			"reporting_supervisor":   supervisors[int(h%uint64(len(supervisors)))],
			"primary_depot_location": depots[int((h>>24)%uint64(len(depots)))],
			"employee_reference":     fmt.Sprintf("EMP-%06d", 300000+(id%700000)),
			"fob_badge_serial":       fmt.Sprintf("FB-%03d-%04d", int(h%900)+100, id%10000),
			"date_of_birth_on_file":  dob.Format("2006-01-02"),
			"approx_hire_date":       hire.Format("2006-01-02"),
			"time_clock_policy":      "rounding to nearest quarter-hour; punches at depot reader or mobile app.",
			"employment_status":      empStatus,
			"separation_date":        sepIface,
		},
		"dot_motor_carrier": map[string]interface{}{
			"cdl_class":                cdl,
			"p_endorsement_school_bus": pEndorse,
			"s_passenger_endorsement":  true,
			"hazmat_exempt_notice":     "Tank vehicle / hazmat NA — pupil transport only.",
			"dot_med_cert_expiry":      meExpiry.Format("2006-01-02"),
			"mvr_review_next_due":      time.Now().AddDate(0, int(h%11)+6, int(h%24)).Format("2006-01-02"),
			"license_state_issue":      licenseStates[h%uint64(len(licenseStates))],
			"license_masked_last4":     fmt.Sprintf("••••%04d", 1000+((h>>3)%9000)),
		},
		"hours_and_turnout": map[string]interface{}{
			"daily_max_spread_hours":   13 + float64(int(h%3)),
			"willing_substitute_cover": (h % 7) != 0,
			"typical_midday_split":     (h % 5) != 0,
			"split_shift_notes_if_any": []string{
				"",
				"Overlaps midday special-ed coverage on Tue/Thu.",
				"Paired with Dispatcher for charter overflow during tournament weeks.",
				"",
			}[int(h%4)],
		},
		"training_completed": map[string]interface{}{
			"defensive_driving_renew_by":           time.Now().AddDate(0, int(h%10)+14, int(h%19)).Format("2006-01-02"),
			"pup_evacuation_drill_credit_year":     time.Now().Year() - int(h%2),
			"first_aid_and_cpr_refresher_due_by":   time.Now().AddDate(1, int(h%6), int(h%11)).Format("2006-01-02"),
			"behavior_de_escalation_module_status": []string{"current", "due Q3", "current"}[int(h%3)],
		},
		"routing_assignment": map[string]interface{}{
			"primary_route_label":    routeLabels[int(h%uint64(len(routeLabels)))],
			"secondary_route_cover":  routeLabels[int((h>>8)%uint64(len(routeLabels)))],
			"dispatch_radio_channel": fmt.Sprintf("%.1f MHz", 152.7+float64(h%45)/100.0),
			"favorite_bus_bay_gate":  fmt.Sprintf("%c-%02d", 'A'+rune(h%6), int(h%31)+1),
		},
		"fleet_use": map[string]interface{}{
			"pre_trip_daily_ok":            true,
			"fuel_card_holder_id_for_demo": fmt.Sprintf("FC-%05d", 12000+(id%88000)),
			"parking_gate_code_hint":       "Printed on bay assignment sheet.",
		},
		"safety_notes": map[string]interface{}{
			"student_boarding_station":            []string{"Curbside A", "Staging lane 4", "Gymnasium circle"}[int(h%3)],
			"ytd_incidents_report_count":          int(h % 5),
			"ytd_near_miss_logged":                (h % 11) != 0,
			"winter_chain_requirement_understood": true,
			"hands_free_phone_policy_ack":         time.Now().AddDate(0, -int(h%8)-1, 0).Format("2006-01-02"),
		},
		"human_resources_snapshot": map[string]interface{}{
			"criminal_registry_check_cleared_as_of": bgCleared.Format("2006-01-02"),
			"drug_screen_pool_last_draw":            drugPA.Format("2006-01-02"),
			"i9_and_work_authorization":             "Citizen — file HR-8842 scanned.",
			"payroll_pay_group_for_demo":            payPayGroups[(h/3)%uint64(len(payPayGroups))],
			"w2_delivery_preference":                "Electronic — ADP Workforce Now.",
			"fmla_eligible_but_no_open_case":        (h % 53) != 0,
			"labor_rep_unit":                        "TranDemo Local 412 (demo)",
		},
		"emergency_contact_demo": map[string]interface{}{
			"name":                               emergencyNames[int(h%uint64(len(emergencyNames)))],
			"relationship_to_employee":           []string{"spouse", "parent", "adult child", "sibling"}[int(h%4)],
			"mobile_masked":                      fmt.Sprintf("+1-555-%04d", 1000+(id%8999)),
			"mobile_secondary":                   emergencyPhones[int((h>>4)%uint64(len(emergencyPhones)))],
			"consent_receive_roster_related_sms": (h % 33) != 0,
			"vaccination_attestation_optional_on_file_since": nil,
		},
	}

	b, err := json.MarshalIndent(obj, "", "  ")
	if err != nil {
		return "{}"
	}
	return string(b)
}
