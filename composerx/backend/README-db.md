# MergeEmailX – Database Setup

This service uses **MySQL** for structured metadata, following `design.md`.

## 1. Create database

```sql
CREATE DATABASE IF NOT EXISTS tran
  CHARACTER SET utf8mb4
  COLLATE utf8mb4_unicode_ci;
```

Make sure your DSN in `TRAN_MYSQL_DSN` points to this DB, e.g.:

```text
TRAN_MYSQL_DSN="root:Dafuq@911@tcp(127.0.0.1:3306)/tran?parseTime=true&charset=utf8mb4"
```

## 2. Apply schema

From the `backend` directory:

```bash
mysql -u root -p tran < schema.sql
```

This creates the following tables:

- `users`
- `email_templates`
- `merge_data_sources`
- `attachment_files`
- `saved_emails`, `saved_email_recipients`
- `email_contacts` (name, email, phone, note ≤ 1000 chars)
- `email_jobs`
- `email_recipients`
- `report_configs`
- `email_report_orders`

For an older database, apply `schema_migration_saved_emails.sql`, `schema_migration_contacts.sql`, etc., as needed.

## 3. Sanity check

After applying the schema, you can verify the tables:

```sql
SHOW TABLES;
DESCRIBE email_templates;
DESCRIBE merge_data_sources;
DESCRIBE attachment_files;
DESCRIBE email_jobs;
DESCRIBE email_recipients;
DESCRIBE report_configs;
DESCRIBE email_report_orders;
```

Once this is in place and MySQL is running, the Go backend will be able to:

- CRUD email templates
- Store merge data file metadata
- Store attachment metadata
- Create email jobs and recipients
- Track report orders

