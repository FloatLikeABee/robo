use sqlx::FromRow;

#[derive(Debug, FromRow)]
struct MorphUserDbRow {
    user_id: i32,
    login_id: Option<String>,
    first_name: Option<String>,
    last_name: String,
    email: Option<String>,
    phone: Option<String>,
    administrator: i8,
}

#[tokio::test]
async fn morph_users_query_decodes() {
    let url = std::env::var("DATABASE_URL").unwrap_or_else(|_| {
        "mysql://root:Dafuq%40911@127.0.0.1:3306/tran?charset=utf8mb4".to_string()
    });
    let pool = sqlx::MySqlPool::connect(&url).await.expect("connect");
    let rows: Vec<MorphUserDbRow> = sqlx::query_as(
        "SELECT UserID AS user_id, LoginID AS login_id, FirstName AS first_name, LastName AS last_name, Email AS email, Phone AS phone, Administrator AS administrator FROM `User` WHERE Deactivated = 0 ORDER BY LastName, FirstName LIMIT 500",
    )
    .fetch_all(&pool)
    .await
    .expect("query");
    assert!(!rows.is_empty(), "expected at least one morph user");
}
