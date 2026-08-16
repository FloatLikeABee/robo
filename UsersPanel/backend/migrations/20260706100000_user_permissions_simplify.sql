-- Per-user app permissions (Admin role still gates Users Panel access via plat_users.roles JSON).
-- MySQL TEXT columns cannot have DEFAULT values, so add nullable, backfill, then enforce NOT NULL.

ALTER TABLE plat_users
    ADD COLUMN permissions TEXT NULL AFTER roles;

UPDATE plat_users
SET permissions = '["morph_util","morph_booki","morph_engi"]'
WHERE permissions IS NULL;

ALTER TABLE plat_users
    MODIFY permissions TEXT NOT NULL;

-- Collapse legacy multi-role assignments to Admin vs non-admin.
UPDATE plat_users
SET roles = '["Admin"]'
WHERE JSON_CONTAINS(COALESCE(roles, '[]'), '"Admin"', '$');

UPDATE plat_users
SET roles = '[]'
WHERE NOT JSON_CONTAINS(COALESCE(roles, '[]'), '"Admin"', '$');
