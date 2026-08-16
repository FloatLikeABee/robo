package demoentitydetail

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// ContactDemoCap is the target number of demo contacts in MorphData.
const ContactDemoCap = 30

// ContactSeed is a predefined unique demo contact row.
type ContactSeed struct {
	FirstName string
	LastName  string
	Email     string
	Phone     string
	Mobile    string
	Role      string
}

// ContactDemoSeeds returns 30 unique demo contacts (deterministic list).
func ContactDemoSeeds() []ContactSeed {
	return []ContactSeed{
		{"Maria", "Garcia", "maria.garcia@email.com", "555-201-4401", "555-301-4401", "Parent"},
		{"James", "Smith", "james.smith@email.com", "555-201-4402", "555-301-4402", "Guardian"},
		{"Patricia", "Johnson", "patricia.johnson@email.com", "555-201-4403", "555-301-4403", "Emergency contact"},
		{"Robert", "Williams", "robert.williams@email.com", "555-201-4404", "555-301-4404", "Parent"},
		{"Jennifer", "Brown", "jennifer.brown@email.com", "555-201-4405", "555-301-4405", "Work contact"},
		{"Michael", "Jones", "michael.jones@email.com", "555-201-4406", "555-301-4406", "Dispatcher liaison"},
		{"Linda", "Davis", "linda.davis@email.com", "555-201-4407", "555-301-4407", "Parent"},
		{"David", "Miller", "david.miller@email.com", "555-201-4408", "555-301-4408", "Guardian"},
		{"Susan", "Wilson", "susan.wilson@email.com", "555-201-4409", "555-301-4409", "Emergency contact"},
		{"William", "Moore", "william.moore@email.com", "555-201-4410", "555-301-4410", "Principal"},
		{"Nancy", "Taylor", "nancy.taylor@email.com", "555-201-4411", "555-301-4411", "Parent"},
		{"Richard", "Anderson", "richard.anderson@email.com", "555-201-4412", "555-301-4412", "Route contact"},
		{"Karen", "Thomas", "karen.thomas@email.com", "555-201-4413", "555-301-4413", "Guardian"},
		{"Joseph", "Jackson", "joseph.jackson@email.com", "555-201-4414", "555-301-4414", "Parent"},
		{"Lisa", "Martin", "lisa.martin@email.com", "555-201-4415", "555-301-4415", "Emergency contact"},
		{"Christopher", "White", "christopher.white@email.com", "555-201-4416", "555-301-4416", "Work contact"},
		{"Barbara", "Harris", "barbara.harris@email.com", "555-201-4417", "555-301-4417", "Parent"},
		{"Daniel", "Clark", "daniel.clark@email.com", "555-201-4418", "555-301-4418", "Guardian"},
		{"Elizabeth", "Lewis", "elizabeth.lewis@email.com", "555-201-4419", "555-301-4419", "Parent"},
		{"Thomas", "Robinson", "thomas.robinson@email.com", "555-201-4420", "555-301-4420", "Dispatcher"},
		{"Angela", "Walker", "angela.walker@email.com", "555-201-4421", "555-301-4421", "Emergency contact"},
		{"Mark", "Young", "mark.young@email.com", "555-201-4422", "555-301-4422", "Parent"},
		{"Michelle", "Allen", "michelle.allen@email.com", "555-201-4423", "555-301-4423", "Guardian"},
		{"Paul", "King", "paul.king@email.com", "555-201-4424", "555-301-4424", "Route contact"},
		{"Laura", "Wright", "laura.wright@email.com", "555-201-4425", "555-301-4425", "Parent"},
		{"Kevin", "Scott", "kevin.scott@email.com", "555-201-4426", "555-301-4426", "Work contact"},
		{"Amy", "Green", "amy.green@email.com", "555-201-4427", "555-301-4427", "Emergency contact"},
		{"Brian", "Adams", "brian.adams@email.com", "555-201-4428", "555-301-4428", "Parent"},
		{"Stephanie", "Baker", "stephanie.baker@email.com", "555-201-4429", "555-301-4429", "Guardian"},
		{"Jason", "Nelson", "jason.nelson@email.com", "555-201-4430", "555-301-4430", "Dispatcher liaison"},
	}
}

func hashContact(id int) uint64 {
	return hashDemoDetail(id ^ 0x434F4E54) // CONT
}

// ContactDisplayKey normalizes a contact name for duplicate detection.
func ContactDisplayKey(first, last string) string {
	fn := strings.TrimSpace(first)
	ln := strings.TrimSpace(last)
	if fn == "" {
		return strings.ToLower(strings.TrimSpace(ln))
	}
	return strings.ToLower(strings.TrimSpace(fn + " " + ln))
}

// ContactDescription builds the MySQL description field for a contact.
func ContactDescription(id int, first, last, role string) string {
	fn := strings.TrimSpace(first)
	ln := strings.TrimSpace(last)
	if fn == "" {
		fn = ln
		ln = ""
	}
	name := strings.TrimSpace(fn + " " + ln)
	if role == "" {
		role = "Contact"
	}
	h := hashContact(id)
	templates := []string{
		"%s serves as the primary %s for transportation notices. Prefers SMS for same-day route changes and email for billing or registration updates.",
		"%s (%s) is authorized to receive pupil release and emergency notifications. Best reached on mobile during school hours.",
		"Listed %s for district routing: %s. Spanish-language correspondence accepted; interpreter line on file when needed.",
		"%s — %s with standing permission to update stop locations through the parent portal. Annual verification completed.",
	}
	tpl := templates[int(h%uint64(len(templates)))]
	return fmt.Sprintf(tpl, name, strings.ToLower(role))
}

// ContactLayoutDetailJSON builds Mongo entity_details.body JSON for entity "contact".
func ContactLayoutDetailJSON(id int, first, last, email, phone, mobile, role string) string {
	h := hashContact(id)
	fn := strings.TrimSpace(first)
	ln := strings.TrimSpace(last)
	if fn == "" {
		fn = ln
		ln = ""
	}
	if role == "" {
		role = "Parent"
	}

	streets := []string{"1420 Cedar Lane", "88 Riverbend Ave", "501 Oak Hollow Dr", "220 Summit Ridge Rd", "915 Meadowbrook Ct"}
	cities := []string{"Riverton", "Cedar Mills", "Oakridge", "Summit Heights", "Prairie View"}
	orgs := []string{"Regional Medical Center", "Summit Logistics", "Oakridge School PTA", "Riverton Fleet Services", "Community Credit Union"}
	languages := []string{"English", "Spanish", "English", "Vietnamese", "English"}

	street := streets[int(h%uint64(len(streets)))]
	city := cities[int((h>>6)%uint64(len(cities)))]
	state := []string{"OR", "WA", "CA", "ID"}[h%4]
	zip := fmt.Sprintf("%05d", 97000+int(h%899))

	obj := map[string]interface{}{
		"contact_profile": map[string]interface{}{
			"display_name":       strings.TrimSpace(fn + " " + ln),
			"primary_role_label": role,
			"preferred_language": languages[int(h%uint64(len(languages)))],
			"organization":       orgs[int((h>>10)%uint64(len(orgs)))],
			"active_since":         time.Now().AddDate(-2-int(h%4), -int(h%11), 0).Format("2006-01-02"),
		},
		"reachability": map[string]interface{}{
			"email_primary":          strings.TrimSpace(email),
			"phone_landline":         strings.TrimSpace(phone),
			"mobile_primary":         strings.TrimSpace(mobile),
			"preferred_channel":      []string{"sms", "email", "phone"}[int(h%3)],
			"ok_robo_call_route_chg": (h%5) != 0,
			"quiet_hours_local":      "21:00–07:00",
		},
		"mailing_address_demo": map[string]interface{}{
			"street1": street,
			"street2": nil,
			"city":    city,
			"state":   state,
			"postal":  zip,
			"country": "US",
		},
		"authorization": map[string]interface{}{
			"may_pick_up_student":        (h % 7) != 0,
			"emergency_only":             role == "Emergency contact",
			"signed_media_release":       (h % 3) == 0,
			"portal_self_service_enabled": true,
			"last_verified_on":           time.Now().AddDate(0, -int(h%10)-1, -int(h%20)).Format("2006-01-02"),
		},
		"notes_internal": map[string]interface{}{
			"dispatch_memo": fmt.Sprintf("Demo contact WM-%04d — update guardian phone before winter break routes.", id%9000+1000),
			"tags":          []string{"demo", strings.ToLower(role), "verified"},
		},
	}

	b, err := json.MarshalIndent(obj, "", "  ")
	if err != nil {
		return "{}"
	}
	return string(b)
}
