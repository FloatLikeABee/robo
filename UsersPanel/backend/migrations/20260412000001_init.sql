-- MySQL 8+ / utf8mb4 (matches database `tran`, user `root`, etc.)
CREATE TABLE IF NOT EXISTS plat_users (
    id VARCHAR(36) NOT NULL PRIMARY KEY,
    email VARCHAR(320) NOT NULL,
    username VARCHAR(255) NOT NULL,
    password_hash VARCHAR(255) NULL,
    google_id VARCHAR(255) NULL,
    is_verified TINYINT(1) NOT NULL DEFAULT 0,
    roles TEXT NOT NULL,
    default_channel_id VARCHAR(64) NOT NULL,
    verification_token VARCHAR(512) NULL,
    verification_expires_at VARCHAR(64) NULL,
    reset_token VARCHAR(512) NULL,
    reset_expires_at VARCHAR(64) NULL,
    created_at VARCHAR(64) NOT NULL,
    updated_at VARCHAR(64) NOT NULL,
    UNIQUE KEY uk_plat_users_email (email),
    UNIQUE KEY uk_plat_users_username (username),
    UNIQUE KEY uk_plat_users_google_id (google_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
