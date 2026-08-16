use serde::{Deserialize, Serialize};

/// MorphData entity types supported by Data Collector.
#[derive(Debug, Clone, Copy, PartialEq, Eq, Hash, Serialize, Deserialize)]
#[serde(rename_all = "snake_case")]
pub enum EntityKind {
    District,
    Facility,
    Member,
    Employee,
    Asset,
    Activity,
    Contact,
    User,
}

impl EntityKind {
    pub fn from_str(s: &str) -> Option<Self> {
        match s.trim().to_lowercase().as_str() {
            "district" | "districts" => Some(Self::District),
            "facility" | "facilities" | "school" | "schools" => Some(Self::Facility),
            "member" | "members" | "participant" | "participants" | "student" | "students" => {
                Some(Self::Member)
            }
            "employee" | "employees" | "staff" => Some(Self::Employee),
            "asset" | "assets" | "vehicle" | "vehicles" => Some(Self::Asset),
            "activity" | "activities" | "trip" | "trips" => Some(Self::Activity),
            "contact" | "contacts" => Some(Self::Contact),
            "user" | "users" => Some(Self::User),
            _ => None,
        }
    }

    pub fn as_str(self) -> &'static str {
        match self {
            Self::District => "district",
            Self::Facility => "facility",
            Self::Member => "member",
            Self::Employee => "employee",
            Self::Asset => "asset",
            Self::Activity => "activity",
            Self::Contact => "contact",
            Self::User => "user",
        }
    }

    pub fn label(self) -> &'static str {
        match self {
            Self::District => "District",
            Self::Facility => "Facility",
            Self::Member => "Participant (Member)",
            Self::Employee => "Employee",
            Self::Asset => "Asset",
            Self::Activity => "Activity",
            Self::Contact => "Contact",
            Self::User => "Morph User",
        }
    }

    pub fn all() -> &'static [EntityKind] {
        &[
            Self::District,
            Self::Facility,
            Self::Member,
            Self::Employee,
            Self::Asset,
            Self::Activity,
            Self::Contact,
            Self::User,
        ]
    }
}

#[derive(Debug, Clone, Serialize)]
pub struct EntitySpec {
    pub kind: EntityKind,
    pub label: String,
    pub description: String,
    pub required_fields: Vec<String>,
    pub optional_fields: Vec<String>,
    pub template_headers: Vec<String>,
    pub csv_template: String,
    pub csv_example: String,
    pub json_example: String,
    pub instructions: Vec<String>,
}

pub fn spec_for(kind: EntityKind) -> EntitySpec {
    let (_required, optional, headers, csv_ex, json_ex) = match kind {
        EntityKind::District => (
            vec!["name"],
            vec!["district_id", "district", "description"],
            vec!["name", "district_id", "district", "description"],
            "Central Region,DR-001,Central Region,Regional district for central facilities",
            r#"[{"name":"Central Region","district_id":"DR-001","district":"Central Region","description":"Regional district"}]"#,
        ),
        EntityKind::Facility => (
            vec!["name"],
            vec![
                "facility_code",
                "district_id",
                "facility_type",
                "description",
                "location",
            ],
            vec![
                "name",
                "facility_code",
                "district_id",
                "facility_type",
                "description",
                "location",
            ],
            "Main Campus,FC-100,1,school,Primary campus building,\"{\"\"city\"\":\"\"Springfield\"\"}\"",
            r#"[{"name":"Main Campus","facility_code":"FC-100","district_id":"1","facility_type":"school","description":"Primary campus"}]"#,
        ),
        EntityKind::Member => (
            vec!["last_name"],
            vec![
                "first_name",
                "middle_name",
                "dob",
                "entry_date",
                "facility",
                "gender",
                "email",
                "participant_type",
                "description",
            ],
            vec![
                "last_name",
                "first_name",
                "middle_name",
                "dob",
                "entry_date",
                "facility",
                "gender",
                "email",
                "participant_type",
                "description",
            ],
            "Doe,Jane,,2005-03-15,2024-09-01,Main Campus,1,jane@example.com,student,Grade 5 participant",
            r#"[{"last_name":"Doe","first_name":"Jane","email":"jane@example.com","facility":"Main Campus","description":"Grade 5 participant"}]"#,
        ),
        EntityKind::Employee => (
            vec!["last_name"],
            vec![
                "first_name",
                "middle_name",
                "staff_guid",
                "email",
                "phone_number",
                "facility_id",
                "employ_type",
                "description",
            ],
            vec![
                "last_name",
                "first_name",
                "middle_name",
                "staff_guid",
                "email",
                "phone_number",
                "facility_id",
                "employ_type",
                "description",
            ],
            "Smith,John,,EMP-42,john@example.com,555-0100,1,full-time,Transport coordinator",
            r#"[{"last_name":"Smith","first_name":"John","email":"john@example.com","facility_id":"1","description":"Transport coordinator"}]"#,
        ),
        EntityKind::Asset => (
            vec!["asset_tag"],
            vec!["description", "AssetType", "ContractorID"],
            vec!["asset_tag", "description", "AssetType", "ContractorID"],
            "BUS-12,School bus type C,bus,",
            r#"[{"asset_tag":"BUS-12","description":"School bus type C","AssetType":"bus"}]"#,
        ),
        EntityKind::Activity => (
            vec!["Name"],
            vec![
                "ActivityType",
                "start_date",
                "end_date",
                "location",
                "GUID",
                "description",
            ],
            vec![
                "Name",
                "ActivityType",
                "start_date",
                "end_date",
                "location",
                "GUID",
                "description",
            ],
            "Field Trip 2026,excursion,2026-05-01,2026-05-01,Museum,ACT-001,Annual museum visit",
            r#"[{"Name":"Field Trip 2026","ActivityType":"excursion","start_date":"2026-05-01","description":"Annual museum visit"}]"#,
        ),
        EntityKind::Contact => (
            vec!["LastName"],
            vec!["FirstName", "Email", "Phone", "Mobile", "description"],
            vec![
                "LastName",
                "FirstName",
                "Email",
                "Phone",
                "Mobile",
                "description",
            ],
            "Doe,Jane,jane@example.com,555-0101,,Billing contact",
            r#"[{"LastName":"Doe","FirstName":"Jane","Email":"jane@example.com","description":"Billing contact"}]"#,
        ),
        EntityKind::User => (
            vec!["login_id", "first_name", "last_name"],
            vec!["email", "phone", "administrator"],
            vec![
                "login_id",
                "first_name",
                "last_name",
                "email",
                "phone",
                "administrator",
            ],
            "jdoe,Jane,Doe,jane@example.com,555-0100,false",
            r#"[{"login_id":"jdoe","first_name":"Jane","last_name":"Doe","email":"jane@example.com","administrator":"false"}]"#,
        ),
    };

    let mut optional_fields: Vec<String> = optional.into_iter().map(String::from).collect();
    if !optional_fields
        .iter()
        .any(|field| field.eq_ignore_ascii_case("title"))
    {
        optional_fields.push("title".to_string());
    }
    let required_fields: Vec<String> = Vec::new();

    let header_line = headers.join(",");
    let csv_template = format!("{header_line}\n");
    let csv_example = format!("{header_line}\n{csv_ex}\n");
    let json_example = format!("{json_ex}\n");

    let description = match kind {
        EntityKind::District => "Import district records into MorphData District table.",
        EntityKind::Facility => "Import facilities (schools/sites) linked to districts.",
        EntityKind::Member => "Import participants (member table).",
        EntityKind::Employee => "Import employees/staff records.",
        EntityKind::Asset => "Import assets (vehicles, equipment).",
        EntityKind::Activity => "Import activities/trips.",
        EntityKind::Contact => "Import contacts.",
        EntityKind::User => "Import Morph User profiles.",
    };

    let instructions = vec![
        "Download the CSV or JSON example and add your rows.".to_string(),
        "Using the template: template columns import into MorphData; any extra columns become JSON Details.".to_string(),
        "Without the template: matching column names import into MorphData; everything else becomes JSON Details.".to_string(),
        "Validate your file, then start the background import.".to_string(),
    ];

    EntitySpec {
        kind,
        label: kind.label().to_string(),
        description: description.to_string(),
        required_fields,
        optional_fields,
        template_headers: headers.into_iter().map(String::from).collect(),
        csv_template,
        csv_example,
        json_example,
        instructions,
    }
}

pub fn all_specs() -> Vec<EntitySpec> {
    EntityKind::all()
        .iter()
        .copied()
        .map(spec_for)
        .collect()
}
