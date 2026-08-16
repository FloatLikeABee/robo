-- Permissions catalog and role ↔ permission mappings (MySQL)

CREATE TABLE IF NOT EXISTS plat_permissions (
    id VARCHAR(36) NOT NULL PRIMARY KEY,
    name VARCHAR(128) NOT NULL,
    description TEXT NULL,
    created_at VARCHAR(64) NOT NULL,
    updated_at VARCHAR(64) NOT NULL,
    UNIQUE KEY uk_plat_permissions_name (name)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS plat_role_permissions (
    role_name VARCHAR(64) NOT NULL,
    permission_id VARCHAR(36) NOT NULL,
    PRIMARY KEY (role_name, permission_id),
    CONSTRAINT fk_prp_permission FOREIGN KEY (permission_id)
        REFERENCES plat_permissions (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- No separate index on role_name: PRIMARY KEY (role_name, permission_id) is enough for lookups by role.

-- Seed permissions (stable UUIDs)
INSERT IGNORE INTO plat_permissions (id, name, description, created_at, updated_at) VALUES
('11111111-1111-4111-8111-111111111101', 'main_panel_access', 'Access the main panel', '2026-04-12T00:00:00Z', '2026-04-12T00:00:00Z'),
('11111111-1111-4111-8111-111111111102', 'create_form', 'Create forms', '2026-04-12T00:00:00Z', '2026-04-12T00:00:00Z'),
('11111111-1111-4111-8111-111111111103', 'broadcast_form', 'Broadcast forms', '2026-04-12T00:00:00Z', '2026-04-12T00:00:00Z'),
('11111111-1111-4111-8111-111111111104', 'compose_email', 'Compose email', '2026-04-12T00:00:00Z', '2026-04-12T00:00:00Z'),
('11111111-1111-4111-8111-111111111105', 'view_reports', 'View reports', '2026-04-12T00:00:00Z', '2026-04-12T00:00:00Z'),
('11111111-1111-4111-8111-111111111106', 'export_reports', 'Export reports (PDF/Excel)', '2026-04-12T00:00:00Z', '2026-04-12T00:00:00Z'),
('11111111-1111-4111-8111-111111111107', 'admin_users', 'Manage users and roles', '2026-04-12T00:00:00Z', '2026-04-12T00:00:00Z'),
('11111111-1111-4111-8111-111111111108', 'admin_roles', 'Configure role metadata', '2026-04-12T00:00:00Z', '2026-04-12T00:00:00Z'),
('11111111-1111-4111-8111-111111111109', 'manage_permissions', 'Manage permission definitions and role assignments', '2026-04-12T00:00:00Z', '2026-04-12T00:00:00Z');

-- Role mappings (ignore duplicates on re-run)
INSERT IGNORE INTO plat_role_permissions (role_name, permission_id) VALUES
('Main Panel', '11111111-1111-4111-8111-111111111101'),
('Forms', '11111111-1111-4111-8111-111111111102'),
('Forms', '11111111-1111-4111-8111-111111111103'),
('Email Composer', '11111111-1111-4111-8111-111111111104'),
('Sharp Reports', '11111111-1111-4111-8111-111111111105'),
('Sharp Reports', '11111111-1111-4111-8111-111111111106'),
('Admin', '11111111-1111-4111-8111-111111111101'),
('Admin', '11111111-1111-4111-8111-111111111102'),
('Admin', '11111111-1111-4111-8111-111111111103'),
('Admin', '11111111-1111-4111-8111-111111111104'),
('Admin', '11111111-1111-4111-8111-111111111105'),
('Admin', '11111111-1111-4111-8111-111111111106'),
('Admin', '11111111-1111-4111-8111-111111111107'),
('Admin', '11111111-1111-4111-8111-111111111108'),
('Admin', '11111111-1111-4111-8111-111111111109');
