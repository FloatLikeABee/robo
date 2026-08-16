use std::collections::HashMap;

use serde_json::Value;
use sqlx::MySqlPool;
use uuid::Uuid;

use crate::importcol::entities::EntityKind;
use crate::importcol::jobs::RowResult;
use crate::importcol::map::MappedRecord;

pub async fn insert_record(
    pool: &MySqlPool,
    kind: EntityKind,
    rec: &MappedRecord,
) -> RowResult {
    match kind {
        EntityKind::District => insert_district(pool, rec).await,
        EntityKind::Facility => insert_facility(pool, rec).await,
        EntityKind::Member => insert_member(pool, rec).await,
        EntityKind::Employee => insert_employee(pool, rec).await,
        EntityKind::Asset => insert_asset(pool, rec).await,
        EntityKind::Activity => insert_activity(pool, rec).await,
        EntityKind::Contact => insert_contact(pool, rec).await,
        EntityKind::User => insert_user(pool, rec).await,
    }
}

fn str_field(fields: &HashMap<String, Value>, keys: &[&str]) -> Option<String> {
    for k in keys {
        if let Some(Value::String(s)) = fields.get(*k) {
            let t = s.trim();
            if !t.is_empty() {
                return Some(t.to_string());
            }
        }
    }
    None
}

async fn insert_district(pool: &MySqlPool, rec: &MappedRecord) -> RowResult {
    let name =
        str_field(&rec.mysql_fields, &["name", "Name", "title"]).unwrap_or_else(|| "Imported district".to_string());
    let district = str_field(&rec.mysql_fields, &["district", "District"]).unwrap_or_else(|| name.clone());
    let district_id = str_field(&rec.mysql_fields, &["district_id", "DistrictID"]);
    let description = str_field(&rec.mysql_fields, &["description"]);

    let res = if let Some(did) = district_id {
        sqlx::query(
            "INSERT INTO District (DBID, DistrictID, District, Name, description) VALUES (1, ?, ?, ?, ?)",
        )
        .bind(&did)
        .bind(&district)
        .bind(&name)
        .bind(&description)
        .execute(pool)
        .await
    } else {
        sqlx::query(
            "INSERT INTO District (DBID, District, Name, description) VALUES (1, ?, ?, ?)",
        )
        .bind(&district)
        .bind(&name)
        .bind(&description)
        .execute(pool)
        .await
    };

    map_insert_result(res, rec)
}

async fn insert_facility(pool: &MySqlPool, rec: &MappedRecord) -> RowResult {
    let name =
        str_field(&rec.mysql_fields, &["name", "Name", "title"]).unwrap_or_else(|| "Imported facility".to_string());
    let facility_code = str_field(&rec.mysql_fields, &["facility_code"]);
    let district_id: Option<i32> = str_field(&rec.mysql_fields, &["district_id"])
        .and_then(|s| s.parse().ok());
    let facility_type = str_field(&rec.mysql_fields, &["facility_type"]);
    let description = str_field(&rec.mysql_fields, &["description"]);
    let location = str_field(&rec.mysql_fields, &["location"]);

    let res = sqlx::query(
        "INSERT INTO `facility` (facility_code, name, district_id, facility_type, description, location) VALUES (?, ?, ?, ?, ?, ?)",
    )
    .bind(&facility_code)
    .bind(&name)
    .bind(district_id)
    .bind(&facility_type)
    .bind(&description)
    .bind(&location)
    .execute(pool)
    .await;

    map_insert_result(res, rec)
}

async fn insert_member(pool: &MySqlPool, rec: &MappedRecord) -> RowResult {
    let last_name =
        str_field(&rec.mysql_fields, &["last_name", "LastName", "title"]).unwrap_or_else(|| "Imported".to_string());
    let first_name = str_field(&rec.mysql_fields, &["first_name", "FirstName"]);
    let middle_name = str_field(&rec.mysql_fields, &["middle_name"]);
    let facility = str_field(&rec.mysql_fields, &["facility"]);
    let email = str_field(&rec.mysql_fields, &["email"]);
    let participant_type = str_field(&rec.mysql_fields, &["participant_type"]);
    let description = str_field(&rec.mysql_fields, &["description"]);
    let gender: Option<i32> = str_field(&rec.mysql_fields, &["gender"]).and_then(|s| s.parse().ok());
    let dob = str_field(&rec.mysql_fields, &["dob"]);
    let entry_date = str_field(&rec.mysql_fields, &["entry_date"]);

    let res = sqlx::query(
        "INSERT INTO `member` (last_name, first_name, middle_name, dob, entry_date, facility, gender, email, participant_type, description) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
    )
    .bind(&last_name)
    .bind(&first_name)
    .bind(&middle_name)
    .bind(&dob)
    .bind(&entry_date)
    .bind(&facility)
    .bind(gender)
    .bind(&email)
    .bind(&participant_type)
    .bind(&description)
    .execute(pool)
    .await;

    map_insert_result(res, rec)
}

async fn insert_employee(pool: &MySqlPool, rec: &MappedRecord) -> RowResult {
    let last_name =
        str_field(&rec.mysql_fields, &["last_name", "LastName", "title"]).unwrap_or_else(|| "Imported".to_string());
    let first_name = str_field(&rec.mysql_fields, &["first_name", "FirstName"]);
    let middle_name = str_field(&rec.mysql_fields, &["middle_name"]);
    let staff_guid = str_field(&rec.mysql_fields, &["staff_guid"]);
    let email = str_field(&rec.mysql_fields, &["email"]);
    let phone_number = str_field(&rec.mysql_fields, &["phone_number", "phone"]);
    let facility_id: Option<i32> = str_field(&rec.mysql_fields, &["facility_id"]).and_then(|s| s.parse().ok());
    let employ_type = str_field(&rec.mysql_fields, &["employ_type"]);
    let description = str_field(&rec.mysql_fields, &["description"]);

    let res = sqlx::query(
        "INSERT INTO `employee` (last_name, first_name, middle_name, staff_guid, email, phone_number, facility_id, employ_type, description) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)",
    )
    .bind(&last_name)
    .bind(&first_name)
    .bind(&middle_name)
    .bind(&staff_guid)
    .bind(&email)
    .bind(&phone_number)
    .bind(facility_id)
    .bind(&employ_type)
    .bind(&description)
    .execute(pool)
    .await;

    map_insert_result(res, rec)
}

async fn insert_asset(pool: &MySqlPool, rec: &MappedRecord) -> RowResult {
    let asset_tag = str_field(&rec.mysql_fields, &["asset_tag", "AssetID", "title"]).unwrap_or_else(|| {
        format!("ASSET-{}", &Uuid::new_v4().to_string()[..8]).to_uppercase()
    });
    let description = str_field(&rec.mysql_fields, &["description"]);
    let asset_type = str_field(&rec.mysql_fields, &["AssetType", "asset_type"]);
    let contractor_id = str_field(&rec.mysql_fields, &["ContractorID", "contractor_id"]);

    let res = sqlx::query(
        "INSERT INTO Asset (asset_tag, description, AssetType, ContractorID) VALUES (?, ?, ?, ?)",
    )
    .bind(&asset_tag)
    .bind(&description)
    .bind(&asset_type)
    .bind(&contractor_id)
    .execute(pool)
    .await;

    map_insert_result(res, rec)
}

async fn insert_activity(pool: &MySqlPool, rec: &MappedRecord) -> RowResult {
    let name =
        str_field(&rec.mysql_fields, &["Name", "name", "title"]).unwrap_or_else(|| "Imported activity".to_string());
    let activity_type = str_field(&rec.mysql_fields, &["ActivityType", "activity_type"]);
    let start_date = str_field(&rec.mysql_fields, &["start_date"]);
    let end_date = str_field(&rec.mysql_fields, &["end_date"]);
    let location = str_field(&rec.mysql_fields, &["location"]);
    let guid = str_field(&rec.mysql_fields, &["GUID", "guid"]);
    let description = str_field(&rec.mysql_fields, &["description"]);

    let res = sqlx::query(
        "INSERT INTO Activity (Name, ActivityType, start_date, end_date, location, GUID, description) VALUES (?, ?, ?, ?, ?, ?, ?)",
    )
    .bind(&name)
    .bind(&activity_type)
    .bind(&start_date)
    .bind(&end_date)
    .bind(&location)
    .bind(&guid)
    .bind(&description)
    .execute(pool)
    .await;

    map_insert_result(res, rec)
}

async fn insert_contact(pool: &MySqlPool, rec: &MappedRecord) -> RowResult {
    let last_name =
        str_field(&rec.mysql_fields, &["LastName", "last_name", "title"]).unwrap_or_else(|| "Imported".to_string());
    let first_name = str_field(&rec.mysql_fields, &["FirstName", "first_name"]);
    let email = str_field(&rec.mysql_fields, &["Email", "email"]);
    let phone = str_field(&rec.mysql_fields, &["Phone", "phone"]);
    let mobile = str_field(&rec.mysql_fields, &["Mobile", "mobile"]);
    let description = str_field(&rec.mysql_fields, &["description"]);

    let res = sqlx::query(
        "INSERT INTO contact (LastName, FirstName, Email, Phone, Mobile, description) VALUES (?, ?, ?, ?, ?, ?)",
    )
    .bind(&last_name)
    .bind(&first_name)
    .bind(&email)
    .bind(&phone)
    .bind(&mobile)
    .bind(&description)
    .execute(pool)
    .await;

    map_insert_result(res, rec)
}

async fn insert_user(pool: &MySqlPool, rec: &MappedRecord) -> RowResult {
    let login_id = str_field(&rec.mysql_fields, &["login_id", "title"]).unwrap_or_else(|| {
        format!("import_{}", &Uuid::new_v4().to_string()[..8])
    });
    let first_name =
        str_field(&rec.mysql_fields, &["first_name", "title"]).unwrap_or_else(|| "Imported".to_string());
    let last_name =
        str_field(&rec.mysql_fields, &["last_name"]).unwrap_or_else(|| "User".to_string());
    let email = str_field(&rec.mysql_fields, &["email"]);
    let phone = str_field(&rec.mysql_fields, &["phone"]);
    let admin: i32 = str_field(&rec.mysql_fields, &["administrator"])
        .map(|s| {
            let l = s.to_lowercase();
            if ["1", "true", "yes"].contains(&l.as_str()) {
                1
            } else {
                0
            }
        })
        .unwrap_or(0);

    let res = sqlx::query(
        "INSERT INTO `User` (login_id, first_name, last_name, email, phone, title, administrator) VALUES (?, ?, ?, ?, ?, ?, ?)",
    )
    .bind(&login_id)
    .bind(&first_name)
    .bind(&last_name)
    .bind(&email)
    .bind(&phone)
    .bind(str_field(&rec.mysql_fields, &["title"]))
    .bind(admin)
    .execute(pool)
    .await;

    map_insert_result(res, rec)
}

fn map_insert_result(
    res: Result<sqlx::mysql::MySqlQueryResult, sqlx::Error>,
    rec: &MappedRecord,
) -> RowResult {
    match res {
        Ok(r) => RowResult {
            row_ref: rec.row_ref.clone(),
            success: true,
            message: "Imported successfully".into(),
            record_id: r.last_insert_id().try_into().ok(),
        },
        Err(e) => RowResult {
            row_ref: rec.row_ref.clone(),
            success: false,
            message: e.to_string(),
            record_id: None,
        },
    }
}

fn fail_row(rec: &MappedRecord, msg: &str) -> RowResult {
    RowResult {
        row_ref: rec.row_ref.clone(),
        success: false,
        message: msg.to_string(),
        record_id: None,
    }
}

pub fn summarize_results(results: &[RowResult]) -> (usize, usize) {
    let imported = results.iter().filter(|r| r.success).count();
    let failed = results.len() - imported;
    (imported, failed)
}
