package demoentitydetail

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

func hashAsset(seed int) uint64 {
	return hashDemoDetail(seed ^ 0x41535421) // 'AST!'
}

// AssetDemoDetailJSON is the canonical Mongo entity_details.body for Tran assets (API entity "vehicle").
func AssetDemoDetailJSON(id int, assetType, assetID, assetTag string) string {
	at := strings.TrimSpace(strings.ToLower(assetType))
	if at == "" {
		at = "bus"
	}
	tag := strings.TrimSpace(assetTag)
	if tag == "" {
		tag = "—"
	}
	aid := strings.TrimSpace(assetID)
	if aid == "" {
		aid = fmt.Sprintf("AST-%06d", id)
	}

	h := hashAsset(id + 701)

	makes := []string{"Blue Bird", "Thomas Built Buses", "IC Bus", "Freightliner Custom Chassis", "Ford", "Chevrolet", "Mercedes-Benz / Freightliner", "RAM", "International"}
	modelsBus := []string{"Vision", "All American RE", "Saf-T-Liner C2", "Saf-T-Liner HDX", "CE Series", "RE Series"}
	modelsVan := []string{"Transit 350 HD", "Sprinter 2500", "ProMaster 3500", "Express 3500", "NV3500"}
	modelsCar := []string{"Impala fleet sedan", "Fusion hybrid pool", "Malibu supervisor unit", "Altima motor-pool", "Outback field supervisor"}
	fuel := []string{"ULSD", "gasoline", "gasoline", "propane autogas", "gasoline"}
	garages := []string{"Central Transportation Annex", "East Valley Yard — Gate C", "West Hills satellite barn", "District HQ — covered slot 12"}

	var makeName, modelName, bodyClass string
	var seating int
	switch at {
	case "van":
		makeName = makes[int(h%uint64(len(makes)))]
		if strings.Contains(strings.ToLower(makeName), "thomas") || strings.Contains(strings.ToLower(makeName), "blue") {
			makeName = "Ford"
		}
		modelName = modelsVan[int((h>>8)%uint64(len(modelsVan)))]
		bodyClass = "Type A-1 / multi-function van"
		seating = 6 + int(h%9) // 6–14
	case "car":
		makeName = []string{"Chevrolet", "Ford", "Nissan", "Subaru", "Toyota"}[int(h%5)]
		modelName = modelsCar[int((h>>8)%uint64(len(modelsCar)))]
		bodyClass = "Sedan / light-duty motor pool"
		seating = 4 + int(h%2)*1
	case "other":
		makeName = "Miscellaneous / specialty"
		modelName = []string{"Chassis-cab with wheelchair lift module", "Route supervisor SUV", "Spare striping / maintenance truck", "Mobile radio test bench hauler", "Parts runner pickup"}[int(h%5)]
		bodyClass = "Non-standard pupil transport support"
		seating = 2 + int(h%5)
	default:
		makeName = makes[int(h%uint64(len(makes)))]
		if makeName == "Ford" || makeName == "Chevrolet" {
			makeName = "Blue Bird"
		}
		modelName = modelsBus[int((h>>8)%uint64(len(modelsBus)))]
		bodyClass = []string{"Type C conventional", "Type D rear engine", "Type C CE"}[int(h%3)]
		seating = 54 + int((h>>16)%28)
	}

	modelYear := 2015 + int((h>>12)%10)
	odometer := 42000 + int(h%220000)
	nextDOT := time.Now().AddDate(0, int((h>>20)%9)+2, int(h%40))
	nextPM := time.Now().AddDate(0, 0, int((h>>4)%40)+15)

	obj := map[string]interface{}{
		"asset_summary": map[string]interface{}{
			"fleet_asset_id":     aid,
			"asset_type_code":    at,
			"vin_or_asset_tag":   tag,
			"demo_seed_record_id": id,
			"body_class":         bodyClass,
			"manufacturer":       makeName,
			"model":              modelName,
			"model_year":         modelYear,
			"rated_seating":       seating,
			"primary_garage":     garages[int((h>>24)%uint64(len(garages)))],
		},
		"powertrain": map[string]interface{}{
			"fuel_type_primary":          fuel[int(h%uint64(len(fuel)))],
			"def_tank_present":           at == "bus" || at == "other",
			"block_heater_installed":   (h%11) != 0 && (at == "bus" || at == "van"),
			"idle_reduction_notes":       "Anti-idle policy except lift ops & extreme weather per TS-4412.",
			"last_emissions_sticker_year": time.Now().Year() - int(h%2),
		},
		"equipment": map[string]interface{}{
			"child_check_button_mounted":      at == "bus" || at == "van",
			"camera_system_generation":        []string{"360° DVR Gen4", "Forward + cabin HD", "Legacy SD archive", "Event-trigger only"}[int(h%4)],
			"two_way_radio_template":          fmt.Sprintf("%.1f MHz primary", 151.5+float64(h%60)/10.0),
			"gps_avl_vendor_placeholder":      []string{"Trapeze AVL", "Zonar V4", "Synovia demo tenant", "Geotab Pupil"}[int((h>>6)%4)],
			"wheelchair_lift_present":         at == "bus" || (at == "van" && (h%5) == 0),
			"fire_extinguisher_inspection_due": nextPM.AddDate(0, 2, 0).Format("2006-01-02"),
		},
		"maintenance": map[string]interface{}{
			"approx_odometer_mi":          odometer,
			"last_pm_completed":           nextPM.AddDate(0, 0, -int((h%25)+30)).Format("2006-01-02"),
			"next_pm_due":                 nextPM.Format("2006-01-02"),
			"annual_dot_inspection_due":   nextDOT.Format("2006-01-02"),
			"oil_sample_program":          (h % 7) == 0,
			"deferred_defects_open_count": int(h % 4),
			"tire_position_notes":         []string{"Steer axles rotated fall service.", "All positions PS2; check inner duals.", "Winter stud-capable rears (not installed)."}[int(h%3)],
		},
		"assignment": map[string]interface{}{
			"typical_use":              []string{"AM/PM home-to-school", "Midday SPED shuttles", "Athletics late block", "Field trip overflow", "Mechanic road-test / spare"}[int((h>>10)%5)],
			"home_base_route_hint":     []string{"R12 Oak Hollow", "R04 Riverbend", "R18 Summit", "Late secondary", "Charter pool"}[int(h%5)],
			"fuel_card_slot":           fmt.Sprintf("FC-%05d", 28000+(id%52000)),
			"winter_chain_kit_onboard": at == "bus" || at == "other",
		},
		"compliance_snapshot": map[string]interface{}{
			"state_inspection_region": []string{"Region 3 — West Hills", "Region 1 — East Valley HQ", "Contract inspection vendor Co-op"}[int(h%3)],
			"chaperone_capacity_note": "Passenger count excludes adult monitors; see activity roster.",
			"medical_transit_exempt":  at != "car",
		},
	}

	b, err := json.MarshalIndent(obj, "", "  ")
	if err != nil {
		return "{}"
	}
	return string(b)
}
