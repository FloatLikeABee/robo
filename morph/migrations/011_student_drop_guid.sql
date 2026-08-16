-- Student: remove legacy GUID column (not used by admin API).
-- IF EXISTS: fresh installs from migrations/001_schema_mysql.sql never had GUID; avoids Error 1054.
ALTER TABLE Student DROP COLUMN IF EXISTS GUID;
