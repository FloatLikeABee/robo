// Package demoentitydetail builds canonical entity_details.body JSON for MongoDB (database e.g. athena,
// collection entity_details) used across Tran demo seeds so participants,
// employees, and other entities share one admin-friendly shape.
package demoentitydetail

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

func hashDemoDetail(seed int) uint64 {
	sum := sha256.Sum256([]byte(fmt.Sprintf("tran-demo-detail-%d", seed)))
	return binary.BigEndian.Uint64(sum[:8])
}

// ParticipantLayoutDetailJSON mirrors the participant (member / student) Mongo detail payload.
// facilityAttached may be empty; trimmed empty strings become "their assigned campus" (same intent as participant rows with no facility).
// typeSlot is participant_type for members (student, passenger, …); for employees pass employ_type (full_time, …) so JSON keys stay aligned.
func ParticipantLayoutDetailJSON(id int, first, last string, gender int, facilityAttached string, typeSlot string) string {
	h := hashDemoDetail(id + 1337)
	routes := []string{"Oak Hollow Loop", "Riverbend Express", "Summit Ridge", "Westgate Shuttle",
		"Meadowlark Connector", "Prairie View Local"}
	stops := []string{"Front circle", "Oak St & 4th Ave", "Community center lot", "Park-n-Ride east",
		"Cedar Lane turnaround", "District depot"}
	internals := []string{
		"Verified guardian contacts annually.",
		"IEP paperwork reviewed by routing coordinator.",
		"No custody restriction fax on file.",
		"Alternate winter stop authorized north entrance.",
		"Summer school routing notes refreshed June 1.",
		"Bilingual household — announcements duplicated when requested.",
	}

	facility := strings.TrimSpace(facilityAttached)
	if facility == "" {
		facility = "their assigned campus"
	}

	obj := map[string]interface{}{
		"participant_summary": map[string]interface{}{
			"display_name":      strings.TrimSpace(fmt.Sprintf("%s %s", first, last)),
			"facility_attached": facility,
			"participant_type":  typeSlot,
			"gender_code":       gender,
			"demo_seed":         id,
		},
		"routing": map[string]interface{}{
			"morning_route_label":   routes[h%uint64(len(routes))],
			"afternoon_route_label": routes[(h>>8)%uint64(len(routes))],
			"primary_stop_landmark": stops[(h>>16)%uint64(len(stops))],
			"approx_pickup_local":   fmt.Sprintf("%02d:%02d", 6+int(h%2), int(h%50)+10),
			"approx_dropoff_local":  fmt.Sprintf("%02d:%02d", 14+int(h%2), int((h>>4)%55)+5),
		},
		"safety_authorization": map[string]interface{}{
			"photo_release_bus_directory":   (h % 5) != 0,
			"alternate_pickup_requires_pin": true,
			"authorized_notes":              internals[int(h%uint64(len(internals)))],
		},
		"services": map[string]interface{}{
			"wheelchair_lift_required":        (h % 37) == 0,
			"behavior_support_plan_active":    (h % 53) == 0,
			"esl_services_flag":               (h % 41) == 0,
			"meal_program_eligible_breakfast": (h % 29) != 0,
		},
		"equipment": map[string]interface{}{
			"assigned_radio_channel":          fmt.Sprintf("CH-%02d", int(h%12)+1),
			"car_seat_or_booster_review_date": time.Now().AddDate(0, int(h%8)+1, int(h%27)).Format("2006-01-02"),
		},
	}

	b, err := json.MarshalIndent(obj, "", "  ")
	if err != nil {
		return "{}"
	}
	return string(b)
}
