-- Remove Student columns: DistanceFromResidSch, PopulationRegionID, Mail_State_Id
-- Requires MySQL 8.0.29+ for DROP COLUMN IF EXISTS.

ALTER TABLE Student
  DROP COLUMN IF EXISTS DistanceFromResidSch,
  DROP COLUMN IF EXISTS PopulationRegionID,
  DROP COLUMN IF EXISTS Mail_State_Id;
