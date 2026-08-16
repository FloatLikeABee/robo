-- Upgrade legacy `Student` to slim schema. Safe to re-run on DBs created from the current
-- migrations/001_schema_mysql.sql (slim): missing indexes/columns are skipped.
-- Requires MySQL 8.0.29+ for DROP INDEX IF EXISTS / DROP COLUMN IF EXISTS.

ALTER TABLE Student DROP INDEX IF EXISTS IX_Student_Local_ID;
ALTER TABLE Student DROP INDEX IF EXISTS IX_Student_LastUpdated;

-- Rename Mi -> Middle_Name only if legacy column Mi still exists (skip if already Middle_Name).
SET @has_mi := (
  SELECT COUNT(*) FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'Student' AND COLUMN_NAME = 'Mi'
);
SET @sql := IF(@has_mi > 0,
  'ALTER TABLE Student CHANGE COLUMN Mi Middle_Name VARCHAR(50) NULL',
  'SELECT 1');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

ALTER TABLE Student
  DROP COLUMN IF EXISTS Aid_Eligible,
  DROP COLUMN IF EXISTS Aide_Req,
  DROP COLUMN IF EXISTS Cohort,
  DROP COLUMN IF EXISTS DistanceFromAMStop,
  DROP COLUMN IF EXISTS DistanceFromPmStop,
  DROP COLUMN IF EXISTS DistanceFromSchl,
  DROP COLUMN IF EXISTS GeoConfidence,
  DROP COLUMN IF EXISTS Geo_City,
  DROP COLUMN IF EXISTS Geo_County,
  DROP COLUMN IF EXISTS Geo_Street,
  DROP COLUMN IF EXISTS Geo_Zip,
  DROP COLUMN IF EXISTS InActive,
  DROP COLUMN IF EXISTS IntGratChar1,
  DROP COLUMN IF EXISTS IntGratChar2,
  DROP COLUMN IF EXISTS IntGratDate1,
  DROP COLUMN IF EXISTS IntGratDate2,
  DROP COLUMN IF EXISTS IntGratNum1,
  DROP COLUMN IF EXISTS IntGratNum2,
  DROP COLUMN IF EXISTS LastUngeocoded,
  DROP COLUMN IF EXISTS LastUngeocodedReason,
  DROP COLUMN IF EXISTS LastUpdated,
  DROP COLUMN IF EXISTS LastUpdatedID,
  DROP COLUMN IF EXISTS LastUpdatedType,
  DROP COLUMN IF EXISTS LoadTime,
  DROP COLUMN IF EXISTS LoadTimeManuallyChanged,
  DROP COLUMN IF EXISTS Local_ID,
  DROP COLUMN IF EXISTS Locked,
  DROP COLUMN IF EXISTS PreRedistSchool,
  DROP COLUMN IF EXISTS PriorSchool,
  DROP COLUMN IF EXISTS ProhibitCross,
  DROP COLUMN IF EXISTS ResidSchool,
  DROP COLUMN IF EXISTS System_City,
  DROP COLUMN IF EXISTS System_State,
  DROP COLUMN IF EXISTS System_Street,
  DROP COLUMN IF EXISTS System_Zip,
  DROP COLUMN IF EXISTS Transported,
  DROP COLUMN IF EXISTS XCoord,
  DROP COLUMN IF EXISTS YCoord;
