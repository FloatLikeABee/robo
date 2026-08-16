-- Canonical roles (referenced by plat_role_permissions.role_name and plat_users.roles JSON)

CREATE TABLE IF NOT EXISTS plat_roles (
    id VARCHAR(36) NOT NULL PRIMARY KEY,
    name VARCHAR(128) NOT NULL,
    description TEXT NULL,
    created_at VARCHAR(64) NOT NULL,
    updated_at VARCHAR(64) NOT NULL,
    UNIQUE KEY uk_plat_roles_name (name)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

INSERT IGNORE INTO plat_roles (id, name, description, created_at, updated_at) VALUES
('22222222-2222-4222-8222-222222222201', 'Main Panel', 'Main panel application access', '2026-04-12T00:00:00Z', '2026-04-12T00:00:00Z'),
('22222222-2222-4222-8222-222222222202', 'Forms', 'Forms creation and broadcast', '2026-04-12T00:00:00Z', '2026-04-12T00:00:00Z'),
('22222222-2222-4222-8222-222222222203', 'Email Composer', 'Email composition', '2026-04-12T00:00:00Z', '2026-04-12T00:00:00Z'),
('22222222-2222-4222-8222-222222222204', 'Sharp Reports', 'Reports and exports', '2026-04-12T00:00:00Z', '2026-04-12T00:00:00Z'),
('22222222-2222-4222-8222-222222222205', 'Admin', 'User, role, and permission administration', '2026-04-12T00:00:00Z', '2026-04-12T00:00:00Z');

-- Enforce role_name values against plat_roles (MySQL 8+). Idempotent: skip if a partial run already added this FK.
SET @fk_prp_plat_role_exists := (
    SELECT COUNT(*) FROM information_schema.TABLE_CONSTRAINTS
    WHERE CONSTRAINT_SCHEMA = DATABASE()
      AND TABLE_NAME = 'plat_role_permissions'
      AND CONSTRAINT_NAME = 'fk_prp_plat_role'
      AND CONSTRAINT_TYPE = 'FOREIGN KEY'
);
SET @fk_prp_plat_role_sql := IF(
    @fk_prp_plat_role_exists = 0,
    'ALTER TABLE plat_role_permissions ADD CONSTRAINT fk_prp_plat_role FOREIGN KEY (role_name) REFERENCES plat_roles (name) ON DELETE CASCADE ON UPDATE CASCADE',
    'SELECT 1'
);
PREPARE fk_prp_plat_role_stmt FROM @fk_prp_plat_role_sql;
EXECUTE fk_prp_plat_role_stmt;
DEALLOCATE PREPARE fk_prp_plat_role_stmt;
